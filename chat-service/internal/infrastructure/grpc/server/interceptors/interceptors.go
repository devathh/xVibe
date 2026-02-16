package interceptors

import (
	"context"

	"github.com/devathh/xvibe/chat/internal/domain/session"
	"github.com/devathh/xvibe/chat/pkg/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PackInterceptors struct {
	jwtMngr     session.JWTManager
	authRequire map[string]bool
}

func New(jwtMngr session.JWTManager, authRequire map[string]bool) *PackInterceptors {
	return &PackInterceptors{
		jwtMngr:     jwtMngr,
		authRequire: authRequire,
	}
}

func (pi *PackInterceptors) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !pi.authRequire[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata")
		}

		token := md.Get("authorization")
		if len(token) < 1 {
			return nil, status.Error(codes.Unauthenticated, "empty token")
		}

		claims, err := pi.jwtMngr.Validate(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
		}

		ctx = context.WithValue(ctx, session.KeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, session.KeyUserEmail, claims.Email)

		return handler(ctx, req)
	}
}
