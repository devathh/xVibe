package userpg

import "github.com/devathh/xvibe/auth-service/internal/domain/user"

func toDomain(model *UserModel) *user.User {
	return user.From(
		model.ID,
		user.Email(model.Email),
		user.PasswordHash(model.PasswordHash),
		model.Firstname,
		model.Lastname,
		user.Username(model.Username),
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func toModel(domain *user.User) *UserModel {
	return &UserModel{
		ID:           domain.ID(),
		Email:        domain.Email().Value(),
		PasswordHash: domain.PasswordHash().Value(),
		Firstname:    domain.Firstname(),
		Lastname:     domain.Lastname(),
		Username:     domain.Username().Value(),
		CreatedAt:    domain.CreatedAt(),
		UpdatedAt:    domain.UpdatedAt(),
	}
}
