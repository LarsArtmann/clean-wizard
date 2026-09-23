package commands

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	errorfamily "github.com/larsartmann/go-error-family"
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
			return errorfamily.WrapRejection(err, "init.confirm_overwrite", "confirmation error")
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
	cfg.Profiles = map[string]*types.Profile{
		"daily": daily,
	}

	err := config.Save(cfg)
	if err != nil {
		return errorfamily.WrapRejection(err, "init.config_save", "failed to save configuration")
	}

	fmt.Println()
	fmt.Println(SuccessStyle.Render("✅ Minimal configuration created successfully!"))
	printConfigSavedNotice()

	printGettingStartedHints()

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
		return errorfamily.WrapRejection(err, "init.config_save", "failed to save configuration")
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
		return "", errorfamily.WrapRejectionf(
			err, "init.setup_mode_select",
			"setup mode selection error for setupMode=%v", setupMode,
		)
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
		return nil, errorfamily.WrapRejectionf(
			err, "init.cleaner_select",
			"cleaner selection error for setupMode=%v", setupMode,
		)
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
			return nil, errorfamily.WrapRejectionf(
				err, "init.docker_warning",
				"docker warning error for includeDockerWarning=%v",
				includeDockerWarning,
			)
		}

		opts.includeDocker = includeDockerWarning
	}

	return &opts, nil
}

func buildConfigFromSetupMode(setupMode string, customOpts *customCleanerOptions) *types.Config {
	cfg := config.GetDefaultConfig()

	switch setupMode {
	case "quick":
		cfg.Profiles = map[string]*types.Profile{
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
		cfg.Profiles = map[string]*types.Profile{
			"daily":      createDailyProfile(),
			"weekly":     createWeeklyProfile(),
			"aggressive": createAggressiveProfile(),
		}
	}

	return cfg
}

func configureSafeMode(cfg *types.Config) error {
	safeMode := true

	if err := newConfirmForm(
		"Enable safe mode?",
		"Safe mode prevents destructive operations and adds confirmation prompts",
		"Yes, enable safe mode (Recommended)",
		"No, disable safe mode",
		&safeMode,
	).Run(); err != nil {
		return errorfamily.WrapRejection(err, "init.safe_mode_select", "safe mode selection error")
	}

	cfg.SafeMode = enums.SafeModeEnabled
	if !safeMode {
		cfg.SafeMode = enums.SafeModeDisabled
	}

	return nil
}

func printConfigSuccess(cfg *types.Config) {
	fmt.Println()
	fmt.Println(SuccessStyle.Render("✅ Configuration created successfully!"))
	printConfigSavedNotice()
	fmt.Println()
	fmt.Println("📋 Available profiles:")

	for name, profile := range cfg.Profiles {
		status := "enabled"
		if profile.Enabled == enums.ProfileStatusDisabled {
			status = "disabled"
		}

		fmt.Printf("   • %s (%s)\n", name, status)
		fmt.Printf("     %s\n", profile.Description)
	}

	printGettingStartedHints()
}

// printConfigSavedNotice prints the standard "Configuration saved to: ~/.clean-wizard.yaml"
// notice with surrounding blank lines.
func printConfigSavedNotice() {
	fmt.Println()
	fmt.Println(InfoStyle.Render("📁 Configuration saved to: ~/.clean-wizard.yaml"))
}

// printGettingStartedHints prints the standard post-setup usage hints.
func printGettingStartedHints() {
	fmt.Println()
	fmt.Println("💡 To get started:")
	fmt.Println("   clean-wizard clean              - Run daily cleanup")
	fmt.Println("   clean-wizard scan               - Scan for cleanable items")
	fmt.Println("   clean-wizard profile list       - List available profiles")
	fmt.Println("   clean-wizard config show        - View current configuration")
}

// createDailyProfile creates the daily cleanup profile.
func createDailyProfile() *types.Profile {
	return &types.Profile{
		Name:        "daily",
		Description: "Quick daily cleanup with safe operations",
		Enabled:     enums.ProfileStatusEnabled,
		Operations: []types.CleanupOperation{
			createLowRiskOperation(
				"temp-files",
				"Clean temporary files",
				operations.OperationTypeTempFiles,
			),
			createLowRiskOperation(
				"go-packages",
				"Clean Go module cache",
				operations.OperationTypeGoPackages,
			),
			createLowRiskOperation(
				"node-packages",
				"Clean Node.js package caches",
				operations.OperationTypeNodePackages,
			),
		},
	}
}

// createWeeklyProfile creates the weekly cleanup profile.
func createWeeklyProfile() *types.Profile {
	return &types.Profile{
		Name:        "weekly",
		Description: "Weekly comprehensive cleanup",
		Enabled:     enums.ProfileStatusEnabled,
		Operations: []types.CleanupOperation{
			createMediumRiskOperation(
				"docker",
				"Clean Docker images, containers, and volumes",
				operations.OperationTypeDocker,
			),
			createLowRiskOperation(
				"go-packages",
				"Clean Go build cache",
				operations.OperationTypeGoPackages,
			),
			createLowRiskOperation(
				"node-packages",
				"Clean Node.js package caches",
				operations.OperationTypeNodePackages,
			),
			createLowRiskOperation(
				"homebrew-cleanup",
				"Clean Homebrew cache",
				operations.OperationTypeHomebrew,
			),
		},
	}
}

// createLowRiskOperation creates a low risk cleanup operation.
func createLowRiskOperation(
	name, description string,
	operationType operations.OperationType,
) types.CleanupOperation {
	return types.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   enums.RiskLevelLowType,
		Enabled:     enums.ProfileStatusEnabled,
		Settings:    operations.DefaultSettings(operationType),
	}
}

// createMediumRiskOperation creates a medium risk cleanup operation.
func createMediumRiskOperation(
	name, description string,
	operationType operations.OperationType,
) types.CleanupOperation {
	return types.CleanupOperation{
		Name:        name,
		Description: description,
		RiskLevel:   enums.RiskLevelMediumType,
		Enabled:     enums.ProfileStatusEnabled,
		Settings:    operations.DefaultSettings(operationType),
	}
}

// createAggressiveProfile creates the aggressive cleanup profile.
func createAggressiveProfile() *types.Profile {
	return &types.Profile{
		Name:        "aggressive",
		Description: "Deep aggressive cleanup (may remove useful items)",
		Enabled:     enums.ProfileStatusDisabled,
		Operations: []types.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   enums.RiskLevelHighType,
				Enabled:     enums.ProfileStatusEnabled,
				Settings:    operations.DefaultSettings(operations.OperationTypeNixGenerations),
			},
			{
				Name:        "homebrew-cleanup",
				Description: "Clean old Homebrew packages",
				RiskLevel:   enums.RiskLevelMediumType,
				Enabled:     enums.ProfileStatusEnabled,
				Settings:    operations.DefaultSettings(operations.OperationTypeHomebrew),
			},
			{
				Name:        "docker",
				Description: "Clean all unused Docker resources",
				RiskLevel:   enums.RiskLevelHighType,
				Enabled:     enums.ProfileStatusEnabled,
				Settings:    operations.DefaultSettings(operations.OperationTypeDocker),
			},
		},
	}
}

// createCustomProfile creates a custom profile based on user selections.
func createCustomProfile(
	includeNix, includeHomebrew, includeDocker, includeNode, includeGo bool,
) map[string]*types.Profile {
	operations := make([]types.CleanupOperation, 0)

	if includeNix {
		operations = append(operations, types.CleanupOperation{
			Name:        "nix-generations",
			Description: "Clean old Nix generations",
			RiskLevel:   enums.RiskLevelMediumType,
			Enabled:     enums.ProfileStatusEnabled,
			Settings:    operations.DefaultSettings(operations.OperationTypeNixGenerations),
		})
	}

	if includeHomebrew {
		operations = append(operations, types.CleanupOperation{
			Name:        "homebrew-cleanup",
			Description: "Clean Homebrew cache and unused packages",
			RiskLevel:   enums.RiskLevelLowType,
			Enabled:     enums.ProfileStatusEnabled,
			Settings:    operations.DefaultSettings(operations.OperationTypeHomebrew),
		})
	}

	if includeDocker {
		operations = append(operations, types.CleanupOperation{
			Name:        "docker",
			Description: "Clean Docker images, containers, and volumes",
			RiskLevel:   enums.RiskLevelMediumType,
			Enabled:     enums.ProfileStatusEnabled,
			Settings:    operations.DefaultSettings(operations.OperationTypeDocker),
		})
	}

	if includeNode {
		operations = append(operations, types.CleanupOperation{
			Name:        "node-packages",
			Description: "Clean Node.js package caches",
			RiskLevel:   enums.RiskLevelLowType,
			Enabled:     enums.ProfileStatusEnabled,
			Settings:    operations.DefaultSettings(operations.OperationTypeNodePackages),
		})
	}

	if includeGo {
		operations = append(operations, types.CleanupOperation{
			Name:        "go-packages",
			Description: "Clean Go module and build caches",
			RiskLevel:   enums.RiskLevelLowType,
			Enabled:     enums.ProfileStatusEnabled,
			Settings:    operations.DefaultSettings(operations.OperationTypeGoPackages),
		})
	}

	// Always include temp files
	operations = append(operations, types.CleanupOperation{
		Name:        "temp-files",
		Description: "Clean temporary files",
		RiskLevel:   enums.RiskLevelLowType,
		Enabled:     enums.ProfileStatusEnabled,
		Settings:    operations.DefaultSettings(operations.OperationTypeTempFiles),
	})

	return map[string]*types.Profile{
		setupModeCustom: {
			Name:        setupModeCustom,
			Description: "Custom cleanup profile",
			Enabled:     enums.ProfileStatusEnabled,
			Operations:  operations,
		},
	}
}

// GetDefaultConfig returns the default configuration.
// This is a wrapper around config.GetDefaultConfig for use in commands.
func GetDefaultConfig() *types.Config {
	return config.GetDefaultConfig()
}
