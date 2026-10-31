// Package logger provides structured logging for Clean Wizard.
//
// This package wraps charm.land/log/v2 (https://github.com/charmbracelet/log)
// to provide beautiful, colorful, structured logging throughout the application.
// It integrates seamlessly with the existing charmbracelet ecosystem (huh, lipgloss, bubbletea).
//
// The process loggers live in atomic pointers, not bare globals: parallel
// tests (and any future concurrent init) swap or read them without data
// races. Init/InitWithLevel are still the single production write points;
// SetSlog exists for tests that capture log output.
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
	"sync/atomic"
	"time"

	"charm.land/log/v2"
)

// Process logger storage. Atomic pointers keep concurrent Init/reads
// race-free; the exported accessor functions are the only supported reads.
var (
	instance   atomic.Pointer[log.Logger]  //nolint:gochecknoglobals
	stdPointer atomic.Pointer[slog.Logger] //nolint:gochecknoglobals
)

// L returns the process logger, or nil before the first Init.
func L() *log.Logger { return instance.Load() }

// Slog returns the slog interoperability logger, or nil before the first Init.
func Slog() *slog.Logger { return stdPointer.Load() }

// SetSlog replaces the slog interoperability logger. Production code should
// not call this; it exists so tests can capture log output.
func SetSlog(l *slog.Logger) { stdPointer.Store(l) }

// Level strings for configuration parsing.
const (
	levelStringDebug = "debug"
	levelStringInfo  = "info"
	levelStringWarn  = "warn"
	levelStringError = "error"
	levelStringFatal = "fatal"
)

// levelFromString converts a level string to the corresponding log.Level.
func levelFromString(levelStr string) log.Level {
	switch levelStr {
	case levelStringDebug:
		return log.DebugLevel
	case levelStringInfo:
		return log.InfoLevel
	case levelStringWarn:
		return log.WarnLevel
	case levelStringError:
		return log.ErrorLevel
	case levelStringFatal:
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

	created := log.NewWithOptions(os.Stdout, log.Options{ //nolint:exhaustruct_v5
		Level:           level,
		Formatter:       formatter,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		Prefix:          "clean-wizard",
	})

	instance.Store(created)
	stdPointer.Store(slog.New(created))
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
	if l := L(); l != nil {
		l.Debug(msg, keyvals...)
	}
}

// Info logs an info message.
func Info(msg string, keyvals ...any) {
	if l := L(); l != nil {
		l.Info(msg, keyvals...)
	}
}

// Warn logs a warning message.
func Warn(msg string, keyvals ...any) {
	if l := L(); l != nil {
		l.Warn(msg, keyvals...)
	}
}

// Error logs an error message.
func Error(msg string, keyvals ...any) {
	if l := L(); l != nil {
		l.Error(msg, keyvals...)
	}
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, keyvals ...any) {
	if l := L(); l != nil {
		l.Fatal(msg, keyvals...)
	} else {
		os.Exit(1)
	}
}

// With creates a child logger with additional fields.
func With(keyvals ...any) *log.Logger {
	if l := L(); l != nil {
		return l.With(keyvals...)
	}

	return nil
}

// WithPrefix creates a child logger with a sub-scope name.
func WithPrefix(name string) *log.Logger {
	if l := L(); l != nil {
		return l.WithPrefix(name)
	}

	// Return a new logger that discards output when uninitialized.
	return log.New(os.Stdout)
}

// CleanerLogger returns a logger for a specific cleaner.
func CleanerLogger(name string) *log.Logger {
	if L() != nil {
		return WithPrefix("cleaner").With("cleaner_name", name)
	}

	return nil
}

// SetLevel changes the log level at runtime.
func SetLevel(level string) {
	if l := L(); l != nil {
		l.SetLevel(levelFromString(level))
	}
}

// GetSlogLogger returns the standard slog.Logger for interoperability.
func GetSlogLogger() *slog.Logger {
	return Slog()
}
