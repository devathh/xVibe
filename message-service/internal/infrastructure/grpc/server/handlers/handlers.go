package handlers

import (
	"context"

	messagepb "github.com/devathh/xvibe/message-service/api/message/v1"
	"github.com/devathh/xvibe/message-service/internal/application/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerAPI struct {
	messagepb.UnimplementedMessageServer
	service services.MessageService
}

func New(service services.MessageService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) CreateMessage(ctx context.Context, req *messagepb.CreateRequest) (*messagepb.MessageModel, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.CreateMessage(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) DeleteMessage(ctx context.Context, req *messagepb.DeleteRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	if err := sapi.service.DeleteMessage(ctx, req); err != nil {
		return nil, err
	}

	return nil, nil
}

func (sapi *ServerAPI) ConnectNewMessages(req *messagepb.ConnectRequest, stream grpc.ServerStreamingServer[messagepb.MessageModel]) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	return sapi.service.ConnectNewMessages(stream.Context(), req, stream)
}

func (sapi *ServerAPI) GetHistory(ctx context.Context, req *messagepb.GetRequest) (*messagepb.MessageModels, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.GetHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
