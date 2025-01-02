package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// Language は対応言語を定義します
type Language string

const (
	// LanguageJa は日本語を表します
	LanguageJa Language = "ja"
	// LanguageEn は英語を表します
	LanguageEn Language = "en"
)

// MessageKey はメッセージキーを定義します
type MessageKey string

const (
	// ErrorInvalidInput は不正な入力エラーのメッセージキーです
	ErrorInvalidInput MessageKey = "error.invalid_input"
	// ErrorNotFound はリソースが見つからないエラーのメッセージキーです
	ErrorNotFound MessageKey = "error.not_found"
	// ErrorAlreadyExists は既に存在するエラーのメッセージキーです
	ErrorAlreadyExists MessageKey = "error.already_exists"
	// ErrorUnauthorized は認証エラーのメッセージキーです
	ErrorUnauthorized MessageKey = "error.unauthorized"
	// ErrorForbidden は権限エラーのメッセージキーです
	ErrorForbidden MessageKey = "error.forbidden"
	// ErrorInternal は内部エラーのメッセージキーです
	ErrorInternal MessageKey = "error.internal"
)

var (
	instance *Translator
	once     sync.Once
)

// Translator は翻訳機能を提供します
type Translator struct {
	messages map[Language]map[MessageKey]string
	mu       sync.RWMutex
}

// GetTranslator は翻訳インスタンスを返します
func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{
			messages: make(map[Language]map[MessageKey]string),
		}
	})
	return instance
}

// LoadMessages はメッセージファイルを読み込みます
func (t *Translator) LoadMessages(lang Language, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read message file: %w", err)
	}

	var messages map[MessageKey]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	t.mu.Lock()
	t.messages[lang] = messages
	t.mu.Unlock()

	return nil
}

// LoadAllMessages は指定されたディレクトリから全言語のメッセージを読み込みます
func (t *Translator) LoadAllMessages(dir string) error {
	languages := []Language{LanguageJa, LanguageEn}
	for _, lang := range languages {
		filename := fmt.Sprintf("messages_%s.json", lang)
		path := filepath.Join(dir, filename)
		if err := t.LoadMessages(lang, path); err != nil {
			return fmt.Errorf("failed to load messages for %s: %w", lang, err)
		}
	}
	return nil
}

// Translate はメッセージを翻訳します
func (t *Translator) Translate(lang Language, key MessageKey, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	messages, ok := t.messages[lang]
	if !ok {
		return string(key)
	}

	msg, ok := messages[key]
	if !ok {
		return string(key)
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
