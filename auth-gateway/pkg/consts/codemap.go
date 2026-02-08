package consts

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	CodeMap = map[codes.Code]int{
		codes.Canceled:         http.StatusRequestTimeout,
		codes.InvalidArgument:  http.StatusBadRequest,
		codes.DeadlineExceeded: http.StatusRequestTimeout,
		codes.AlreadyExists:    http.StatusConflict,
		codes.Internal:         http.StatusInternalServerError,
		codes.NotFound:         http.StatusNotFound,
		codes.Unauthenticated:  http.StatusUnauthorized,
		codes.PermissionDenied: http.StatusForbidden,
	}
)
