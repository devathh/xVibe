package tests

import (
	"errors"
	"testing"

	"github.com/devathh/xvibe/auth-service/internal/domain/user"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Success(t *testing.T) {
	user, err := user.New(
		user.Email("mail@example.com"),
		"",
		"Firstname",
		"Lastname",
		user.Username("username"),
	)

	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, user.Email().Value(), "mail@example.com")
	assert.Equal(t, user.Firstname(), "Firstname")
	assert.Equal(t, user.Lastname(), "Lastname")
	assert.Equal(t, user.Username().Value(), "username")
}

func TestNewUser_InvalidEmail(t *testing.T) {
	testCases := []struct {
		name       string
		inputEmail string
		expected   error
	}{
		{"empty", "", consts.ErrInvalidEmail},
		{"spaces", "   ", consts.ErrInvalidEmail},
		{"invalid_domain_1", "mail", consts.ErrInvalidEmail},
		{"invalid_domain_2", "mail@", consts.ErrInvalidEmail},
		{"invalid_mail_1", "@example.com", consts.ErrInvalidEmail},
		{"invalid_mail_2", ".@example", consts.ErrInvalidEmail},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := user.New(user.Email(tc.inputEmail), "", "firstname", "", "username"); !errors.Is(err, tc.expected) {
				t.Errorf("want %v got %v", tc.expected, err)
			}
		})
	}
}

func TestNewUser_InvalidUsername(t *testing.T) {
	testCases := []struct {
		name          string
		inputUsername string
		expected      error
	}{
		{"empty", "", consts.ErrInvalidUsername},
		{"spaces", "   ", consts.ErrInvalidUsername},
		{"too_little", "w", consts.ErrInvalidUsername},
		{"too_long", "too_long_username_for_validation", consts.ErrInvalidUsername},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := user.New("mail@example.com", "", "firstname", "", user.Username(tc.inputUsername)); !errors.Is(err, tc.expected) {
				t.Errorf("want %v got %v", tc.expected, err)
			}
		})
	}
}
