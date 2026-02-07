package app

import (
	"fmt"
	"log/slog"
	"os"

	authpb "github.com/devathh/xvibe/auth-service/api/auth/v1"
	"github.com/devathh/xvibe/auth-service/internal/application/services"
	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/devathh/xvibe/auth-service/internal/domain/user"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/cache/redis"
	sessionredis "github.com/devathh/xvibe/auth-service/internal/infrastructure/cache/redis/session"
	userredis "github.com/devathh/xvibe/auth-service/internal/infrastructure/cache/redis/user"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	grpcserver "github.com/devathh/xvibe/auth-service/internal/infrastructure/grpc"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/grpc/handlers"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/grpc/interceptors"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/persistence/postgres"
	userpg "github.com/devathh/xvibe/auth-service/internal/infrastructure/persistence/postgres/user"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/security/mtls"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/session/jwt"
	"github.com/devathh/xvibe/auth-service/pkg/log"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

// The main structure of the entire server,
// it controls the start/stop
type App struct {
	log    *slog.Logger
	server *grpcserver.Server
}

// Start the server
func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

// Graceful stop the server
func (a *App) GracefulStop() {
	a.log.Info("graceful stop")
	a.server.GraceulStop()
}

// Assembling all components for server initialization,
// connecting dependencies, etc
func New() (*App, func(), error) {
	// .env file
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	// config
	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// the custom logger
	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup log's handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	db, userRepo, err := providePersistence(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide persistence: %w", err)
	}

	redisClient, sessionRepo, userCache, err := provideCache(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide cache: %w", err)
	}

	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load jwt manager: %w", err)
	}

	grpcServer, err := provideServer(
		cfg,
		log,
		userRepo,
		userCache,
		sessionRepo,
		jwtMngr,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide grpc server: %w", err)
	}

	return &App{
			log:    log,
			server: grpcserver.New(cfg, grpcServer),
		}, func() {
			if err := postgres.Close(db); err != nil {
				log.Error("failed to close connection with postgres", slog.String("error", err.Error()))
			} else {
				log.Info("connection with postgres was closed")
			}

			if err := redis.Close(redisClient); err != nil {
				log.Error("failed to close connection with redis", slog.String("error", err.Error()))
			} else {
				log.Info("connection with redis was closed")
			}
		}, nil
}

// Initiating a grpc server
// and creating a service for it
func provideServer(cfg *config.Config, log *slog.Logger, userRepo user.UserRepository, userCache user.UserCache, sessionRepo session.SessionRepository, jwtMngr session.JwtManager) (*grpc.Server, error) {
	service := services.New(
		cfg,
		log,
		userRepo,
		userCache,
		sessionRepo,
		jwtMngr,
	)

	api := handlers.New(service)

	interceptorPack := interceptors.New(jwtMngr, map[string]bool{
		authpb.Auth_Update_FullMethodName:    true,
		authpb.Auth_GetSelf_FullMethodName:   true,
		authpb.Auth_LogoutAll_FullMethodName: true,
	})
	sOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptorPack.BaseInterceptor(),
			interceptorPack.AuthInterceptor(),
		),
	}

	if cfg.Server.GRPC.TLS.Enable {
		tlsConfig, err := mtls.LoadMTLSConfig(cfg)
		if err != nil {
			return nil, err
		}

		sOpts = append(sOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	server := grpc.NewServer(sOpts...)
	authpb.RegisterAuthServer(server, api)
	return server, nil
}

// Connecting to redis,
// as well as initializing repositories
func provideCache(cfg *config.Config) (*redissdk.Client, *sessionredis.SessionRepository, *userredis.UserCache, error) {
	client, err := redis.Connect(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	return client, sessionredis.New(cfg, client), userredis.New(cfg, client), nil
}

// Connecting to postgres,
// as well as initializing repositories
func providePersistence(cfg *config.Config) (*gorm.DB, *userpg.UserRepository, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	if err := postgres.Migrate(db); err != nil {
		return nil, nil, err
	}

	userRepo := userpg.New(db)

	return db, userRepo, nil
}
