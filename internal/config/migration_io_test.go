package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/knadh/koanf/parsers/yaml"
	errorfamily "github.com/larsartmann/go-error-family"
)

// writeTestConfigFile renders a configuration through the production save
// shape and writes it to path, so tests exercise the real file format.
func writeTestConfigFile(t *testing.T, path string, config map[string]any) {
	t.Helper()

	data, err := yaml.Parser().Marshal(config)
	if err != nil {
		t.Fatalf("marshal test config: %v", err)
	}

	if err := os.WriteFile(path, data, ConfigFilePermission); err != nil {
		t.Fatalf("write test config: %v", err)
	}
}

func simpleConfigFileMap(version string) map[string]any {
	return map[string]any{
		"version":                version,
		"safe_mode":              enums.SafeModeEnabled.String(),
		"max_disk_usage_percent": 50,
		"protected":              []string{"/System"},
		"last_clean":             time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC),
		"updated":                time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC),
		"profiles": map[string]any{
			"daily": map[string]any{
				"name":        "daily",
				"description": "Quick daily cleanup",
				"enabled":     int(enums.ProfileStatusEnabled),
				"operations": []any{
					map[string]any{
						"name":        "nix-generations",
						"description": "Clean old Nix generations",
						"risk_level":  int(enums.RiskLevelLowType),
						"enabled":     int(enums.ProfileStatusEnabled),
					},
				},
			},
		},
	}
}

// errExplode is the injected migration failure for rollback tests.
var errExplode = errors.New("transform exploded")

func TestMigrateConfigFilePreviewFailure(t *testing.T) { //nolint:paralleltest
	withMigrationChain(t, Migration{
		From:        FormatVersion{Major: 1, Minor: 0, Patch: 0},
		To:          FormatVersion{Major: 1, Minor: 1, Patch: 0},
		Description: "explode in preview",
		Apply: func(*types.Config) error {
			return errExplode
		},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	_, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err == nil {
		t.Fatal("MigrateConfigFile expected error, got nil")
	}

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection (preview failure touches nothing)", errorfamily.Classify(err))
	}
}

func TestMigrateConfigFileMissing(t *testing.T) { //nolint:paralleltest
	missing := filepath.Join(t.TempDir(), "absent.yaml")

	_, err := MigrateConfigFile(context.Background(), missing, MigrateOptions{})
	if err == nil {
		t.Fatal("MigrateConfigFile expected error for missing file, got nil")
	}

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection", errorfamily.Classify(err))
	}
}

func TestMigrateConfigFileAlreadyCurrent(t *testing.T) { //nolint:paralleltest
	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	report, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err != nil {
		t.Fatalf("MigrateConfigFile error = %v", err)
	}

	if !report.AlreadyCurrent {
		t.Error("AlreadyCurrent = false, want true")
	}

	if report.BackupPath != "" {
		t.Errorf("BackupPath = %q, want empty (nothing written)", report.BackupPath)
	}
}

func TestMigrateConfigFileEmptyVersionIsCurrent(t *testing.T) { //nolint:paralleltest
	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap(""))

	report, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err != nil {
		t.Fatalf("MigrateConfigFile error = %v", err)
	}

	if !report.AlreadyCurrent {
		t.Error("AlreadyCurrent = false, want true for a pre-versioning config")
	}
}

func TestMigrateConfigFileFromFuture(t *testing.T) { //nolint:paralleltest
	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("9.9.9"))

	_, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err == nil {
		t.Fatal("MigrateConfigFile expected error for future version, got nil")
	}

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection", errorfamily.Classify(err))
	}

	if !strings.Contains(err.Error(), "upgrade clean-wizard") {
		t.Errorf("error = %v, want guidance to upgrade the binary", err)
	}
}

func TestMigrateConfigFileNoMigrationPath(t *testing.T) { //nolint:paralleltest
	withMigrationChain(t) // current becomes 1.1.0, but no steps registered

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	_, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err == nil {
		t.Fatal("MigrateConfigFile expected error for missing chain, got nil")
	}

	if !strings.Contains(err.Error(), "no migration path") {
		t.Errorf("error = %v, want the missing-path explanation", err)
	}
}

func TestMigrateConfigFileSuccess(t *testing.T) { //nolint:paralleltest
	target := FormatVersion{Major: 1, Minor: 1, Patch: 0}
	withMigrationChain(t, Migration{
		From:        FormatVersion{Major: 1, Minor: 0, Patch: 0},
		To:          target,
		Description: "protect /home",
		Apply: func(config *types.Config) error {
			config.Protected = append(config.Protected, "/home")

			return nil
		},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	report, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err != nil {
		t.Fatalf("MigrateConfigFile error = %v", err)
	}

	if len(report.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(report.Records))
	}

	if len(report.Records[0].Changes) != 1 {
		t.Errorf("changes = %d, want 1 (the added protected path)", len(report.Records[0].Changes))
	}

	if report.BackupPath == "" {
		t.Error("BackupPath empty, want a backup beside the config")
	}

	if _, statErr := os.Stat(report.BackupPath); statErr != nil {
		t.Errorf("backup missing: %v", statErr)
	}

	if len(report.Findings) != 1 {
		t.Errorf("findings = %d, want 1", len(report.Findings))
	}

	migrated, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("reloading migrated config: %v", err)
	}

	if migrated.Version != "1.1.0" {
		t.Errorf("migrated version = %q, want 1.1.0", migrated.Version)
	}

	if len(migrated.Protected) != 2 {
		t.Errorf("migrated protected = %v, want 2 entries", migrated.Protected)
	}
}

func TestMigrateConfigFileAborted(t *testing.T) { //nolint:paralleltest
	withMigrationChain(t, Migration{
		From:        FormatVersion{Major: 1, Minor: 0, Patch: 0},
		To:          FormatVersion{Major: 1, Minor: 1, Patch: 0},
		Description: "protect /home",
		Apply: func(config *types.Config) error {
			config.Protected = append(config.Protected, "/home")

			return nil
		},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	decline := func(MigrationPreview) bool { return false }

	_, err := MigrateConfigFile(context.Background(), path, MigrateOptions{Confirm: decline})
	if !errors.Is(err, ErrMigrationAborted) {
		t.Fatalf("error = %v, want ErrMigrationAborted", err)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}

	if strings.Contains(string(data), "1.1.0") {
		t.Error("config was rewritten despite abort")
	}
}

func TestMigrateConfigFileRollbackOnFailure(t *testing.T) { //nolint:paralleltest
	calls := 0

	withMigrationChain(t, Migration{
		From:        FormatVersion{Major: 1, Minor: 0, Patch: 0},
		To:          FormatVersion{Major: 1, Minor: 1, Patch: 0},
		Description: "explode after preview",
		Apply: func(config *types.Config) error {
			calls++ // the preview run must succeed; only the real apply explodes
			if calls > 1 {
				return errExplode
			}

			config.Protected = append(config.Protected, "/home")

			return nil
		},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	_, err := MigrateConfigFile(context.Background(), path, MigrateOptions{})
	if err == nil {
		t.Fatal("MigrateConfigFile expected error, got nil")
	}

	if errorfamily.Classify(err) != errorfamily.Corruption {
		t.Errorf("family = %v, want Corruption", errorfamily.Classify(err))
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}

	if strings.Contains(string(data), "1.1.0") {
		t.Error("config not restored after failed migration")
	}
}

func TestCheckConfigVersionGatesLoad(t *testing.T) { //nolint:paralleltest
	withMigrationChain(t) // current becomes 1.1.0

	dir := t.TempDir()
	path := filepath.Join(dir, ".clean-wizard.yaml")
	writeTestConfigFile(t, path, simpleConfigFileMap("1.0.0"))

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("LoadFromPath expected error for outdated config, got nil")
	}

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Errorf("family = %v, want Rejection", errorfamily.Classify(err))
	}

	if !strings.Contains(err.Error(), "clean-wizard config migrate") {
		t.Errorf("error = %v, want a pointer to 'clean-wizard config migrate'", err)
	}
}
