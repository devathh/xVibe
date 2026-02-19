package services

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	authpb "github.com/devathh/xvibe/auth-gateway/api/auth/v1"
	"github.com/devathh/xvibe/auth-gateway/internal/application/services/dtos"
	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-gateway/pkg/consts"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type AuthGatewayService interface {
	Register(ctx *gin.Context, req *dtos.RegisterRequest) (*dtos.Token, int, error)
	Login(ctx *gin.Context, req *dtos.LoginRequest) (*dtos.Token, int, error)
	Refresh(ctx *gin.Context, req *dtos.RefreshRequest) (*dtos.Token, int, error)
	Update(ctx *gin.Context, req *dtos.UpdateRequest) (*dtos.User, int, error)
	LogoutAll(ctx *gin.Context) (int, error)
	GetSelf(ctx *gin.Context) (*dtos.User, int, error)
	GetUserByID(ctx *gin.Context, id string) (*dtos.User, int, error)
}

type authGatewayService struct {
	cfg        *config.Config
	log        *slog.Logger
	authClient authpb.AuthClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	authClient authpb.AuthClient,
) AuthGatewayService {
	return &authGatewayService{
		cfg:        cfg,
		log:        log,
		authClient: authClient,
	}
}

// This method calls register from auth-service
// /api/v1/register POST
func (ags *authGatewayService) Register(ctx *gin.Context, req *dtos.RegisterRequest) (*dtos.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)

	resp, err := ags.authClient.Register(metadata.NewOutgoingContext(ctx, md), &authpb.RegisterRequest{
		Email:     strings.TrimSpace(req.Email),
		Password:  strings.TrimSpace(req.Password),
		Firstname: strings.TrimSpace(req.Firstname),
		Lastname:  strings.TrimSpace(req.Lastname),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
	}, http.StatusCreated, nil
}

// This method calls login from auth-service
// /api/v1/login POST
func (ags *authGatewayService) Login(ctx *gin.Context, req *dtos.LoginRequest) (*dtos.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)

	resp, err := ags.authClient.Login(metadata.NewOutgoingContext(ctx, md), &authpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

// This method calls refresh from auth-service
// /api/v1/refresh POST
func (ags *authGatewayService) Refresh(ctx *gin.Context, req *dtos.RefreshRequest) (*dtos.Token, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)

	resp, err := ags.authClient.Refresh(metadata.NewOutgoingContext(ctx, md), &authpb.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

// This method calls update from auth-service
// requires jwt-token (ac in cookies)
// /api/v1/user PATCH
func (ags *authGatewayService) Update(ctx *gin.Context, req *dtos.UpdateRequest) (*dtos.User, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)
	if err := ags.tokenMD(&md, ctx); err != nil {
		return nil, http.StatusUnauthorized, consts.ErrInvalidToken
	}

	resp, err := ags.authClient.Update(metadata.NewOutgoingContext(ctx, md), &authpb.UpdateRequest{
		Updates: &authpb.Updates{
			Firstname: req.Updates.Firstname,
			Lastname:  req.Updates.Lastname,
			Username:  req.Updates.Username,
			Email:     req.Updates.Email,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: req.UpdateMask,
		},
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		Firstname: resp.Firstname,
		Lastname:  resp.Lastname,
		Username:  resp.Username,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
		UpdatedAt: resp.UpdatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

// This method calls logout-all from auth-service
// require jwt-token (ac in cookies)
// /api/v1/logout DELETE
func (ags *authGatewayService) LogoutAll(ctx *gin.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)
	if err := ags.tokenMD(&md, ctx); err != nil {
		return http.StatusUnauthorized, consts.ErrInvalidToken
	}

	if _, err := ags.authClient.LogoutAll(metadata.NewOutgoingContext(ctx, md), nil); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return http.StatusInternalServerError, consts.ErrInternalServer
	}

	return http.StatusNoContent, nil
}

// This method calls get self from auth-service
// require jwt-token (ac in cookies)
// /api/v1/user GET
func (ags *authGatewayService) GetSelf(ctx *gin.Context) (*dtos.User, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)
	if err := ags.tokenMD(&md, ctx); err != nil {
		return nil, http.StatusUnauthorized, consts.ErrInvalidToken
	}

	resp, err := ags.authClient.GetSelf(metadata.NewOutgoingContext(ctx, md), nil)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		Firstname: resp.Firstname,
		Lastname:  resp.Lastname,
		Username:  resp.Username,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
		UpdatedAt: resp.UpdatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

// This method calls get user by id from auth-service
// /api/v1/user?id=123abc
func (ags *authGatewayService) GetUserByID(ctx *gin.Context, id string) (*dtos.User, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	md := metadata.MD{}
	ags.baseMD(&md, ctx)

	resp, err := ags.authClient.GetUserByID(metadata.NewOutgoingContext(ctx, md), &authpb.GetByIDRequest{
		Id: id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if code, ok := consts.CodeMap[st.Code()]; ok {
			return nil, code, errors.New(st.Message())
		}

		ags.log.Error("unknown code has receibed from xvibe-auth",
			slog.String("code_string", st.Code().String()),
			slog.Int("code_int", int(st.Code())))
		return nil, http.StatusInternalServerError, consts.ErrInternalServer
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		Firstname: resp.Firstname,
		Lastname:  resp.Lastname,
		Username:  resp.Username,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
		UpdatedAt: resp.UpdatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (ags *authGatewayService) baseMD(md *metadata.MD, ctx *gin.Context) {
	md.Set("x-client-ip", ctx.ClientIP())
	md.Set("x-client-user-agent", ctx.Request.UserAgent())
}

func (ags *authGatewayService) tokenMD(md *metadata.MD, ctx *gin.Context) error {
	token := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if token == "" {
		return consts.ErrInvalidToken
	}

	md.Set("authorization", token)
	return nil
}
