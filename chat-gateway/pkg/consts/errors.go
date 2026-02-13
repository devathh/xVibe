package consts

import "errors"

var (
	// General's
	ErrInvalidChatType = errors.New("invalid type of chat")
	ErrInvalidToken    = errors.New("invalid token")
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrEmptyMembers    = errors.New("members must be contains at least one member")

	// Service's
	ErrBadGateway     = errors.New("bad gateway")
	ErrInternalServer = errors.New("internal server error")
)
