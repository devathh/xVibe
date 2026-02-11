package redis

import (
	"context"
	"fmt"

	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Secrets.Redis.Host,
			cfg.Secrets.Redis.Port,
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
		return fmt.Errorf("failed to close redis client: %w", err)
	}

	return nil
}
