package messageredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/devathh/xvibe/message-service/internal/domain/message"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MessagePubsub struct {
	log    *slog.Logger
	client *redis.Client
}

func New(log *slog.Logger, client *redis.Client) *MessagePubsub {
	return &MessagePubsub{
		log:    log,
		client: client,
	}
}

func (mps *MessagePubsub) Publish(ctx context.Context, msg *message.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	bytes, err := toModel(msg)
	if err != nil {
		return err
	}

	key := mps.getChannel(msg.ChatID())
	if err := mps.client.Publish(ctx, key, bytes).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (mps *MessagePubsub) Subscribe(ctx context.Context, chatID uuid.UUID, handler func(context.Context, *message.Message)) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := mps.getChannel(chatID)
	pubsub := mps.client.Subscribe(ctx, key)

	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return errors.New("pubsub channel closed")
			}

			domain, err := toDomain([]byte(msg.Payload))
			if err != nil {
				mps.log.Warn("failed to parse message", slog.String("error", err.Error()))
				continue
			}

			go func(m *message.Message) {
				handler(ctx, m)
			}(domain)
		}
	}
}

func (mps *MessagePubsub) getChannel(chatID uuid.UUID) string {
	return fmt.Sprintf("messages:%s", chatID)
}
