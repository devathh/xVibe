package services

import (
	"context"
	"log/slog"

	messagepb "github.com/devathh/xvibe/message-gateway/api/message/v1"
	"github.com/devathh/xvibe/message-gateway/internal/application/dtos"
	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/config"
	"google.golang.org/grpc/metadata"
)

type MessageGatewayService interface {
	CreateMessage(ctx context.Context, req *dtos.CreateRequest, token string) (*dtos.MessageModel, error)
	DeleteMessage(ctx context.Context, req *dtos.DeleteRequest, token string) error
	GetHistory(ctx context.Context, req *dtos.GetRequest, token string) (*dtos.MessageModels, error)
	ConnectNewMessages(ctx context.Context, chatID, token string, handle func(m *dtos.MessageModel) error) error
}

type messageGatewayService struct {
	cfg           *config.Config
	log           *slog.Logger
	messageClient messagepb.MessageClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	messageClient messagepb.MessageClient,
) MessageGatewayService {
	return &messageGatewayService{
		cfg:           cfg,
		log:           log,
		messageClient: messageClient,
	}
}

// Create a new message in chat
// calls CreateMessage method in message-service
func (mgs *messageGatewayService) CreateMessage(ctx context.Context, req *dtos.CreateRequest, token string) (*dtos.MessageModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	md := mgs.tokenMD(token)
	resp, err := mgs.messageClient.CreateMessage(metadata.NewOutgoingContext(ctx, md), &messagepb.CreateRequest{
		ChatId: req.ChatID,
		Body:   req.Body,
	})
	if err != nil {
		return nil, err
	}

	response := &dtos.MessageModel{
		ID:       resp.Id,
		ChatID:   resp.ChatId,
		AuthorID: resp.AuthorId,
		Body:     resp.Body,
	}
	if resp.SentAt != nil {
		response.SentAt = resp.SentAt.AsTime().UnixMilli()
	} else {
		mgs.log.Warn("sent time of message is nil")
	}

	return response, nil
}

// Delete the message from chat
// calls DeleteMessage method in message-service
func (mgs *messageGatewayService) DeleteMessage(ctx context.Context, req *dtos.DeleteRequest, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	md := mgs.tokenMD(token)
	_, err := mgs.messageClient.DeleteMessage(metadata.NewOutgoingContext(ctx, md), &messagepb.DeleteRequest{
		MsgId: req.MsgID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (mgs *messageGatewayService) GetHistory(ctx context.Context, req *dtos.GetRequest, token string) (*dtos.MessageModels, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	md := mgs.tokenMD(token)
	resp, err := mgs.messageClient.GetHistory(metadata.NewOutgoingContext(ctx, md), &messagepb.GetRequest{
		ChatId:   req.ChatID,
		Limit:    req.Limit,
		BeforeId: req.BeforeID,
	})
	if err != nil {
		return nil, err
	}

	response := dtos.MessageModels{
		Messages: make([]dtos.MessageModel, len(resp.Messages)),
		HasMore:  resp.HasMore,
	}
	for idx, msg := range resp.Messages {
		response.Messages[idx] = dtos.MessageModel{
			ID:       msg.Id,
			ChatID:   msg.ChatId,
			AuthorID: msg.AuthorId,
			Body:     msg.Body,
		}
		if msg.SentAt != nil {
			response.Messages[idx].SentAt = msg.SentAt.AsTime().UnixMilli()
		} else {
			mgs.log.Warn("sent time of msg is empty")
		}
	}

	return &response, nil
}

func (mgs *messageGatewayService) ConnectNewMessages(ctx context.Context, chatID, token string, handle func(m *dtos.MessageModel) error) error {
	md := mgs.tokenMD(token)
	stream, err := mgs.messageClient.ConnectNewMessages(
		metadata.NewOutgoingContext(ctx, md),
		&messagepb.ConnectRequest{
			ChatId: chatID,
		},
	)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	msgCh := make(chan *dtos.MessageModel, 10)

	go func() {
		defer close(msgCh)

		for {
			msg, err := stream.Recv()
			if err != nil {
				mgs.log.Error("received error from stream", slog.String("error", err.Error()))
				return
			}

			model := &dtos.MessageModel{
				ID:       msg.Id,
				ChatID:   msg.ChatId,
				AuthorID: msg.AuthorId,
				Body:     msg.Body,
			}
			if msg.SentAt != nil {
				model.SentAt = msg.SentAt.AsTime().UnixMilli()
			} else {
				mgs.log.Warn("sent time of msg is empty")
			}

			msgCh <- model
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgCh:
			if !ok {
				return nil
			}

			if err := handle(msg); err != nil {
				return err
			}
		}
	}
}

func (mgs *messageGatewayService) tokenMD(token string) metadata.MD {
	return metadata.New(map[string]string{
		"authorization": token,
	})
}
