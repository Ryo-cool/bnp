package validator

import (
	"reflect"

	"github.com/my-backend-project/internal/pkg/apperrors"
)

// CrossFieldValidator はクロスフィールドバリデーションを行うインターフェースです
type CrossFieldValidator interface {
	ValidateCrossField(s interface{}) error
}

// PasswordConfirmation はパスワードと確認用パスワードのバリデーションを行います
type PasswordConfirmation struct {
	Password        string `validate:"required,min=8"`
	PasswordConfirm string `validate:"required"`
}

func (p *PasswordConfirmation) ValidateCrossField(_ interface{}) error {
	if p.Password != p.PasswordConfirm {
		return &apperrors.ValidationError{
			Field:   "PasswordConfirm",
			Message: "パスワードが一致しません",
		}
	}
	return nil
}

// DateRange は日付範囲のバリデーションを行います
type DateRange struct {
	StartDate string `validate:"required"`
	EndDate   string `validate:"required"`
}

func (d *DateRange) ValidateCrossField(_ interface{}) error {
	if d.StartDate > d.EndDate {
		return &apperrors.ValidationError{
			Field:   "DateRange",
			Message: "開始日は終了日より前である必要があります",
		}
	}
	return nil
}

// validateCrossField はクロスフィールドバリデーションを行います
func (v *Validator) validateCrossField(s interface{}) error {
	// 直接のバリデーション
	if validator, ok := s.(CrossFieldValidator); ok {
		if err := validator.ValidateCrossField(s); err != nil {
			return err
		}
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	// フィールドのバリデーション
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 埋め込みフィールドの場合
		if fieldType.Anonymous {
			if err := v.validateCrossField(field.Interface()); err != nil {
				return err
			}
			continue
		}

		// 通常の構造体フィールドの場合
		if field.Kind() == reflect.Struct {
			if err := v.validateCrossField(field.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}
