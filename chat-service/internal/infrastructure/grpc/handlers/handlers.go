package handlers

import (
	"context"

	chatpb "github.com/devathh/xvibe/chat/api/chat/v1"
	"github.com/devathh/xvibe/chat/internal/application/services"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerAPI struct {
	chatpb.UnimplementedChatServer
	service services.ChatService
}

func New(service services.ChatService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) AddMembers(ctx context.Context, req *chatpb.MembersRequest) (*chatpb.ChatWithMembers, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.AddMembers(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) Create(ctx context.Context, req *chatpb.CreateRequest) (*chatpb.ChatModel, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) Delete(ctx context.Context, req *chatpb.DeleteRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Delete(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) DeleteMembers(ctx context.Context, req *chatpb.MembersRequest) (*empty.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.DeleteMembers(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetChat(ctx context.Context, req *chatpb.GetRequest) (*chatpb.ChatWithMembers, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.GetChat(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetSelfChats(ctx context.Context, _ *emptypb.Empty) (*chatpb.ChatModels, error) {
	resp, err := sapi.service.GetSelfChats(ctx)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) UpdateGroup(ctx context.Context, req *chatpb.UpdateRequest) (*chatpb.ChatWithMembers, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.UpdateGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
