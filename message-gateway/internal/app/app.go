package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	messagepb "github.com/devathh/xvibe/message-gateway/api/message/v1"
	"github.com/devathh/xvibe/message-gateway/internal/application/services"
	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/config"
	xvibegrpc "github.com/devathh/xvibe/message-gateway/internal/infrastructure/grpc/xvibe"
	httpserver "github.com/devathh/xvibe/message-gateway/internal/infrastructure/http"
	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/http/handlers"
	"github.com/devathh/xvibe/message-gateway/pkg/log"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
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
	a.log.Info("server shutdown")
	return a.server.Shutdown(ctx)
}

func New() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	messageConn, messageClient, err := provideGRPC(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide grpc: %w", err)
	}

	server, err := provideServer(cfg, log, messageClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide server: %w", err)
	}

	return &App{
			log:    log,
			server: server,
		}, func() {
			if err := xvibegrpc.Close(messageConn); err != nil {
				log.Error("failed to close message conn", slog.String("error", err.Error()))
			} else {
				log.Info("message conn was closed")
			}
		}, nil
}

func provideServer(cfg *config.Config, log *slog.Logger, messageClient messagepb.MessageClient) (*httpserver.Server, error) {
	service := services.New(cfg, log, messageClient)
	handler, err := handlers.New(cfg, log, service)
	if err != nil {
		return nil, err
	}

	return httpserver.New(cfg, handler), nil
}

func provideGRPC(cfg *config.Config) (*grpc.ClientConn, messagepb.MessageClient, error) {
	var (
		client *grpc.ClientConn
		err    error
	)

	if cfg.Services.XVBMessage.TLS.Enable {
		client, err = xvibegrpc.ConnectMTLSMessage(cfg)
	} else {
		client, err = xvibegrpc.ConnectInsecureMessage(cfg)
	}

	if err != nil {
		return nil, nil, err
	}

	return client, messagepb.NewMessageClient(client), nil
}
