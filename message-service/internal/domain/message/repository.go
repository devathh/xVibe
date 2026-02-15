package message

import (
	"context"

	"github.com/google/uuid"
)

type MessageRepository interface {
	GetWrappedDEK(context.Context, uuid.UUID) ([]byte, error)                                // chat's id
	Save(context.Context, *Message) (*Message, error)                                        // with nonce
	Delete(context.Context, uuid.UUID, uuid.UUID) error                                      // message's id, sender's id
	GetHistory(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uint32) ([]*Message, error) // chat's id, user's id, before id
	IsUserMember(context.Context, uuid.UUID, uuid.UUID) bool                                 // chat's id, user's id
}

type MessageCacheRepository interface {
	Publish(ctx context.Context, msg *Message) error
	Subscribe(ctx context.Context, chatID uuid.UUID, handler func(context.Context, *Message)) error
}
