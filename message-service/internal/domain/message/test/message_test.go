package test

import (
	"errors"
	"testing"
	"time"

	"github.com/devathh/xvibe/message-service/internal/domain/message"
	"github.com/devathh/xvibe/message-service/pkg/consts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage_Success(t *testing.T) {
	chatID := uuid.New()
	authorID := uuid.New()
	encryptedBody := []byte("very secret message")

	now := time.Now().UTC()
	msg, err := message.New(chatID, authorID, encryptedBody, nil)

	require.NoError(t, err)
	require.NotNil(t, msg)

	assert.Equal(t, msg.ChatID(), chatID)
	assert.Equal(t, msg.AuthorID(), authorID)
	assert.Equal(t, msg.EncryptedBody(), encryptedBody)
	assert.WithinDuration(t, msg.SentAt(), now, 100*time.Millisecond)
}

func TestNewMessage_InvalidParams(t *testing.T) {
	type input struct {
		chatID, authorID uuid.UUID
		encryptedBody    []byte
	}

	testCases := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "empty_chat_id",
			input: input{
				chatID:        uuid.Nil,
				authorID:      uuid.New(),
				encryptedBody: []byte("well well well"),
			},
			expected: consts.ErrInvalidChatID,
		},
		{
			name: "empty_author_id",
			input: input{
				chatID:        uuid.New(),
				authorID:      uuid.Nil,
				encryptedBody: []byte("well well well"),
			},
			expected: consts.ErrInvalidAuthorID,
		},
		{
			name: "empty_encrypted_body",
			input: input{
				chatID:        uuid.New(),
				authorID:      uuid.New(),
				encryptedBody: []byte{},
			},
			expected: consts.ErrInvalidEncryptedBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := message.New(
				tc.input.chatID,
				tc.input.authorID,
				tc.input.encryptedBody,
				nil,
			); !errors.Is(err, tc.expected) {
				t.Errorf("want %v, got %v", tc.expected, err)
			}
		})
	}
}
