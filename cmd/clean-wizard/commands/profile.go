package commands

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
)

// NewProfileCommand creates a profile management command.
func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage cleaning profiles",
		Long:  `Manage cleaning profiles - list, show, create, and delete profiles.`,
	}

	cmd.AddCommand(NewProfileListCommand())
	cmd.AddCommand(NewProfileShowCommand())
	cmd.AddCommand(NewProfileCreateCommand())
	cmd.AddCommand(NewProfileDeleteCommand())

	return cmd
}

// NewProfileListCommand creates a command to list all profiles.
func NewProfileListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		Long:  `List all available cleaning profiles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileListCommand(cmd, args)
		},
	}
}

// runProfileListCommand executes the profile list command.
func runProfileListCommand(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		// If config doesn't exist, show default profiles
		cfg = config.GetDefaultConfig()
	}

	fmt.Println("📋 Available Profiles")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if len(cfg.Profiles) == 0 {
		fmt.Println("   No profiles configured.")
		fmt.Println("   Run 'clean-wizard init' to create profiles.")
		return nil
	}

	for name, profile := range cfg.Profiles {
		status := "✅"
		if profile.Enabled == domain.ProfileStatusDisabled {
			status = "⚪"
		}
		fmt.Printf("  %s %s\n", status, name)
		fmt.Printf("     %s\n", profile.Description)
		fmt.Printf("     Operations: %d\n", len(profile.Operations))
		fmt.Println()
	}

	return nil
}

// NewProfileShowCommand creates a command to show profile details.
func NewProfileShowCommand() *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "show [profile]",
		Short: "Show profile details",
		Long:  `Show detailed information about a specific profile.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileShowCommand(cmd, args, detailed)
		},
	}

	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed operation information")

	return cmd
}

// runProfileShowCommand executes the profile show command.
func runProfileShowCommand(cmd *cobra.Command, args []string, detailed bool) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	profile, ok := cfg.Profiles[profileName]
	if !ok {
		// Check if it exists in default config
		defaultCfg := config.GetDefaultConfig()
		if p, exists := defaultCfg.Profiles[profileName]; exists {
			profile = p
		} else {
			fmt.Printf("❌ Profile '%s' not found\n", profileName)
			fmt.Println()
			fmt.Println("Available profiles:")
			for name := range defaultCfg.Profiles {
				fmt.Printf("   • %s\n", name)
			}
			return nil
		}
	}

	fmt.Printf("📄 Profile: %s\n", profileName)
	fmt.Println("━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("Description: %s\n", profile.Description)
	fmt.Printf("Status: %s\n", formatProfileStatus(profile.Enabled))
	fmt.Printf("Operations: %d\n", len(profile.Operations))
	fmt.Println()

	if detailed {
		fmt.Println("Operations:")
		fmt.Println("━━━━━━━━━━━━")
		for i, op := range profile.Operations {
			fmt.Printf("\n  %d. %s\n", i+1, op.Name)
			fmt.Printf("     Description: %s\n", op.Description)
			fmt.Printf("     Risk Level: %s\n", op.RiskLevel.String())
			fmt.Printf("     Enabled: %s\n", op.Enabled.String())
		}
	}

	return nil
}

// NewProfileCreateCommand creates a command to create a new profile.
func NewProfileCreateCommand() *cobra.Command {
	var name string
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new profile",
		Long:  `Create a new cleaning profile interactively.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileCreateCommand(cmd, args, name, description)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	cmd.Flags().StringVar(&description, "description", "", "Profile description")

	return cmd
}

// runProfileCreateCommand executes the profile create command.
func runProfileCreateCommand(cmd *cobra.Command, args []string, name, description string) error {
	// Interactive prompt for name if not provided
	if name == "" {
		fmt.Println("Creating new profile...")
		fmt.Print("Enter profile name: ")
		_, err := fmt.Scanln(&name)
		if err != nil {
			// Use default name if scan fails
			name = "custom"
		}
	}

	if description == "" {
		fmt.Print("Enter profile description: ")
		_, err := fmt.Scanln(&description)
		if err != nil {
			description = "Custom cleaning profile"
		}
	}

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	// Check if profile already exists
	if _, exists := cfg.Profiles[name]; exists {
		fmt.Printf("❌ Profile '%s' already exists\n", name)
		fmt.Printf("   Use 'clean-wizard profile delete %s' to remove it first.\n", name)
		return nil
	}

	// Create new profile with default operations
	cfg.Profiles[name] = &domain.Profile{
		Name:        name,
		Description: description,
		Enabled:     domain.ProfileStatusEnabled,
		Operations: []domain.CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Clean old Nix generations",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations),
			},
			{
				Name:        "temp-files",
				Description: "Clean temporary files",
				RiskLevel:   domain.RiskLevelType(domain.RiskLevelLowType),
				Enabled:     domain.ProfileStatusEnabled,
				Settings:    domain.DefaultSettings(domain.OperationTypeTempFiles),
			},
		},
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Profile '%s' created successfully!\n", name)
	fmt.Println()
	fmt.Println("💡 To use this profile:")
	fmt.Printf("   clean-wizard clean --profile %s\n", name)

	return nil
}

// NewProfileDeleteCommand creates a command to delete a profile.
func NewProfileDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [profile]",
		Short: "Delete a profile",
		Long:  `Delete a cleaning profile.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileDeleteCommand(cmd, args, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	return cmd
}

// runProfileDeleteCommand executes the profile delete command.
func runProfileDeleteCommand(cmd *cobra.Command, args []string, force bool) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ No configuration found.")
		fmt.Println("   Run 'clean-wizard init' to create configuration first.")
		os.Exit(0)
	}

	_, ok := cfg.Profiles[profileName]
	if !ok {
		fmt.Printf("❌ Profile '%s' not found\n", profileName)
		return nil
	}

	// Don't allow deleting the last profile
	if len(cfg.Profiles) <= 1 {
		fmt.Println("❌ Cannot delete the last profile.")
		fmt.Println("   At least one profile must exist.")
		return nil
	}

	// Confirm deletion
	if !force {
		fmt.Printf("Delete profile '%s'? This cannot be undone.\n", profileName)
		fmt.Print("Type 'yes' to confirm: ")
		var confirm string
		_, err := fmt.Scanln(&confirm)
		if err != nil || confirm != "yes" {
			fmt.Println("   Deletion cancelled.")
			return nil
		}
	}

	delete(cfg.Profiles, profileName)

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Profile '%s' deleted successfully!\n", profileName)

	return nil
}

// formatProfileStatus formats profile status for display.
func formatProfileStatus(status domain.ProfileStatus) string {
	switch status {
	case domain.ProfileStatusEnabled:
		return "✅ Enabled"
	case domain.ProfileStatusDisabled:
		return "⚪ Disabled"
	default:
		return "Unknown"
	}
}
