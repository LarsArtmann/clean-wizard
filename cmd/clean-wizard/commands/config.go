package commands

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates a configuration management command.
func NewConfigCommand() *cobra.Command {
	return newParentCommand(
		"config",
		"Manage configuration",
		"Manage configuration files - show, edit, validate, migrate, and reset.",
		NewConfigShowCommand,
		NewConfigEditCommand,
		NewConfigValidateCommand,
		NewConfigMigrateCommand,
		NewConfigResetCommand,
	)
}

// NewConfigMigrateCommand creates a command to migrate the configuration to
// the current format version.
func NewConfigMigrateCommand() *cobra.Command {
	var (
		assumeYes bool
		backupDir string
		sarifOut  string
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration to the current format version",
		Long: `Bring an outdated configuration file up to the current format version.

The migration runs as a dry-run first and shows every change before asking
for confirmation. The original file is backed up and restored automatically
if the migrated configuration fails validation or cannot be written.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigMigrateCommand(assumeYes, backupDir, sarifOut)
		},
	}

	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip the confirmation prompt")
	cmd.Flags().StringVar(&backupDir, "backup-dir", "",
		"Directory for the pre-migration backup (default: <config dir>/.clean-wizard-backups)")
	cmd.Flags().StringVar(&sarifOut, "sarif", "",
		"Write migration findings as SARIF 2.1.0 to this file ('-' for stdout)")

	return cmd
}

// runConfigMigrateCommand executes the config migrate command.
func runConfigMigrateCommand(assumeYes bool, backupDir, sarifOut string) error {
	configPath := getConfigPath()

	var confirm func(config.MigrationPreview) bool
	if !assumeYes {
		confirm = func(preview config.MigrationPreview) bool {
			fmt.Println("\nMigration plan:")

			for _, step := range preview.Plan {
				fmt.Printf("  %s: %s\n", step.Step(), step.Description)
			}

			if preview.Diff != "" {
				fmt.Printf("\nChanges:\n%s\n", preview.Diff)
			}

			fmt.Printf("\nThis rewrites %s. A backup is created first.\n", configPath)

			return promptForConfirmation("Migrate configuration now?", "Migration cancelled.")
		}
	}

	report, err := config.MigrateConfigFile(context.Background(), configPath, config.MigrateOptions{
		Confirm:   confirm,
		BackupDir: backupDir,
	})
	if errors.Is(err, config.ErrMigrationAborted) {
		fmt.Println("ℹ️  Configuration unchanged.")

		return nil
	}

	if err != nil {
		// Engine errors arrive pre-classified; this wrap adds context only so
		// the original family survives to the CLI boundary.
		return fmt.Errorf("config migrate: %w", err)
	}

	if report.AlreadyCurrent {
		fmt.Printf("✅ Configuration is already at format version %s; nothing to do.\n",
			config.CurrentFormatVersion.String())

		return nil
	}

	for _, record := range report.Records {
		fmt.Printf("✅ %s: %s (%d changes)\n", record.Step(), record.Description, len(record.Changes))
	}

	if report.Diff != "" {
		fmt.Printf("\n%s\n", report.Diff)
	}

	fmt.Printf("Backup: %s\n", report.BackupPath)

	return writeMigrationSarif(sarifOut, report)
}

// writeMigrationSarif emits the migration report as SARIF when requested.
func writeMigrationSarif(sarifOut string, report config.MigrationReport) error {
	if sarifOut == "" {
		return nil
	}

	data, err := format.FindingsToSARIF(report.Findings)
	if err != nil {
		return fmt.Errorf("config migrate: %w", err)
	}

	if sarifOut == "-" {
		fmt.Println(string(data))

		return nil
	}

	if writeErr := os.WriteFile(sarifOut, data, 0o600); writeErr != nil {
		return errorfamily.WrapRejection(writeErr, "config.sarif_output", "failed to write SARIF output: "+sarifOut)
	}

	fmt.Printf("SARIF report: %s\n", sarifOut)

	return nil
}

// NewConfigShowCommand creates a command to show the configuration.
func NewConfigShowCommand() *cobra.Command {
	var json bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current configuration file contents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShowCommand(cmd, args, json)
		},
	}

	cmd.Flags().BoolVar(&json, "json", false, "Output in JSON format")

	return cmd
}

// runConfigShowCommand executes the config show command.
func runConfigShowCommand(_ *cobra.Command, _ []string, jsonOutput bool) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ No configuration found.")
		fmt.Println()
		fmt.Println("💡 To create configuration, run:")
		fmt.Println("   clean-wizard init           - Interactive setup")
		fmt.Println("   clean-wizard init --minimal  - Minimal setup")

		return nil //nolint:nilerr // intentional: missing config shows help, not error
	}

	if jsonOutput {
		return showConfigJSON(cfg)
	}

	fmt.Println("📄 Current Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("Version: %s\n", cfg.Version)
	fmt.Printf("Safe Mode: %s\n", cfg.SafeMode.String())
	fmt.Printf("Max Disk Usage: %d%%\n", cfg.MaxDiskUsage)
	fmt.Printf("Protected Paths: %d\n", len(cfg.Protected))
	fmt.Println()
	fmt.Println("Protected Paths:")

	for _, path := range cfg.Protected {
		fmt.Printf("   • %s\n", path)
	}

	fmt.Println()
	fmt.Printf("Profiles: %d\n", len(cfg.Profiles))
	fmt.Println()

	// Show profiles summary
	PrintProfileSummaries(cfg.Profiles, ProfileFormatText)

	return nil
}

// showConfigJSON outputs config in JSON format.
func showConfigJSON(cfg *types.Config) error {
	// Use proper JSON marshaling for complete output
	jsonBytes, err := json.Marshal(cfg, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return errorfamily.WrapCorruption(err, "config.json_marshal", "failed to marshal configuration to JSON")
	}

	fmt.Println(string(jsonBytes))

	return nil
}

// NewConfigEditCommand creates a command to edit the configuration.
func NewConfigEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration",
		Long:  `Open configuration file in the default editor.`,
		RunE:  runConfigEditCommand,
	}

	return cmd
}

// runConfigEditCommand executes the config edit command.
func runConfigEditCommand(_ *cobra.Command, _ []string) error {
	configPath := getConfigPath()

	// Check if config exists, create if not
	_, err := config.Load()
	if err != nil {
		fmt.Println("Creating new configuration...")

		cfg := config.GetDefaultConfig()

		err := config.Save(cfg)
		if err != nil {
			return errorfamily.WrapRejection(err, "config.create", "failed to create configuration")
		}

		fmt.Println("✅ Configuration created.")
		fmt.Println()
	}

	// Get editor
	editor := getEditor()
	if editor == "" {
		fmt.Println("❌ No editor found.")
		fmt.Println("   Set the EDITOR environment variable or configure a default editor.")
		fmt.Println()
		fmt.Println("💡 To edit manually:")
		fmt.Printf("   %s %s\n", getPreferredEditor(), configPath)

		return nil
	}

	fmt.Printf("Opening configuration in %s...\n", editor)
	fmt.Println()

	// Open editor
	editCmd := exec.CommandContext(context.Background(), editor, configPath)
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	editCmd.Stdin = os.Stdin

	if err := editCmd.Run(); err != nil {
		fmt.Printf("Editor exited with error: %v\n", err)
		fmt.Println()
		fmt.Println("💡 To validate the edited configuration:")
		fmt.Println("   clean-wizard config validate")
	}

	return nil
}

// NewConfigValidateCommand creates a command to validate the configuration.
func NewConfigValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  `Validate the configuration file syntax and values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigValidateCommand(cmd, args, configPath)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")

	return cmd
}

// runConfigValidateCommand executes the config validate command.
func runConfigValidateCommand(_ *cobra.Command, _ []string, _ string) error {
	fmt.Println("🔍 Validating configuration...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ Configuration validation failed!")
		fmt.Printf("   Error: %v\n", err)
		fmt.Println()
		fmt.Println("💡 To create a valid configuration:")
		fmt.Println("   clean-wizard init --minimal")

		return errorfamily.WrapRejection(err, "config.load", "configuration validation could not load the config")
	}

	fmt.Println("✅ Configuration is valid!")
	fmt.Println()
	fmt.Printf("  Version: %s\n", cfg.Version)
	fmt.Printf("  Profiles: %d\n", len(cfg.Profiles))
	fmt.Printf("  Safe Mode: %s\n", cfg.SafeMode.String())

	return nil
}

// NewConfigResetCommand creates a command to reset the configuration.
func NewConfigResetCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  `Reset configuration file to default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigResetCommand(cmd, args, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Reset without confirmation")

	return cmd
}

// runConfigResetCommand executes the config reset command.
func runConfigResetCommand(_ *cobra.Command, _ []string, force bool) error {
	configPath := getConfigPath()

	// Check if config exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		fmt.Println("ℹ️  No configuration file exists.")
		fmt.Println("   Creating default configuration...")
	} else if !force {
		if !promptForConfirmation(
			"Reset configuration to defaults? This will overwrite "+configPath,
			"Reset cancelled.",
		) {
			return nil
		}
	}

	// Create default configuration
	cfg := config.GetDefaultConfig()

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return errorfamily.WrapRejection(err, "config.save", "failed to save configuration")
	}

	fmt.Println("✅ Configuration reset to defaults!")
	fmt.Println()
	fmt.Println("📁 Configuration saved to:", configPath)
	fmt.Println()
	fmt.Println("💡 To customize:")
	fmt.Println("   clean-wizard config edit")
	fmt.Println("   clean-wizard profile create --name custom --description 'My custom profile'")

	return nil
}

// loadConfigFromPath loads configuration from an explicit path, falling back
// to the default discovery path when configPath is empty.
func loadConfigFromPath(configPath string) (*types.Config, error) {
	if configPath != "" {
		cfg, err := config.LoadFromPath(configPath)
		if err != nil {
			return nil, errorfamily.WrapRejectionf(
				err,
				"config.load_from_path",
				"failed to load config from %s",
				configPath,
			)
		}

		return cfg, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, errorfamily.WrapRejection(err, "config.load", "failed to load config")
	}

	return cfg, nil
}

// promptForConfirmation prompts the user to confirm an action by typing 'yes'.
// Returns true if confirmed, false if cancelled.
func promptForConfirmation(promptMsg, cancelMsg string) bool {
	fmt.Printf("%s\n", promptMsg)
	fmt.Print("Type 'yes' to confirm: ")

	var confirm string

	_, err := fmt.Scanln(&confirm)
	if err != nil || confirm != "yes" {
		fmt.Printf("   %s\n", cancelMsg)

		return false
	}

	return true
}

// getConfigPath returns the path to the configuration file.
func getConfigPath() string {
	return os.Getenv("HOME") + "/.clean-wizard.yaml"
}

// getEditor returns the editor to use for editing configuration.
func getEditor() string {
	// Check environment variables
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}

	editor = os.Getenv("VISUAL")
	if editor != "" {
		return editor
	}

	// Check for common editors
	editors := []string{"vim", "vi", "nano", "code", "vscode"}
	for _, e := range editors {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}

	return ""
}

// getPreferredEditor returns the preferred editor name for display.
func getPreferredEditor() string {
	return "nano" // Fallback suggestion
}
