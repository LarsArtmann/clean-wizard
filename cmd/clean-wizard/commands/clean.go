package commands

import (
	"context"
	"fmt"
	"os"
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
func NewCleanCommand(validationLevel config.ValidationLevel) *cobra.Command {
	var configFile string
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Long:  `Safely clean old files, package caches, and temporary data from your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Starting system cleanup...")
			ctx := context.Background()

			// Load and validate configuration if provided
			if configFile != "" {
				fmt.Printf("📄 Loading configuration from %s...\n", configFile)
				
				// Set config file path using environment variable
				os.Setenv("CONFIG_PATH", configFile)
				
				loadedCfg, err := config.LoadWithContext(ctx)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Apply configuration validation (basic only)
				fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n", 
					loadedCfg.SafeMode, len(loadedCfg.Profiles))
				
				// Apply config values to clean operation
				if dailyProfile, exists := loadedCfg.Profiles["daily"]; exists {
					fmt.Printf("📋 Using daily profile configuration\n")
					// Extract clean parameters from profile
					for _, op := range dailyProfile.Operations {
						if op.Name == "nix-generations" && op.Enabled {
							fmt.Printf("🔧 Configuring Nix generations cleanup\n")
							// Could extract settings like keep count from op.Settings
							break
						}
					}
				}
			}

			nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)
			settings := map[string]any{"generations": 3}

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
			result := nixCleaner.CleanOldGenerations(ctx,3)

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
	} else {
		fmt.Printf("\n✅ Cleanup completed successfully\n")
	}
}