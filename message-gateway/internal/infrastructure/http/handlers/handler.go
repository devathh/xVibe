package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/devathh/xvibe/message-gateway/internal/application/services"
	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, log *slog.Logger, service services.MessageGatewayService) (http.Handler, error) {
	var router *gin.Engine
	switch cfg.Env {
	case "prod":
		router = gin.New()
		router.Use(gin.Recovery())
	case "dev", "local":
		router = gin.Default()
	default:
		return nil, errors.New("invalid environment")
	}

	routes := Routes{
		log:     log,
		service: service,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/message", routes.CreateMessage())
			v1.DELETE("/message", routes.DeleteMessage())
			v1.GET("/:chatid/history", routes.GetHistory())

			v1.GET("/:chatid/messages", routes.ConnectNewMessages()) // ws
		}
	}

	return router, nil
}
