package sessionredis

import (
	"encoding/json"
	"fmt"

	"github.com/devathh/xvibe/auth-service/internal/domain/session"
)

func toDomain(bytes []byte) (*session.Session, error) {
	var model SessionModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}

	return session.From(
		model.UserID,
		model.Email,
		session.Fingerprint(model.Fingerprint),
		model.CreatedAt,
	), nil
}

func toModel(domain *session.Session) ([]byte, error) {
	bytes, err := json.Marshal(SessionModel{
		UserID:      domain.UserID(),
		Email:       domain.Email(),
		Fingerprint: domain.Fingerprint().Value(),
		CreatedAt:   domain.CreatedAt(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to bytes: %w", err)
	}

	return bytes, nil
}
