package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(logLevel, logFile string, maxSize, maxBackups, maxAge int) (*zap.Logger, error) {
	// Parse log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create console encoder for development
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// JSON encoder for file
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Check if running in container (simplified log rotation)
	inContainer := os.Getenv("CONTAINER") == "true" || os.Getenv("IN_DOCKER") == "true"

	var fileWriter zapcore.WriteSyncer
	if inContainer {
		// In container: use simple file without rotation
		// Docker/Podman handles log rotation
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to stdout if file creation fails
			fileWriter = zapcore.AddSync(os.Stdout)
		} else {
			fileWriter = zapcore.AddSync(file)
		}
	} else {
		// Outside container: use lumberjack for rotation
		lumberjackLogger := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxSize,    // megabytes
			MaxBackups: maxBackups, // number of backups
			MaxAge:     maxAge,     // days
			Compress:   true,
		}
		fileWriter = zapcore.AddSync(lumberjackLogger)
	}

	// Create multi-writer (console + file)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(fileEncoder, fileWriter, level),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
