package apperrors

// ValidationError はバリデーションエラーを表します
type ValidationError struct {
	Field   string
	Message string
}

// Error はエラーメッセージを返します
func (e *ValidationError) Error() string {
	return e.Message
}
