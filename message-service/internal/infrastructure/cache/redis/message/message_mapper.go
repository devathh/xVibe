package messageredis

import (
	"encoding/json"
	"fmt"

	"github.com/devathh/xvibe/message-service/internal/domain/message"
)

func toDomain(bytes []byte) (*message.Message, error) {
	var model MessageModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}

	return message.From(
		model.ID,
		model.ChatID,
		model.AuthorID,
		model.EncryptedBody,
		model.Nonce,
		model.SentAt,
	), nil
}

func toModel(domain *message.Message) ([]byte, error) {
	bytes, err := json.Marshal(MessageModel{
		ID:            domain.ID(),
		ChatID:        domain.ChatID(),
		AuthorID:      domain.AuthorID(),
		EncryptedBody: domain.EncryptedBody(),
		Nonce:         domain.Nonce(),
		SentAt:        domain.SentAt(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model: %w", err)
	}

	return bytes, nil
}
