package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	// テスト用のログディレクトリを作成
	tmpDir := os.TempDir()
	logDir := filepath.Join(tmpDir, "test-logs")
	logFile := filepath.Join(logDir, "test.log")

	// テスト終了後にクリーンアップ
	defer os.RemoveAll(logDir)

	// ロガーの設定
	cfg := &Config{
		Level:      "debug",
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     1,
		Compress:   true,
		Console:    true,
	}

	// ロガーの初期化
	err := Init(cfg)
	assert.NoError(t, err)

	// 各ログレベルのテスト
	t.Run("log levels", func(t *testing.T) {
		Debug("debug message", zap.String("key", "value"))
		Info("info message", zap.Int("count", 1))
		Warn("warn message", zap.Bool("flag", true))
		Error("error message", zap.Float64("value", 3.14))
	})

	// WithFieldsのテスト
	t.Run("with fields", func(t *testing.T) {
		logger := WithFields(
			zap.String("service", "test"),
			zap.String("environment", "testing"),
		)
		assert.NotNil(t, logger)
		logger.Info("test message with fields")
	})

	// ログファイルの存在確認
	t.Run("log file exists", func(t *testing.T) {
		_, err := os.Stat(logFile)
		assert.NoError(t, err)
	})

	// Syncのテスト（コンソール出力無効）
	t.Run("sync", func(t *testing.T) {
		// コンソール出力を無効にした設定で再初期化
		cfg.Console = false
		err := Init(cfg)
		assert.NoError(t, err)

		err = Sync()
		assert.NoError(t, err)
	})
}

func TestNewTestLogger(t *testing.T) {
	logger, err := NewTestLogger()
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("test message")
}

func TestLoggerWithInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "invalid log level",
			cfg: &Config{
				Level:   "invalid",
				Console: true,
			},
			wantErr: true,
		},
		{
			name: "invalid file path",
			cfg: &Config{
				Level:    "info",
				Filename: "/invalid/path/that/does/not/exist/log.txt",
				Console:  false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
