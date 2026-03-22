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
	errNotFound     = errors.New("not found")
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
	errConflict     = errors.New("conflict")
	errValidation   = errors.New("validation error")
)

func ErrNotFound() *AppError {
	return &AppError{Code: http.StatusNotFound, Message: errNotFound.Error()}
}

func ErrUnauthorized() *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: errUnauthorized.Error()}
}

func ErrForbidden() *AppError {
	return &AppError{Code: http.StatusForbidden, Message: errForbidden.Error()}
}

func ErrConflict() *AppError {
	return &AppError{Code: http.StatusConflict, Message: errConflict.Error()}
}

func ErrValidation() *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: errValidation.Error()}
}

func NewNotFound(detail string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: errNotFound.Error(), Detail: detail}
}

func NewValidation(detail string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: errValidation.Error(), Detail: detail}
}

func NewConflict(detail string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: errConflict.Error(), Detail: detail}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
