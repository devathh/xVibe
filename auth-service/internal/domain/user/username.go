package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Username string

func (u Username) IsValid() bool {
	runes := []rune(u)
	return len(runes) < 15 && len(runes) > 2
}

func (u Username) Value() string {
	return string(u)
}

func NewUsername() (Username, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate username: %w", err)
	}

	return Username(hex.EncodeToString(bytes)), nil
}
