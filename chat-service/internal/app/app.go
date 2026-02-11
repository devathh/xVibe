package app

import (
	"fmt"
	"log/slog"
	"os"

	chatpb "github.com/devathh/xvibe/chat/api/chat/v1"
	"github.com/devathh/xvibe/chat/internal/application/services"
	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/cache/redis"
	chatredis "github.com/devathh/xvibe/chat/internal/infrastructure/cache/redis/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	grpcserver "github.com/devathh/xvibe/chat/internal/infrastructure/grpc"
	"github.com/devathh/xvibe/chat/internal/infrastructure/grpc/handlers"
	"github.com/devathh/xvibe/chat/internal/infrastructure/grpc/interceptors"
	"github.com/devathh/xvibe/chat/internal/infrastructure/persistence/postgres"
	chatpg "github.com/devathh/xvibe/chat/internal/infrastructure/persistence/postgres/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/session/jwt"
	"github.com/devathh/xvibe/chat/pkg/log"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	log    *slog.Logger
	server *grpcserver.Server
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) GracefulStop() {
	a.log.Info("graceful shutdown...")
	a.server.Close()
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

	db, chatRepo, err := providePostgres(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide postgres: %w", err)
	}

	redisClient, chatCacheRepo, err := provideRedis(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide redis: %w", err)
	}

	grpcServer, err := provideServer(
		cfg,
		log,
		chatRepo,
		chatCacheRepo,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide server: %w", err)
	}

	return &App{
			log:    log,
			server: grpcserver.New(cfg, grpcServer),
		}, func() {
			if err := postgres.Close(db); err != nil {
				log.Error("failed to close postgres connection", slog.String("error", err.Error()))
			} else {
				log.Info("connection with postgres was closed")
			}

			if err := redis.Close(redisClient); err != nil {
				log.Error("failed to close redis connection", slog.String("error", err.Error()))
			} else {
				log.Info("connection with redis was closed")
			}
		}, nil
}

// TODO: mtls-connection
func provideServer(cfg *config.Config, log *slog.Logger, chatRepo chat.ChatRepository, chatCacheRepo chat.ChatCacheRepository) (*grpc.Server, error) {
	service := services.New(cfg, log, chatRepo, chatCacheRepo)
	api := handlers.New(service)

	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, err
	}

	packInterceptors := interceptors.New(jwtMngr, map[string]bool{
		chatpb.Chat_Create_FullMethodName:       true,
		chatpb.Chat_Delete_FullMethodName:       true,
		chatpb.Chat_UpdateGroup_FullMethodName:  true,
		chatpb.Chat_GetSelfChats_FullMethodName: true,
		chatpb.Chat_GetChat_FullMethodName:      true,
		chatpb.Chat_AddMembers_FullMethodName:   true,
	})
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		packInterceptors.AuthInterceptor(),
	))
	chatpb.RegisterChatServer(grpcServer, api)

	return grpcServer, nil
}

func provideRedis(cfg *config.Config) (*redissdk.Client, *chatredis.ChatCacheRepository, error) {
	client, err := redis.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	return client, chatredis.New(cfg, client), nil
}

func providePostgres(cfg *config.Config) (*gorm.DB, *chatpg.ChatRepository, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	if err := postgres.Migrate(db); err != nil {
		return nil, nil, err
	}

	return db, chatpg.New(db), nil
}
