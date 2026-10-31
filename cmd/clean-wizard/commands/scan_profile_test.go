package commands

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestFilterCleanersByProfile(t *testing.T) {
	t.Parallel()

	configs := []CleanerConfig{
		{Name: cleaner.CleanerNix, Available: CleanerAvailabilityAvailable},
		{Name: cleaner.CleanerDocker, Available: CleanerAvailabilityAvailable},
		{Name: cleaner.CleanerCargo, Available: CleanerAvailabilityAvailable},
	}

	cfg := &types.Config{
		Profiles: map[string]*types.Profile{
			"weekly": {
				Operations: []types.CleanupOperation{
					{Name: cleaner.CleanerNix},
					{Name: cleaner.CleanerDocker},
					{Name: "not-a-cleaner"},
				},
			},
			"empty": {
				Operations: []types.CleanupOperation{
					{Name: "not-a-cleaner"},
				},
			},
		},
	}

	t.Run("known profile narrows to listed cleaners", func(t *testing.T) {
		t.Parallel()

		filtered := filterCleanersByProfile(configs, cfg, "weekly")

		assert.Len(t, filtered, 2)
		assert.Equal(t, cleaner.CleanerNix, filtered[0].Name)
		assert.Equal(t, cleaner.CleanerDocker, filtered[1].Name)
	})

	t.Run("unknown profile is a no-op", func(t *testing.T) {
		t.Parallel()

		filtered := filterCleanersByProfile(configs, cfg, "missing")

		assert.Equal(t, configs, filtered)
	})

	t.Run("profile matching nothing yields empty slice", func(t *testing.T) {
		t.Parallel()

		filtered := filterCleanersByProfile(configs, cfg, "empty")

		assert.Empty(t, filtered)
	})
}
