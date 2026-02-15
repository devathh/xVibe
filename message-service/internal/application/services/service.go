package services

import (
	"context"
	"errors"
	"log/slog"

	messagepb "github.com/devathh/xvibe/message-service/api/message/v1"
	"github.com/devathh/xvibe/message-service/internal/domain/crypto"
	"github.com/devathh/xvibe/message-service/internal/domain/message"
	"github.com/devathh/xvibe/message-service/internal/domain/session"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/message-service/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MessageService interface {
	CreateMessage(ctx context.Context, req *messagepb.CreateRequest) (*messagepb.MessageModel, error)
	DeleteMessage(ctx context.Context, req *messagepb.DeleteRequest) error
	ConnectNewMessages(ctx context.Context, req *messagepb.ConnectRequest, stream grpc.ServerStreamingServer[messagepb.MessageModel]) error
	GetHistory(ctx context.Context, req *messagepb.GetRequest) (*messagepb.MessageModels, error)
}

type messageService struct {
	cfg        *config.Config
	log        *slog.Logger
	msgRepo    message.MessageRepository
	msgCache   message.MessageCacheRepository
	wrapperDEK crypto.WrapperDEK
	aesgcmEnc  crypto.AESGCMEncryptor
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	msgRepo message.MessageRepository,
	msgCache message.MessageCacheRepository,
	wrapperDEK crypto.WrapperDEK,
	aesgcmEnc crypto.AESGCMEncryptor,
) MessageService {
	return &messageService{
		cfg:        cfg,
		log:        log,
		msgRepo:    msgRepo,
		msgCache:   msgCache,
		wrapperDEK: wrapperDEK,
		aesgcmEnc:  aesgcmEnc,
	}
}

// Create a new message in chat
// message body will be encoded by DEK
// REQUIRES: jwt-token of sender
func (ms *messageService) CreateMessage(ctx context.Context, req *messagepb.CreateRequest) (*messagepb.MessageModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := ms.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
	defer cancel()

	encryptedBody, nonce, err := ms.getEncryptedBody(ctxTimeout, []byte(req.Body), chatID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		ms.log.Error("failed to encrypt body", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	msg, err := message.New(
		chatID,
		userID,
		encryptedBody,
		nonce,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	savedMsg, err := ms.msgRepo.Save(ctxTimeout, msg)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		ms.log.Error("failed to save msg into db", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go ms.publishMessage(context.Background(), savedMsg)
	return &messagepb.MessageModel{
		Id:       savedMsg.ID().String(),
		ChatId:   savedMsg.ChatID().String(),
		AuthorId: savedMsg.AuthorID().String(),
		Body:     req.Body,
		SentAt:   timestamppb.New(savedMsg.SentAt()),
	}, nil
}

// Delete the message
// message can be deleted only by sender
// REQUIRES: jwt-token of sender
func (ms *messageService) DeleteMessage(ctx context.Context, req *messagepb.DeleteRequest) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	userID, err := ms.getUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	msgID, err := uuid.Parse(req.MsgId)
	if err != nil {
		return status.Error(codes.InvalidArgument, consts.ErrInvalidMsgID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
	defer cancel()

	if err := ms.msgRepo.Delete(ctxTimeout, msgID, userID); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrNoEnoughRights) {
			return status.Error(codes.PermissionDenied, err.Error())
		}

		if errors.Is(err, consts.ErrMsgNotFound) {
			return status.Error(codes.NotFound, err.Error())
		}

		ms.log.Error("failed to delete message", slog.String("error", err.Error()))
		return status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return nil
}

// Subscribe to get new messages from chat
// only members of chat can subscribe
// REQUIRES: jwt-token of user
func (ms *messageService) ConnectNewMessages(ctx context.Context, req *messagepb.ConnectRequest, stream grpc.ServerStreamingServer[messagepb.MessageModel]) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, err.Error())
	}

	userID, err := ms.getUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
	defer cancel()

	if !ms.msgRepo.IsUserMember(ctxTimeout, chatID, userID) {
		return status.Error(codes.PermissionDenied, consts.ErrNoEnoughRights.Error())
	}

	ms.msgCache.Subscribe(ctx, chatID, func(ctx context.Context, m *message.Message) {
		ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
		defer cancel()

		encodedBody, err := ms.getDecodedBody(ctxTimeout, m.EncryptedBody(), m.Nonce(), m.ChatID())
		if err != nil {
			ms.log.Warn("failed to decode message", slog.String("error", err.Error()))
			return
		}

		stream.Send(&messagepb.MessageModel{
			Id:       m.ID().String(),
			ChatId:   m.ChatID().String(),
			AuthorId: m.AuthorID().String(),
			Body:     string(encodedBody),
			SentAt:   timestamppb.New(m.SentAt()),
		})
	})

	return nil
}

// Get history of chat
// id is uuid v7
// REQUIRES: jwt-token of user
func (ms *messageService) GetHistory(ctx context.Context, req *messagepb.GetRequest) (*messagepb.MessageModels, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := ms.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	msgID := uuid.Nil
	if req.BeforeId != "" {
		msgID, err = uuid.Parse(req.BeforeId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidMsgID.Error())
		}
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
	defer cancel()

	history, err := ms.msgRepo.GetHistory(ctxTimeout, chatID, userID, msgID, req.Limit)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrNoEnoughRights) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		if errors.Is(err, consts.ErrTooBigLimit) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		ms.log.Error("failed to get history of chat", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := messagepb.MessageModels{
		Messages: make([]*messagepb.MessageModel, len(history)),
		HasMore:  len(history) == int(req.Limit),
	}
	for idx, msg := range history {
		encryptedBody, err := ms.getDecodedBody(ctxTimeout, msg.EncryptedBody(), msg.Nonce(), msg.ChatID())
		if err != nil {
			ms.log.Warn("failed to decode message", slog.String("error", err.Error()))
			continue
		}

		response.Messages[idx] = &messagepb.MessageModel{
			Id:       msg.ID().String(),
			ChatId:   msg.ChatID().String(),
			AuthorId: msg.AuthorID().String(),
			Body:     string(encryptedBody),
			SentAt:   timestamppb.New(msg.SentAt()),
		}
	}

	return &response, nil
}

func (ms *messageService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(session.KeyUserID)
	if raw == nil {
		return uuid.Nil, consts.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, consts.ErrInvalidToken
}

// This method returns encryptedBody, nonce and error
func (ms *messageService) getEncryptedBody(ctx context.Context, body []byte, chatID uuid.UUID) ([]byte, []byte, error) {
	wrappedDEK, err := ms.msgRepo.GetWrappedDEK(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}

	// FIXME: change hardcode KEK to KMS
	dek, err := ms.wrapperDEK.UnwrapDEK(wrappedDEK, []byte("4af1ecd1f5f734471fb942f7078c8786"))
	if err != nil {
		return nil, nil, err
	}

	return ms.aesgcmEnc.Encode(body, dek)
}

func (ms *messageService) getDecodedBody(ctx context.Context, encryptedBody, nonce []byte, chatID uuid.UUID) ([]byte, error) {
	wrappedDEK, err := ms.msgRepo.GetWrappedDEK(ctx, chatID)
	if err != nil {
		return nil, err
	}

	// FIXME: change hardcode KEK to KMS
	dek, err := ms.wrapperDEK.UnwrapDEK(wrappedDEK, []byte("4af1ecd1f5f734471fb942f7078c8786"))
	if err != nil {
		return nil, err
	}

	return ms.aesgcmEnc.Decode(encryptedBody, nonce, dek)
}

func (ms *messageService) publishMessage(ctx context.Context, msg *message.Message) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ms.cfg.Server.Timeout)
	defer cancel()

	if err := ms.msgCache.Publish(ctxTimeout, msg); err != nil {
		ms.log.Error("failed to publish message", slog.String("error", err.Error()))
	}
}
