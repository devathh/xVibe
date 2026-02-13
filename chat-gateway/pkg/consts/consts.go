package consts

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	TYPE_SELF  = 1
	TYPE_GROUP = 2
)

var (
	CodeMap = map[codes.Code]int{
		codes.Canceled:         http.StatusRequestTimeout,
		codes.Unauthenticated:  http.StatusUnauthorized,
		codes.InvalidArgument:  http.StatusBadRequest,
		codes.DeadlineExceeded: http.StatusRequestTimeout,
		codes.NotFound:         http.StatusNotFound,
		codes.Internal:         http.StatusInternalServerError,
		codes.PermissionDenied: http.StatusForbidden,
	}
)
