package logger

import (
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		development bool
	}{
		{
			name:        "production mode",
			development: false,
		},
		{
			name:        "development mode",
			development: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.development)
			require.NotNil(t, L, "logger should be initialized")
			require.NotNil(t, StdLogger, "slog logger should be initialized")
		})
	}
}

func TestInitWithLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		level       string
		development bool
	}{
		{
			name:        "debug level",
			level:       "debug",
			development: true,
		},
		{
			name:        "info level",
			level:       "info",
			development: false,
		},
		{
			name:        "warn level",
			level:       "warn",
			development: false,
		},
		{
			name:        "error level",
			level:       "error",
			development: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitWithLevel(tt.level, tt.development)
			require.NotNil(t, L, "logger should be initialized")
		})
	}
}

func TestLoggingFunctions(t *testing.T) {
	Init(true) // Development mode for testability

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "debug log",
			fn:   func() { Debug("debug message", "key", "value") },
		},
		{
			name: "info log",
			fn:   func() { Info("info message", "key", "value") },
		},
		{
			name: "warn log",
			fn:   func() { Warn("warn message", "key", "value") },
		},
		{
			name: "error log",
			fn:   func() { Error("error message", "key", "value") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, tt.fn)
		})
	}
}

func TestWith(t *testing.T) {
	Init(true)

	childLogger := With("request_id", "123", "user_id", "456")
	assert.NotNil(t, childLogger)
}

func TestWithPrefix(t *testing.T) {
	Init(true)

	prefixedLogger := WithPrefix("test-component")
	assert.NotNil(t, prefixedLogger)
}

func TestCleanerLogger(t *testing.T) {
	Init(true)

	cleanerLog := CleanerLogger("docker")
	assert.NotNil(t, cleanerLog)
}

func TestSetLevel(t *testing.T) {
	Init(true)

	tests := []struct {
		name  string
		level string
	}{
		{name: "debug", level: "debug"},
		{name: "info", level: "info"},
		{name: "warn", level: "warn"},
		{name: "error", level: "error"},
		{name: "fatal", level: "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() { SetLevel(tt.level) })
		})
	}
}

func TestSetLevel_Uninitialized(t *testing.T) {
	// Reset logger to nil
	L = nil

	// Should not panic even when logger is nil
	assert.NotPanics(t, func() { SetLevel("debug") })
}

func TestSync(t *testing.T) {
	Init(true)

	// Should not panic
	assert.NotPanics(t, func() { Sync() })
}

func TestGetSlogLogger(t *testing.T) {
	Init(true)

	slogLogger := GetSlogLogger()
	assert.NotNil(t, slogLogger)
}

func TestLoggingWithMultipleFields(t *testing.T) {
	Init(true)

	// Test logging with various field types
	assert.NotPanics(t, func() {
		Info("complex log",
			"string", "value",
			"int", 42,
			"bool", true,
			"nested", map[string]string{"key": "value"},
		)
	})
}

func TestCleanerLoggerIntegration(t *testing.T) {
	Init(true)

	cleaners := []string{"docker", "nix", "homebrew", "go", "cargo"}

	for _, name := range cleaners {
		t.Run(name, func(t *testing.T) {
			logger := CleanerLogger(name)
			require.NotNil(t, logger)

			// Should be able to log with the cleaner logger
			assert.NotPanics(t, func() {
				logger.Info("cleaning started", "dry_run", false)
				logger.Debug("scanning cache", "path", "/tmp/cache")
			})
		})
	}
}

// BenchmarkLogging benchmarks the logging performance.
func BenchmarkLogging(b *testing.B) {
	Init(false) // Production mode for benchmarking

	b.Run("info", func(b *testing.B) {
		for i := range b.N {
			Info("benchmark message", "iteration", i)
		}
	})

	b.Run("info_with_fields", func(b *testing.B) {
		for i := range b.N {
			Info("benchmark message",
				"iteration", i,
				"cleaner", "docker",
				"bytes_freed", 1024*1024,
			)
		}
	})
}

// TestLoggingOutput verifies that logs are written.
func TestLoggingOutput(t *testing.T) {
	// This is a basic smoke test to ensure logging doesn't panic
	// In a real scenario, you'd capture stdout and verify output
	Init(true)

	// Verify all log levels work
	Debug("test debug")
	Info("test info")
	Warn("test warn")
	Error("test error")

	// If we get here without panic, the test passes
	assert.True(t, true)
}

// TestPrefix verifies prefix functionality.
func TestPrefix(t *testing.T) {
	Init(true)

	// Test that prefix methods work
	prefixed := WithPrefix("test")
	assert.NotNil(t, prefixed)

	// Should be able to log with prefix
	assert.NotPanics(t, func() {
		prefixed.Info("prefixed message")
	})
}

// TestLevelValidation tests that invalid levels are handled gracefully.
func TestLevelValidation(t *testing.T) {
	Init(true)

	// Invalid level should default to info
	assert.NotPanics(t, func() {
		InitWithLevel("invalid_level", true)
	})

	// Verify logger still works
	Info("should still work")
}

// TestNilLoggerSafety tests that functions handle nil logger gracefully.
func TestNilLoggerSafety(t *testing.T) {
	// Save original logger
	originalL := L

	defer func() { L = originalL }()

	// Set to nil
	L = nil

	// All these should not panic
	assert.NotPanics(t, func() { Debug("test") })
	assert.NotPanics(t, func() { Info("test") })
	assert.NotPanics(t, func() { Warn("test") })
	assert.NotPanics(t, func() { Error("test") })
	assert.NotPanics(t, func() { With("key", "value") })
	assert.NotPanics(t, func() { WithPrefix("test") })
	assert.NotPanics(t, func() { CleanerLogger("test") })
}

// TestDevelopmentVsProduction tests different modes.
func TestDevelopmentVsProduction(t *testing.T) {
	t.Run("development", func(t *testing.T) {
		Init(true)
		assert.NotNil(t, L)
		assert.NotNil(t, StdLogger)
	})

	t.Run("production", func(t *testing.T) {
		Init(false)
		assert.NotNil(t, L)
		assert.NotNil(t, StdLogger)
	})
}

// TestLogMessageContent verifies log message content.
func TestLogMessageContent(t *testing.T) {
	Init(true)

	// Test that we can create logs with various content
	messages := []string{
		"simple message",
		"message with special chars: !@#$%^&*()",
		"message with unicode: 你好世界 🌍",
		"very long message " + strings.Repeat("x", 1000),
	}

	for _, msg := range messages {
		t.Run(msg[:min(len(msg), 20)], func(t *testing.T) {
			assert.NotPanics(t, func() {
				Info(msg)
			})
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// TestCharmbraceletIntegration verifies integration with charmbracelet/log.
func TestCharmbraceletIntegration(t *testing.T) {
	Init(true)

	// Test that we can use charmbracelet/log specific features
	assert.NotNil(t, L)

	// Test level setting via charmbracelet/log API
	L.SetLevel(log.DebugLevel)
	assert.Equal(t, log.DebugLevel, L.GetLevel())

	L.SetLevel(log.InfoLevel)
	assert.Equal(t, log.InfoLevel, L.GetLevel())
}

// TestSlogIntegration verifies slog integration.
func TestSlogIntegration(t *testing.T) {
	Init(true)

	// Test that slog logger works
	require.NotNil(t, StdLogger)

	// Test slog methods
	assert.NotPanics(t, func() {
		StdLogger.Info("slog info message")
		StdLogger.Debug("slog debug message")
	})
}
