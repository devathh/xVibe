package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/devathh/xvibe/chat/internal/domain/session"
	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat/pkg/consts"
	"github.com/golang-jwt/jwt/v5"
)

type JwtManager struct {
	cfg    *config.Config
	public *rsa.PublicKey
}

func New(cfg *config.Config) (*JwtManager, error) {
	public, err := loadPublic(cfg)
	if err != nil {
		return nil, err
	}

	return &JwtManager{
		cfg:    cfg,
		public: public,
	}, nil
}

func (j *JwtManager) Validate(tokenString string) (*session.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &session.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, consts.ErrInvalidToken
		}

		return j.public, nil
	})
	if err != nil {
		return nil, consts.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*session.Claims); ok {
		return claims, nil
	}

	return nil, consts.ErrInvalidToken
}

func loadPublic(cfg *config.Config) (*rsa.PublicKey, error) {
	bytes, err := os.ReadFile(cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	return key, nil
}
