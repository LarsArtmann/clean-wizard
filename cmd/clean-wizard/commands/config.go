package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates config command.
func NewConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration files",
		Long:  `Manage and manipulate configuration files for clean-wizard settings.`,
	}

	// Add subcommands
	configCmd.AddCommand(
		NewConfigShowCommand(),
		NewConfigEditCommand(),
		NewConfigValidateCommand(),
		NewConfigResetCommand(),
	)

	return configCmd
}

// NewConfigShowCommand creates config show command.
func NewConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long:  `Display the current configuration settings from ~/.clean-wizard.yaml`,
		Args:  cobra.NoArgs,
		RunE:  runConfigShow,
	}
}

// NewConfigEditCommand creates config edit command.
func NewConfigEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration in default editor",
		Long:  `Open the configuration file in your default text editor for manual editing.`,
		Args:  cobra.NoArgs,
		RunE:  runConfigEdit,
	}
}

// NewConfigValidateCommand creates config validate command.
func NewConfigValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  `Validate the configuration file syntax and structure for errors.`,
		Args:  cobra.NoArgs,
		RunE:  runConfigValidate,
	}
}

// NewConfigResetCommand creates config reset command.
func NewConfigResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  `Reset the configuration file to default settings, backup existing configuration first.`,
		Args:  cobra.NoArgs,
		RunE:  runConfigReset,
	}
}

// runConfigShow displays current configuration.
func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("📄 Configuration: %s\n", getConfigPath())
	fmt.Printf("🔤 Version: %s\n", cfg.Version)
	fmt.Printf("🛡️  Safe Mode: %v\n", cfg.SafeMode)
	fmt.Printf("💾 Max Disk Usage: %d%%\n", cfg.MaxDiskUsage)

	if cfg.CurrentProfile != "" {
		fmt.Printf("🏷️  Current Profile: %s\n", cfg.CurrentProfile)
	}

	fmt.Printf("🚫 Protected Paths: %d\n", len(cfg.Protected))
	if len(cfg.Protected) > 0 {
		fmt.Println("   Protected paths:")
		for _, path := range cfg.Protected {
			fmt.Printf("   • %s\n", path)
		}
	}

	fmt.Printf("📋 Profiles: %d\n", len(cfg.Profiles))
	if len(cfg.Profiles) > 0 {
		for name, profile := range cfg.Profiles {
			status := "✅"
			if !profile.Enabled.IsEnabled() {
				status = "❌"
			}
			fmt.Printf("   %s %s - %s (%d operations)\n", status, name, profile.Description, len(profile.Operations))
		}
	}

	return nil
}

// runConfigEdit opens config in default editor.
func runConfigEdit(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	// Check if editor is available
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try to find default editor
		if _, err := exec.LookPath("code"); err == nil {
			editor = "code"
		} else if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else {
			return errors.New("no editor found. Please set EDITOR environment variable")
		}
	}

	fmt.Printf("📝 Opening %s with %s...\n", configPath, editor)

	editCmd := exec.Command(editor, configPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	return editCmd.Run()
}

// runConfigValidate validates configuration file.
func runConfigValidate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Validating configuration...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Configuration is valid!\n")
	fmt.Printf("📋 Configuration has:\n")
	fmt.Printf("   • %d profiles\n", len(cfg.Profiles))
	fmt.Printf("   • %d protected paths\n", len(cfg.Protected))
	fmt.Printf("   • Safe mode: %v\n", cfg.SafeMode)
	fmt.Printf("   • Max disk usage: %d%%\n", cfg.MaxDiskUsage)

	return nil
}

// runConfigReset resets configuration to defaults.
func runConfigReset(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	// Backup existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup"
		fmt.Printf("💾 Creating backup at %s...\n", backupPath)

		err := copyFile(configPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		fmt.Printf("✅ Backup created successfully\n")
	}

	fmt.Println("🔄 Resetting configuration to defaults...")

	// Get default config
	defaultCfg := config.GetDefaultConfig()

	// Save default config
	if err := config.Save(defaultCfg); err != nil {
		return fmt.Errorf("failed to save default configuration: %w", err)
	}

	fmt.Printf("✅ Configuration reset to defaults\n")
	fmt.Printf("📝 Edit with: clean-wizard config edit\n")
	fmt.Printf("💡 Restore backup with: cp %s.backup %s\n", configPath, configPath)

	return nil
}

// getConfigPath returns the path to configuration file.
func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	return os.ExpandEnv("$HOME/.clean-wizard.yaml")
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0o644)
}
