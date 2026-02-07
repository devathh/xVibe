package sessionredis

import (
	"time"

	"github.com/google/uuid"
)

type SessionModel struct {
	UserID      uuid.UUID `yaml:"user_id"`
	Email       string    `yaml:"email"`
	Fingerprint string    `yaml:"fingerprint"`
	CreatedAt   time.Time `yaml:"created_at"`
}
