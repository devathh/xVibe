package messageredis

import (
	"time"

	"github.com/google/uuid"
)

type MessageModel struct {
	ID            uuid.UUID `json:"id"`
	ChatID        uuid.UUID `json:"chat_id"`
	AuthorID      uuid.UUID `json:"author_id"`
	EncryptedBody []byte    `json:"encrypted_body"`
	Nonce         []byte    `json:"nonce"`
	SentAt        time.Time `json:"sent_at"`
}
