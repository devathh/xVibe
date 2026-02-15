package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
)

type AESGCMEncryptor struct {
	cfg *config.Config
}

func NewAESGCMEncryptor(cfg *config.Config) *AESGCMEncryptor {
	return &AESGCMEncryptor{
		cfg: cfg,
	}
}

func (age *AESGCMEncryptor) Encode(plaintext, dek []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	cipherText := gcm.Seal(nil, nonce, plaintext, nil)
	return cipherText, nonce, nil
}

func (age *AESGCMEncryptor) Decode(cipherText, nonce, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}

	if gcm.NonceSize() != len(nonce) {
		return nil, fmt.Errorf("invalid len of nonce: %w", err)
	}

	encodedText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open gcm: %w", err)
	}

	return encodedText, nil
}
