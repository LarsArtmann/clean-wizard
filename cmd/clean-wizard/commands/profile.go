package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
)

// NewProfileCommand creates the profile command
func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long:  `Manage and view cleanup configuration profiles for different cleaning strategies.`,
	}

	// Add subcommands
	cmd.AddCommand(
		NewProfileListCommand(),
		NewProfileSelectCommand(),
		NewProfileInfoCommand(),
	)

	return cmd
}

// NewProfileListCommand creates the profile list command
func NewProfileListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		Long:  `List all available configuration profiles with their descriptions and status.`,
		Args:  cobra.NoArgs,
		RunE:  runProfileList,
	}
}

// NewProfileSelectCommand creates the profile select command
func NewProfileSelectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "select [profile-name]",
		Short: "Select a profile as the default",
		Long:  `Select a profile as the default for future clean and scan operations.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runProfileSelect,
	}
}

// NewProfileInfoCommand creates the profile info command
func NewProfileInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info [profile-name]",
		Short: "Show detailed information about a profile",
		Long:  `Show detailed information about a specific profile including all operations and their settings.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runProfileInfo,
	}
}

// runProfileList lists all available profiles
func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("❌ No profiles found in configuration")
		return nil
	}

	// Sort profile names for consistent display
	var profileNames []string
	for name := range cfg.Profiles {
		profileNames = append(profileNames, name)
	}
	sort.Strings(profileNames)

	fmt.Printf("📋 Available Profiles:\n\n")

	for _, name := range profileNames {
		profile := cfg.Profiles[name]
		status := "✅ Enabled"
		if !profile.Enabled {
			status = "❌ Disabled"
		}

		fmt.Printf("🏷️  %s\n", name)
		fmt.Printf("   Description: %s\n", profile.Description)
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Operations: %d\n", len(profile.Operations))

		// Show risk level summary
		riskLevels := make(map[domain.RiskLevel]int)
		for _, op := range profile.Operations {
			riskLevels[op.RiskLevel]++
		}

		var riskSummary []string
		if riskLevels[domain.RiskLow] > 0 {
			riskSummary = append(riskSummary, fmt.Sprintf("%d low", riskLevels[domain.RiskLow]))
		}
		if riskLevels[domain.RiskMedium] > 0 {
			riskSummary = append(riskSummary, fmt.Sprintf("%d medium", riskLevels[domain.RiskMedium]))
		}
		if riskLevels[domain.RiskHigh] > 0 {
			riskSummary = append(riskSummary, fmt.Sprintf("%d high", riskLevels[domain.RiskHigh]))
		}
		if riskLevels[domain.RiskCritical] > 0 {
			riskSummary = append(riskSummary, fmt.Sprintf("%d critical", riskLevels[domain.RiskCritical]))
		}

		if len(riskSummary) > 0 {
			fmt.Printf("   Risk levels: %s\n", strings.Join(riskSummary, ", "))
		}

		fmt.Println()
	}

	return nil
}

// runProfileSelect selects a profile as default
func runProfileSelect(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if profile exists
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		// Show available profiles
		var names []string
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		return fmt.Errorf("profile '%s' not found. Available profiles: %s",
			profileName, strings.Join(names, ", "))
	}

	// Check if profile is enabled
	if !profile.Enabled {
		return fmt.Errorf("profile '%s' is disabled. Enable it first before selecting.", profileName)
	}

	// Update current profile in config
	cfg.CurrentProfile = profileName
	cfg.Updated = config.GetCurrentTime()

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Profile '%s' selected as default\n", profileName)
	fmt.Printf("📝 Configuration updated in ~/.clean-wizard.yaml\n")

	return nil
}

// runProfileInfo shows detailed information about a profile
func runProfileInfo(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var profileName string
	if len(args) == 1 {
		profileName = args[0]
	} else {
		profileName = cfg.CurrentProfile
		if profileName == "" {
			profileName = "daily" // Default fallback
		}
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		var names []string
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)
		return fmt.Errorf("profile '%s' not found. Available profiles: %s",
			profileName, strings.Join(names, ", "))
	}

	fmt.Printf("📋 Profile Details: %s\n", profile.Name)
	fmt.Printf("%s\n", strings.Repeat("=", len(profile.Name)+20))

	// Basic info
	status := "✅ Enabled"
	if !profile.Enabled {
		status = "❌ Disabled"
	}

	fmt.Printf("Description: %s\n", profile.Description)
	fmt.Printf("Status: %s\n", status)

	// Check if this is the current profile
	isCurrent := (cfg.CurrentProfile == profileName)
	if isCurrent {
		fmt.Printf("Current: ✅ Yes (default profile)\n")
	} else {
		fmt.Printf("Current: ❌ No (current: %s)\n", cfg.CurrentProfile)
	}

	fmt.Println()

	// Operations
	if len(profile.Operations) == 0 {
		fmt.Printf("Operations: None configured\n")
	} else {
		fmt.Printf("Operations (%d total):\n", len(profile.Operations))
		fmt.Println()

		for i, op := range profile.Operations {
			status := "✅ Enabled"
			if !op.Enabled {
				status = "❌ Disabled"
			}

			riskColor := getRiskColor(op.RiskLevel)
			fmt.Printf("  %d. %s\n", i+1, op.Name)
			fmt.Printf("     Description: %s\n", op.Description)
			fmt.Printf("     Status: %s\n", status)
			fmt.Printf("     Risk Level: %s%s\n", riskColor, op.RiskLevel.String())

			// Show settings if any
			if op.Settings != nil {
				fmt.Printf("     Settings:\n")
				if op.Settings.NixGenerations != nil {
					fmt.Printf("       • generations: %d\n", op.Settings.NixGenerations.Generations)
					fmt.Printf("       • optimize: %t\n", op.Settings.NixGenerations.Optimize)
				}
				if op.Settings.TempFiles != nil {
					fmt.Printf("       • older_than: %s\n", op.Settings.TempFiles.OlderThan)
					if len(op.Settings.TempFiles.Excludes) > 0 {
						fmt.Printf("       • excludes: %v\n", op.Settings.TempFiles.Excludes)
					}
				}
				if op.Settings.Homebrew != nil {
					fmt.Printf("       • unused_only: %t\n", op.Settings.Homebrew.UnusedOnly)
				}
				if op.Settings.SystemTemp != nil {
					fmt.Printf("       • paths: %v\n", op.Settings.SystemTemp.Paths)
					fmt.Printf("       • older_than: %s\n", op.Settings.SystemTemp.OlderThan)
				}
			}

			if i < len(profile.Operations)-1 {
				fmt.Println()
			}
		}
	}

	fmt.Println()

	// Usage hint
	if !isCurrent && profile.Enabled {
		fmt.Printf("💡 To use this profile: clean-wizard profile select %s\n", profileName)
	} else if !profile.Enabled {
		fmt.Printf("⚠️  This profile is disabled. Enable it in the configuration to use it.\n")
	}

	return nil
}

// getRiskColor returns color code for risk level
func getRiskColor(level domain.RiskLevel) string {
	switch level {
	case domain.RiskLow:
		return "🟢 " // Green
	case domain.RiskMedium:
		return "🟡 " // Yellow
	case domain.RiskHigh:
		return "🟠 " // Orange
	case domain.RiskCritical:
		return "🔴 " // Red
	default:
		return "⚪ " // White
	}
}
