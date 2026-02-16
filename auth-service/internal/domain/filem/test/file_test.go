package test

import (
	"errors"
	"testing"

	"github.com/devathh/xvibe/auth-service/internal/domain/filem"
	"github.com/devathh/xvibe/auth-service/pkg/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile_Success(t *testing.T) {
	name := "name.file"
	content := []byte("ola minioun")

	file, err := filem.NewFile(name, content)

	require.NoError(t, err)
	require.NotNil(t, file)

	assert.Equal(t, file.Name(), name)
	assert.Equal(t, file.Content(), content)
}

func TestNewFile_Invalid(t *testing.T) {
	type input struct {
		name    string
		content []byte
	}

	testcases := []struct {
		name     string
		input    input
		expected error
	}{
		{
			name: "empty_name",
			input: input{
				name:    "",
				content: []byte("content"),
			},
			expected: consts.ErrInvalidFilename,
		},
		{
			name: "empty_content",
			input: input{
				name:    "name.file",
				content: []byte{},
			},
			expected: consts.ErrInvalidContent,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := filem.NewFile(tc.input.name, tc.input.content); !errors.Is(err, tc.expected) {
				t.Errorf("want %v, got %v", tc.expected, err)
			}
		})
	}
}
