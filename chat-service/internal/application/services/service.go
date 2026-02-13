package services

import (
	"context"
	"errors"
	"log/slog"

	chatpb "github.com/devathh/xvibe/chat/api/chat/v1"
	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/domain/member"
	"github.com/devathh/xvibe/chat/internal/domain/session"
	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatService interface {
	Create(context.Context, *chatpb.CreateRequest) (*chatpb.ChatModel, error)
	Delete(context.Context, *chatpb.DeleteRequest) (*emptypb.Empty, error)
	UpdateGroup(context.Context, *chatpb.UpdateRequest) (*chatpb.ChatWithMembers, error)
	AddMembers(context.Context, *chatpb.MembersRequest) (*chatpb.ChatWithMembers, error)
	DeleteMembers(context.Context, *chatpb.MembersRequest) (*emptypb.Empty, error)
	GetSelfChats(context.Context) (*chatpb.ChatModels, error)
	GetChat(context.Context, *chatpb.GetRequest) (*chatpb.ChatWithMembers, error)
}

type chatService struct {
	cfg       *config.Config
	log       *slog.Logger
	chatRepo  chat.ChatRepository
	chatCache chat.ChatCacheRepository
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	chatRepo chat.ChatRepository,
	chatCache chat.ChatCacheRepository,
) ChatService {
	return &chatService{
		cfg:       cfg,
		log:       log,
		chatRepo:  chatRepo,
		chatCache: chatCache,
	}
}

// Add new member into group
// if user is already exist in group, nothing happened
// REQUIRES: jwt-token (to check is owner)
func (c *chatService) AddMembers(ctx context.Context, req *chatpb.MembersRequest) (*chatpb.ChatWithMembers, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	memberIds := make([]uuid.UUID, len(req.MemberIds))
	for idx, id := range req.MemberIds {
		memberIds[idx], err = uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidMemberID.Error())
		}
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	chat, members, err := c.chatRepo.AddMembers(
		ctxTimeout,
		memberIds,
		chatID,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, consts.ErrUsersDontExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		c.log.Error("failed to add members into group", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := chatpb.ChatWithMembers{
		Chat: &chatpb.ChatModel{
			Id:        chat.ID().String(),
			OwnerId:   chat.OwnerID().String(),
			Title:     chat.Title(),
			Typ:       chatpb.Type(chat.Type().Value()),
			CreatedAt: timestamppb.New(chat.CreatedAt()),
		},
		MemberIds:      make([]*chatpb.Member, len(members)),
		IsCurrentOwner: chat.OwnerID() == userID,
		Timestamp:      timestamppb.Now(),
	}
	for idx, member := range members {
		response.MemberIds[idx] = &chatpb.Member{
			UserId:    member.ID().String(),
			Firstname: member.Firstname(),
			Lastname:  member.Lastname(),
			IsOwner:   member.IsOwner(),
		}
	}

	return &response, nil
}

// Create new chat into db
// (can be private or group by req)
// Owner's id is user id from token
// REQUIRES: jwt-token
func (c *chatService) Create(ctx context.Context, req *chatpb.CreateRequest) (*chatpb.ChatModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chat, members, err := c.createChat(req, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	savedChat, err := c.chatRepo.Save(ctxTimeout, chat, members)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUsersDontExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		c.log.Error("failed to save chat into db", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	if err := c.chatCache.Del(ctxTimeout, []*member.Member{member.New(
		userID,
		"",
		"",
		false,
	)}); err != nil {
		c.log.Error("failed to clear cache", slog.String("error", err.Error()))
	}

	return &chatpb.ChatModel{
		Id:        savedChat.ID().String(),
		OwnerId:   savedChat.OwnerID().String(),
		Title:     savedChat.Title(),
		Typ:       chatpb.Type(savedChat.Type().Value()),
		CreatedAt: timestamppb.New(savedChat.CreatedAt()),
	}, nil
}

// Delete chat
// can be deleted only by owner
// REQUIRES: jwt-token
func (c *chatService) Delete(ctx context.Context, req *chatpb.DeleteRequest) (*emptypb.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	if err := c.chatRepo.Delete(ctxTimeout, chatID, userID); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrUserIsntOwner) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		c.log.Error("failed to delete chat", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	if err := c.chatCache.Del(ctxTimeout, []*member.Member{member.New(
		userID,
		"",
		"",
		false,
	)}); err != nil {
		c.log.Error("failed to clear cache", slog.String("error", err.Error()))
	}

	return nil, nil
}

// Delete members from group
// you can't delete owner of current group
func (c *chatService) DeleteMembers(ctx context.Context, req *chatpb.MembersRequest) (*emptypb.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	memberIds := make([]uuid.UUID, len(req.MemberIds))
	for idx, id := range req.MemberIds {
		memberIds[idx], err = uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidMemberID.Error())
		}
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	if err := c.chatRepo.DeleteMembers(
		ctxTimeout,
		chatID,
		memberIds,
	); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, consts.ErrDeleteOwner) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		c.log.Error("failed to delete members from group", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return nil, nil
}

// Get one chat by id
func (c *chatService) GetChat(ctx context.Context, req *chatpb.GetRequest) (*chatpb.ChatWithMembers, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	chat, members, err := c.chatRepo.GetChat(ctxTimeout, chatID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		c.log.Error("failed to get chat", slog.String("chat_id", chatID.String()), slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := chatpb.ChatWithMembers{
		Chat: &chatpb.ChatModel{
			Id:        chat.ID().String(),
			OwnerId:   chat.OwnerID().String(),
			Title:     chat.Title(),
			Typ:       chatpb.Type(chat.Type().Value()),
			CreatedAt: timestamppb.New(chat.CreatedAt()),
		},
		MemberIds:      make([]*chatpb.Member, len(members)),
		IsCurrentOwner: chat.OwnerID() == userID,
		Timestamp:      timestamppb.Now(),
	}
	for idx, member := range members {
		response.MemberIds[idx] = &chatpb.Member{
			UserId:    member.ID().String(),
			Firstname: member.Firstname(),
			Lastname:  member.Lastname(),
			IsOwner:   member.IsOwner(),
		}
	}

	return &response, nil
}

func (c *chatService) GetSelfChats(ctx context.Context) (*chatpb.ChatModels, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	if chats, err := c.chatCache.Get(ctxTimeout, userID); err == nil {
		c.log.Debug("chats are taken from cache")
		return c.returnChats(chats), nil
	}

	chats, err := c.chatRepo.GetSelfChats(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		c.log.Error("failed to get self chats", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go c.saveChatsCache(context.Background(), chats, userID)
	return c.returnChats(chats), nil
}

// Update title of group
// can be updated only by owner
// REQUIRES: jwt-token
func (c *chatService) UpdateGroup(ctx context.Context, req *chatpb.UpdateRequest) (*chatpb.ChatWithMembers, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidChatID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	chat, members, err := c.chatRepo.Update(
		ctxTimeout,
		chatID,
		userID,
		req.Title,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrChatNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, consts.ErrUserIsntOwner) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		c.log.Error("failed to update title of group", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := chatpb.ChatWithMembers{
		Chat: &chatpb.ChatModel{
			Id:        chat.ID().String(),
			OwnerId:   chat.OwnerID().String(),
			Title:     chat.Title(),
			Typ:       chatpb.Type(chat.Type().Value()),
			CreatedAt: timestamppb.New(chat.CreatedAt()),
		},
		MemberIds:      make([]*chatpb.Member, len(members)),
		IsCurrentOwner: chat.OwnerID() == userID,
		Timestamp:      timestamppb.Now(),
	}
	for idx, member := range members {
		response.MemberIds[idx] = &chatpb.Member{
			UserId:    member.ID().String(),
			Firstname: member.Firstname(),
			Lastname:  member.Lastname(),
			IsOwner:   member.IsOwner(),
		}
	}

	go c.clearCache(context.Background(), members)
	return &response, nil
}

// Create group or private chat
// Must be used only in create method
// userID is taken from jwt-token
func (c *chatService) createChat(req *chatpb.CreateRequest, userID uuid.UUID) (*chat.ChatModel, []uuid.UUID, error) {
	typ, err := chat.NewTypeRaw(int(req.Typ))
	if err != nil {
		return nil, nil, err
	}

	switch payload := req.Creates.(type) {
	// Try to create the group
	case *chatpb.CreateRequest_CreateGroup:
		if typ.Value() != chat.TYPE_GROUP {
			return nil, nil, consts.ErrInvalidType
		}

		chatModel, err := chat.New(
			payload.CreateGroup.Title,
			typ,
			userID,
		)
		if err != nil {
			return nil, nil, err
		}

		memberIds := make([]uuid.UUID, len(payload.CreateGroup.MemberIds)+1)
		for i, id := range payload.CreateGroup.MemberIds {
			if userID.String() == id {
				continue
			}

			memberIds[i], err = uuid.Parse(id)
			if err != nil {
				return nil, nil, consts.ErrInvalidMemberID
			}
		}
		memberIds[len(memberIds)-1] = userID

		return chatModel, memberIds, nil
	// Try to create self chat (one to one)
	case *chatpb.CreateRequest_CreateSelf:
		if typ.Value() != chat.TYPE_SELF {
			return nil, nil, consts.ErrInvalidType
		}

		chatModel, err := chat.New(
			"",
			typ,
			userID,
		)
		if err != nil {
			return nil, nil, err
		}

		memberID, err := uuid.Parse(payload.CreateSelf.MemberId)
		if err != nil {
			return nil, nil, err
		}

		if memberID == userID {
			return nil, nil, consts.ErrSelfChat
		}

		return chatModel, []uuid.UUID{
			memberID,
			userID,
		}, nil
	}

	return nil, nil, consts.ErrInvalidRequest
}

// Get user's id from context
// user's id context is passed through the interceptor
func (c *chatService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(session.KeyUserID)
	if raw == nil {
		return uuid.Nil, consts.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, consts.ErrInvalidToken
}

// Convert domain to response
func (c *chatService) returnChats(chats []*chat.ChatModel) *chatpb.ChatModels {
	response := chatpb.ChatModels{
		Chats:     make([]*chatpb.ChatModel, len(chats)),
		Timestamp: timestamppb.Now(),
	}
	for idx, chat := range chats {
		response.Chats[idx] = &chatpb.ChatModel{
			Id:        chat.ID().String(),
			OwnerId:   chat.OwnerID().String(),
			Title:     chat.Title(),
			Typ:       chatpb.Type(chat.Type().Value()),
			CreatedAt: timestamppb.New(chat.CreatedAt()),
		}
	}

	return &response
}

func (c *chatService) saveChatsCache(ctx context.Context, chats []*chat.ChatModel, userID uuid.UUID) {
	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	if err := c.chatCache.Save(ctxTimeout, userID, chats); err != nil {
		c.log.Error("failed to save chats into cache", slog.String("error", err.Error()))
	}
}

func (c *chatService) clearCache(ctx context.Context, members []*member.Member) {
	ctxTimeout, cancel := context.WithTimeout(ctx, c.cfg.Server.Timeout)
	defer cancel()

	if err := c.chatCache.Del(ctxTimeout, members); err != nil {
		c.log.Error("failed to delete all cache", slog.String("error", err.Error()))
	}
}
