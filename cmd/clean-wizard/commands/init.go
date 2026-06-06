package commands

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
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

// newConfirmForm creates a confirmation form with the given parameters.
// This consolidates the duplicate huh.NewForm + huh.NewConfirm pattern.
func newConfirmForm(title, description, affirmative, negative string, value *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative(affirmative).
				Negative(negative).
				Value(value),
		),
	)
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

		confirmForm := newConfirmForm(
			"Overwrite existing configuration?",
			"This will replace your current configuration",
			"Yes, overwrite",
			"No, cancel",
			&overwrite,
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
		"daily": daily, //nolint:goconst
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

const setupModeCustom = "custom"

// createInteractiveConfig creates a configuration interactively using huh forms.
func createInteractiveConfig() error {
	fmt.Println(
		InfoStyle.Render("Let's create the perfect cleaning configuration for your system!"),
	)
	fmt.Println()

	setupMode, err := selectSetupMode()
	if err != nil {
		return err
	}

	customOpts, err := maybeSelectCustomCleaners(setupMode)
	if err != nil {
		return err
	}

	cfg := buildConfigFromSetupMode(setupMode, customOpts)

	if err := configureSafeMode(cfg); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printConfigSuccess(cfg)

	return nil
}

type customCleanerOptions struct {
	includeNix      bool
	includeHomebrew bool
	includeDocker   bool
	includeNode     bool
	includeGo       bool
}

func selectSetupMode() (string, error) {
	var setupMode string

	modeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How would you like to configure Clean Wizard?").
				Options(
					huh.NewOption("🎯 Quick Setup (Recommended)", "quick"),
					huh.NewOption("⚙️  Custom Setup (Choose cleaners)", setupModeCustom),
					huh.NewOption("📦 Full Setup (All profiles)", "full"),
				).
				Value(&setupMode),
		),
	)

	if err := modeForm.Run(); err != nil {
		return "", fmt.Errorf("setup mode selection error for setupMode=%v: %w", setupMode, err)
	}

	return setupMode, nil
}

func maybeSelectCustomCleaners(setupMode string) (*customCleanerOptions, error) {
	if setupMode != setupModeCustom {
		return &customCleanerOptions{}, nil //nolint:exhaustruct
	}

	var opts customCleanerOptions

	cleanerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include Nix cleaner?").
				Description("Clean old Nix store generations").
				Value(&opts.includeNix),
			huh.NewConfirm().
				Title("Include Homebrew cleaner?").
				Description("Clean Homebrew cache and unused packages").
				Value(&opts.includeHomebrew),
			huh.NewConfirm().
				Title("Include Docker cleaner?").
				Description("Clean Docker images, containers, and volumes").
				Value(&opts.includeDocker),
			huh.NewConfirm().
				Title("Include Node.js cleaner?").
				Description("Clean npm, pnpm, yarn, bun caches").
				Value(&opts.includeNode),
			huh.NewConfirm().
				Title("Include Go cleaner?").
				Description("Clean Go module and build caches").
				Value(&opts.includeGo),
		),
	)

	if err := cleanerForm.Run(); err != nil {
		return nil, fmt.Errorf("cleaner selection error for setupMode=%v: %w", setupMode, err)
	}

	if opts.includeDocker {
		var includeDockerWarning bool

		err := newConfirmForm(
			"⚠️  Docker cleaner warning",
			"This will remove unused Docker images and volumes. Continue?",
			"Yes, I understand",
			"No, skip Docker",
			&includeDockerWarning,
		).Run()
		if err != nil {
			return nil, fmt.Errorf(
				"docker warning error for includeDockerWarning=%v: %w",
				includeDockerWarning,
				err,
			)
		}

		opts.includeDocker = includeDockerWarning
	}

	return &opts, nil
}

func buildConfigFromSetupMode(setupMode string, customOpts *customCleanerOptions) *domain.Config {
	cfg := config.GetDefaultConfig()

	switch setupMode {
	case "quick":
		cfg.Profiles = map[string]*domain.Profile{
			"daily": createDailyProfile(),
		}
	case setupModeCustom:
		cfg.Profiles = createCustomProfile(
			customOpts.includeNix,
			customOpts.includeHomebrew,
			customOpts.includeDocker,
			customOpts.includeNode,
			customOpts.includeGo,
		)
	case "full":
		cfg.Profiles = map[string]*domain.Profile{
			"daily":      createDailyProfile(),
			"weekly":     createWeeklyProfile(),
			"aggressive": createAggressiveProfile(), //nolint:goconst
		}
	}

	return cfg
}

func configureSafeMode(cfg *domain.Config) error {
	safeMode := true

	if err := newConfirmForm(
		"Enable safe mode?",
		"Safe mode prevents destructive operations and adds confirmation prompts",
		"Yes, enable safe mode (Recommended)",
		"No, disable safe mode",
		&safeMode,
	).Run(); err != nil {
		return fmt.Errorf("safe mode selection error: %w", err)
	}

	cfg.SafeMode = domain.SafeModeEnabled
	if !safeMode {
		cfg.SafeMode = domain.SafeModeDisabled
	}

	return nil
}

func printConfigSuccess(cfg *domain.Config) {
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
}

// createDailyProfile creates the daily cleanup profile.
func createDailyProfile() *domain.Profile {
	return &domain.Profile{
		Name:        "daily",
		Description: "Quick daily cleanup with safe operations",
		Enabled:     domain.ProfileStatusEnabled,
		Operations: []domain.CleanupOperation{
			createLowRiskOperation(
				"temp-files",
				"Clean temporary files",
				domain.OperationTypeTempFiles,
			),
			createLowRiskOperation(
				"go-packages",
				"Clean Go module cache",
				domain.OperationTypeGoPackages,
			),
			createLowRiskOperation(
				"node-packages",
				"Clean Node.js package caches",
				domain.OperationTypeNodePackages,
			),
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
			createMediumRiskOperation(
				"docker",
				"Clean Docker images, containers, and volumes",
				domain.OperationTypeDocker,
			),
			createLowRiskOperation(
				"go-packages",
				"Clean Go build cache",
				domain.OperationTypeGoPackages,
			),
			createLowRiskOperation(
				"node-packages",
				"Clean Node.js package caches",
				domain.OperationTypeNodePackages,
			),
			createLowRiskOperation(
				"homebrew-cleanup",
				"Clean Homebrew cache",
				domain.OperationTypeHomebrew,
			),
		},
	}
}

// createLowRiskOperation creates a low risk cleanup operation.
func createLowRiskOperation(
	name, description string,
	operationType domain.OperationType,
) domain.CleanupOperation {
	return domain.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   domain.RiskLevelLowType,
		Enabled:     domain.ProfileStatusEnabled,
		Settings:    domain.DefaultSettings(operationType),
	}
}

// createMediumRiskOperation creates a medium risk cleanup operation.
func createMediumRiskOperation(
	name, description string,
	operationType domain.OperationType,
) domain.CleanupOperation {
	return domain.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   domain.RiskLevelMediumType,
		Enabled:     domain.ProfileStatusEnabled,
		Settings:    domain.DefaultSettings(operationType),
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
				Name:        "nix-generations",           //nolint:goconst
				Description: "Clean old Nix generations", //nolint:goconst
				RiskLevel:   domain.RiskLevelHighType,
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean old Homebrew packages",
				RiskLevel:   domain.RiskLevelMediumType,
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
			},
			{
				Name:        "docker", //nolint:goconst
				Description: "Clean all unused Docker resources",
				RiskLevel:   domain.RiskLevelHighType,
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
			RiskLevel:   domain.RiskLevelMediumType,
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
		})
	}

	if includeHomebrew {
		operations = append(operations, domain.CleanupOperation{
			Name:        "homebrew-cleanup",
			Description: "Clean Homebrew cache and unused packages",
			RiskLevel:   domain.RiskLevelLowType,
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeHomebrew),
		})
	}

	if includeDocker {
		operations = append(operations, domain.CleanupOperation{
			Name:        "docker",
			Description: "Clean Docker images, containers, and volumes",
			RiskLevel:   domain.RiskLevelMediumType,
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeDocker),
		})
	}

	if includeNode {
		operations = append(operations, domain.CleanupOperation{
			Name:        "node-packages",
			Description: "Clean Node.js package caches",
			RiskLevel:   domain.RiskLevelLowType,
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeNodePackages),
		})
	}

	if includeGo {
		operations = append(operations, domain.CleanupOperation{
			Name:        "go-packages",
			Description: "Clean Go module and build caches",
			RiskLevel:   domain.RiskLevelLowType,
			Enabled:     domain.ProfileStatusEnabled,
			Settings:    domain.DefaultSettings(domain.OperationTypeGoPackages),
		})
	}

	// Always include temp files
	operations = append(operations, domain.CleanupOperation{
		Name:        "temp-files",
		Description: "Clean temporary files",
		RiskLevel:   domain.RiskLevelLowType,
		Enabled:     domain.ProfileStatusEnabled,
		Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
	})

	return map[string]*domain.Profile{
		setupModeCustom: {
			Name:        setupModeCustom,
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
