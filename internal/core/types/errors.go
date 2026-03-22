package types

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

var (
	ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "not found"}
	ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &AppError{Code: http.StatusForbidden, Message: "forbidden"}
	ErrConflict     = &AppError{Code: http.StatusConflict, Message: "conflict"}
	ErrValidation   = &AppError{Code: http.StatusUnprocessableEntity, Message: "validation error"}
)

func NewNotFound(detail string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: "not found", Detail: detail}
}

func NewValidation(detail string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: "validation error", Detail: detail}
}

func NewConflict(detail string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: "conflict", Detail: detail}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
