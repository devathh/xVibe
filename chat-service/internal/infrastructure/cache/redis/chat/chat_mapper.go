package chatredis

import (
	"encoding/json"

	"github.com/devathh/xvibe/chat/internal/domain/chat"
)

func toDomain(bytes []byte) ([]*chat.ChatModel, error) {
	var models ChatModels
	if err := json.Unmarshal(bytes, &models); err != nil {
		return nil, err
	}

	domainModels := make([]*chat.ChatModel, len(models.Chats))
	for idx, model := range models.Chats {
		typ, _ := chat.NewTypeRaw(model.TypeID)
		domainModels[idx] = chat.From(
			model.ID,
			model.OwnerID,
			model.Title,
			typ,
			model.CreatedAt,
		)
	}

	return domainModels, nil
}

func toModel(chats []*chat.ChatModel) ([]byte, error) {
	models := ChatModels{
		Chats: make([]ChatModel, len(chats)),
	}

	for idx, model := range chats {
		models.Chats[idx] = ChatModel{
			ID:        model.ID(),
			OwnerID:   model.OwnerID(),
			Title:     model.Title(),
			TypeID:    int(model.Type().Value()),
			CreatedAt: model.CreatedAt(),
		}
	}

	return json.Marshal(&models)
}
