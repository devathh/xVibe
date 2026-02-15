package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
)

var (
	ErrInvalidKEK = errors.New("invalid size of kek")
)

type WrapperDEK struct {
	cfg *config.Config
}

func New(cfg *config.Config) *WrapperDEK {
	return &WrapperDEK{
		cfg: cfg,
	}
}

func (cd *WrapperDEK) WrapDEK(dek, kek []byte) ([]byte, error) {
	if len(kek) != 32 {
		return nil, ErrInvalidKEK
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := make([]byte, cd.cfg.Secrets.Cipher.NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, dek, nil)
	return append(nonce, ciphertext...), nil
}

func (cd *WrapperDEK) UnwrapDEK(dek, kek []byte) ([]byte, error) {
	if len(kek) != 32 {
		return nil, ErrInvalidKEK
	}
	if len(dek) < cd.cfg.Secrets.Cipher.NonceSize {
		return nil, errors.New("invalid size of dek")
	}

	nonce := dek[:cd.cfg.Secrets.Cipher.NonceSize]
	ciphertext := dek[cd.cfg.Secrets.Cipher.NonceSize:]

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	unwrappedDEK, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open gcm: %w", err)
	}

	return unwrappedDEK, nil
}
