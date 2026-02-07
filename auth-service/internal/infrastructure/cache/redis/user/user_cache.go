package userredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/devathh/xvibe/auth-service/internal/domain/user"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	cfg    *config.Config
	client *redis.Client
}

func New(cfg *config.Config, client *redis.Client) *UserCache {
	return &UserCache{
		cfg:    cfg,
		client: client,
	}
}

func (uc *UserCache) Save(ctx context.Context, user *user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	bytes, err := toBytes(user)
	if err != nil {
		return err
	}

	key := uc.getKey(user.ID())
	if err := uc.client.Set(ctx, key, bytes, uc.cfg.Service.Cache.UserTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (uc *UserCache) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := uc.getKey(id)
	bytes, err := uc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, err
	}

	return toDomain(bytes)
}

func (uc *UserCache) getKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}
