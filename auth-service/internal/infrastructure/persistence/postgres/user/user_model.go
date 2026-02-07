package userpg

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID           uuid.UUID `gorm:"primarykey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Firstname    string    `gorm:"not null"`
	Lastname     string
	Username     string `gorm:"uniqueIndex;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
