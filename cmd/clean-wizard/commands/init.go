package commands

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
)

// NewInitCommand creates an interactive setup wizard command.
func NewInitCommand() *cobra.Command {
	var force bool
	var minimal bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long:  `Interactive setup wizard that creates a comprehensive cleaning configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCommand(cmd, args, force, minimal)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&minimal, "minimal", false, "Create minimal configuration")

	return cmd
}

// runInitCommand executes the init command.
func runInitCommand(cmd *cobra.Command, args []string, force, minimal bool) error {
	fmt.Println("🧹 Clean Wizard Setup")
	fmt.Println("======================")
	fmt.Println()

	// Check if config already exists
	cfg, err := config.Load()
	if err == nil && !force {
		fmt.Println("⚠️  Configuration already exists!")
		fmt.Println("   Use --force to overwrite or --minimal to create a basic config.")
		fmt.Printf("\n   Current profiles: ")
		for name := range cfg.Profiles {
			fmt.Printf("%s ", name)
		}
		fmt.Println()
		return nil
	}

	if minimal {
		return createMinimalConfig()
	}

	return createInteractiveConfig()
}

// createMinimalConfig creates a minimal configuration.
func createMinimalConfig() error {
	fmt.Println("Creating minimal configuration...")

	cfg := config.GetDefaultConfig()
	// Keep only the daily profile for minimal config
	daily := cfg.Profiles["daily"]
	cfg.Profiles = map[string]*domain.Profile{
		"daily": daily,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Minimal configuration created successfully!")
	fmt.Println()
	fmt.Println("📁 Configuration saved to: ~/.clean-wizard.yaml")
	fmt.Println()
	fmt.Println("💡 To get started:")
	fmt.Println("   clean-wizard clean              - Run daily cleanup")
	fmt.Println("   clean-wizard scan               - Scan for cleanable items")
	fmt.Println("   clean-wizard profile list       - List available profiles")
	fmt.Println("   clean-wizard config show        - View current configuration")

	return nil
}

// createInteractiveConfig creates a configuration interactively.
func createInteractiveConfig() error {
	fmt.Println("Let's create the perfect cleaning configuration for your system!")
	fmt.Println()

	// For now, just use default config with all profiles
	fmt.Println("Creating comprehensive configuration with all profiles...")
	fmt.Println()

	cfg := config.GetDefaultConfig()

	// Add comprehensive profiles
	cfg.Profiles["aggressive"] = &domain.Profile{
		Name:        "aggressive",
		Description: "Deep aggressive cleanup",
		Enabled:     domain.ProfileStatusEnabled,
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelHighType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean old Homebrew packages",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelMediumType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
			},
		},
	}

	cfg.Profiles["weekly"] = &domain.Profile{
		Name:        "weekly",
		Description: "Weekly comprehensive cleanup",
		Enabled:     domain.ProfileStatusEnabled,
		Operations: []domain.CleanupOperation{
			{
				Name:        "docker",
				Description: "Clean Docker images, containers, and volumes",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelMediumType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeDocker),
			},
			{
				Name:        "go-packages",
				Description: "Clean Go build cache",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeGoPackages),
			},
			{
				Name:        "node-packages",
				Description: "Clean Node.js package caches",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeNodePackages),
			},
		},
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Configuration created successfully!")
	fmt.Println()
	fmt.Println("📁 Configuration saved to: ~/.clean-wizard.yaml")
	fmt.Println()
	fmt.Println("📋 Available profiles:")
	for name, profile := range cfg.Profiles {
		status := "enabled"
		if profile.Enabled == domain.ProfileStatusDisabled {
			status = "disabled"
		}
		fmt.Printf("   • %s (%s)\n", name, status)
		fmt.Printf("     %s\n", profile.Description)
	}
	fmt.Println()
	fmt.Println("💡 To get started:")
	fmt.Println("   clean-wizard clean              - Run daily cleanup")
	fmt.Println("   clean-wizard scan               - Scan for cleanable items")
	fmt.Println("   clean-wizard profile list       - List available profiles")
	fmt.Println("   clean-wizard config show        - View current configuration")

	return nil
}
