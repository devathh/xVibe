package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	chatpb "github.com/devathh/xvibe/chat-gateway/api/chat/v1"
	"github.com/devathh/xvibe/chat-gateway/internal/application/services"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/grpc/xvibe"
	httpserver "github.com/devathh/xvibe/chat-gateway/internal/infrastructure/http"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/http/handlers"
	"github.com/devathh/xvibe/chat-gateway/pkg/log"
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
	a.log.Info("shutdown...")
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
		return nil, nil, fmt.Errorf("failed to setup log handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	clientConn, chatClient, err := provideGRPC(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide grpc: %w", err)
	}

	service := services.New(cfg, log, chatClient)
	handler, err := handlers.New(cfg, service)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load handlers: %w", err)
	}

	return &App{
			log:    log,
			server: httpserver.New(cfg, handler),
		}, func() {
			if err := xvibe.Close(clientConn); err != nil {
				log.Error("failed to close client connection", slog.String("error", err.Error()))
			} else {
				log.Info("connection with xvibe-chat was closed")
			}
		}, nil
}

func provideGRPC(cfg *config.Config) (*grpc.ClientConn, chatpb.ChatClient, error) {
	var (
		client *grpc.ClientConn
		err    error
	)

	if !cfg.Services.XvibeChat.TLS.Enable {
		client, err = xvibe.ConnectInsecureXvibeChat(cfg)
	} else {
		client, err = xvibe.ConnectMTLSXvibeChat(cfg)
	}
	if err != nil {
		return nil, nil, err
	}

	return client, chatpb.NewChatClient(client), nil
}
