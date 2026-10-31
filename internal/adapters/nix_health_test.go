package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nixScriptDir builds a PATH whose only entry is dir. Callers create a fake
// `nix` binary inside dir to control what HealthCheck observes, keeping the
// tests hermetic on machines with and without a real nix installation.
func nixScriptDir(t *testing.T, script string) {
	t.Helper()

	dir := t.TempDir()

	if script != "" {
		path := filepath.Join(dir, "nix")
		require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	}

	t.Setenv("PATH", dir)
}

func TestNixAdapter_HealthCheck_HealthyBinary(t *testing.T) {
	nixScriptDir(t, "#!/bin/sh\necho \"nix (Nix) 2.34.7\"\n")

	adapter := NewNixAdapter(0, 0)

	err := adapter.HealthCheck(t.Context())

	require.NoError(t, err)
}

func TestNixAdapter_HealthCheck_FailingBinary(t *testing.T) {
	nixScriptDir(t, "#!/bin/sh\necho \"boom\" >&2\nexit 1\n")

	adapter := NewNixAdapter(0, 0)

	err := adapter.HealthCheck(t.Context())

	require.Error(t, err)

	notAvailable := &cleaner.NotAvailableError{}
	require.ErrorAs(t, err, &notAvailable)
	assert.Equal(t, "nix", notAvailable.CleanerName)
	assert.Equal(t, "cleaner.nix.not_available", notAvailable.Code)

	assert.True(t, cleaner.IsNotAvailableError(err), "health failure must classify as Infrastructure/not-available")
	assert.Equal(t, errorfamily.Infrastructure, errorfamily.Classify(err))
	assert.False(t, errorfamily.IsRetryable(err))
}

func TestNixAdapter_HealthCheck_MissingBinary(t *testing.T) {
	nixScriptDir(t, "")

	adapter := NewNixAdapter(0, 0)

	err := adapter.HealthCheck(t.Context())

	require.Error(t, err)
	assert.True(t, cleaner.IsNotAvailableError(err))
}

func TestNixAdapter_HealthCheck_RespectsDeadlinelessContext(t *testing.T) {
	nixScriptDir(t, "#!/bin/sh\nsleep 5\n")

	adapter := NewNixAdapter(50*1_000_000, 0) // 50ms timeout

	err := adapter.HealthCheck(t.Context())

	require.Error(t, err, "the adapter timeout must fire even without a caller deadline (ticket #56 pattern)")
}
