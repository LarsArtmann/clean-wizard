// Package logger provides structured logging for Clean Wizard.
//
// This package wraps charm.land/log/v2 (https://github.com/charmbracelet/log)
// to provide beautiful, colorful, structured logging throughout the application.
// It integrates seamlessly with the existing charmbracelet ecosystem (huh, lipgloss, bubbletea).
//
// Usage:
//
//	import "github.com/LarsArtmann/clean-wizard/internal/logger"
//
//	func main() {
//	    logger.Init(false) // or true for development
//	    defer logger.Sync()
//
//	    logger.Info("application started",
//	        "version", version.Version,
//	    )
//	}
package logger

import (
	"log/slog"
	"os"
	"time"

	"charm.land/log/v2"
)

// L is the global logger instance.
// Use this for all logging throughout the application.
var L *log.Logger

// StdLogger provides a standard slog.Logger for interoperability.
var StdLogger *slog.Logger

// levelFromString converts a level string to the corresponding log.Level.
func levelFromString(levelStr string) log.Level {
	switch levelStr {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}

// setupLogger creates the logger with the given level and development mode.
func setupLogger(level log.Level, development bool) {
	formatter := log.JSONFormatter
	if development {
		formatter = log.TextFormatter
	}

	L = log.NewWithOptions(os.Stdout, log.Options{
		Level:           level,
		Formatter:       formatter,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		Prefix:          "clean-wizard",
	})

	StdLogger = slog.New(L)
}

// Init initializes the global logger.
//
// In development mode, logs are colorful and human-readable.
// In production mode, logs are JSON-formatted for structured parsing.
//
// Example:
//
//	logger.Init(true)  // Development: colorful console output
//	logger.Init(false) // Production: JSON output
func Init(development bool) {
	level := log.InfoLevel
	if development {
		level = log.DebugLevel
	}

	setupLogger(level, development)
}

// InitWithLevel initializes the logger with a specific level.
//
// Valid levels: debug, info, warn, error, fatal.
func InitWithLevel(levelStr string, development bool) {
	level := levelFromString(levelStr)
	setupLogger(level, development)
}

// Sync flushes any buffered log entries.
// Call this before program exit.
func Sync() {
	// charmbracelet/log doesn't require explicit flushing
}

// Debug logs a debug message.
func Debug(msg string, keyvals ...any) {
	if L != nil {
		L.Debug(msg, keyvals...)
	}
}

// Info logs an info message.
func Info(msg string, keyvals ...any) {
	if L != nil {
		L.Info(msg, keyvals...)
	}
}

// Warn logs a warning message.
func Warn(msg string, keyvals ...any) {
	if L != nil {
		L.Warn(msg, keyvals...)
	}
}

// Error logs an error message.
func Error(msg string, keyvals ...any) {
	if L != nil {
		L.Error(msg, keyvals...)
	}
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, keyvals ...any) {
	if L != nil {
		L.Fatal(msg, keyvals...)
	} else {
		os.Exit(1)
	}
}

// With creates a child logger with additional fields.
func With(keyvals ...any) *log.Logger {
	if L != nil {
		return L.With(keyvals...)
	}

	return nil
}

// WithPrefix creates a child logger with a sub-scope name.
func WithPrefix(name string) *log.Logger {
	if L != nil {
		return L.WithPrefix(name)
	}
	// Return a new logger that discards output when L is nil
	return log.New(os.Stdout)
}

// CleanerLogger returns a logger for a specific cleaner.
func CleanerLogger(name string) *log.Logger {
	if L != nil {
		return WithPrefix("cleaner").With("cleaner_name", name)
	}

	return nil
}

// SetLevel changes the log level at runtime.
func SetLevel(level string) {
	if L == nil {
		return
	}

	L.SetLevel(levelFromString(level))
}

// GetSlogLogger returns the standard slog.Logger for interoperability.
func GetSlogLogger() *slog.Logger {
	return StdLogger
}
