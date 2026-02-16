package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/devathh/xvibe/rate-limiter/internal/infrastructure/config"
	httpserver "github.com/devathh/xvibe/rate-limiter/internal/infrastructure/http"
	"github.com/devathh/xvibe/rate-limiter/internal/infrastructure/http/handlers"
	"github.com/devathh/xvibe/rate-limiter/pkg/log"
	"github.com/joho/godotenv"
)

type App struct {
	log    *slog.Logger
	server *httpserver.Server
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("server shutdown...")
	return a.server.Shutdown(ctx)
}

func New() (*App, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("failed to setup log's handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	server, err := provideServer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to provide server: %w", err)
	}

	return &App{
		log:    log,
		server: server,
	}, nil
}

func provideServer(cfg *config.Config) (*httpserver.Server, error) {
	handler, err := handlers.New(cfg)
	if err != nil {
		return nil, err
	}

	return httpserver.New(cfg, handler), nil
}
