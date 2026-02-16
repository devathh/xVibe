package userpg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devathh/xvibe/auth-service/internal/domain/user"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ur *UserRepository) Save(ctx context.Context, user *user.User) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	model := toModel(user)
	if err := ur.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		if ur.isUniqueError(err) {
			return nil, consts.ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return toDomain(model), nil
}

func (ur *UserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model UserModel
	if err := ur.db.WithContext(ctx).First(&model, "email = ?", email.Value()).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return toDomain(&model), nil
}

func (ur *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model UserModel
	if err := ur.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return toDomain(&model), nil
}

func (ur *UserRepository) GetByUsername(ctx context.Context, like string) ([]*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	like = strings.TrimSpace(like)
	if like == "" {
		return nil, consts.ErrInvalidUsername
	}

	var models []UserModel
	if err := ur.db.WithContext(ctx).Limit(100).Where("username ILIKE ?", "%"+like+"%").Find(&models).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to get users like username: %w", err)
	}

	users := make([]*user.User, len(models))
	for idx, user := range models {
		users[idx] = toDomain(&user)
	}

	return users, nil
}

func (ur *UserRepository) Update(ctx context.Context, updUser *user.User, mask []string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	updateMap := map[string]any{
		"firstname": updUser.Firstname(),
		"lastname":  updUser.Lastname(),
		"username":  updUser.Username().Value(),
		"email":     updUser.Email().Value(),
	}

	var model UserModel
	if err := ur.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(&model, updUser.ID()).Error; err != nil {
			return err
		}

		updates := make(map[string]any, 5)
		for _, field := range mask {
			updates[field] = updateMap[field]
		}

		if err := tx.Model(&model).Where("id = ?", updUser.ID()).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrUserDoesntExist
		}

		if ur.isUniqueError(err) {
			return nil, consts.ErrUserAlreadyTaken
		}

		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return toDomain(&model), nil
}

func (ur *UserRepository) isUniqueError(err error) bool {
	return errors.Is(postgres.Dialector{}.Translate(err), gorm.ErrDuplicatedKey)
}
