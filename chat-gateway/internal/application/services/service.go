package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	chatpb "github.com/devathh/xvibe/chat-gateway/api/chat/v1"
	"github.com/devathh/xvibe/chat-gateway/internal/application/dtos"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat-gateway/pkg/consts"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ChatGatewayService interface {
	Create(ctx context.Context, token string, req *dtos.CreateRequest) (*dtos.ChatModel, int, error)
	Delete(ctx context.Context, token string, req *dtos.DeleteRequest) (int, error)
	UpdateGroup(ctx context.Context, token string, req *dtos.UpdateRequest) (*dtos.ChatWithMembers, int, error)
	AddMembers(ctx context.Context, token string, req *dtos.MembersRequest) (*dtos.ChatWithMembers, int, error)
	DeleteMembers(ctx context.Context, req *dtos.MembersRequest) (int, error)
	GetSelfChats(ctx context.Context, token string) (*dtos.ChatModels, int, error)
	GetChat(ctx context.Context, token, chatID string) (*dtos.ChatWithMembers, int, error)
}

type chatGatewayService struct {
	cfg        *config.Config
	log        *slog.Logger
	chatClient chatpb.ChatClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	chatClient chatpb.ChatClient,
) ChatGatewayService {
	return &chatGatewayService{
		cfg:        cfg,
		log:        log,
		chatClient: chatClient,
	}
}

// Create method of current gateway service
// calls Create method in chat-service
// REQUIRES: jwt-token
func (cgs *chatGatewayService) Create(ctx context.Context, token string, req *dtos.CreateRequest) (*dtos.ChatModel, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	if (req.TypeID == consts.TYPE_SELF && req.CreateSelf == nil) ||
		(req.TypeID == consts.TYPE_GROUP && req.CreateGroup == nil) {
		return nil, http.StatusBadRequest, consts.ErrInvalidChatType
	}

	request, err := cgs.generateCreateRequest(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	response, err := cgs.chatClient.Create(metadata.NewOutgoingContext(ctx, md), request)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		cgs.log.Error("unknown code from chat-service has recieved",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.ChatModel{
		ID:         response.Id,
		OwnerID:    response.OwnerId,
		Title:      response.Title,
		TypeID:     int(response.Typ),
		TypeString: response.Typ.String(),
		CreatedAt:  response.CreatedAt.AsTime().UnixMilli(),
	}, http.StatusCreated, nil
}

// Delete method of current gateway service
// calls Delete method in chat-service
// REQUIRES: jwt-token (only owner can delete group)
func (cgs *chatGatewayService) Delete(ctx context.Context, token string, req *dtos.DeleteRequest) (int, error) {
	if err := ctx.Err(); err != nil {
		return http.StatusRequestTimeout, err
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	if _, err := cgs.chatClient.Delete(metadata.NewOutgoingContext(ctx, md), &chatpb.DeleteRequest{
		ChatId: req.ChatID,
	}); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return http.StatusInternalServerError, consts.ErrInternalServer
	}

	return http.StatusNoContent, nil
}

// UpdateGroup method of current gateway service
// calls UpdateGroup method in chat-service
// REQUIRES: jwt-token (only owner can update title of group)
func (cgs *chatGatewayService) UpdateGroup(ctx context.Context, token string, req *dtos.UpdateRequest) (*dtos.ChatWithMembers, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return nil, http.StatusBadRequest, consts.ErrEmptyTitle
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	resp, err := cgs.chatClient.UpdateGroup(metadata.NewOutgoingContext(ctx, md), &chatpb.UpdateRequest{
		ChatId: req.ChatID,
		Title:  req.Title,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	response := dtos.ChatWithMembers{
		ChatModel: dtos.ChatModel{
			ID:         resp.Chat.Id,
			OwnerID:    resp.Chat.OwnerId,
			Title:      resp.Chat.Title,
			TypeID:     int(resp.Chat.Typ),
			TypeString: resp.Chat.Typ.String(),
			CreatedAt:  resp.Chat.CreatedAt.AsTime().UnixMilli(),
		},
		Members:        make([]dtos.Member, len(resp.MemberIds)),
		IsCurrentOwner: resp.IsCurrentOwner,
		Timestamp:      resp.Timestamp.AsTime().UnixMilli(),
	}
	for idx, member := range resp.MemberIds {
		response.Members[idx] = dtos.Member{
			UserID:    member.UserId,
			Firstname: member.Firstname,
			Lastname:  member.Lastname,
			IsOwner:   member.IsOwner,
		}
	}

	return &response, http.StatusOK, nil
}

// AddMembers method of current gateway service
// calls AddMembers method in chat-service
// REQUIRES: jwt-token (to fill is_owner field)
func (cgs *chatGatewayService) AddMembers(ctx context.Context, token string, req *dtos.MembersRequest) (*dtos.ChatWithMembers, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	if len(req.MemberIds) < 1 {
		return nil, http.StatusBadRequest, consts.ErrEmptyMembers
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	resp, err := cgs.chatClient.AddMembers(metadata.NewOutgoingContext(ctx, md), &chatpb.MembersRequest{
		ChatId:    req.ChatID,
		MemberIds: req.MemberIds,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	response := dtos.ChatWithMembers{
		ChatModel: dtos.ChatModel{
			ID:         resp.Chat.Id,
			OwnerID:    resp.Chat.OwnerId,
			Title:      resp.Chat.Title,
			TypeID:     int(resp.Chat.Typ),
			TypeString: resp.Chat.Typ.String(),
			CreatedAt:  resp.Chat.CreatedAt.AsTime().UnixMilli(),
		},
		Members:        make([]dtos.Member, len(resp.MemberIds)),
		IsCurrentOwner: resp.IsCurrentOwner,
		Timestamp:      resp.Timestamp.AsTime().UnixMilli(),
	}
	for idx, member := range resp.MemberIds {
		response.Members[idx] = dtos.Member{
			UserID:    member.UserId,
			Firstname: member.Firstname,
			Lastname:  member.Lastname,
			IsOwner:   member.IsOwner,
		}
	}

	return &response, http.StatusOK, nil
}

// DeleteMembers method of current gateway service
// calls DeleteMembers method in chat-service
// REQUIRES: jwt-token
func (cgs *chatGatewayService) DeleteMembers(ctx context.Context, req *dtos.MembersRequest) (int, error) {
	if err := ctx.Err(); err != nil {
		return http.StatusRequestTimeout, err
	}

	if len(req.MemberIds) < 1 {
		return http.StatusBadRequest, consts.ErrEmptyMembers
	}

	if _, err := cgs.chatClient.DeleteMembers(ctx, &chatpb.MembersRequest{
		ChatId:    req.ChatID,
		MemberIds: req.MemberIds,
	}); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return http.StatusInternalServerError, consts.ErrInternalServer
	}

	return http.StatusNoContent, nil
}

// GetSelfChats method of current gateway service
// calls GetSelfChats method in chat-service
// REQUIRES: jwt-token (to get chats of THIS user)
func (cgs *chatGatewayService) GetSelfChats(ctx context.Context, token string) (*dtos.ChatModels, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	resp, err := cgs.chatClient.GetSelfChats(metadata.NewOutgoingContext(ctx, md), nil)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	response := dtos.ChatModels{
		Chats:     make([]dtos.ChatModel, len(resp.Chats)),
		Timestamp: resp.Timestamp.AsTime().UnixMilli(),
	}

	for idx, chat := range resp.Chats {
		response.Chats[idx] = dtos.ChatModel{
			ID:         chat.Id,
			OwnerID:    chat.OwnerId,
			Title:      chat.Title,
			TypeID:     int(chat.Typ),
			TypeString: chat.Typ.String(),
			CreatedAt:  chat.CreatedAt.AsTime().UnixMilli(),
		}
	}

	return &response, http.StatusOK, nil
}

// GetChat method of current gateway service
// calls GetChat method in chat-service
// REQUIRES: jwt-token
func (cgs *chatGatewayService) GetChat(ctx context.Context, token, chatID string) (*dtos.ChatWithMembers, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	resp, err := cgs.chatClient.GetChat(metadata.NewOutgoingContext(ctx, md), &chatpb.GetRequest{
		ChatId: chatID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		cgs.log.Error("has recieved unknown code from chat-service",
			slog.String("error", st.Message()),
			slog.String("code", st.Code().String()),
		)
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	response := dtos.ChatWithMembers{
		ChatModel: dtos.ChatModel{
			ID:         resp.Chat.Id,
			OwnerID:    resp.Chat.OwnerId,
			Title:      resp.Chat.Title,
			TypeID:     int(resp.Chat.Typ),
			TypeString: resp.Chat.Typ.String(),
			CreatedAt:  resp.Chat.CreatedAt.AsTime().UnixMilli(),
		},
		Members:        make([]dtos.Member, len(resp.MemberIds)),
		IsCurrentOwner: resp.IsCurrentOwner,
		Timestamp:      resp.Timestamp.AsTime().UnixMilli(),
	}
	for idx, member := range resp.MemberIds {
		response.Members[idx] = dtos.Member{
			UserID:    member.UserId,
			Firstname: member.Firstname,
			Lastname:  member.Lastname,
			IsOwner:   member.IsOwner,
		}
	}

	return &response, http.StatusOK, nil
}

func (cgs *chatGatewayService) generateCreateRequest(req *dtos.CreateRequest) (*chatpb.CreateRequest, error) {
	request := chatpb.CreateRequest{
		Typ: chatpb.Type(req.TypeID),
	}
	switch req.TypeID {
	case consts.TYPE_GROUP:
		request.Creates = &chatpb.CreateRequest_CreateGroup{
			CreateGroup: &chatpb.CreateGroup{
				Title:     req.CreateGroup.Title,
				MemberIds: req.CreateGroup.MemberIds,
			},
		}
	case consts.TYPE_SELF:
		request.Creates = &chatpb.CreateRequest_CreateSelf{
			CreateSelf: &chatpb.CreateSelf{
				MemberId: req.CreateSelf.MemberID,
			},
		}
	default:
		return nil, consts.ErrInvalidChatType
	}

	return &request, nil
}
