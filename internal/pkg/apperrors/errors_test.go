package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		want    string
		wantErr bool
	}{
		{
			name: "with wrapped error",
			err: &AppError{
				Message: "invalid input",
				Err:     errors.New("field required"),
			},
			want: "invalid input: field required",
		},
		{
			name: "without wrapped error",
			err: &AppError{
				Message: "not found",
			},
			want: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppError_GRPCStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		wantCode codes.Code
	}{
		{
			name:     "invalid input error",
			err:      NewInvalidInputError("invalid input", nil),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "not found error",
			err:      NewNotFoundError("not found", nil),
			wantCode: codes.NotFound,
		},
		{
			name:     "unauthorized error",
			err:      NewUnauthorizedError("unauthorized", nil),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "internal error",
			err:      NewInternalError("internal error", nil),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := status.Convert(tt.err)
			assert.Equal(t, tt.wantCode, status.Code())
		})
	}
}
