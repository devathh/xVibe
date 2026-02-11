package chat

import "github.com/devathh/xvibe/chat/pkg/consts"

type typ int

const (
	TYPE_UNKNOWN typ = iota
	TYPE_SELF
	TYPE_GROUP
)

type Type struct {
	typ typ
}

func (t Type) Value() typ {
	return t.typ
}

func (t Type) String() string {
	switch t.typ {
	case TYPE_SELF:
		return "self"
	case TYPE_GROUP:
		return "group"
	}

	return "unknown"
}

func NewTypeRaw(raw int) (Type, error) {
	typ := typ(raw)
	if typ <= TYPE_UNKNOWN || typ > TYPE_GROUP {
		return Type{}, consts.ErrInvalidType
	}

	return Type{
		typ: typ,
	}, nil
}

func NewType(typ typ) Type {
	return Type{
		typ: typ,
	}
}
