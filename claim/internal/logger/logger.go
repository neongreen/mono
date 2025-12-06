package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var log *slog.Logger

func init() {
	// Default: only show warnings and errors with tint formatter
	log = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelWarn,
		TimeFormat: "15:04:05",
	}))
}

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	level := slog.LevelWarn
	if enabled {
		level = slog.LevelDebug
	}
	log = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: "15:04:05",
	}))
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	log.Error(msg, args...)
}
