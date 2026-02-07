package user

import (
	"strings"
	"time"

	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
)

type User struct {
	id           uuid.UUID
	email        Email
	passwordHash PasswordHash
	firstname    string
	lastname     string
	username     Username
	createdAt    time.Time
	updatedAt    time.Time
}

func New(
	email Email,
	passwordHash PasswordHash,
	firstname, lastname string,
	username Username,
) (*User, error) {
	firstname = strings.TrimSpace(firstname)
	if firstname == "" {
		return nil, consts.ErrInvalidFirstname
	}
	// lastname can be empty
	lastname = strings.TrimSpace(lastname)

	email = Email(strings.TrimSpace(email.Value()))
	if !email.IsValid() {
		return nil, consts.ErrInvalidEmail
	}

	username = Username(strings.TrimSpace(username.Value()))
	if !username.IsValid() {
		return nil, consts.ErrInvalidUsername
	}

	return &User{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		firstname:    firstname,
		lastname:     lastname,
		username:     username,
		createdAt:    time.Now().UTC(),
		updatedAt:    time.Now().UTC(),
	}, nil
}

// This function must be called only from persistence
func From(
	id uuid.UUID,
	email Email,
	passwordHash PasswordHash,
	firstname, lastname string,
	username Username,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		firstname:    firstname,
		lastname:     lastname,
		username:     username,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func ForUpdate(
	id uuid.UUID,
	firstname, lastname string,
	username Username,
	email Email,
) *User {
	return &User{
		id:        id,
		firstname: firstname,
		lastname:  lastname,
		username:  username,
		email:     email,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Firstname() string {
	return u.firstname
}

func (u *User) Lastname() string {
	return u.lastname
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}
