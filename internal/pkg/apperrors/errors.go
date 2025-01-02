package apperrors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorType string

const (
	NotFound     ErrorType = "not_found"
	InvalidInput ErrorType = "invalid_input"
	Internal     ErrorType = "internal"
	Unauthorized ErrorType = "unauthorized"
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

func (e *AppError) GRPCStatus() *status.Status {
	var code codes.Code
	switch e.Type {
	case NotFound:
		code = codes.NotFound
	case InvalidInput:
		code = codes.InvalidArgument
	case Unauthorized:
		code = codes.Unauthenticated
	default:
		code = codes.Internal
	}
	return status.New(code, e.Error())
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
		Type:    Unauthorized,
		Message: message,
		Err:     err,
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
