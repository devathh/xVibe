package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtManager struct {
	cfg     *config.Config
	public  *rsa.PublicKey
	private *rsa.PrivateKey
}

func New(cfg *config.Config) (*JwtManager, error) {
	private, err := loadPrivate(cfg)
	if err != nil {
		return nil, err
	}

	public, err := loadPublic(cfg)
	if err != nil {
		return nil, err
	}

	return &JwtManager{
		cfg:     cfg,
		private: private,
		public:  public,
	}, nil
}

func (j *JwtManager) GenerateAccess(userID uuid.UUID, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, session.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "xvibe",
			Subject:   "xvibe-client",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.cfg.Service.Session.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString(j.private)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (j *JwtManager) GenerateRefresh() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to read empty bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
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

func loadPrivate(cfg *config.Config) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(cfg.Secrets.JWT.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
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
