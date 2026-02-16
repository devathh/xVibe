package filem

import (
	"context"
	"os"

	"github.com/devathh/xvibe/auth-service/internal/domain/filem"
	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
)

type FilemRepository struct {
	cfg *config.Config
}

func New(cfg *config.Config) *FilemRepository {
	return &FilemRepository{
		cfg: cfg,
	}
}

func (fm *FilemRepository) GetPublicKey(ctx context.Context) (*filem.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	bytes, err := os.ReadFile(fm.cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return filem.NewFile(
		fm.cfg.Secrets.JWT.PublicKeyPath,
		bytes,
	)
}
