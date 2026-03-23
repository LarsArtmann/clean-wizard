package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	t.Parallel()

	// Should not panic
	Init(true)
	assert.NotNil(t, L)
	assert.NotNil(t, SugaredLogger)

	// Cleanup
	Sync()
}

func TestInitWithLevel(t *testing.T) {
	t.Parallel()

	// Should not panic with valid level
	InitWithLevel("debug", true)
	assert.NotNil(t, L)

	// Cleanup
	Sync()
}

func TestLoggerMethods(t *testing.T) {
	// Initialize logger for testing
	Init(true)
	defer Sync()

	// These should not panic even without actual output checking
	t.Run("Debug", func(t *testing.T) {
		t.Parallel()
		Debug("debug message", zap.String("key", "value"))
	})

	t.Run("Info", func(t *testing.T) {
		t.Parallel()
		Info("info message", zap.String("key", "value"))
	})

	t.Run("Warn", func(t *testing.T) {
		t.Parallel()
		Warn("warn message", zap.String("key", "value"))
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		Error("error message", zap.String("key", "value"))
	})
}

func TestWith(t *testing.T) {
	Init(true)
	defer Sync()

	child := With(zap.String("context", "test"))
	assert.NotNil(t, child)
}

func TestNamed(t *testing.T) {
	Init(true)
	defer Sync()

	named := Named("test-scope")
	assert.NotNil(t, named)
}

func TestCleanerLogger(t *testing.T) {
	Init(true)
	defer Sync()

	cleanerLog := CleanerLogger("docker")
	assert.NotNil(t, cleanerLog)
}

func TestFieldHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    zap.Field
		expected string
	}{
		{"String", String("key", "value"), "key"},
		{"Int", Int("count", 42), "count"},
		{"Int64", Int64("size", 1024), "size"},
		{"Uint64", Uint64("bytes", 2048), "bytes"},
		{"Bool", Bool("enabled", true), "enabled"},
		{"Path", Path("path", "/tmp/test"), "path"},
		{"ByteSize", ByteSize("freed", 1024), "freed_bytes"},
		{"Operation", Operation("docker"), "operation"},
		{"DryRun", DryRun(true), "dry_run"},
		{"DurationMillis", DurationMillis("elapsed", 100), "elapsed_ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.field.Key)
		})
	}
}

func TestErrorField(t *testing.T) {
	t.Parallel()

	err := assert.AnError
	field := ErrorField(err)
	assert.Equal(t, "error", field.Key)
	assert.Equal(t, err, field.Interface)
}

func TestSyncWithNil(t *testing.T) {
	t.Parallel()

	// Should not panic even when L is nil
	originalL := L
	L = nil
	defer func() { L = originalL }()

	Sync() // Should not panic
}

func TestDebugWithNil(t *testing.T) {
	t.Parallel()

	// Should not panic even when L is nil
	originalL := L
	L = nil
	defer func() { L = originalL }()

	Debug("test message") // Should not panic
}
