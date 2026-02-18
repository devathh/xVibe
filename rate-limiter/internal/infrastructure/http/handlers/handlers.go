package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/devathh/xvibe/rate-limiter/internal/infrastructure/config"
	"github.com/devathh/xvibe/rate-limiter/internal/infrastructure/http/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) (http.Handler, error) {
	router, err := loadRouter(cfg)
	if err != nil {
		return nil, err
	}

	proxy, err := loadProxy(cfg)
	if err != nil {
		return nil, err
	}

	router.NoRoute(func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return router, nil
}

// Create new router with level of logger
func loadRouter(cfg *config.Config) (*gin.Engine, error) {
	var router *gin.Engine

	switch cfg.Env {
	case "prod":
		router = gin.New()
		router.Use(gin.Recovery())
	case "dev", "local":
		router = gin.Default()
	default:
		return nil, errors.New("invalid env")
	}

	router.Use(middlewares.NewRateLimitMiddleware(
		cfg.Service.RateLimit.Limit,
		cfg.Service.RateLimit.Window,
	))
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"POST", "GET", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowHeaders:    []string{"Access-Control-Allow-Headers", "Content-Type", "Authorization"},
	}))

	return router, nil
}

// Load http proxy, that only change host
func loadProxy(cfg *config.Config) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(cfg.Secrets.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(
		targetURL,
	)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
	}

	return proxy, nil
}
