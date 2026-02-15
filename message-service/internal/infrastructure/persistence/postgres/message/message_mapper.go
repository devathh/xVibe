package messagepg

import "github.com/devathh/xvibe/message-service/internal/domain/message"

func toModel(domain *message.Message) *MessageModel {
	return &MessageModel{
		ID:            domain.ID(),
		ChatID:        domain.ChatID(),
		AuthorID:      domain.AuthorID(),
		EncryptedBody: domain.EncryptedBody(),
		Nonce:         domain.Nonce(),
		SentAt:        domain.SentAt(),
	}
}

func toDomain(model *MessageModel) *message.Message {
	return message.From(
		model.ID,
		model.ChatID,
		model.AuthorID,
		model.EncryptedBody,
		model.Nonce,
		model.SentAt,
	)
}
