package di

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/logger"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swapDebugLogger redirects the logger package's slog sink into buf for the
// test's lifetime and restores it afterwards. Tests touching the package-level
// logger must not run in parallel with each other.
func swapDebugLogger(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	original := logger.StdLogger
	logger.StdLogger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	t.Cleanup(func() { logger.StdLogger = original })

	return &buf
}

// TestCleanupLogsShutdownErrorsAtDebug verifies the deferred cleanup logs a
// failing service's shutdown error at Debug instead of silently discarding it.
func TestCleanupLogsShutdownErrorsAtDebug(t *testing.T) {
	buf := swapDebugLogger(t)

	container, cleanup := New()

	do.Provide(container.Injector(), func(do.Injector) (do.ShutdownerWithError, error) {
		return failingShutdowner{}, nil
	})

	// do is lazy: the service only registers for shutdown once instantiated.
	shutdowner := do.MustInvoke[do.ShutdownerWithError](container.Injector())
	require.NotNil(t, shutdowner)

	cleanup()

	logs := buf.String()
	assert.Contains(t, logs, "DI shutdown reported service errors")
	assert.Contains(t, logs, "shutdown boom")
}

// TestCleanupSilentWhenNoShutdownErrors keeps the happy path log-free.
func TestCleanupSilentWhenNoShutdownErrors(t *testing.T) {
	buf := swapDebugLogger(t)

	container, cleanup := New()
	require.NotNil(t, container)

	cleanup()

	assert.False(t, strings.Contains(buf.String(), "DI shutdown"),
		"clean shutdown must not log: %s", buf.String())
}

type failingShutdowner struct{}

func (failingShutdowner) Shutdown() error {
	return errShutdownBoom
}

var errShutdownBoom = errors.New("shutdown boom")
