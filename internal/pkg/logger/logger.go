package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// defaultLogger はデフォルトのロガーインスタンスです
	defaultLogger *zap.Logger
)

// Config はロガーの設定を定義します
type Config struct {
	Level      string `json:"level" yaml:"level"`           // ログレベル
	Filename   string `json:"filename" yaml:"filename"`     // ログファイル名
	MaxSize    int    `json:"maxsize" yaml:"maxsize"`       // ログファイルの最大サイズ（MB）
	MaxBackups int    `json:"maxbackups" yaml:"maxbackups"` // 保持する古いログファイルの最大数
	MaxAge     int    `json:"maxage" yaml:"maxage"`         // 古いログファイルを保持する最大日数
	Compress   bool   `json:"compress" yaml:"compress"`     // 古いログファイルを圧縮するかどうか
	Console    bool   `json:"console" yaml:"console"`       // 標準出力にも出力するかどうか
}

// Init はロガーを初期化します
func Init(cfg *Config) error {
	// ログレベルの設定
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return fmt.Errorf("failed to parse log level: %v", err)
	}

	// エンコーダーの設定
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 出力先の設定
	var cores []zapcore.Core

	// ファイル出力の設定
	if cfg.Filename != "" {
		// ディレクトリの作成
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// ログローテーションの設定
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			w,
			level,
		)
		cores = append(cores, core)
	}

	// コンソール出力の設定
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, core)
	}

	// コアの結合
	core := zapcore.NewTee(cores...)

	// ロガーの作成
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	defaultLogger = logger
	return nil
}

// Debug はデバッグレベルのログを出力します
func Debug(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

// Info は情報レベルのログを出力します
func Info(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

// Warn は警告レベルのログを出力します
func Warn(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

// Error はエラーレベルのログを出力します
func Error(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

// Fatal は致命的なエラーレベルのログを出力します
func Fatal(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	}
}

// WithFields は指定されたフィールドを持つ新しいロガーを返します
func WithFields(fields ...zap.Field) *zap.Logger {
	if defaultLogger != nil {
		return defaultLogger.With(fields...)
	}
	return nil
}

// Sync はバッファされたログをフラッシュします
func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}

// NewTestLogger はテスト用のロガーを作成します
func NewTestLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return config.Build()
}
