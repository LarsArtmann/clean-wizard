package config

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// withMigrationChain swaps the global format version and migration registry
// for a synthetic chain and restores both on cleanup. Tests using it must not
// run in parallel. An empty chain registers no migrations; the format version
// is advanced past the last step's To, or to 1.1.0 when the chain is empty.
func withMigrationChain(t *testing.T, chain ...Migration) {
	t.Helper()

	originalVersion := CurrentFormatVersion
	originalMigrations := migrations

	target := FormatVersion{Major: 1, Minor: 1, Patch: 0}
	if len(chain) > 0 {
		target = chain[len(chain)-1].To
	}

	CurrentFormatVersion = target
	migrations = chain

	t.Cleanup(func() {
		CurrentFormatVersion = originalVersion
		migrations = originalMigrations
	})
}

func migrationTestConfig(version string) *types.Config {
	return &types.Config{ //nolint:exhaustruct
		Version:      version,
		SafeMode:     enums.SafeModeEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/System"},
		Profiles: map[string]*types.Profile{
			"daily": {
				Name:        "daily",
				Description: "Quick daily cleanup",
				Enabled:     enums.ProfileStatusEnabled,
				Operations: []types.CleanupOperation{
					{
						Name:      "nix-generations",
						RiskLevel: enums.RiskLevelLowType,
						Enabled:   enums.ProfileStatusEnabled,
					},
				},
			},
		},
		LastClean: time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC),
		Updated:   time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC),
	}
}

func TestPlanMigration(t *testing.T) {
	t.Run("current version needs no plan", func(t *testing.T) {
		plan, err := PlanMigration(CurrentFormatVersion)
		if err != nil {
			t.Fatalf("PlanMigration error = %v", err)
		}

		if len(plan) != 0 {
			t.Errorf("PlanMigration = %d steps, want 0", len(plan))
		}
	})

	t.Run("chains multiple steps in order", func(t *testing.T) {
		withMigrationChain(t,
			Migration{From: FormatVersion{1, 0, 0}, To: FormatVersion{1, 1, 0}, Description: "step one"},
			Migration{From: FormatVersion{1, 1, 0}, To: FormatVersion{1, 2, 0}, Description: "step two"},
		)

		plan, err := PlanMigration(FormatVersion{1, 0, 0})
		if err != nil {
			t.Fatalf("PlanMigration error = %v", err)
		}

		if len(plan) != 2 {
			t.Fatalf("PlanMigration = %d steps, want 2", len(plan))
		}

		if plan[0].Description != "step one" || plan[1].Description != "step two" {
			t.Errorf("PlanMigration order = [%s, %s], want [step one, step two]",
				plan[0].Description, plan[1].Description)
		}
	})

	t.Run("missing link rejected", func(t *testing.T) {
		withMigrationChain(t) // no migrations registered

		_, err := PlanMigration(FormatVersion{1, 0, 0})
		if err == nil {
			t.Fatal("PlanMigration expected error for missing chain, got nil")
		}

		if !strings.Contains(err.Error(), "no migration path") {
			t.Errorf("error = %v, want it to mention the missing path", err)
		}
	})

	t.Run("non progressive step is skipped", func(t *testing.T) {
		withMigrationChain(t,
			Migration{From: FormatVersion{1, 0, 0}, To: FormatVersion{1, 0, 0}, Description: "self loop"},
		)

		_, err := PlanMigration(FormatVersion{1, 0, 0})
		if err == nil {
			t.Fatal("PlanMigration expected error for self-loop chain, got nil")
		}
	})
}

func TestApplyMigrations(t *testing.T) {
	addPath := func(path string) func(*types.Config) error {
		return func(config *types.Config) error {
			config.Protected = append(config.Protected, path)

			return nil
		}
	}

	t.Run("applies steps in order and stamps version", func(t *testing.T) {
		withMigrationChain(t,
			Migration{
				From: FormatVersion{1, 0, 0}, To: FormatVersion{1, 1, 0},
				Description: "protect home",
				Apply:       addPath("/home"),
			},
			Migration{
				From: FormatVersion{1, 1, 0}, To: FormatVersion{1, 2, 0},
				Description: "protect etc",
				Apply:       addPath("/etc"),
			},
		)

		config := migrationTestConfig("1.0.0")
		records, err := ApplyMigrations(config, Migrations())
		if err != nil {
			t.Fatalf("ApplyMigrations error = %v", err)
		}

		if config.Version != "1.2.0" {
			t.Errorf("Version = %q, want 1.2.0", config.Version)
		}

		if len(records) != 2 {
			t.Fatalf("records = %d, want 2", len(records))
		}

		if len(config.Protected) != 3 {
			t.Fatalf("protected = %v, want 3 entries", config.Protected)
		}

		if records[1].Changes == nil || len(records[1].Changes) != 1 {
			t.Fatalf("second record changes = %v, want exactly the new protected path", records[1].Changes)
		}

		if records[1].Changes[0].Path != "protected" {
			t.Errorf("change path = %q, want protected", records[1].Changes[0].Path)
		}
	})

	t.Run("apply failure aborts with corruption classification", func(t *testing.T) {
		withMigrationChain(t,
			Migration{
				From: FormatVersion{1, 0, 0}, To: FormatVersion{1, 1, 0},
				Description: "explode",
				Apply: func(*types.Config) error {
					return errors.New("boom")
				},
			},
		)

		config := migrationTestConfig("1.0.0")
		_, err := ApplyMigrations(config, Migrations())
		if err == nil {
			t.Fatal("ApplyMigrations expected error, got nil")
		}

		if !strings.Contains(err.Error(), "boom") {
			t.Errorf("error = %v, want the cause to be preserved", err)
		}
	})
}

func TestFlattenConfigDeterministic(t *testing.T) {
	config := migrationTestConfig("1.0.0")

	first := flattenConfig(config)
	second := flattenConfig(config)

	if len(first) != len(second) {
		t.Fatalf("flatten sizes differ: %d vs %d", len(first), len(second))
	}

	for key, value := range first {
		if second[key] != value {
			t.Errorf("key %q unstable: %q vs %q", key, value, second[key])
		}
	}
}
