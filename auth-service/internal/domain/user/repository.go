package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, like string) ([]*User, error)
	Update(ctx context.Context, updUser *User, mask []string) (*User, error)
}

type UserCache interface {
	Save(ctx context.Context, user *User) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Del(ctx context.Context, id uuid.UUID) error
}
