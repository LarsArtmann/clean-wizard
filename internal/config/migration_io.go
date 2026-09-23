package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/logger"
	"github.com/knadh/koanf/parsers/yaml"
	atomicwrite "github.com/larsartmann/go-atomic-write"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/linter-autoconfigure-sdk"
	yamlv3 "gopkg.in/yaml.v3"
)

// configFilePermissionFallback is used when the config file's own permissions
// cannot be read before the atomic rewrite.
const configFilePermissionFallback = ConfigFilePermission

// ErrMigrationAborted is returned by MigrateConfigFile when the user (or the
// Confirm callback) declines the migration. The configuration is untouched.
var ErrMigrationAborted = errors.New("configuration migration aborted")

// MigrationPreview is the dry-run result shown before a migration is applied:
// the planned steps, the per-step change records, and the flattened diff.
type MigrationPreview struct {
	Plan    []Migration
	Records []MigrationRecord
	Diff    string
}

// MigrationReport describes a completed migration.
type MigrationReport struct {
	Records        []MigrationRecord
	Diff           string
	BackupPath     string
	AlreadyCurrent bool
	Findings       []finding.Finding
}

// MigrateOptions controls MigrateConfigFile.
type MigrateOptions struct {
	// Confirm receives the dry-run preview and decides whether to proceed.
	// nil proceeds without asking.
	Confirm func(MigrationPreview) bool

	// BackupDir overrides the default backup directory
	// (<config dir>/.clean-wizard-backups).
	BackupDir string
}

// DefaultBackupDir returns the default backup directory for a config path.
func DefaultBackupDir(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), ".clean-wizard-backups")
}

// MigrateConfigFile migrates the configuration file at configPath to
// CurrentFormatVersion: dry-run the chain, ask for confirmation, back up the
// file, apply, validate, and atomically rewrite it. Any failure after the
// backup restores the original file.
func MigrateConfigFile(
	ctx context.Context,
	configPath string,
	opts MigrateOptions,
) (MigrationReport, error) {
	if err := ctx.Err(); err != nil {
		return MigrationReport{}, err
	}

	if _, err := os.Stat(configPath); err != nil {
		return MigrationReport{}, errorfamily.WrapRejection(err, "config.migrate", "configuration file not found: "+configPath)
	}

	config, err := loadConfigForMigration(ctx, configPath)
	if err != nil {
		return MigrationReport{}, err
	}

	from, err := NormalizeFormatVersion(config.Version)
	if err != nil {
		return MigrationReport{}, err
	}

	switch {
	case from.Compare(CurrentFormatVersion) > 0:
		return MigrationReport{}, errorfamily.NewRejection(
			"config.migrate",
			"configuration format "+from.String()+" is newer than this binary supports ("+CurrentFormatVersion.String()+"); upgrade clean-wizard",
		)
	case from.Compare(CurrentFormatVersion) == 0:
		return MigrationReport{AlreadyCurrent: true}, nil
	}

	plan, err := PlanMigration(from)
	if err != nil {
		return MigrationReport{}, err
	}

	preview, err := previewMigration(config, plan)
	if err != nil {
		return MigrationReport{}, err
	}

	if opts.Confirm != nil && !opts.Confirm(preview) {
		return MigrationReport{}, ErrMigrationAborted
	}

	return applyMigration(configPath, config, plan, preview, opts)
}

// CheckConfigVersion gates the load path: configurations from a future format
// or an outdated format are rejected with actionable guidance instead of being
// silently rewritten behind read-only commands.
func CheckConfigVersion(config *types.Config) error {
	version, err := NormalizeFormatVersion(config.Version)
	if err != nil {
		return err
	}

	switch {
	case version.Compare(CurrentFormatVersion) > 0:
		return errorfamily.NewRejection(
			"config.load",
			"configuration format "+version.String()+" is newer than this binary supports ("+CurrentFormatVersion.String()+"); upgrade clean-wizard",
		)
	case version.Compare(CurrentFormatVersion) < 0:
		return errorfamily.NewRejection(
			"config.load",
			"configuration format "+version.String()+" is outdated (current: "+CurrentFormatVersion.String()+"); run 'clean-wizard config migrate' to upgrade it in place",
		)
	}

	return nil
}

func loadConfigForMigration(ctx context.Context, configPath string) (*types.Config, error) {
	k := setupKoanf()

	// A non-error return means the file is missing and defaults were
	// substituted — there is nothing on disk to migrate.
	if _, err := readConfigFileFromPath(ctx, k, configPath); err != nil {
		if errors.Is(err, ErrConfigShouldUnmarshal) {
			return parseConfig(k)
		}

		return nil, err
	}

	return nil, errorfamily.NewRejection("config.migrate", "configuration file not found: "+configPath)
}

func previewMigration(config *types.Config, plan []Migration) (MigrationPreview, error) {
	clone, err := cloneConfig(config)
	if err != nil {
		return MigrationPreview{}, err
	}

	records, err := ApplyMigrations(clone, plan)
	if err != nil {
		return MigrationPreview{}, errorfamily.WrapRejection(err, "config.migrate", "migration preview failed")
	}

	return MigrationPreview{
		Plan:    plan,
		Records: records,
		Diff:    autoconfigure.FormatDiff(recordChanges(records)),
	}, nil
}

func applyMigration(
	configPath string,
	config *types.Config,
	plan []Migration,
	preview MigrationPreview,
	opts MigrateOptions,
) (MigrationReport, error) {
	backupDir := opts.BackupDir
	if backupDir == "" {
		backupDir = DefaultBackupDir(configPath)
	}

	backup := pipeline.NewFileBackup(backupDir)

	if err := backup.Backup(configPath); err != nil {
		return MigrationReport{}, errorfamily.WrapTransient(err, "config.migrate", "failed to back up configuration before migration")
	}

	restore := func(cause error, message string) (MigrationReport, error) {
		if restoreErr := backup.Restore(configPath); restoreErr != nil {
			return MigrationReport{}, errorfamily.WrapCorruption(
				errors.Join(cause, restoreErr),
				"config.migrate",
				message+"; the backup itself could not be restored — original preserved at "+backup.BackupPath(configPath),
			)
		}

		return MigrationReport{}, errorfamily.WrapCorruption(cause, "config.migrate", message)
	}

	if _, err := ApplyMigrations(config, plan); err != nil {
		return restore(err, "migration failed; original configuration restored from backup")
	}

	if err := validateLoadedConfig(config); err != nil {
		return restore(err, "migrated configuration failed validation; original configuration restored from backup")
	}

	permission := os.FileMode(configFilePermissionFallback)
	if info, statErr := os.Stat(configPath); statErr == nil {
		permission = info.Mode().Perm()
	}

	yamlData, err := yaml.Parser().Marshal(configYAMLMap(config))
	if err != nil {
		return restore(err, "failed to encode migrated configuration; original configuration restored from backup")
	}

	if err := atomicwrite.WriteWithPerm(configPath, yamlData, permission); err != nil {
		return restore(err, "failed to write migrated configuration; original configuration restored from backup")
	}

	report := MigrationReport{
		Records:    preview.Records,
		Diff:       preview.Diff,
		BackupPath: backup.BackupPath(configPath),
	}

	report.Findings = migrationFindings(configPath, report)

	logger.Info("Configuration migrated",
		"from", plan[0].From.String(),
		"to", CurrentFormatVersion.String(),
		"backup", report.BackupPath)

	return report, nil
}

// recordChanges flattens the per-step change lists into one deterministic slice.
func recordChanges(records []MigrationRecord) []autoconfigure.Change {
	changes := []autoconfigure.Change{}

	for _, record := range records {
		changes = append(changes, record.Changes...)
	}

	return changes
}

// cloneConfig deep-copies the configuration through its on-disk YAML shape so
// the dry-run cannot mutate the original.
func cloneConfig(config *types.Config) (*types.Config, error) {
	data, err := yamlv3.Marshal(configYAMLMap(config))
	if err != nil {
		return nil, errorfamily.WrapCorruption(err, "config.migrate", "failed to encode configuration for migration preview")
	}

	clone := &types.Config{}
	if err := yamlv3.Unmarshal(data, clone); err != nil {
		return nil, errorfamily.WrapCorruption(err, "config.migrate", "failed to decode configuration for migration preview")
	}

	return clone, nil
}

// migrationFindings renders a completed migration as go-finding findings
// (category "migration", one per applied step) for structured/SARIF output.
func migrationFindings(configPath string, report MigrationReport) []finding.Finding {
	findings := make([]finding.Finding, 0, len(report.Records))

	for _, record := range report.Records {
		built, err := finding.NewBuilder(
			finding.RuleName("config-migration"),
			finding.ToolName("clean-wizard"),
			fmt.Sprintf("configuration format migrated %s: %s (%d changes)",
				record.Step(), record.Description, len(record.Changes)),
			finding.SeverityInfo,
			finding.FilePos(finding.FilePath(configPath)),
		).
			WithCategory(finding.CategoryMigration).
			WithConfidence(finding.ConfidenceFull).
			WithFixStrategy(finding.FixStrategySuggest).
			WithSuggestion("review the migration diff; restore the backup at " + report.BackupPath + " to revert").
			Build()
		if err != nil {
			logger.Error("Failed to build migration finding", "error", err, "step", record.Step())

			continue
		}

		findings = append(findings, built)
	}

	return findings
}
