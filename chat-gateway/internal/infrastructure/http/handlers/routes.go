package handlers

import (
	"net/http"

	"github.com/devathh/xvibe/chat-gateway/internal/application/dtos"
	"github.com/devathh/xvibe/chat-gateway/internal/application/services"
	"github.com/devathh/xvibe/chat-gateway/pkg/consts"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	service services.ChatGatewayService
}

func NewRoutes(service services.ChatGatewayService) *Routes {
	return &Routes{
		service: service,
	}
}

func (r *Routes) Create() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.CreateRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		resp, code, err := r.service.Create(ctx, token, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.DeleteRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		code, err := r.service.Delete(ctx, token, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, nil)
	}
}

func (r *Routes) UpdateGroup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.UpdateRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		resp, code, err := r.service.UpdateGroup(ctx, token, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) AddMembers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.MembersRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		resp, code, err := r.service.AddMembers(ctx, token, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) DeleteMembers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.MembersRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		code, err := r.service.DeleteMembers(ctx, &req)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, nil)
	}
}

func (r *Routes) GetSelfChats() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		resp, code, err := r.service.GetSelfChats(ctx, token)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) GetChat() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		chatID := ctx.Query("chatid")
		if chatID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid chat's id",
			})
			return
		}

		token, err := ctx.Cookie("ac")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": consts.ErrInvalidToken.Error(),
			})
			return
		}

		resp, code, err := r.service.GetChat(ctx, token, chatID)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}
