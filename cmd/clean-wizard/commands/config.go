package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates a configuration management command.
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `Manage configuration files - show, edit, validate, and reset.`,
	}

	cmd.AddCommand(NewConfigShowCommand())
	cmd.AddCommand(NewConfigEditCommand())
	cmd.AddCommand(NewConfigValidateCommand())
	cmd.AddCommand(NewConfigResetCommand())

	return cmd
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
func runConfigShowCommand(cmd *cobra.Command, args []string, jsonOutput bool) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ No configuration found.")
		fmt.Println()
		fmt.Println("💡 To create configuration, run:")
		fmt.Println("   clean-wizard init           - Interactive setup")
		fmt.Println("   clean-wizard init --minimal  - Minimal setup")
		return nil
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
	for name, profile := range cfg.Profiles {
		status := "enabled"
		if profile.Enabled == domain.ProfileStatusDisabled {
			status = "disabled"
		}
		fmt.Printf("  📁 %s (%s)\n", name, status)
		fmt.Printf("     %s\n", profile.Description)
		fmt.Printf("     Operations: %d\n", len(profile.Operations))
		fmt.Println()
	}

	return nil
}

// showConfigJSON outputs config in JSON format.
func showConfigJSON(cfg *domain.Config) error {
	// Simple JSON output - could be enhanced with proper JSON marshaling
	fmt.Println("{")
	fmt.Printf("  \"version\": \"%s\",\n", cfg.Version)
	fmt.Printf("  \"safe_mode\": \"%s\",\n", cfg.SafeMode.String())
	fmt.Printf("  \"max_disk_usage\": %d,\n", cfg.MaxDiskUsage)
	fmt.Printf("  \"protected\": [")
	for i, path := range cfg.Protected {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("\"%s\"", path)
	}
	fmt.Println("],")
	fmt.Printf("  \"profiles\": {")
	fmt.Println("}")
	return nil
}

// NewConfigEditCommand creates a command to edit the configuration.
func NewConfigEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration",
		Long:  `Open configuration file in the default editor.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigEditCommand(cmd, args)
		},
	}
	return cmd
}

// runConfigEditCommand executes the config edit command.
func runConfigEditCommand(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	// Check if config exists, create if not
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Creating new configuration...")
		cfg = config.GetDefaultConfig()
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to create configuration: %w", err)
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
	editCmd := exec.Command(editor, configPath)
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
func runConfigValidateCommand(cmd *cobra.Command, args []string, configPath string) error {
	fmt.Println("🔍 Validating configuration...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ Configuration validation failed!")
		fmt.Printf("   Error: %v\n", err)
		fmt.Println()
		fmt.Println("💡 To create a valid configuration:")
		fmt.Println("   clean-wizard init --minimal")
		return err
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
func runConfigResetCommand(cmd *cobra.Command, args []string, force bool) error {
	configPath := getConfigPath()

	// Check if config exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		fmt.Println("ℹ️  No configuration file exists.")
		fmt.Println("   Creating default configuration...")
	} else if !force {
		fmt.Printf("Reset configuration to defaults? This will overwrite %s\n", configPath)
		fmt.Print("Type 'yes' to confirm: ")
		var confirm string
		_, err = fmt.Scanln(&confirm)
		if err != nil || confirm != "yes" {
			fmt.Println("   Reset cancelled.")
			return nil
		}
	}

	// Create default configuration
	cfg := config.GetDefaultConfig()

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
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
