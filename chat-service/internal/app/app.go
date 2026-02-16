package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	authpb "github.com/devathh/xvibe/chat/api/auth/v1"
	chatpb "github.com/devathh/xvibe/chat/api/chat/v1"
	"github.com/devathh/xvibe/chat/internal/application/services"
	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/cache/redis"
	chatredis "github.com/devathh/xvibe/chat/internal/infrastructure/cache/redis/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	grpcserver "github.com/devathh/xvibe/chat/internal/infrastructure/grpc/server"
	"github.com/devathh/xvibe/chat/internal/infrastructure/grpc/server/handlers"
	"github.com/devathh/xvibe/chat/internal/infrastructure/grpc/server/interceptors"
	"github.com/devathh/xvibe/chat/internal/infrastructure/grpc/xvibe"

	"github.com/devathh/xvibe/chat/internal/infrastructure/persistence/filem"
	"github.com/devathh/xvibe/chat/internal/infrastructure/persistence/postgres"
	chatpg "github.com/devathh/xvibe/chat/internal/infrastructure/persistence/postgres/chat"
	"github.com/devathh/xvibe/chat/internal/infrastructure/security/crypto"
	"github.com/devathh/xvibe/chat/internal/infrastructure/security/mtls"
	"github.com/devathh/xvibe/chat/internal/infrastructure/session/jwt"
	"github.com/devathh/xvibe/chat/pkg/log"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	if err := loadPublicKey(cfg); err != nil {
		return nil, nil, fmt.Errorf("faield to load public key: %w", err)
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

func provideServer(cfg *config.Config, log *slog.Logger, chatRepo chat.ChatRepository, chatCacheRepo chat.ChatCacheRepository) (*grpc.Server, error) {
	wrapperDEK := crypto.New(cfg)
	service := services.New(cfg, log, chatRepo, chatCacheRepo, wrapperDEK)
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

	sOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			packInterceptors.AuthInterceptor(),
		),
	}
	if cfg.Server.GRPC.TLS.Enable {
		tlsConfig, err := mtls.LoadMTLSConfig(cfg)
		if err != nil {
			return nil, err
		}

		sOpts = append(sOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcServer := grpc.NewServer(sOpts...)
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

func loadPublicKey(cfg *config.Config) error {
	var (
		client *grpc.ClientConn
		err    error
	)

	if cfg.Services.XvibeAuth.TLS.Enable {
		client, err = xvibe.ConnectMTLSAuth(cfg)
	} else {
		client, err = xvibe.ConnectInsecureAuth(cfg)
	}

	if err != nil {
		return err
	}
	defer xvibe.Close(client)

	authClient := authpb.NewAuthClient(client)

	publicKey, err := authClient.GetPublicKey(context.Background(), nil)
	if err != nil {
		return err
	}

	return filem.New(cfg).Save(publicKey.GetContent())
}
