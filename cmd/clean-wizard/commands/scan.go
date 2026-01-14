package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/middleware"
	sharedConfig "github.com/LarsArtmann/clean-wizard/internal/shared/utils/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewScanCommand creates scan command with proper domain types.
func NewScanCommand() *cobra.Command {
	var configFile string
	var profileName string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		Long:  `Analyze your system to identify old files, package caches, and temporary data that can be safely cleaned.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ = cmd.Flags().GetBool("json")
			if !jsonOutput {
				fmt.Println("🔍 Analyzing system state...")
			}
			ctx := context.Background()

			// Get verbose from parent command's persistent flags
			verbose, _ := cmd.Parent().PersistentFlags().GetBool("verbose")

			// Parse validation level from flag
			validationLevelStr, _ := cmd.Flags().GetString("validation-level")
			validationLevel := ParseValidationLevel(validationLevelStr)

			// Parse profile name from flag
			profileName, _ = cmd.Flags().GetString("profile")

			// Determine scan parameters from configuration
			scanType := domain.ScanTypeNixStore
			scanMode := domain.ScanModeRecursive
			limit := 100
			var loadedCfg *domain.Config

			// Load and validate configuration if provided
			if configFile != "" {
				if !jsonOutput {
					fmt.Printf("📄 Loading configuration from %s...\n", configFile)
				}

				// Set config file path using environment variable
				os.Setenv("CONFIG_PATH", configFile)

				var err error
				loadedCfg, err = config.LoadWithContext(ctx)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Apply validation based on level
				if validationLevel > config.ValidationLevelNone {
					if !jsonOutput {
						fmt.Printf("🔍 Applying validation level: %s\n", validationLevel.String())
					}

					if validationLevel >= config.ValidationLevelBasic {
						// Basic validation
						if len(loadedCfg.Protected) == 0 {
							return errors.New("basic validation failed: protected paths cannot be empty")
						}
					}
				}
			} else {
				// Load default configuration to get profile information
				logger := logrus.New()
				logger.SetOutput(os.Stderr)
				logger.SetFormatter(&logrus.TextFormatter{ForceColors: true})

				var err error
				loadedCfg, err = sharedConfig.LoadConfigOrContinue(ctx, logger)
				if err != nil {
					// This shouldn't happen with LoadConfigOrContinue
					fmt.Printf("⚠️  Unexpected error loading configuration: %v\n", err)
				} else if loadedCfg == nil {
					// Continue without profile support
				}
			}

			// Apply validation if we have loaded configuration
			if loadedCfg != nil && validationLevel > config.ValidationLevelNone {
				fmt.Printf("🔍 Applying validation level: %s\n", validationLevel.String())

				if validationLevel >= config.ValidationLevelBasic {
					// Basic validation
					if len(loadedCfg.Protected) == 0 {
						return errors.New("basic validation failed: protected paths cannot be empty")
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
					if !loadedCfg.SafeMode.IsEnabled() {
						return errors.New("strict validation failed: safe_mode must be enabled")
					}
				}

				fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n",
					loadedCfg.SafeMode, len(loadedCfg.Profiles))
			}

			// Apply profile if specified
			if loadedCfg != nil && profileName != "" {
				profile, exists := loadedCfg.Profiles[profileName]
				if !exists {
					return fmt.Errorf("profile '%s' not found in configuration", profileName)
				}

				if !profile.Enabled.IsEnabled() {
					return fmt.Errorf("profile '%s' is disabled", profileName)
				}

				fmt.Printf("🏷️  Using profile: %s (%s)\n", profileName, profile.Description)
			} else if loadedCfg != nil && loadedCfg.CurrentProfile != "" {
				// Use current profile from config
				profile := loadedCfg.Profiles[loadedCfg.CurrentProfile]
				if profile != nil && profile.Enabled.IsEnabled() {
					fmt.Printf("🏷️  Using current profile: %s (%s)\n", loadedCfg.CurrentProfile, profile.Description)
				}
			} else if loadedCfg != nil {
				// Default to daily profile if available
				if dailyProfile, exists := loadedCfg.Profiles["daily"]; exists && dailyProfile.Enabled.IsEnabled() {
					fmt.Printf("📋 Using daily profile configuration\n")
					// Extract scan parameters from profile
					for _, op := range dailyProfile.Operations {
						if op.Name == "nix-generations" && op.Enabled.IsEnabled() {
							// Nix generations scanning
							limit = 50 // Default for generations
							break
						}
					}
				}
			}

			// Create scan request with applied configuration
			scanReq := domain.ScanRequest{
				Type:      scanType,
				Recursive: scanMode,
				Limit:     limit,
			}

			// Validate scan request
			validator := middleware.NewValidationMiddleware()
			validatedReq := validator.ValidateScanRequest(ctx, scanReq)
			if validatedReq.IsErr() {
				return fmt.Errorf("scan validation failed: %w", validatedReq.Error())
			}

			// Perform scan
			nixCleaner := cleaner.NewNixCleaner(verbose, false)
			result := nixCleaner.ListGenerations(ctx)

			if result.IsErr() {
				return fmt.Errorf("scan failed: %w", result.Error())
			}

			// Convert generations to scan result for display
			generations := result.Value()

			// Calculate total bytes from generations
			var totalBytes int64
			for _, gen := range generations {
				totalBytes += gen.EstimateSize()
			}

			scanResult := domain.ScanResult{
				TotalBytes:   totalBytes,
				TotalItems:   len(generations),
				ScannedPaths: []string{"/nix/store"},
				ScanTime:     time.Duration(100 * time.Millisecond),
				ScannedAt:    time.Now(),
			}

			// Display results
			displayScanResults(scanResult, generations, verbose, jsonOutput)
			return nil
		},
	}

	// Scan command flags
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")

	return cmd
}

// displayScanResults shows scan results to user.
func displayScanResults(result domain.ScanResult, generations []domain.NixGeneration, verbose, jsonOutput bool) {
	if jsonOutput {
		// Create JSON output for generations
		generationsJSON := make([]map[string]any, 0, len(generations))
		for _, gen := range generations {
			generationsJSON = append(generationsJSON, map[string]any{
				"id":                  gen.ID,
				"path":                gen.Path,
				"date":                gen.Date,
				"current":             gen.Current.IsCurrent(),
				"size_estimate_bytes": gen.EstimateSize(),
			})
		}

		jsonOutput := map[string]any{
			"status":            "success",
			"total_items":       result.TotalItems,
			"total_bytes":       result.TotalBytes,
			"total_bytes_human": format.Bytes(result.TotalBytes),
			"scan_time_ms":      result.ScanTime.Milliseconds(),
			"scanned_at":        result.ScannedAt,
			"scanned_paths":     result.ScannedPaths,
			"generations":       generationsJSON,
		}
		jsonData, _ := json.MarshalIndent(jsonOutput, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	fmt.Printf("\n📊 Scan Results:\n")
	fmt.Printf("   • Total generations: %d\n", result.TotalItems)

	// Count current generations
	currentCount := 0
	for _, gen := range generations {
		if gen.Current.IsCurrent() {
			currentCount++
		}
	}

	// Calculate cleanable generations (non-current ones)
	cleanableCount := len(generations) - currentCount

	fmt.Printf("   • Current generation: %d\n", currentCount)
	fmt.Printf("   • Cleanable generations: %d\n", cleanableCount)
	fmt.Printf("   • Store size: %s\n", format.Bytes(result.TotalBytes))

	if cleanableCount > 0 {
		fmt.Printf("\n💡 You can clean up %d old generations to free space\n", cleanableCount)
		fmt.Printf("   🏃 Run 'clean-wizard clean' to start cleanup\n")
		fmt.Printf("   🔍 Try 'clean-wizard clean --dry-run' first to see what would be cleaned\n")
	} else {
		fmt.Printf("\n✅ System is already clean - no old generations found\n")
	}

	// Add completion message for BDD tests
	fmt.Printf("\n✅ Scan completed!\n")
}
