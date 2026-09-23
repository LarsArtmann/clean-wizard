package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// loadKoanfFromYAML loads a koanf instance from a YAML string, mirroring the production load path.
func loadKoanfFromYAML(t *testing.T, yamlContent string) *koanf.Koanf {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.yaml")

	err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	k := koanf.New(".")

	err = k.Load(file.Provider(configPath), yaml.Parser())
	if err != nil {
		t.Fatalf("failed to load koanf: %v", err)
	}

	return k
}

func TestUnmarshalOperationSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yamlContent    string
		wantSettings   bool
		validateResult func(t *testing.T, settings *operations.OperationSettings)
	}{
		{
			name: "nix generations settings are parsed (string and integer enums)",
			yamlContent: `
profiles:
  daily:
    operations:
      - name: nix-generations
        settings:
          nix_generations:
            generations: 3
            optimize: "ENABLED"
            dry_run: 2
`,
			wantSettings: true,
			validateResult: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				nix := settings.NixGenerations
				if nix == nil {
					t.Fatal("NixGenerations settings not parsed")
				}

				if nix.Generations != 3 {
					t.Errorf("Generations = %d, want 3", nix.Generations)
				}

				if nix.Optimize != enums.OptimizationModeEnabled {
					t.Errorf("Optimize = %v, want ENABLED", nix.Optimize)
				}

				if nix.DryRun != enums.ExecutionModeForce {
					t.Errorf("DryRun = %v, want FORCE", nix.DryRun)
				}
			},
		},
		{
			name: "docker prune mode settings are parsed",
			yamlContent: `
profiles:
  daily:
    operations:
      - name: docker
        settings:
          docker:
            prune_mode: "VOLUMES"
`,
			wantSettings: true,
			validateResult: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				if settings.Docker == nil {
					t.Fatal("Docker settings not parsed")
				}

				if settings.Docker.PruneMode != enums.DockerPruneVolumes {
					t.Errorf("PruneMode = %v, want VOLUMES", settings.Docker.PruneMode)
				}
			},
		},
		{
			name: "temp files settings are parsed",
			yamlContent: `
profiles:
  daily:
    operations:
      - name: temp-files
        settings:
          temp_files:
            older_than: "14d"
            excludes:
              - "/tmp/keep"
              - "/var/tmp/keep"
`,
			wantSettings: true,
			validateResult: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				temp := settings.TempFiles
				if temp == nil {
					t.Fatal("TempFiles settings not parsed")
				}

				if temp.OlderThan != "14d" {
					t.Errorf("OlderThan = %q, want %q", temp.OlderThan, "14d")
				}

				if len(temp.Excludes) != 2 || temp.Excludes[0] != "/tmp/keep" {
					t.Errorf("Excludes = %v, want [/tmp/keep /var/tmp/keep]", temp.Excludes)
				}
			},
		},
		{
			name: "unknown settings sections are ignored without error",
			yamlContent: `
profiles:
  daily:
    operations:
      - name: docker
        settings:
          docker:
            prune_mode: 0
          unknown_cleaner:
            some_key: true
`,
			wantSettings: true,
			validateResult: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()

				if settings.Docker == nil {
					t.Fatal("Docker settings not parsed")
				}

				if settings.Docker.PruneMode != enums.DockerPruneAll {
					t.Errorf("PruneMode = %v, want ALL", settings.Docker.PruneMode)
				}
			},
		},
		{
			name: "invalid enum value leaves settings unset",
			yamlContent: `
profiles:
  daily:
    operations:
      - name: docker
        settings:
          docker:
            prune_mode: "NOT_A_MODE"
`,
			wantSettings: false,
			validateResult: func(t *testing.T, settings *operations.OperationSettings) {
				t.Helper()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			k := loadKoanfFromYAML(t, tt.yamlContent)
			op := types.CleanupOperation{Name: "test-operation"}

			unmarshalOperationSettings(k, "daily", 0, &op)

			if !tt.wantSettings {
				if op.Settings != nil {
					t.Fatalf("Settings = %+v, want nil", op.Settings)
				}

				return
			}

			if op.Settings == nil {
				t.Fatal("Settings not populated")
			}

			tt.validateResult(t, op.Settings)
		})
	}
}

func TestUnmarshalOperationSettings_NoSettingsBlock(t *testing.T) {
	t.Parallel()

	k := loadKoanfFromYAML(t, `
profiles:
  daily:
    operations:
      - name: docker
`)
	op := types.CleanupOperation{Name: "docker"}

	unmarshalOperationSettings(k, "daily", 0, &op)

	if op.Settings != nil {
		t.Errorf("Settings = %+v, want nil when no settings block exists", op.Settings)
	}
}
