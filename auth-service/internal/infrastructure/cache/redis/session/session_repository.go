package sessionredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	cfg    *config.Config
	client *redis.Client
}

func New(cfg *config.Config, client *redis.Client) *SessionRepository {
	return &SessionRepository{
		cfg:    cfg,
		client: client,
	}
}

func (sr *SessionRepository) Save(ctx context.Context, refresh string, session *session.Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	modelBytes, err := toModel(session)
	if err != nil {
		return err
	}

	pipe := sr.client.TxPipeline()

	key := sr.getRefreshKey(refresh)
	pipe.Set(ctx, key, modelBytes, sr.cfg.Service.Session.RefreshTTL)

	key = sr.getSessionsKey(session.UserID())
	pipe.SAdd(ctx, key, refresh)
	pipe.Expire(ctx, key, sr.cfg.Service.Session.RefreshTTL*2)

	if _, err := pipe.Exec(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to exec redis pipe: %w", err)
	}

	return nil
}

func (sr *SessionRepository) Get(ctx context.Context, refresh string) (*session.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := sr.getRefreshKey(refresh)
	bytes, err := sr.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, redis.Nil) {
			return nil, consts.ErrSessionDoesntExist
		}

		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return toDomain(bytes)
}

func (sr *SessionRepository) Del(ctx context.Context, refresh string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := sr.getRefreshKey(refresh)
	if err := sr.client.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, redis.Nil) {
			return consts.ErrSessionDoesntExist
		}

		return fmt.Errorf("failed to del session: %w", err)
	}

	return nil
}

func (sr *SessionRepository) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	key := sr.getSessionsKey(userID)
	refreshKeys, err := sr.client.SMembers(ctx, key).Result()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		if errors.Is(err, redis.Nil) {
			return consts.ErrSessionDoesntExist
		}

		return fmt.Errorf("failed to get all session members: %w", err)
	}

	oldKeys := make([]string, len(refreshKeys))
	for idx, oldKey := range refreshKeys {
		oldKeys[idx] = sr.getRefreshKey(oldKey)
	}

	pipe := sr.client.TxPipeline()
	if len(oldKeys) > 0 {
		pipe.Del(ctx, oldKeys...)
	}

	pipe.Del(ctx, key)

	if _, err := pipe.Exec(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return err
		}

		return fmt.Errorf("failed to exec pipe to del: %w", err)
	}

	return nil
}

func (sr *SessionRepository) getSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("sessions:%s", userID.String())
}

func (sr *SessionRepository) getRefreshKey(refresh string) string {
	return fmt.Sprintf("rf:session:%s", refresh)
}
