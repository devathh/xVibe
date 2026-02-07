package userredis

import (
	"encoding/json"

	"github.com/devathh/xvibe/auth-service/internal/domain/user"
)

func toDomain(bytes []byte) (*user.User, error) {
	var model UserModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, err
	}

	return user.From(
		model.ID,
		user.Email(model.Email),
		"",
		model.Firstname,
		model.Lastname,
		user.Username(model.Username),
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func toBytes(domain *user.User) ([]byte, error) {
	bytes, err := json.Marshal(&UserModel{
		ID:        domain.ID(),
		Email:     domain.Email().Value(),
		Firstname: domain.Firstname(),
		Lastname:  domain.Lastname(),
		Username:  domain.Username().Value(),
		CreatedAt: domain.CreatedAt(),
		UpdatedAt: domain.UpdatedAt(),
	})
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
