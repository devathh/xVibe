package tests

import (
	"testing"
	"time"

	"github.com/devathh/xvibe/auth-service/internal/domain/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession_Success(t *testing.T) {
	userID := uuid.New()
	email := "mail@example.com"
	fingerprint := session.NewFingerPrint(
		"0.0.0.0",
		"Mozilla",
	)

	now := time.Now().UTC()

	session, err := session.Create(
		userID,
		email,
		session.Fingerprint(fingerprint),
	)

	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, session.UserID(), userID)
	assert.Equal(t, session.Email(), email)
	assert.Equal(t, session.Fingerprint(), fingerprint)
	assert.True(t, session.Fingerprint().Compare("0.0.0.0", "Mozilla"))
	assert.WithinDuration(t, session.CreatedAt(), now, 100*time.Millisecond)
}
