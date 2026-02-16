package consts

import (
	"errors"
)

var (
	// Domain's
	ErrInvalidFirstname = errors.New("len of firstname must be more than 2")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidUsername  = errors.New("username must be more than 2 and less than 15")
	ErrInvalidPassword  = errors.New("password must be more than 6 letters")
	ErrInvalidFilename  = errors.New("invalid name of file")
	ErrInvalidContent   = errors.New("content of file cannot be empty")

	// General's
	ErrInvalidToken = errors.New("invalid token")
	ErrChanClosed   = errors.New("chan is closed")

	// Repository's
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserDoesntExist    = errors.New("user doesn't exist")
	ErrUserAlreadyTaken   = errors.New("email or username is already taken")
	ErrSessionDoesntExist = errors.New("session doesn't exist")

	// Service's
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidClientIP    = errors.New("invalid client ip")
	ErrInvalidUserAgent   = errors.New("invalid user's agent")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUserID      = errors.New("invalid user's id")
)
