package handlers

import (
	"errors"
	"net/http"

	"github.com/devathh/xvibe/chat-gateway/internal/application/services"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, service services.ChatGatewayService) (http.Handler, error) {
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

	routes := NewRoutes(service)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/chat", routes.Create())
			v1.DELETE("/chat", routes.Delete())
			v1.PATCH("/chat", routes.UpdateGroup())
			v1.GET("/chat", routes.GetChat()) // ?chatid=123abc
			v1.GET("/chats", routes.GetSelfChats())

			v1.POST("/members", routes.AddMembers())
			v1.DELETE("/members", routes.DeleteMembers())
		}
	}

	return router, nil
}
