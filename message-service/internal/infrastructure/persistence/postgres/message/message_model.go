package messagepg

import (
	"time"

	"github.com/google/uuid"
)

type MessageModel struct {
	ID            uuid.UUID `gorm:"primarykey"`
	ChatID        uuid.UUID `gorm:"not null;index"`
	AuthorID      uuid.UUID `gorm:"not null;index"`
	EncryptedBody []byte    `gorm:"type:bytea;not null"`
	Nonce         []byte    `gorm:"type:bytea;not null"`
	SentAt        time.Time `gorm:"not null"`
}

func (mm *MessageModel) TableName() string {
	return "message_models"
}
