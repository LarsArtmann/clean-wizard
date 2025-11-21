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
				if err := ApplyValidationToConfig(loadedCfg, validationLevel); err != nil {
					return err
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

			// Resolve profile to use
			usedProfile, err := ResolveProfile(loadedCfg, profileName)
			if err != nil && profileName != "" {
				return err
			}

			// Fallback: try daily profile if ResolveProfile failed
			if err != nil && loadedCfg != nil {
				if dailyProfile, exists := loadedCfg.Profiles["daily"]; exists && dailyProfile.Status == domain.StatusEnabled {
					fmt.Printf("📋 Using daily profile configuration\n")
					usedProfile = dailyProfile
					err = nil
				}
			}

			// Extract settings from profile if available and build cleaner list
			var settings *domain.OperationSettings
			var cleaners []domain.Cleaner
			
			if usedProfile != nil {
				fmt.Printf("🏷️  Using profile: %s (%s)\n", usedProfile.Name, usedProfile.Description)
				
				// Process each operation in the profile
				for _, op := range usedProfile.Operations {
					if op.Status != domain.StatusEnabled {
						continue // Skip disabled operations
					}
					
					// Get settings for operation
					var opSettings *domain.OperationSettings
					if op.Settings != nil {
						opSettings = op.Settings
					} else {
						// Get default settings based on operation type
						switch op.Name {
						case "nix-generations":
							opSettings = domain.DefaultSettings(domain.OperationTypeNixGenerations)
						case "homebrew":
							opSettings = domain.DefaultSettings(domain.OperationTypeHomebrew)
						case "npm-cache":
							opSettings = domain.DefaultSettings(domain.OperationTypePackageCache)
						case "pnpm-store":
							opSettings = domain.DefaultSettings(domain.OperationTypePackageCache)
						case "temp-files":
							opSettings = domain.DefaultSettings(domain.OperationTypeTempFiles)
						default:
							opSettings = nil // No specific settings
						}
					}
					
					// Create appropriate cleaner for operation
					switch op.Name {
					case "nix-generations":
						fmt.Printf("🔧 Configuring Nix generations cleanup\n")
						nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)
						cleaners = append(cleaners, nixCleaner)
						if opSettings != nil {
							settings = opSettings
						}
						
					case "homebrew":
						fmt.Printf("🍺 Configuring Homebrew cleanup\n")
						homebrewCleaner := cleaner.NewHomebrewCleaner(cleanVerbose, cleanDryRun)
						cleaners = append(cleaners, homebrewCleaner)
						
					case "npm-cache":
						fmt.Printf("📦 Configuring npm cache cleanup\n")
						npmCleaner := cleaner.NewNpmCleaner(cleanVerbose, cleanDryRun)
						cleaners = append(cleaners, npmCleaner)
						
					case "pnpm-store":
						fmt.Printf("📦 Configuring pnpm store cleanup\n")
						pnpmCleaner := cleaner.NewPnpmCleaner(cleanVerbose, cleanDryRun)
						cleaners = append(cleaners, pnpmCleaner)
						
					case "temp-files":
						fmt.Printf("🗑️  Configuring temporary file cleanup\n")
						tempCleaner := cleaner.NewTempFileCleaner(cleanVerbose, cleanDryRun)
						cleaners = append(cleaners, tempCleaner)
						
					default:
						fmt.Printf("⚠️  Unknown operation: %s\n", op.Name)
					}
				}
			}

			// If no cleaners found, use Nix cleaner as fallback
			if len(cleaners) == 0 {
				fmt.Printf("🔧 Configuring Nix generations cleanup (fallback)\n")
				nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)
				cleaners = append(cleaners, nixCleaner)
			}

			startTime := time.Now()

			if cleanDryRun {
				fmt.Println("🔍 Running in DRY-RUN mode - no files will be deleted")
			}

			// Execute all cleaners and aggregate results
			var totalFreedBytes uint64
			var totalItemsRemoved uint
			var totalItemsFailed uint

			for _, cleaner := range cleaners {
				// Validate cleaner settings
				validator := middleware.NewValidationMiddleware()
				validatedSettings := validator.ValidateCleanerSettings(ctx, cleaner, settings)
				if validatedSettings.IsErr() {
					fmt.Printf("⚠️  Cleaner validation failed: %v\n", validatedSettings.Error())
					totalItemsFailed++
					continue
				}

				// Execute cleanup
				result := cleaner.Cleanup(ctx, settings)
				if result.IsErr() {
					fmt.Printf("⚠️  Cleaner failed: %v\n", result.Error())
					totalItemsFailed++
					continue
				}

				cleanResult := result.Value()
				totalFreedBytes += cleanResult.FreedBytes
				totalItemsRemoved += cleanResult.ItemsRemoved
				totalItemsFailed += cleanResult.ItemsFailed
			}

			duration := time.Since(startTime)

			// Create aggregated result for display
			aggregatedResult := domain.CleanResult{
				FreedBytes:   totalFreedBytes,
				ItemsRemoved: totalItemsRemoved,
				ItemsFailed:  totalItemsFailed,
				CleanTime:    duration,
				CleanedAt:    time.Now(),
				Strategy:     domain.StrategyConservative,
			}

			if cleanDryRun {
				aggregatedResult.Strategy = domain.StrategyDryRun
			}

			displayCleanResults(aggregatedResult, cleanVerbose, duration, cleanDryRun)
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
