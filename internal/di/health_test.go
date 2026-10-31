package di

import (
	"context"
	"errors"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRegisteredContainer builds a container with all production services.
func newRegisteredContainer(t *testing.T) (*Container, func()) {
	t.Helper()

	container, cleanup := New()

	cfg := &types.Config{}
	err := RegisterAllServices(container.Injector(), cfg, RunSettings{Verbose: false, DryRun: true})
	require.NoError(t, err)

	return container, cleanup
}

func TestHealthReport_CoversEveryRegisteredService(t *testing.T) {
	t.Parallel()

	container, cleanup := newRegisteredContainer(t)
	defer cleanup()

	report, err := HealthReport(t.Context(), container.Injector())
	require.NoError(t, err)
	require.NotEmpty(t, report)

	// Every registered service appears exactly once and is healthy.
	seen := map[string]bool{}
	for _, entry := range report {
		require.False(t, seen[entry.Service], "duplicate service %s", entry.Service)
		seen[entry.Service] = true
		assert.True(t, entry.Healthy(), "service %s should be healthy: %v", entry.Service, entry.Error)
	}

	// Spot-check the interesting seams: alias, cleaners, registry.
	assert.Contains(t, seen, do.NameOf[adapters.NixStore]())
	assert.Contains(t, seen, cleanerServiceName(cleaner.CleanerNix))
	assert.Contains(t, seen, do.NameOf[*cleaner.Registry]())
}

func TestHealthReport_MergesHealthCheckFailureIntoAlias(t *testing.T) {
	t.Parallel()

	container, cleanup := newRegisteredContainer(t)
	defer cleanup()

	unhealthy := &fakeNixStore{healthErr: errors.New("nix probe failed")}
	do.Override(container.Injector(), func(i do.Injector) (adapters.NixStore, error) {
		return unhealthy, nil
	})

	report, err := HealthReport(t.Context(), container.Injector())
	require.NoError(t, err)

	aliasName := do.NameOf[adapters.NixStore]()

	var aliasEntry *HealthEntry
	for i := range report {
		if report[i].Service == aliasName {
			aliasEntry = &report[i]
		}
	}

	require.NotNil(t, aliasEntry, "alias service missing from report")
	assert.False(t, aliasEntry.Healthy())
	assert.ErrorContains(t, aliasEntry.Error, "nix probe failed")
}

func TestSummarize_CountsUnhealthy(t *testing.T) {
	t.Parallel()

	entries := []HealthEntry{
		{Service: "a"},
		{Service: "b", Error: errors.New("boom")},
		{Service: "c"},
	}

	summary := Summarize(entries)

	assert.Equal(t, 3, summary.Total)
	assert.Equal(t, 1, summary.Unhealthy)
}

func TestUnhealthyError_ClassifiesAsInfrastructure(t *testing.T) {
	t.Parallel()

	err := UnhealthyError(HealthSummary{Total: 18, Unhealthy: 2})

	require.Error(t, err)
	assert.Equal(t, errorfamily.Infrastructure, errorfamily.Classify(err))
	assert.Equal(t, "doctor.unhealthy", errorfamily.Code(err))
	assert.Equal(t, 69, errorfamily.ExitCode(err))
	assert.False(t, errorfamily.IsRetryable(err))
}
