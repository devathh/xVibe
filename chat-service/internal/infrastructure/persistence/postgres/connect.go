package postgres

import (
	"fmt"

	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	chatpg "github.com/devathh/xvibe/chat/internal/infrastructure/persistence/postgres/chat"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d sslmode=%s user=%s password=%s dbname=%s",
		cfg.Secrets.Postgres.Host,
		cfg.Secrets.Postgres.Port,
		cfg.Secrets.Postgres.SSLMode,
		cfg.Secrets.Postgres.Auth.User,
		cfg.Secrets.Postgres.Auth.Password,
		cfg.Secrets.Postgres.Auth.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	sqlDB.SetConnMaxIdleTime(cfg.Secrets.Postgres.Conn.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.Secrets.Postgres.Conn.MaxLifetime)
	sqlDB.SetMaxIdleConns(cfg.Secrets.Postgres.Conn.MaxIdles)
	sqlDB.SetMaxOpenConns(cfg.Secrets.Postgres.Conn.MaxOpens)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&chatpg.ChatModel{}, &chatpg.ChatMembers{}); err != nil {
		return fmt.Errorf("failed to migrate postgres: %w", err)
	}

	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}
