package filem

import (
	"fmt"
	"os"

	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
)

type FilemRepository struct {
	cfg *config.Config
}

func New(cfg *config.Config) *FilemRepository {
	return &FilemRepository{
		cfg: cfg,
	}
}

func (fm *FilemRepository) Save(bytes []byte) error {
	file, err := os.Create(fm.cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}

	return nil
}
