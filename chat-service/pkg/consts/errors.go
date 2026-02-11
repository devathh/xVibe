package consts

import "errors"

var (
	// Domain's
	ErrInvalidType   = errors.New("invalid type of chat")
	ErrTitleInSelf   = errors.New("impossible to create title of self group")
	ErrInvalidTitle  = errors.New("invalid title of group")
	ErrInvalidChatID = errors.New("invalid chat's id")

	// General's
	ErrInvalidToken = errors.New("invalid token")

	// Repository's
	ErrInvalidMemberID = errors.New("invalid member's id")
	ErrUserIsntOwner   = errors.New("user isn't owner of this chat")
	ErrChatNotFound    = errors.New("chat not found")
	ErrChatsNotFound   = errors.New("there isn't any chats")
	ErrUsersDontExist  = errors.New("some users don't exist")
	ErrDeleteOwner     = errors.New("you can't delete the owner of chat")

	// Service's
	ErrInvalidRequest = errors.New("invalid request")
	ErrSelfChat       = errors.New("you can't create a chat with yourself")
	ErrInternalServer = errors.New("internal server error")
)
