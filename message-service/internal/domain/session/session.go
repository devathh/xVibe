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

type ctxKey string

const (
	KeyUserID    ctxKey = "user-id"
	KeyUserEmail ctxKey = "user-email"
)
