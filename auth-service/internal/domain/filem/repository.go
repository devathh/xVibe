package filem

import "context"

type FilemRepository interface {
	GetPublicKey(context.Context) (*File, error)
}
