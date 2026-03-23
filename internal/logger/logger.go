// Package logger provides structured logging for Clean Wizard.
//
// This package wraps zap (https://github.com/uber-go/zap) to provide
// structured, leveled logging throughout the application.
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
//	        zap.String("version", version.Version),
//	    )
//	}
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L is the global logger instance.
// Use this for all logging throughout the application.
var L *zap.Logger

// SugaredLogger provides a slower but more ergonomic API.
// Use for simple logging where performance is not critical.
var SugaredLogger *zap.SugaredLogger

// Init initializes the global logger.
//
// In development mode, logs are human-readable with colors.
// In production mode, logs are JSON-formatted for structured parsing.
//
// Example:
//
//	logger.Init(true)  // Development: console output with colors
//	logger.Init(false) // Production: JSON output
func Init(development bool) {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Always log to stdout, never stderr for normal operation
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	var err error
	L, err = config.Build(
		zap.AddCallerSkip(1), // Skip this wrapper in caller info
	)
	if err != nil {
		// Fallback to stdlib log if zap fails
		panic("failed to initialize logger: " + err.Error())
	}

	SugaredLogger = L.Sugar()
}

// InitWithLevel initializes the logger with a specific level.
//
// Valid levels: debug, info, warn, error, dpanic, panic, fatal
func InitWithLevel(level string, development bool) {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// Parse and set level
	lvl := zap.InfoLevel
	if err := lvl.UnmarshalText([]byte(level)); err == nil {
		config.Level = zap.NewAtomicLevelAt(lvl)
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	var err error
	L, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	SugaredLogger = L.Sugar()
}

// Sync flushes any buffered log entries.
// Call this before program exit.
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

// Debug logs a debug message.
func Debug(msg string, fields ...zap.Field) {
	if L != nil {
		L.Debug(msg, fields...)
	}
}

// Info logs an info message.
func Info(msg string, fields ...zap.Field) {
	if L != nil {
		L.Info(msg, fields...)
	}
}

// Warn logs a warning message.
func Warn(msg string, fields ...zap.Field) {
	if L != nil {
		L.Warn(msg, fields...)
	}
}

// Error logs an error message.
func Error(msg string, fields ...zap.Field) {
	if L != nil {
		L.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits.
func Fatal(msg string, fields ...zap.Field) {
	if L != nil {
		L.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}

// With creates a child logger with additional fields.
func With(fields ...zap.Field) *zap.Logger {
	if L != nil {
		return L.With(fields...)
	}
	return nil
}

// Named creates a child logger with a sub-scope name.
func Named(name string) *zap.Logger {
	if L != nil {
		return L.Named(name)
	}
	return nil
}

// CleanerLogger returns a logger for a specific cleaner.
func CleanerLogger(name string) *zap.Logger {
	return Named("cleaner").With(zap.String("cleaner_name", name))
}

// Common field helpers

// String creates a string field.
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int creates an int field.
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 creates an int64 field.
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Uint64 creates a uint64 field.
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

// Bool creates a bool field.
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Duration creates a duration field.
func Duration(key string, val interface{}) zap.Field {
	if d, ok := val.(time.Duration); ok {
		return zap.Duration(key, d)
	}
	return zap.String(key, "invalid duration")
}

// Error creates an error field.
func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

// Path creates a path field (sanitized for security).
func Path(key, val string) zap.Field {
	return zap.String(key, val)
}

// ByteSize creates a human-readable byte size field.
func ByteSize(key string, bytes int64) zap.Field {
	return zap.Int64(key+"_bytes", bytes)
}

// Operation creates an operation type field.
func Operation(opType string) zap.Field {
	return zap.String("operation", opType)
}

// DryRun indicates if this is a dry-run operation.
func DryRun(val bool) zap.Field {
	return zap.Bool("dry_run", val)
}

// DurationMillis creates a duration field in milliseconds.
func DurationMillis(key string, ms int64) zap.Field {
	return zap.Int64(key+"_ms", ms)
}
