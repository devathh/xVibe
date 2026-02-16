package handlers

import (
	"context"

	authpb "github.com/devathh/xvibe/auth-service/api/auth/v1"
	"github.com/devathh/xvibe/auth-service/internal/application/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerAPI struct {
	authpb.UnimplementedAuthServer
	service services.AuthService
}

func New(service services.AuthService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Refresh(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) Update(ctx context.Context, req *authpb.UpdateRequest) (*authpb.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetSelf(ctx context.Context, _ *emptypb.Empty) (*authpb.User, error) {
	resp, err := sapi.service.GetSelf(ctx)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetUserByID(ctx context.Context, req *authpb.GetByIDRequest) (*authpb.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.GetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetUsersByUsername(ctx context.Context, req *authpb.GetByUsernameRequest) (*authpb.Users, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	resp, err := sapi.service.GetUsersByUsername(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetPublicKey(ctx context.Context, _ *emptypb.Empty) (*authpb.PublicKey, error) {
	resp, err := sapi.service.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) LogoutAll(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := sapi.service.LogoutAll(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}
