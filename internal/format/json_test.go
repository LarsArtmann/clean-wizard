package format

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanResultsToJSON_IncludesFamilyAndCode(t *testing.T) {
	t.Parallel()

	naErr := errorfamily.NewInfrastructure("cleaner.cargo.not_available", "cargo not available")
	results := map[string]domain.CleanResult{
		"nix": {FreedBytes: 1024, ItemsRemoved: 3},
	}
	skipped := map[string]error{
		"cargo": naErr,
	}
	failed := map[string]error{
		"docker": errors.New("docker daemon not running"),
	}

	data, err := CleanResultsToJSON(results, 0, false, skipped, failed)
	require.NoError(t, err)

	var output JSONOutput
	require.NoError(t, json.Unmarshal(data, &output))

	byName := make(map[string]CleanerResult, len(output.Cleaners))
	for _, c := range output.Cleaners {
		byName[c.Name] = c
	}

	// Skipped cleaner should carry family/code/retryable from errorfamily classification.
	cargo := byName["cargo"]
	assert.Equal(t, "skipped", cargo.Status)
	assert.Equal(t, "infrastructure", cargo.Family)
	assert.Equal(t, "cleaner.cargo.not_available", cargo.Code)
	assert.False(t, cargo.Retryable)

	// Failed cleaner should carry family from classification (defaults to Transient).
	docker := byName["docker"]
	assert.Equal(t, "failed", docker.Status)
	assert.Equal(t, "transient", docker.Family)
	assert.True(t, docker.Retryable)
}

func TestCleanResultsToJSON_DeterministicOrdering(t *testing.T) {
	t.Parallel()

	results := map[string]domain.CleanResult{
		"zebra":  {FreedBytes: 1},
		"alpha":  {FreedBytes: 2},
		"middle": {FreedBytes: 3},
	}

	// Run multiple times — the cleaners array must always be sorted by name.
	// (The cleaned_at timestamp varies, so we compare structurally, not raw bytes.)
	for range 5 {
		data, err := CleanResultsToJSON(results, 0, false, nil, nil)
		require.NoError(t, err)

		var output JSONOutput
		require.NoError(t, json.Unmarshal(data, &output))

		require.Len(t, output.Cleaners, 3)
		assert.Equal(t, "alpha", output.Cleaners[0].Name)
		assert.Equal(t, "middle", output.Cleaners[1].Name)
		assert.Equal(t, "zebra", output.Cleaners[2].Name)
	}
}
