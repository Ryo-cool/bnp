package apperrors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorCode はアプリケーション固有のエラーコードを定義します
type ErrorCode int

const (
	// ErrUnknown は未知のエラーを表します
	ErrUnknown ErrorCode = iota
	// ErrInvalidInput は入力値が不正な場合のエラーを表します
	ErrInvalidInput
	// ErrNotFound はリソースが見つからない場合のエラーを表します
	ErrNotFound
	// ErrAlreadyExists は既に存在するリソースを作成しようとした場合のエラーを表します
	ErrAlreadyExists
	// ErrUnauthorized は認証が必要な場合のエラーを表します
	ErrUnauthorized
	// ErrForbidden は権限が不足している場合のエラーを表します
	ErrForbidden
	// ErrInternal は内部エラーを表します
	ErrInternal
)

// CustomError はアプリケーション固有のエラー情報を保持します
type CustomError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error はエラーメッセージを返します
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap は内部エラーを返します
func (e *CustomError) Unwrap() error {
	return e.Err
}

// New は新しいCustomErrorを作成します
func New(code ErrorCode, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ToGRPCError はCustomErrorをgRPCのstatusエラーに変換します
func (e *CustomError) ToGRPCError() error {
	var grpcCode codes.Code
	switch e.Code {
	case ErrInvalidInput:
		grpcCode = codes.InvalidArgument
	case ErrNotFound:
		grpcCode = codes.NotFound
	case ErrAlreadyExists:
		grpcCode = codes.AlreadyExists
	case ErrUnauthorized:
		grpcCode = codes.Unauthenticated
	case ErrForbidden:
		grpcCode = codes.PermissionDenied
	case ErrInternal:
		grpcCode = codes.Internal
	default:
		grpcCode = codes.Unknown
	}

	st := status.New(grpcCode, e.Message)
	return st.Err()
}

// FromGRPCError はgRPCのstatusエラーからCustomErrorを作成します
func FromGRPCError(err error) *CustomError {
	st, ok := status.FromError(err)
	if !ok {
		return New(ErrUnknown, "Unknown error", err)
	}

	var code ErrorCode
	switch st.Code() {
	case codes.InvalidArgument:
		code = ErrInvalidInput
	case codes.NotFound:
		code = ErrNotFound
	case codes.AlreadyExists:
		code = ErrAlreadyExists
	case codes.Unauthenticated:
		code = ErrUnauthorized
	case codes.PermissionDenied:
		code = ErrForbidden
	case codes.Internal:
		code = ErrInternal
	default:
		code = ErrUnknown
	}

	return New(code, st.Message(), nil)
}
