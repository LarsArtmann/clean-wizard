package commands

import (
	"encoding/json/v2"
	"errors"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/di"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newHealthTestContainer builds a fully registered container for gate tests.
func newHealthTestContainer(t *testing.T) *di.Container {
	t.Helper()

	container, cleanup := di.New()
	t.Cleanup(cleanup)

	err := di.RegisterAllServices(
		container.Injector(),
		&types.Config{},
		di.RunSettings{Verbose: false, DryRun: true, MaxConcurrency: 0, Profile: ""}, //nolint:exhaustruct // test defaults
	)
	require.NoError(t, err)

	return container
}

// blockPATH points PATH at an empty dir so no nix binary is found; t.Setenv
// forbids parallel execution for the calling test.
func blockPATH(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", t.TempDir())
}

func TestEnforceCleanHealthGate_PassesOnHealthyContainer(t *testing.T) {
	t.Parallel()

	container := newHealthTestContainer(t)

	err := enforceCleanHealthGate(t.Context(), container.Injector())

	require.NoError(t, err)
}

func TestEnforceCleanHealthGate_BlocksOnUnhealthyContainer(t *testing.T) {
	// t.Setenv forbids parallel execution.
	container := newHealthTestContainer(t)

	blockPATH(t)

	err := enforceCleanHealthGate(t.Context(), container.Injector())

	require.Error(t, err)
	assert.ErrorIs(t, err, di.ErrDoctorUnhealthy)
	assert.Equal(t, errorfamily.Infrastructure, errorfamily.Classify(err))
	assert.Equal(t, 69, errorfamily.ExitCode(err))
}

func TestRunDoctorCommand_RejectsJSONAndSARIFTogether(t *testing.T) {
	t.Parallel()

	err := runDoctorCommand(doctorOptions{jsonOutput: true, sarifPath: "-"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDoctorJSONWithSARIF)
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
}

func TestBuildDoctorJSON_ShapeSnapshot(t *testing.T) {
	t.Parallel()

	report := []di.HealthEntry{
		{Service: "cleaner.nix"},
		{Service: "adapters.NixStore", Error: di.ErrDoctorUnhealthy},
	}

	data, err := buildDoctorJSON(report)
	require.NoError(t, err)

	var decoded struct {
		Services []struct {
			Service   string `json:"service"`
			Healthy   bool   `json:"healthy"`
			Error     string `json:"error"`
			Family    string `json:"family"`
			Code      string `json:"code"`
			Retryable bool   `json:"retryable"`
		} `json:"services"`
		Summary struct {
			Total     int  `json:"total"`
			Unhealthy int  `json:"unhealthy"`
			Healthy   bool `json:"healthy"`
		} `json:"summary"`
	}

	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Len(t, decoded.Services, 2)

	healthy := decoded.Services[0]
	assert.Equal(t, "cleaner.nix", healthy.Service)
	assert.True(t, healthy.Healthy)
	assert.Empty(t, healthy.Error)

	unhealthy := decoded.Services[1]
	assert.Equal(t, "adapters.NixStore", unhealthy.Service)
	assert.False(t, unhealthy.Healthy)
	assert.NotEmpty(t, unhealthy.Error)
	assert.Equal(t, "infrastructure", unhealthy.Family)
	assert.Equal(t, "doctor.unhealthy", unhealthy.Code)

	assert.Equal(t, 2, decoded.Summary.Total)
	assert.Equal(t, 1, decoded.Summary.Unhealthy)
	assert.False(t, decoded.Summary.Healthy)
}

func TestDoctorUnhealthySentinel_ExitCodeIs69(t *testing.T) {
	t.Parallel()

	wrapped := errors.Join(ErrDoctorUnhealthy)

	assert.Equal(t, 69, errorfamily.ExitCode(wrapped))
	assert.False(t, errorfamily.IsRetryable(wrapped))
}
