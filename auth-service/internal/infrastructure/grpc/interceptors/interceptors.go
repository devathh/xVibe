package interceptors

import (
	"context"

	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InterceptorsPack struct {
	jwtMngr     session.JwtManager
	authRequire map[string]bool
}

func New(jwtMngr session.JwtManager, authRequire map[string]bool) *InterceptorsPack {
	return &InterceptorsPack{
		jwtMngr:     jwtMngr,
		authRequire: authRequire,
	}
}

func (ip *InterceptorsPack) BaseInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata")
		}

		clientIP := md.Get("x-client-ip")
		if len(clientIP) < 1 {
			return nil, status.Error(codes.InvalidArgument, "ip is required")
		}

		userAgent := md.Get("x-client-user-agent")
		if len(userAgent) < 1 {
			return nil, status.Error(codes.InvalidArgument, "user-agent is required")
		}

		ctx = context.WithValue(ctx, session.ClientIPKey, clientIP[0])
		ctx = context.WithValue(ctx, session.UserAgentKey, userAgent[0])

		return handler(ctx, req)
	}
}

func (ip *InterceptorsPack) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !ip.authRequire[info.FullMethod] {
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

		claims, err := ip.jwtMngr.Validate(token[0])
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidToken.Error())
		}

		ctx = context.WithValue(ctx, session.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, session.UserEmailKey, claims.Email)

		return handler(ctx, req)
	}
}
