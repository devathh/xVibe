package chatredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/domain/member"
	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ChatCacheRepository struct {
	cfg    *config.Config
	client *redis.Client
}

func New(cfg *config.Config, client *redis.Client) *ChatCacheRepository {
	return &ChatCacheRepository{
		cfg:    cfg,
		client: client,
	}
}

func (ccr *ChatCacheRepository) Save(
	ctx context.Context,
	userID uuid.UUID,
	chats []*chat.ChatModel,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	bytes, err := toModel(chats)
	if err != nil {
		return fmt.Errorf("failed to convert domain into model: %w", err)
	}

	key := ccr.getKey(userID)
	if err := ccr.client.Set(ctx, key, bytes, ccr.cfg.Service.Cache.ChatsTTL).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to save user's chats into cache: %w", err)
	}

	return nil
}

func (ccr *ChatCacheRepository) Get(ctx context.Context, userID uuid.UUID) ([]*chat.ChatModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := ccr.getKey(userID)
	bytes, err := ccr.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, redis.Nil) {
			return nil, err
		}

		return nil, err
	}

	model, err := toDomain(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return model, nil
}

func (ccr *ChatCacheRepository) Del(ctx context.Context, memberIds []*member.Member) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	keys := make([]string, len(memberIds))
	for idx, id := range memberIds {
		keys[idx] = ccr.getKey(id.ID())
	}

	if err := ccr.client.Del(ctx, keys...).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to delete all user's chats cache: %w", err)
	}

	return nil
}

func (ccr *ChatCacheRepository) getKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:chats:%s", userID.String())
}
