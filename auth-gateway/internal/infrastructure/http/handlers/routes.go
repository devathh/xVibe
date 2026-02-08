package handlers

import (
	"net/http"

	"github.com/devathh/xvibe/auth-gateway/internal/application/services"
	"github.com/devathh/xvibe/auth-gateway/internal/application/services/dtos"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	service services.AuthGatewayService
}

func NewRoutes(service services.AuthGatewayService) *Routes {
	return &Routes{
		service: service,
	}
}

func (r *Routes) Register() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.RegisterRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Register(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.LoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Login(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Refresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.RefreshRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Refresh(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.UpdateRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		resp, code, err := r.service.Update(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) LogoutAll() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code, err := r.service.LogoutAll(ctx)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, nil)
	}
}

func (r *Routes) GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Query("id")

		var (
			resp *dtos.User
			code int
			err  error
		)
		if id == "" {
			resp, code, err = r.service.GetSelf(ctx)
		} else {
			resp, code, err = r.service.GetUserByID(ctx, id)
		}

		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}
