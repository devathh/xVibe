package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	authpb "github.com/devathh/xvibe/auth-gateway/api/auth/v1"
	"github.com/devathh/xvibe/auth-gateway/internal/application/services"
	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/config"
	xvibegrpc "github.com/devathh/xvibe/auth-gateway/internal/infrastructure/grpc/xvibe"
	httpserver "github.com/devathh/xvibe/auth-gateway/internal/infrastructure/http"
	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/http/handlers"
	"github.com/devathh/xvibe/auth-gateway/pkg/log"
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
		return nil, nil, fmt.Errorf("failed to setup log's handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	clientConn, authClient, err := provideXvibe(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide xvibe")
	}

	server, err := provideServer(cfg, log, authClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide server: %w", err)
	}

	return &App{
			log:    log,
			server: server,
		}, func() {
			if err := xvibegrpc.CloseAuth(clientConn); err != nil {
				log.Error("failed to close xvibe-auth", slog.String("error", err.Error()))
			} else {
				log.Info("connection with xvibe-auth was closed")
			}
		}, nil
}

func provideServer(cfg *config.Config, log *slog.Logger, authClient authpb.AuthClient) (*httpserver.Server, error) {
	service := services.New(cfg, log, authClient)
	handler, err := handlers.New(cfg, service)
	if err != nil {
		return nil, err
	}

	return httpserver.New(cfg, handler), nil
}

func provideXvibe(cfg *config.Config) (*grpc.ClientConn, authpb.AuthClient, error) {
	var (
		conn *grpc.ClientConn
		err  error
	)

	if !cfg.Services.XvibeAuth.MTLS.Enable {
		conn, err = xvibegrpc.ConnectInsecureAuth(cfg)
	} else {
		conn, err = xvibegrpc.ConnectMTLSAuth(cfg)
	}
	if err != nil {
		return nil, nil, err
	}

	return conn, authpb.NewAuthClient(conn), nil
}
