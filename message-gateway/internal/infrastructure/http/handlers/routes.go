package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/devathh/xvibe/message-gateway/internal/application/dtos"
	"github.com/devathh/xvibe/message-gateway/internal/application/services"
	"github.com/devathh/xvibe/message-gateway/pkg/consts"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/status"
)

type Routes struct {
	log     *slog.Logger
	service services.MessageGatewayService
}

func (r *Routes) CreateMessage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.CreateRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		if strings.TrimSpace(req.Body) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "body of message cannot be empty",
			})
			return
		}

		token, err := r.getToken(ctx)
		if err != nil || token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		resp, err := r.service.CreateMessage(ctx, &req, token)
		if err != nil {
			code, msg := r.returnErrCode(err)
			ctx.JSON(code, gin.H{
				"error": msg,
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (r *Routes) DeleteMessage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dtos.DeleteRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		token, err := r.getToken(ctx)
		if err != nil || token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = r.service.DeleteMessage(ctx, &req, token)
		if err != nil {
			code, msg := r.returnErrCode(err)
			ctx.JSON(code, gin.H{
				"error": msg,
			})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func (r *Routes) GetHistory() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := r.getToken(ctx)
		if err != nil || token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		chatID := ctx.Param("chatid")
		if chatID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid chat's id",
			})
			return
		}

		limit, err := strconv.Atoi(ctx.Query("limit"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit",
			})
			return
		}

		beforeID := ctx.Query("beforeid")

		resp, err := r.service.GetHistory(ctx, &dtos.GetRequest{
			ChatID:   chatID,
			Limit:    uint32(limit),
			BeforeID: beforeID,
		}, token)
		if err != nil {
			code, msg := r.returnErrCode(err)
			ctx.JSON(code, gin.H{
				"error": msg,
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (r *Routes) ConnectNewMessages() gin.HandlerFunc {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		chatID := ctx.Param("chatid")
		if chatID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid chat's id",
			})
			return
		}

		token, err := r.getToken(ctx)
		if err != nil || token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		r.service.ConnectNewMessages(ctx, chatID, token, func(m *dtos.MessageModel) error {
			msgBytes, err := json.Marshal(m)
			if err != nil {
				return err
			}

			return conn.WriteMessage(websocket.TextMessage, msgBytes)
		})
	}
}

func (r *Routes) returnErrCode(err error) (int, string) {
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusBadGateway, consts.ErrBadGateway.Error()
	}

	if code, ok := consts.CodeMap[st.Code()]; ok {
		return code, st.Message()
	}

	r.log.Error("unknown code has received from message-service",
		slog.String("error", st.Message()),
		slog.String("code", st.Code().String()),
	)
	return http.StatusInternalServerError, consts.ErrInternalServer.Error()
}

func (r *Routes) getToken(ctx *gin.Context) (string, error) {
	token := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if token == "" {
		return "", consts.ErrInvalidToken
	}

	return token, nil
}
