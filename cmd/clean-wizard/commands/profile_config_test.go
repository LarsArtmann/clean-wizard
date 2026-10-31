package commands

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTestConfigHome points CONFIG_PATH at a temp file so profile/config
// command runs never touch the user's real configuration.
func withTestConfigHome(t *testing.T) string {
	t.Helper()

	path := t.TempDir() + "/clean-wizard.yaml"
	t.Setenv("CONFIG_PATH", path)

	return path
}

func TestRunProfileListCommand_EmptyConfig(t *testing.T) {

	withTestConfigHome(t)

	err := runProfileListCommand(nil, nil)

	require.NoError(t, err, "listing profiles of a missing config falls back to defaults")
}

func TestRunProfileShowCommand_RejectsMissingProfile(t *testing.T) {

	withTestConfigHome(t)

	err := runProfileShowCommand(nil, []string{"does-not-exist"}, false)

	require.Error(t, err)
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
}

func TestRunProfileCreateAndDelete_RoundTrip(t *testing.T) {

	configPath := withTestConfigHome(t)

	err := runProfileCreateCommand(nil, nil, "test-profile", "created by test")
	require.NoError(t, err)

	saved, err := config.LoadFromPath(configPath)
	require.NoError(t, err)
	assert.Contains(t, saved.Profiles, "test-profile")

	err = runProfileDeleteCommand(nil, []string{"test-profile"}, true)
	require.NoError(t, err)

	saved, err = config.LoadFromPath(configPath)
	require.NoError(t, err)
	assert.NotContains(t, saved.Profiles, "test-profile")
}

func TestRunProfileDeleteCommand_RejectsMissingProfile(t *testing.T) {

	withTestConfigHome(t)

	err := runProfileDeleteCommand(nil, []string{"ghost"}, true)

	require.Error(t, err)
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
}

func TestRunConfigShowCommand_MissingConfigFallsBackToDefaults(t *testing.T) {

	withTestConfigHome(t)

	err := runConfigShowCommand(nil, nil, false)

	require.NoError(t, err, "show on a missing config renders defaults instead of failing")
}

func TestRunConfigValidateCommand_MissingConfigFallsBackToDefaults(t *testing.T) {

	withTestConfigHome(t)

	err := runConfigValidateCommand(nil, nil, "")

	require.NoError(t, err, "default config must validate")
}

func TestRunConfigMigrateCommand_RejectsMissingConfig(t *testing.T) {
	withTestConfigHome(t)

	err := runConfigMigrateCommand(true, "", "")

	require.Error(t, err, "migrate refuses a missing config instead of inventing one")
	assert.Equal(t, errorfamily.Rejection, errorfamily.Classify(err))
}
