package consts

import "errors"

var (
	// Domain's
	ErrInvalidEncryptedBody = errors.New("body of message cannot be empty")
	ErrInvalidChatID        = errors.New("invalid chat's id")
	ErrInvalidMsgID         = errors.New("invalid msg's id")
	ErrInvalidAuthorID      = errors.New("invalid author's id")

	// General's
	ErrTooBigLimit  = errors.New("too big limit")
	ErrInvalidToken = errors.New("invalid token")

	// Repository's
	ErrChatNotFound   = errors.New("chat not found")
	ErrMsgNotFound    = errors.New("message not found")
	ErrNoEnoughRights = errors.New("user does not have sufficient rights")

	// Service's
	ErrInternalServer = errors.New("internal server error")
)
