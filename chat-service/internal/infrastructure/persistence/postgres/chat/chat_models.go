package chatpg

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatModel struct {
	ID        uuid.UUID `gorm:"id"`
	OwnerID   uuid.UUID `gorm:"not null"`
	Title     string
	TypeID    int
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type ChatMembers struct {
	ID     uuid.UUID `gorm:"primarykey"`
	ChatID uuid.UUID `gorm:"index:idx_members_chat_user,unique;type:uuid"`
	UserID uuid.UUID `gorm:"index:idx_members_chat_user,unique;type:uuid"`
}

type UserMember struct {
	ID        uuid.UUID `gorm:"column:id"`
	Firstname string    `gorm:"column:firstname"`
	Lastname  string    `gorm:"column:lastname"`
}
