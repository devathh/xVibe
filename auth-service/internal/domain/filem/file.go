package filem

import (
	"strings"

	"github.com/devathh/xvibe/auth-service/pkg/consts"
)

type File struct {
	name    string
	content []byte
}

func NewFile(name string, content []byte) (*File, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, consts.ErrInvalidFilename
	}

	if len(content) < 1 {
		return nil, consts.ErrInvalidContent
	}

	return &File{
		name:    name,
		content: content,
	}, nil
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Content() []byte {
	return f.content
}
