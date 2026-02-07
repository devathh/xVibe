package session

import (
	"net/mail"
	"time"

	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/google/uuid"
)

type Session struct {
	userID      uuid.UUID
	email       string
	fingerprint Fingerprint
	createdAt   time.Time
}

func Create(
	userID uuid.UUID,
	email string,
	fingerprint Fingerprint,
) (*Session, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, consts.ErrInvalidEmail
	}

	return &Session{
		userID:      userID,
		email:       email,
		fingerprint: fingerprint,
		createdAt:   time.Now().UTC(),
	}, nil
}

func From(
	userID uuid.UUID,
	email string,
	fingerprint Fingerprint,
	createdAt time.Time,
) *Session {
	return &Session{
		userID:      userID,
		email:       email,
		fingerprint: fingerprint,
		createdAt:   createdAt,
	}
}

func (s *Session) UserID() uuid.UUID {
	return s.userID
}

func (s *Session) Email() string {
	return s.email
}

func (s *Session) Fingerprint() Fingerprint {
	return s.fingerprint
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}
