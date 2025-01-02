package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/my-backend-project/internal/pkg/apperrors"
	"github.com/my-backend-project/internal/pkg/i18n"
)

// Validator はバリデーション機能を提供します
type Validator struct {
	validator  *validator.Validate
	translator *i18n.Translator
}

// New は新しいValidatorを作成します
func New(translator *i18n.Translator) *Validator {
	return &Validator{
		validator:  validator.New(),
		translator: translator,
	}
}

// Validate は構造体のバリデーションを行います
func (v *Validator) Validate(s interface{}, lang i18n.Language) error {
	if err := v.validator.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// 最初のバリデーションエラーのみを返す
			firstErr := validationErrors[0]
			message := v.translator.Translate(lang, i18n.MessageKey("validation."+firstErr.Tag()))

			// パラメータがある場合はフォーマットする
			if firstErr.Param() != "" {
				message = fmt.Sprintf(message, firstErr.Param())
			}

			return &apperrors.ValidationError{
				Field:   firstErr.Field(),
				Message: message,
			}
		}
		return err
	}
	return nil
}

// ValidateAll は通常のバリデーションとクロスフィールドバリデーションを行います
func (v *Validator) ValidateAll(s interface{}, lang i18n.Language) error {
	// 通常のバリデーション
	if err := v.Validate(s, lang); err != nil {
		return err
	}

	// クロスフィールドバリデーション
	return v.validateCrossField(s)
}

// CustomValidator はEcho用のバリデーターラッパーです
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator は新しいCustomValidatorを作成します
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate はEchoのValidatorインターフェースを実装します
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// 最初のバリデーションエラーのみを返す
			firstErr := validationErrors[0]
			return &apperrors.ValidationError{
				Field:   firstErr.Field(),
				Message: fmt.Sprintf("%s のバリデーションに失敗しました: %s", firstErr.Field(), firstErr.Tag()),
			}
		}
		return err
	}
	return nil
}
