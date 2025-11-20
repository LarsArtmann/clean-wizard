package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/middleware"
	"github.com/spf13/cobra"
)

var (
	cleanDryRun  bool
	cleanVerbose bool
)

// NewCleanCommand creates clean command with proper domain types
func NewCleanCommand() *cobra.Command {
	var configFile string
	var profileName string

	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Long:  `Safely clean old files, package caches, and temporary data from your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Starting system cleanup...")
			ctx := context.Background()

			// Parse validation level from flag
			validationLevelStr, _ := cmd.Flags().GetString("validation-level")
			validationLevel := ParseValidationLevel(validationLevelStr)

			// Parse profile name from flag
			profileName, _ = cmd.Flags().GetString("profile")

			// Load and validate configuration if provided
			var loadedCfg *domain.Config
			if configFile != "" {
				fmt.Printf("📄 Loading configuration from %s...\n", configFile)

				var err error
				loadedCfg, err = config.LoadWithContextAndPath(ctx, configFile)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Apply validation based on level
				if validationLevel > config.ValidationLevelNone {
					fmt.Printf("🔍 Applying validation level: %s\n", validationLevel.String())

					if validationLevel >= config.ValidationLevelBasic {
						// Basic validation
						if len(loadedCfg.Protected) == 0 {
							return fmt.Errorf("basic validation failed: protected paths cannot be empty")
						}
					}

					if validationLevel >= config.ValidationLevelComprehensive {
						// Comprehensive validation
						if err := loadedCfg.Validate(); err != nil {
							return fmt.Errorf("comprehensive validation failed: %w", err)
						}
					}

					if validationLevel >= config.ValidationLevelStrict {
						// Strict validation
						if loadedCfg.SafetyLevel == domain.SafetyLevelDisabled {
							return fmt.Errorf("strict validation failed: safe_mode must be enabled")
						}
					}
				}

				fmt.Printf("✅ Configuration applied: safety_level=%v, profiles=%d\n",
					loadedCfg.SafetyLevel, len(loadedCfg.Profiles))
			} else {
				// Load default configuration to get profile information
				var err error
				loadedCfg, err = config.LoadWithContext(ctx)
				if err != nil {
					fmt.Printf("⚠️  Could not load default configuration: %v\n", err)
					// Continue without profile support
				} else {
					fmt.Printf("📋 Using configuration from ~/.clean-wizard.yaml\n")
				}
			}

			// Apply profile if specified
			var usedProfile *domain.Profile
			if loadedCfg != nil && profileName != "" {
				profile, exists := loadedCfg.Profiles[profileName]
				if !exists {
					return fmt.Errorf("profile '%s' not found in configuration", profileName)
				}

				if profile.Status == domain.StatusDisabled {
					return fmt.Errorf("profile '%s' is disabled", profileName)
				}

				fmt.Printf("🏷️  Using profile: %s (%s)\n", profileName, profile.Description)
				usedProfile = profile
			} else if loadedCfg != nil && loadedCfg.CurrentProfile != "" {
				// Use current profile from config
				profile := loadedCfg.Profiles[loadedCfg.CurrentProfile]
				if profile != nil && profile.Status == domain.StatusEnabled {
					fmt.Printf("🏷️  Using current profile: %s (%s)\n", loadedCfg.CurrentProfile, profile.Description)
					usedProfile = profile
				}
			} else if loadedCfg != nil {
				// Default to daily profile if available
				if dailyProfile, exists := loadedCfg.Profiles["daily"]; exists && dailyProfile.Status == domain.StatusEnabled {
					fmt.Printf("📋 Using daily profile configuration\n")
					usedProfile = dailyProfile
				}
			}

			// Extract settings from profile if available
			var settings *domain.OperationSettings
			if usedProfile != nil {
				for _, op := range usedProfile.Operations {
					if op.Name == "nix-generations" && op.Status == domain.StatusEnabled {
						fmt.Printf("🔧 Configuring Nix generations cleanup\n")
						if op.Settings != nil {
							settings = op.Settings
						} else {
							settings = domain.DefaultSettings(domain.OperationTypeNixGenerations)
						}
						break
					}
				}
			}

			nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)

			validator := middleware.NewValidationMiddleware()
			validatedSettings := validator.ValidateCleanerSettings(ctx, nixCleaner, settings)
			if validatedSettings.IsErr() {
				return fmt.Errorf("cleaner validation failed: %w", validatedSettings.Error())
			}

			startTime := time.Now()

			if cleanDryRun {
				fmt.Println("🔍 Running in DRY-RUN mode - no files will be deleted")
			}

			// Clean old generations (keep last 3)
			result := nixCleaner.CleanOldGenerations(ctx, 3)

			if result.IsErr() {
				_, err := result.Unwrap()
				return handleCleanError(err, cleanDryRun)
			}

			duration := time.Since(startTime)
			displayCleanResults(result.Value(), cleanVerbose, duration, cleanDryRun)
			return nil
		},
	}

	// Clean command flags
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be cleaned without doing it")
	cleanCmd.Flags().BoolVar(&cleanVerbose, "verbose", false, "Show detailed output")
	cleanCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")

	return cleanCmd
}

// handleCleanError provides user-friendly error messages
func handleCleanError(err error, isDryRun bool) error {
	if isDryRun {
		fmt.Printf("🔍 Dry run encountered issues: %s\n", err)
		return nil
	}

	return fmt.Errorf("cleanup failed: %w", err)
}

// displayCleanResults shows cleanup results to user
func displayCleanResults(result domain.CleanResult, verbose bool, duration time.Duration, isDryRun bool) {
	status := "SUCCESS"
	if !result.IsValid() {
		status = "FAILED"
	}

	action := "cleaned"
	if isDryRun {
		action = "would be cleaned"
	}

	fmt.Printf("\n🎯 Cleanup Results (%s):\n", status)
	fmt.Printf("   • Duration: %s\n", duration.String())

	if result.IsValid() {
		fmt.Printf("   • Status: %d items %s\n", result.ItemsRemoved, action)
		if result.FreedBytes > 0 {
			fmt.Printf("   • Space freed: %s\n", format.Bytes(result.FreedBytes))
		}

		if verbose {
			fmt.Printf("\n📋 Details:\n")
			fmt.Printf("   - Strategy: %s\n", result.Strategy)
			fmt.Printf("   - Items failed: %d\n", result.ItemsFailed)
		}
	}

	if isDryRun {
		fmt.Printf("\n💡 This was a dry run - no files were actually deleted\n")
		fmt.Printf("   🏃 Run 'clean-wizard clean' without --dry-run to perform cleanup\n")
	}

	if result.IsValid() {
		fmt.Printf("\n✅ Cleanup completed successfully\n")
	}
}
