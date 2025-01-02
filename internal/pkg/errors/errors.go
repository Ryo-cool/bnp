package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
)

type ErrorType string

const (
	NotFound     ErrorType = "not_found"
	InvalidInput ErrorType = "invalid_input"
	Internal     ErrorType = "internal"
)

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
	Code    codes.Code
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Type:    NotFound,
		Message: message,
		Err:     err,
	}
}

func NewInvalidInputError(message string, err error) *AppError {
	return &AppError{
		Type:    InvalidInput,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    Internal,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Message: message,
		Err:     err,
		Code:    codes.Unauthenticated,
	}
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func IsNotFound(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if As(err, &appErr) {
		return appErr.Type == NotFound
	}
	return false
}

func IsInvalidInput(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if As(err, &appErr) {
		return appErr.Type == InvalidInput
	}
	return false
}

func IsInternal(err error) bool {
	var appErr *AppError
	if err == nil {
		return false
	}
	if As(err, &appErr) {
		return appErr.Type == Internal
	}
	return false
}
