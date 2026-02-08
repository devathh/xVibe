package handlers

import (
	"errors"
	"net/http"

	"github.com/devathh/xvibe/auth-gateway/internal/application/services"
	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, service services.AuthGatewayService) (http.Handler, error) {
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
			v1.POST("/register", routes.Register())
			v1.POST("/login", routes.Login())
			v1.POST("/refresh", routes.Refresh())
			v1.DELETE("/logout", routes.LogoutAll())

			v1.PATCH("/user", routes.Update())
			v1.GET("/user", routes.GetUser()) // id can be empty, or ex. ?id=123ab...
		}
	}

	return router, nil
}
