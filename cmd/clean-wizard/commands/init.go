package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewInitCommand creates init command
func NewInitCommand() *cobra.Command {
	var force bool
	var minimal bool

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard",
		Long:  `Interactive setup wizard that creates a comprehensive cleaning configuration tailored to your needs.`,
		Run: func(cmd *cobra.Command, args []string) {
			configPath := filepath.Join(os.Getenv("HOME"), ".clean-wizard.yaml")
			
			// Check if config already exists
			if _, err := os.Stat(configPath); err == nil && !force {
				fmt.Printf("⚠️  Configuration file already exists at %s\n", configPath)
				fmt.Println("Use --force to overwrite or --minimal to create a minimal configuration")
				return
			}

			fmt.Println("🧹 Clean Wizard Setup")
			fmt.Println("======================")
			fmt.Println("Let's create the perfect cleaning configuration for your system!")

			// Create default configuration
			config := createDefaultConfig()

			if !minimal {
				// Interactive setup
				err := runInteractiveSetup(config)
				if err != nil {
					fmt.Printf("❌ Error during interactive setup: %v\n", err)
					return
				}
			}

			// Write configuration
			err := writeConfiguration(config, configPath)
			if err != nil {
				fmt.Printf("❌ Failed to write configuration: %v\n", err)
				return
			}

			fmt.Printf("✅ Configuration created successfully at %s\n", configPath)
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  clean-wizard scan        # See what can be cleaned")
			fmt.Println("  clean-wizard clean --dry-run  # Test without changes")
			fmt.Println("  clean-wizard clean        # Perform cleanup")
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	initCmd.Flags().BoolVar(&minimal, "minimal", false, "Create minimal configuration")

	return initCmd
}

func runInteractiveSetup(config *domain.Config) error {
	reader := bufio.NewReader(os.Stdin)

	// Safe mode
	safeMode := askYesNo(reader, "Enable safe mode?", true)
	config.SafeMode = safeMode

	// Dry run by default
	dryRun := askYesNo(reader, "Enable dry run by default?", true)
	
	// Automatic backups
	_ = askYesNo(reader, "Enable automatic backups?", true) // TODO: Implement backup functionality

	// Maximum disk usage percentage
	maxDiskUsage := askNumber(reader, "Maximum disk usage percentage?", 10, 95, 90)
	config.MaxDiskUsage = maxDiskUsage

	// Configure profiles based on safety preferences
	configureProfiles(config, dryRun, safeMode)

	return nil
}

func askYesNo(reader *bufio.Reader, question string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}
	
	for {
		fmt.Printf("? %s › %s ", question, defaultStr)
		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return defaultValue
		}
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please answer 'y' or 'n'")
	}
}

func askNumber(reader *bufio.Reader, question string, min, max, defaultValue int) int {
	for {
		fmt.Printf("? %s › %d ", question, defaultValue)
		input, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			return defaultValue
		}
		
		var value int
		_, err = fmt.Sscanf(input, "%d", &value)
		if err != nil {
			fmt.Printf("Please enter a number between %d and %d\n", min, max)
			continue
		}
		
		if value < min || value > max {
			fmt.Printf("Please enter a number between %d and %d\n", min, max)
			continue
		}
		
		return value
	}
}

func configureProfiles(config *domain.Config, dryRun, safeMode bool) {
	// Daily profile
	dailySettings := &domain.OperationSettings{
		NixGenerations: &domain.NixGenerationsSettings{
			Generations: 3,
			Optimize:    true,
			DryRun:      dryRun,
		},
	}

	// Comprehensive profile
	comprehensiveSettings := &domain.OperationSettings{
		NixGenerations: &domain.NixGenerationsSettings{
			Generations: 1,
			Optimize:    true,
			DryRun:      dryRun,
		},
		Homebrew: &domain.HomebrewSettings{
			UnusedOnly: true,
			Prune:      "recent",
		},
	}

	// Aggressive profile (only if not in safe mode)
	aggressiveSettings := &domain.OperationSettings{
		NixGenerations: &domain.NixGenerationsSettings{
			Generations: 0, // Remove all but current
			Optimize:    true,
			DryRun:      dryRun,
		},
		Homebrew: &domain.HomebrewSettings{
			UnusedOnly: false,
			Prune:      "all",
		},
	}

	// Create profiles
	config.Profiles = map[string]*domain.Profile{
		"daily": {
			Name:        "daily",
			Description: "Quick daily cleanup for routine maintenance",
			Enabled:     true,
			Operations: []domain.CleanupOperation{
				{
					Name:        "nix-generations",
					Description: "Clean old Nix generations",
					RiskLevel:   domain.RiskLow,
					Enabled:     true,
					Settings:    dailySettings,
				},
			},
		},
		"comprehensive": {
			Name:        "comprehensive",
			Description: "Complete system cleanup for weekly maintenance",
			Enabled:     true,
			Operations: []domain.CleanupOperation{
				{
					Name:        "nix-generations",
					Description: "Comprehensive Nix cleanup",
					RiskLevel:   domain.RiskLow,
					Enabled:     true,
					Settings:    comprehensiveSettings,
				},
			},
		},
	}

	// Only add aggressive profile if not in safe mode
	if !safeMode {
		config.Profiles["aggressive"] = &domain.Profile{
			Name:        "aggressive",
			Description: "Nuclear option - everything that can be cleaned",
			Enabled:     true,
			Operations: []domain.CleanupOperation{
				{
					Name:        "nix-generations",
					Description: "Aggressive Nix cleanup",
					RiskLevel:   domain.RiskHigh,
					Enabled:     true,
					Settings:    aggressiveSettings,
				},
			},
		}
	}

	// Set default profile
	config.CurrentProfile = "daily"
}

func createDefaultConfig() *domain.Config {
	return &domain.Config{
		Version:    "1.0.0",
		SafeMode:   true,
		MaxDiskUsage: 90,
		Protected: []string{
			"/",
			"/System",
			"/Library",
			"/Applications",
			"/Users",
			"/usr",
			"/etc",
			"/var",
			"/bin",
			"/sbin",
			"/nix/store",
		},
		Profiles: make(map[string]*domain.Profile),
	}
}

func writeConfiguration(config *domain.Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
