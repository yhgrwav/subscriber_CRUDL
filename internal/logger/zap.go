package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger - конструктор для логгера, принимающий уровень из конфигурации .env
// DEBUG
// INFO
// WARN
// ERROR
func NewLogger(logLevel string) (*zap.Logger, func() error, error) {
	lvl := zap.NewAtomicLevel()

	// читаем логлвл
	if err := lvl.UnmarshalText([]byte(logLevel)); err != nil {
		return nil, nil, fmt.Errorf("ошибка анмаршала уровня логов")
	}

	// создаём папку для логов с полными правами для создателя
	// и частичными правами для других пользователей
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, nil, fmt.Errorf("ошибка создания файла для логов:%w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	//создаём файл с названием времени создания файла для опциональной фильтрации по дате
	logFilePath := filepath.Join("logs", fmt.Sprintf("%s.log", timestamp))

	// если файла нет - он создаётся, также он доступен только для письма в него
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка открытия logs:%w", err)
	}

	//создаем конфиг для логгера и указываем формат времени как в timestamp
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")

	encoder := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl),
		zapcore.NewCore(encoder, zapcore.AddSync(logFile), lvl),
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	return logger, logFile.Close, nil
}
