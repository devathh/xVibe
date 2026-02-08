package consts

import "errors"

var (
	// General's
	ErrInvalidToken = errors.New("invalid token")

	// Service's
	ErrInternalServer = errors.New("internal server error")
	ErrBadGateway     = errors.New("bad gateway")
)
