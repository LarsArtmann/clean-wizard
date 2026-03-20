package commands

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewInitCommand creates an interactive setup wizard command.
func NewInitCommand() *cobra.Command {
	var (
		force   bool
		minimal bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long:  `Interactive setup wizard that creates a comprehensive cleaning configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCommand(force, minimal)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&minimal, "minimal", false, "Create minimal configuration")

	return cmd
}

// runInitCommand executes the init command.
func runInitCommand(force, minimal bool) error {
	fmt.Println(TitleStyle.Render("🧹 Clean Wizard Setup"))
	fmt.Println()

	// Check if config already exists
	cfg, err := config.Load()
	if err == nil && !force {
		fmt.Println("⚠️  Configuration already exists!")
		fmt.Printf("   Current profiles: ")

		for name := range cfg.Profiles {
			fmt.Printf("%s ", name)
		}

		fmt.Println()
		fmt.Println()

		var overwrite bool

		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Overwrite existing configuration?").
					Description("This will replace your current configuration").
					Affirmative("Yes, overwrite").
					Negative("No, cancel").
					Value(&overwrite),
			),
		)

		err := confirmForm.Run()
		if err != nil {
			return fmt.Errorf("confirmation error: %w", err)
		}

		if !overwrite {
			fmt.Println("❌ Cancelled. No changes made.")

			return nil
		}
	}

	if minimal {
		return createMinimalConfig()
	}

	return createInteractiveConfig()
}

// createMinimalConfig creates a minimal configuration.
func createMinimalConfig() error {
	fmt.Println(InfoStyle.Render("Creating minimal configuration..."))

	cfg := config.GetDefaultConfig()
	// Keep only the daily profile for minimal config
	daily := cfg.Profiles["daily"]
	cfg.Profiles = map[string]*domain.Profile{
		"daily": daily,
	}

	err := config.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println(SuccessStyle.Render("✅ Minimal configuration created successfully!"))
	fmt.Println()
	fmt.Println(InfoStyle.Render("📁 Configuration saved to: ~/.clean-wizard.yaml"))
	fmt.Println()
	fmt.Println("💡 To get started:")
	fmt.Println("   clean-wizard clean              - Run daily cleanup")
	fmt.Println("   clean-wizard scan               - Scan for cleanable items")
	fmt.Println("   clean-wizard profile list       - List available profiles")
	fmt.Println("   clean-wizard config show        - View current configuration")

	return nil
}

// createInteractiveConfig creates a configuration interactively using huh forms.
func createInteractiveConfig() error {
	fmt.Println(
		InfoStyle.Render("Let's create the perfect cleaning configuration for your system!"),
	)
	fmt.Println()

	// Interactive form for configuration options
	var (
		setupMode            string
		includeNix           bool
		includeHomebrew      bool
		includeDocker        bool
		includeNode          bool
		includeGo            bool
		includeDockerWarning bool
	)

	// Select setup mode
	modeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How would you like to configure Clean Wizard?").
				Options(
					huh.NewOption("🎯 Quick Setup (Recommended)", "quick"),
					huh.NewOption("⚙️  Custom Setup (Choose cleaners)", "custom"),
					huh.NewOption("📦 Full Setup (All profiles)", "full"),
				).
				Value(&setupMode),
		),
	)

	if err := modeForm.Run(); err != nil {
		return fmt.Errorf("setup mode selection error: %w", err)
	}

	// If custom mode, let user select cleaners
	if setupMode == "custom" {
		cleanerForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Include Nix cleaner?").
					Description("Clean old Nix store generations").
					Value(&includeNix),
				huh.NewConfirm().
					Title("Include Homebrew cleaner?").
					Description("Clean Homebrew cache and unused packages").
					Value(&includeHomebrew),
				huh.NewConfirm().
					Title("Include Docker cleaner?").
					Description("Clean Docker images, containers, and volumes").
					Value(&includeDocker),
				huh.NewConfirm().
					Title("Include Node.js cleaner?").
					Description("Clean npm, pnpm, yarn, bun caches").
					Value(&includeNode),
				huh.NewConfirm().
					Title("Include Go cleaner?").
					Description("Clean Go module and build caches").
					Value(&includeGo),
			),
		)

		err := cleanerForm.Run()
		if err != nil {
			return fmt.Errorf("cleaner selection error: %w", err)
		}

		// Warn about Docker if selected
		if includeDocker {
			warningForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("⚠️  Docker cleaner warning").
						Description("This will remove unused Docker images and volumes. Continue?").
						Affirmative("Yes, I understand").
						Negative("No, skip Docker").
						Value(&includeDockerWarning),
				),
			)

			err := warningForm.Run()
			if err != nil {
				return fmt.Errorf("docker warning error: %w", err)
			}

			includeDocker = includeDockerWarning
		}
	}

	// Create configuration based on selections
	cfg := config.GetDefaultConfig()

	switch setupMode {
	case "quick":
		// Quick setup: daily profile only with safe cleaners
		cfg.Profiles = map[string]*domain.Profile{
			"daily": createDailyProfile(),
		}
	case "custom":
		cfg.Profiles = createCustomProfile(
			includeNix,
			includeHomebrew,
			includeDocker,
			includeNode,
			includeGo,
		)
	case "full":
		// Full setup: all profiles
		cfg.Profiles = map[string]*domain.Profile{
			"daily":      createDailyProfile(),
			"weekly":     createWeeklyProfile(),
			"aggressive": createAggressiveProfile(),
		}
	}

	// Ask about safe mode
	safeMode := true

	safeModeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable safe mode?").
				Description("Safe mode prevents destructive operations and adds confirmation prompts").
				Affirmative("Yes, enable safe mode (Recommended)").
				Negative("No, disable safe mode").
				Value(&safeMode),
		),
	)

	if err := safeModeForm.Run(); err != nil {
		return fmt.Errorf("safe mode selection error: %w", err)
	}

	cfg.SafeMode = domain.SafeModeEnabled
	if !safeMode {
		cfg.SafeMode = domain.SafeModeDisabled
	}

	// Save configuration
	err := config.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show success message
	fmt.Println()
	fmt.Println(SuccessStyle.Render("✅ Configuration created successfully!"))
	fmt.Println()
	fmt.Println(InfoStyle.Render("📁 Configuration saved to: ~/.clean-wizard.yaml"))
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

// createDailyProfile creates the daily cleanup profile.
func createDailyProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "daily",
		Description: "Quick daily cleanup with safe operations",
		Enabled:     domain.ProfileStatusEnabled,
		Operations: []domain.CleanupOperation{
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
			},
			{
				Name:        "go-packages",
				Description: "Clean Go module cache",
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
}

// createWeeklyProfile creates the weekly cleanup profile.
func createWeeklyProfile() *domain.Profile {
	return &domain.Profile{
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
			{
				Name:        "homebrew-cleanup",
				Description: "Clean Homebrew cache",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
			},
		},
	}
}

// createAggressiveProfile creates the aggressive cleanup profile.
func createAggressiveProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "aggressive",
		Description: "Deep aggressive cleanup (may remove useful items)",
		Enabled:     domain.ProfileStatusDisabled,
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
			{
				Name:        "docker",
				Description: "Clean all unused Docker resources",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelHighType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeDocker),
			},
		},
	}
}

// createCustomProfile creates a custom profile based on user selections.
func createCustomProfile(
	includeNix, includeHomebrew, includeDocker, includeNode, includeGo bool,
) map[string]*domain.Profile {
	operations := make([]domain.CleanupOperation, 0)

	if includeNix {
		operations = append(operations, domain.CleanupOperation{
			Name:        "nix-generations",
			Description: "Clean old Nix generations",
			RiskLevel:   domain.RiskLevelType(domain.RiskLevelMediumType),
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
		})
	}

	if includeHomebrew {
		operations = append(operations, domain.CleanupOperation{
			Name:        "homebrew-cleanup",
			Description: "Clean Homebrew cache and unused packages",
			RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
		})
	}

	if includeDocker {
		operations = append(operations, domain.CleanupOperation{
			Name:        "docker",
			Description: "Clean Docker images, containers, and volumes",
			RiskLevel:   domain.RiskLevelType(domain.RiskLevelMediumType),
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeDocker),
		})
	}

	if includeNode {
		operations = append(operations, domain.CleanupOperation{
			Name:        "node-packages",
			Description: "Clean Node.js package caches",
			RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeNodePackages),
		})
	}

	if includeGo {
		operations = append(operations, domain.CleanupOperation{
			Name:        "go-packages",
			Description: "Clean Go module and build caches",
			RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeGoPackages),
		})
	}

	// Always include temp files
	operations = append(operations, domain.CleanupOperation{
		Name:        "temp-files",
		Description: "Clean temporary files",
		RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
		Enabled:     domain.ProfileStatusEnabled,
		Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
	})

	return map[string]*domain.Profile{
		"custom": {
			Name:        "custom",
			Description: "Custom cleanup profile",
			Enabled:     domain.ProfileStatusEnabled,
			Operations:  operations,
		},
	}
}

// GetDefaultConfig returns the default configuration.
// This is a wrapper around config.GetDefaultConfig for use in commands.
func GetDefaultConfig() *domain.Config {
	return config.GetDefaultConfig()
}

// init ensures we check for terminal support.
func init() {
	// Check if we're in an interactive terminal
	if fi, err := os.Stdin.Stat(); err == nil {
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// Non-interactive mode, use minimal config by default
		}
	}
}
