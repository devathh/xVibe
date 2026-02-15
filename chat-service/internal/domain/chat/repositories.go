package chat

import (
	"context"

	"github.com/devathh/xvibe/chat/internal/domain/member"
	"github.com/google/uuid"
)

type ChatCacheRepository interface {
	Save(
		ctx context.Context,
		userID uuid.UUID,
		chats []*ChatModel,
	) error

	Get(ctx context.Context, userID uuid.UUID) ([]*ChatModel, error)
	Del(ctx context.Context, memberIds []*member.Member) error
}

type ChatRepository interface {
	Save(
		ctx context.Context,
		chat *ChatModel,
		memberIds []uuid.UUID,
		wrappedDEK []byte,
	) (*ChatModel, error)

	Delete(
		ctx context.Context,
		chatID uuid.UUID,
		userID uuid.UUID,
	) error

	Update(
		ctx context.Context,
		chatID uuid.UUID,
		userID uuid.UUID,
		title string,
	) (*ChatModel, []*member.Member, error)

	AddMembers(
		ctx context.Context,
		memberIds []uuid.UUID,
		chatID uuid.UUID,
	) (*ChatModel, []*member.Member, error)

	DeleteMembers(
		ctx context.Context,
		chatID uuid.UUID,
		memberIds []uuid.UUID,
	) error

	GetSelfChats(
		ctx context.Context,
		userID uuid.UUID,
	) ([]*ChatModel, error)

	GetChat(
		ctx context.Context,
		chatID uuid.UUID,
	) (*ChatModel, []*member.Member, error)
}
