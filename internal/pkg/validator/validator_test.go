package validator

import (
	"testing"

	"github.com/my-backend-project/internal/pkg/apperrors"
	"github.com/my-backend-project/internal/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=32"`
}

type TestCrossFieldStruct struct {
	Password        string `validate:"required,min=8"`
	PasswordConfirm string `validate:"required"`
	StartDate       string `validate:"required"`
	EndDate         string `validate:"required"`
}

func (t *TestCrossFieldStruct) ValidateCrossField(_ interface{}) error {
	if t.Password != t.PasswordConfirm {
		return &apperrors.ValidationError{
			Field:   "PasswordConfirm",
			Message: "パスワードが一致しません",
		}
	}
	if t.StartDate > t.EndDate {
		return &apperrors.ValidationError{
			Field:   "DateRange",
			Message: "開始日は終了日より前である必要があります",
		}
	}
	return nil
}

func setupTest(t *testing.T) *Validator {
	translator := i18n.GetTranslator()
	err := translator.LoadMessages(i18n.LanguageJa, "../i18n/messages_ja.json")
	assert.NoError(t, err)
	return New(translator)
}

func TestValidator_Validate(t *testing.T) {
	v := setupTest(t)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "password too short",
			input: TestStruct{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "pass",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input, i18n.LanguageJa)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	v := setupTest(t)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid cross field validation",
			input: &TestCrossFieldStruct{
				Password:        "password123",
				PasswordConfirm: "password123",
				StartDate:       "2024-01-01",
				EndDate:         "2024-12-31",
			},
			wantErr: false,
		},
		{
			name: "password mismatch",
			input: &TestCrossFieldStruct{
				Password:        "password123",
				PasswordConfirm: "password124",
				StartDate:       "2024-01-01",
				EndDate:         "2024-12-31",
			},
			wantErr: true,
		},
		{
			name: "invalid date range",
			input: &TestCrossFieldStruct{
				Password:        "password123",
				PasswordConfirm: "password123",
				StartDate:       "2024-12-31",
				EndDate:         "2024-01-01",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateAll(tt.input, i18n.LanguageJa)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &apperrors.ValidationError{
		Field:   "test",
		Message: "error message",
	}
	assert.Equal(t, "error message", err.Error())
}
