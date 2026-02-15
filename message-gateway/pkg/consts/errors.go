package consts

import "errors"

var (
	// General's
	ErrInvalidToken = errors.New("invalid token")

	// Service's
	ErrBadGateway     = errors.New("bad gateway")
	ErrInternalServer = errors.New("internal server error")
)
