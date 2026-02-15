package chatpg

import (
	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/internal/domain/member"
	"github.com/google/uuid"
)

func toModelChatModel(domain *chat.ChatModel, wrappedDEK []byte) *ChatModel {
	return &ChatModel{
		ID:         domain.ID(),
		OwnerID:    domain.OwnerID(),
		Title:      domain.Title(),
		TypeID:     int(domain.Type().Value()),
		WrappedDEK: wrappedDEK,
		CreatedAt:  domain.CreatedAt(),
	}
}

func toDomainChatModel(model *ChatModel) *chat.ChatModel {
	typ, _ := chat.NewTypeRaw(model.TypeID)
	return chat.From(
		model.ID,
		model.OwnerID,
		model.Title,
		typ,
		model.CreatedAt,
	)
}

func toModelMembers(memberIds []uuid.UUID, chatID uuid.UUID) []ChatMembers {
	members := make([]ChatMembers, len(memberIds))
	for idx, member := range memberIds {
		members[idx] = ChatMembers{
			ID:     uuid.New(),
			UserID: member,
			ChatID: chatID,
		}
	}

	return members
}

func toDomainMembers(members []ChatMembers) []uuid.UUID {
	memberIds := make([]uuid.UUID, len(members))
	for idx, member := range members {
		memberIds[idx] = member.UserID
	}

	return memberIds
}

func toDomainUserMembers(memberModels []UserMember, ownerID uuid.UUID) []*member.Member {
	members := make([]*member.Member, len(memberModels))
	for idx, model := range memberModels {
		members[idx] = member.New(
			model.ID,
			model.Firstname,
			model.Lastname,
			ownerID == model.ID,
		)
	}

	return members
}
