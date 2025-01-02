package apperrors

import (
	"github.com/my-backend-project/internal/pkg/i18n"
)

// ErrorWrapper はエラーとi18nを統合するラッパーです
type ErrorWrapper struct {
	translator *i18n.Translator
}

// NewErrorWrapper は新しいErrorWrapperを作成します
func NewErrorWrapper(translator *i18n.Translator) *ErrorWrapper {
	return &ErrorWrapper{
		translator: translator,
	}
}

// WrapError はエラーをi18n対応のエラーに変換します
func (w *ErrorWrapper) WrapError(err *CustomError, lang i18n.Language, args ...interface{}) error {
	var key i18n.MessageKey
	switch err.Code {
	case ErrInvalidInput:
		key = i18n.ErrorInvalidInput
	case ErrNotFound:
		key = i18n.ErrorNotFound
	case ErrAlreadyExists:
		key = i18n.ErrorAlreadyExists
	case ErrUnauthorized:
		key = i18n.ErrorUnauthorized
	case ErrForbidden:
		key = i18n.ErrorForbidden
	case ErrInternal:
		key = i18n.ErrorInternal
	default:
		return err
	}

	message := w.translator.Translate(lang, key, args...)
	return &CustomError{
		Code:    err.Code,
		Message: message,
		Err:     err.Err,
	}
}

// WrapValidationError はバリデーションエラーをi18n対応のエラーに変換します
func (w *ErrorWrapper) WrapValidationError(field string, lang i18n.Language) error {
	message := w.translator.Translate(lang, i18n.MessageKey("validation.required"), field)
	return &CustomError{
		Code:    ErrInvalidInput,
		Message: message,
		Err:     nil,
	}
}

// WrapAuthError は認証エラーをi18n対応のエラーに変換します
func (w *ErrorWrapper) WrapAuthError(key i18n.MessageKey, lang i18n.Language) error {
	message := w.translator.Translate(lang, key)
	return &CustomError{
		Code:    ErrUnauthorized,
		Message: message,
		Err:     nil,
	}
}

// WrapTaskError はタスク関連のエラーをi18n対応のエラーに変換します
func (w *ErrorWrapper) WrapTaskError(key i18n.MessageKey, lang i18n.Language, args ...interface{}) error {
	message := w.translator.Translate(lang, key, args...)
	var code ErrorCode
	switch key {
	case i18n.MessageKey("task.not_found"):
		code = ErrNotFound
	case i18n.MessageKey("task.invalid_status"):
		code = ErrInvalidInput
	default:
		code = ErrInternal
	}
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     nil,
	}
}

// WrapUserError はユーザー関連のエラーをi18n対応のエラーに変換します
func (w *ErrorWrapper) WrapUserError(key i18n.MessageKey, lang i18n.Language, args ...interface{}) error {
	message := w.translator.Translate(lang, key, args...)
	var code ErrorCode
	switch key {
	case i18n.MessageKey("user.not_found"):
		code = ErrNotFound
	case i18n.MessageKey("user.invalid_email"), i18n.MessageKey("user.invalid_password"):
		code = ErrInvalidInput
	case i18n.MessageKey("user.already_exists"):
		code = ErrAlreadyExists
	default:
		code = ErrInternal
	}
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     nil,
	}
}
