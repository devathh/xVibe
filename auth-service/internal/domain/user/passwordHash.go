package user

import (
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHash string

func (ph PasswordHash) Value() string {
	return string(ph)
}

func (ph PasswordHash) Compare(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(ph), []byte(password)) == nil
}

func NewPasswordHash(password string) (PasswordHash, error) {
	if len([]rune(password)) < 6 {
		return "", consts.ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", consts.ErrInvalidPassword
	}

	return PasswordHash(hash), nil
}
