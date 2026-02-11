package chat

import (
	"strings"
	"time"

	"github.com/devathh/xvibe/chat/pkg/consts"
	"github.com/google/uuid"
)

type ChatModel struct {
	id        uuid.UUID
	ownerID   uuid.UUID
	title     string
	typ       Type
	createdAt time.Time
}

func New(title string, typ Type, ownerID uuid.UUID) (*ChatModel, error) {
	if typ.Value() == TYPE_UNKNOWN {
		return nil, consts.ErrInvalidType
	}

	title = strings.TrimSpace(title)
	if typ.Value() == TYPE_SELF && title != "" {
		return nil, consts.ErrTitleInSelf
	}

	if typ.Value() == TYPE_GROUP && title == "" {
		return nil, consts.ErrInvalidTitle
	}

	return &ChatModel{
		id:        uuid.New(),
		ownerID:   ownerID,
		title:     title,
		typ:       typ,
		createdAt: time.Now().UTC(),
	}, nil
}

func From(
	id, ownerID uuid.UUID,
	title string,
	typ Type,
	createdAt time.Time,
) *ChatModel {
	return &ChatModel{
		id:        id,
		ownerID:   ownerID,
		title:     title,
		typ:       typ,
		createdAt: createdAt,
	}
}

func (cm *ChatModel) ID() uuid.UUID {
	return cm.id
}

func (cm *ChatModel) OwnerID() uuid.UUID {
	return cm.ownerID
}

func (cm *ChatModel) Title() string {
	return cm.title
}

func (cm *ChatModel) Type() Type {
	return cm.typ
}

func (cm *ChatModel) CreatedAt() time.Time {
	return cm.createdAt
}
