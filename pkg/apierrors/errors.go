package apierrors

import "net/http"

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string { return e.Message }

var (
	ErrNotFound     = &APIError{Code: http.StatusNotFound, Message: "not found"}
	ErrBadRequest   = &APIError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrInternal     = &APIError{Code: http.StatusInternalServerError, Message: "internal error"}
	ErrConflict     = &APIError{Code: http.StatusConflict, Message: "already exists"}
	ErrUnauthorized = &APIError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &APIError{Code: http.StatusForbidden, Message: "forbidden"}
)
