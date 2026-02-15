package interceptors

import (
	"context"

	"github.com/devathh/xvibe/message-service/internal/domain/session"
	"github.com/devathh/xvibe/message-service/pkg/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewAuthInterceptor(jwtMngr session.JWTManager, noAuth map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if noAuth[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata")
		}

		token := md.Get("authorization")
		if len(token) < 1 {
			return nil, status.Error(codes.InvalidArgument, "token is required")
		}

		claims, err := jwtMngr.Validate(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
		}

		ctx = context.WithValue(ctx, session.KeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, session.KeyUserEmail, claims.Email)

		return handler(ctx, req)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func NewAuthStreamInterceptor(jwtMngr session.JWTManager, noAuth map[string]bool) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if noAuth[info.FullMethod] {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.InvalidArgument, "invalid metadata")
		}

		token := md.Get("authorization")
		if len(token) < 1 {
			return status.Error(codes.InvalidArgument, "token is required")
		}

		claims, err := jwtMngr.Validate(token[0])
		if err != nil {
			return status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
		}

		ctx := context.WithValue(ss.Context(), session.KeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, session.KeyUserEmail, claims.Email)

		return handler(srv, &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}
