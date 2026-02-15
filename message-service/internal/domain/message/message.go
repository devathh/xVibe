package message

import (
	"fmt"
	"time"

	"github.com/devathh/xvibe/message-service/pkg/consts"
	"github.com/google/uuid"
)

type Message struct {
	id            uuid.UUID // uuid v7
	chatID        uuid.UUID
	authorID      uuid.UUID
	encryptedBody []byte
	nonce         []byte
	sentAt        time.Time
}

func New(
	chatID, authorID uuid.UUID,
	encryptedBody, nonce []byte,
) (*Message, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	if chatID == uuid.Nil {
		return nil, consts.ErrInvalidChatID
	}

	if authorID == uuid.Nil {
		return nil, consts.ErrInvalidAuthorID
	}

	if len(encryptedBody) == 0 {
		return nil, consts.ErrInvalidEncryptedBody
	}

	return &Message{
		id:            id,
		chatID:        chatID,
		authorID:      authorID,
		encryptedBody: encryptedBody,
		nonce:         nonce,
		sentAt:        time.Now().UTC(),
	}, nil
}

func From(
	id uuid.UUID,
	chatID, authorID uuid.UUID,
	encryptedBody, nonce []byte,
	sentAt time.Time,
) *Message {
	return &Message{
		id:            id,
		chatID:        chatID,
		authorID:      authorID,
		encryptedBody: encryptedBody,
		nonce:         nonce,
		sentAt:        sentAt,
	}
}

func (m *Message) ID() uuid.UUID {
	return m.id
}

func (m *Message) ChatID() uuid.UUID {
	return m.chatID
}

func (m *Message) AuthorID() uuid.UUID {
	return m.authorID
}

func (m *Message) EncryptedBody() []byte {
	return m.encryptedBody
}

func (m *Message) Nonce() []byte {
	return m.nonce
}

func (m *Message) SentAt() time.Time {
	return m.sentAt
}
