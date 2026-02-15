package app

import (
	"fmt"
	"log/slog"
	"os"

	messagepb "github.com/devathh/xvibe/message-service/api/message/v1"
	"github.com/devathh/xvibe/message-service/internal/application/services"
	"github.com/devathh/xvibe/message-service/internal/domain/message"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/cache/redis"
	messageredis "github.com/devathh/xvibe/message-service/internal/infrastructure/cache/redis/message"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/crypto"
	grpcserver "github.com/devathh/xvibe/message-service/internal/infrastructure/grpc"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/grpc/handlers"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/grpc/interceptors"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/persistence/postgres"
	messagepg "github.com/devathh/xvibe/message-service/internal/infrastructure/persistence/postgres/message"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/security/mtls"
	"github.com/devathh/xvibe/message-service/internal/infrastructure/session/jwt"
	"github.com/devathh/xvibe/message-service/pkg/log"
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

func (a *App) GracefulShutdown() {
	a.log.Info("graceful stop")
	a.server.GracefulShutdown()
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

	log.Info("config was loaded", slog.Any("config", cfg.Server), slog.Any("app", cfg.App))

	db, msgRepo, err := providePostgres(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide postgres: %w", err)
	}

	rClient, msgCache, err := provideRedis(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide redis: %w", err)
	}

	grpcServer, err := provideServer(
		cfg,
		log,
		msgRepo,
		msgCache,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide grpc server: %w", err)
	}

	return &App{
			log:    log,
			server: grpcserver.New(cfg, grpcServer),
		}, func() {
			if err := postgres.Close(db); err != nil {
				log.Error("failed to close postgres", slog.String("error", err.Error()))
			} else {
				log.Info("postgres was closed")
			}

			if err := redis.Close(rClient); err != nil {
				log.Error("failed to close redis", slog.String("error", err.Error()))
			} else {
				log.Info("redis was closed")
			}
		}, nil
}

func provideServer(cfg *config.Config, log *slog.Logger, msgRepo message.MessageRepository, msgCache message.MessageCacheRepository) (*grpc.Server, error) {
	wrapperDEK := crypto.New(cfg)
	aesgcmEnc := crypto.NewAESGCMEncryptor(cfg)
	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, err
	}

	service := services.New(
		cfg,
		log,
		msgRepo,
		msgCache,
		wrapperDEK,
		aesgcmEnc,
	)

	api := handlers.New(service)
	sOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.NewAuthInterceptor(jwtMngr, map[string]bool{}),
		),
		grpc.ChainStreamInterceptor(
			interceptors.NewAuthStreamInterceptor(jwtMngr, map[string]bool{}),
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
	messagepb.RegisterMessageServer(grpcServer, api)
	return grpcServer, nil
}

func providePostgres(cfg *config.Config) (*gorm.DB, *messagepg.MessageRepository, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	if err := postgres.Migrate(db); err != nil {
		return nil, nil, err
	}

	return db, messagepg.New(db), nil
}

func provideRedis(cfg *config.Config, log *slog.Logger) (*redissdk.Client, *messageredis.MessagePubsub, error) {
	rClient, err := redis.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	return rClient, messageredis.New(log, rClient), nil
}
