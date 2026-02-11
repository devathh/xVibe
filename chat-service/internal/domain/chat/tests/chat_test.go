package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/devathh/xvibe/chat/internal/domain/chat"
	"github.com/devathh/xvibe/chat/pkg/consts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSelfChat_Success(t *testing.T) {
	typ := chat.NewType(chat.TYPE_SELF)
	ownerID := uuid.New()

	now := time.Now().UTC()
	chat, err := chat.New(
		"",
		typ,
		ownerID,
	)

	require.NoError(t, err)
	require.NotNil(t, chat)

	assert.Equal(t, chat.OwnerID(), ownerID)
	assert.Equal(t, chat.Type().Value(), typ.Value())
	assert.Empty(t, chat.Title())
	assert.WithinDuration(t, chat.CreatedAt(), now, 100*time.Millisecond)
}

func TestNewGroupChat_Success(t *testing.T) {
	typ := chat.NewType(chat.TYPE_GROUP)
	ownerID := uuid.New()
	title := "group"

	now := time.Now().UTC()
	chat, err := chat.New(
		title,
		typ,
		ownerID,
	)

	require.NoError(t, err)
	require.NotNil(t, chat)

	assert.Equal(t, chat.OwnerID(), ownerID)
	assert.Equal(t, chat.Type().Value(), typ.Value())
	assert.Equal(t, chat.Title(), title)
	assert.WithinDuration(t, chat.CreatedAt(), now, 100*time.Millisecond)
}

func TestNewChat_Invalids(t *testing.T) {
	type input struct {
		title   string
		typ     chat.Type
		ownerID uuid.UUID
	}

	testCases := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "selfchat_title",
			input: input{
				title:   "not_empty",
				typ:     chat.NewType(chat.TYPE_SELF),
				ownerID: uuid.New(),
			},
			expected: consts.ErrTitleInSelf,
		},
		{
			name: "chat_invalid_type",
			input: input{
				title:   "",
				typ:     chat.NewType(chat.TYPE_UNKNOWN),
				ownerID: uuid.New(),
			},
			expected: consts.ErrInvalidType,
		},
		{
			name: "groupchat_empty_title",
			input: input{
				title:   "",
				typ:     chat.NewType(chat.TYPE_GROUP),
				ownerID: uuid.New(),
			},
			expected: consts.ErrInvalidTitle,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := chat.New(
				tc.input.title,
				tc.input.typ,
				tc.input.ownerID,
			); !errors.Is(err, tc.expected) {
				t.Errorf("want %v, got %v", tc.expected, err)
			}
		})
	}
}
