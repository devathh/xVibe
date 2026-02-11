package chatredis

import (
	"time"

	"github.com/google/uuid"
)

type ChatModel struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Title     string    `json:"title"`
	TypeID    int       `json:"type_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatModels struct {
	Chats []ChatModel `json:"chat_models"`
}
