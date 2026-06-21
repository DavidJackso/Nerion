package apierrors

import "net/http"

type APIError struct {
	Code      int
	ErrorCode string
	Message   string
	Fields    map[string]string
}

func (e *APIError) Error() string { return e.Message }

var (
	ErrNotFound     = &APIError{Code: http.StatusNotFound, ErrorCode: "not_found", Message: "Запись не найдена"}
	ErrBadRequest   = &APIError{Code: http.StatusBadRequest, ErrorCode: "bad_request", Message: "Некорректный запрос"}
	ErrInternal     = &APIError{Code: http.StatusInternalServerError, ErrorCode: "internal_error", Message: "Внутренняя ошибка сервера"}
	ErrConflict     = &APIError{Code: http.StatusConflict, ErrorCode: "conflict", Message: "Уже существует"}
	ErrUnauthorized = &APIError{Code: http.StatusUnauthorized, ErrorCode: "unauthorized", Message: "Необходима авторизация"}
	ErrForbidden    = &APIError{Code: http.StatusForbidden, ErrorCode: "forbidden", Message: "Доступ запрещён"}
)

func NewError(code int, errorCode, message string) *APIError {
	return &APIError{Code: code, ErrorCode: errorCode, Message: message}
}

func NewValidationError(fields map[string]string) *APIError {
	return &APIError{
		Code:      http.StatusUnprocessableEntity,
		ErrorCode: "validation_failed",
		Message:   "Ошибка валидации",
		Fields:    fields,
	}
}
