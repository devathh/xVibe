package redis

import (
	"context"
	"fmt"
	"net"

	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(
			cfg.Secrets.Redis.Host,
			fmt.Sprint(cfg.Secrets.Redis.Port),
		),
		Username: cfg.Secrets.Redis.Auth.Username,
		Password: cfg.Secrets.Redis.Auth.Password,
		DB:       cfg.Secrets.Redis.Auth.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func Close(client *redis.Client) error {
	if err := client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}

	return nil
}
