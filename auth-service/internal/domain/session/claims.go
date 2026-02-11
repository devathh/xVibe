package session

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type JwtManager interface {
	GenerateAccess(userID uuid.UUID, email string) (string, error)
	GenerateRefresh() (string, error)
	Validate(tokenString string) (*Claims, error)
}

type ctxKey string

const (
	ClientIPKey  ctxKey = "x-client-ip"
	UserAgentKey ctxKey = "x-client-user-agent"
	UserEmailKey ctxKey = "user-email"
	UserIDKey    ctxKey = "user-id"
)
