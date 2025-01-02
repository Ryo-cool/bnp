package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorCode int

const (
	Unknown ErrorCode = iota
	InvalidInput
	NotFound
	AlreadyExists
	Unauthorized
	Internal
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
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

// GRPCStatus converts AppError to gRPC status
func (e *AppError) GRPCStatus() *status.Status {
	var code codes.Code
	switch e.Code {
	case InvalidInput:
		code = codes.InvalidArgument
	case NotFound:
		code = codes.NotFound
	case AlreadyExists:
		code = codes.AlreadyExists
	case Unauthorized:
		code = codes.Unauthenticated
	case Internal:
		code = codes.Internal
	default:
		code = codes.Unknown
	}
	return status.New(code, e.Error())
}

// Error constructors
func NewInvalidInputError(message string, err error) *AppError {
	return &AppError{
		Code:    InvalidInput,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Code:    NotFound,
		Message: message,
		Err:     err,
	}
}

func NewAlreadyExistsError(message string, err error) *AppError {
	return &AppError{
		Code:    AlreadyExists,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Code:    Unauthorized,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    Internal,
		Message: message,
		Err:     err,
	}
}
