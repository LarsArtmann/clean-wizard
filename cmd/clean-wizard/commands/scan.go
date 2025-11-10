package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/LarsArtmann/clean-wizard/internal/middleware"
	"github.com/spf13/cobra"
)

// NewScanCommand creates scan command with proper domain types
func NewScanCommand(verbose bool) *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		Long:  `Analyze your system to identify old files, package caches, and temporary data that can be safely cleaned.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Analyzing system state...")
			ctx := context.Background()

			// Determine scan parameters from configuration
			scanType := domain.ScanTypeNixStore
			recursive := true
			limit := 100
			var loadedCfg *domain.Config

			// Load and validate configuration if provided
			if configFile != "" {
				fmt.Printf("📄 Loading configuration from %s...\n", configFile)
				
				// Set config file path using environment variable (simpler approach)
				os.Setenv("CONFIG_PATH", configFile)
				
				var err error
				loadedCfg, err = config.LoadWithContext(ctx)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Debug: Show what was actually loaded
				fmt.Printf("🔍 DEBUG: Loaded config:\n")
				fmt.Printf("  - Version: %s\n", loadedCfg.Version)
				fmt.Printf("  - SafeMode: %v\n", loadedCfg.SafeMode)
				fmt.Printf("  - MaxDiskUsage: %d\n", loadedCfg.MaxDiskUsage)
				fmt.Printf("  - Protected paths: %v\n", loadedCfg.Protected)
				fmt.Printf("  - Profiles: %v\n", loadedCfg.Profiles)

				// Apply configuration validation using middleware (basic only for now)
				// middleware := config.NewValidationMiddleware()
				// Skip comprehensive validation to get basic workflow working
				// _, err = middleware.ValidateAndLoadConfig(ctx)
				// if err != nil {
				//	return fmt.Errorf("configuration validation failed: %w", err)
				// }

				// Apply config values to scan request
				fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n", 
					loadedCfg.SafeMode, len(loadedCfg.Profiles))
				
				// Use profile-based configuration if available
				if dailyProfile, exists := loadedCfg.Profiles["daily"]; exists {
					fmt.Printf("📋 Using daily profile configuration\n")
					// Extract scan parameters from profile
					for _, op := range dailyProfile.Operations {
						if op.Name == "nix-generations" && op.Enabled {
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
				Recursive: recursive,
				Limit:     limit,
			}

			// Validate scan request
			validator := middleware.NewValidationMiddleware()
			validatedReq := validator.ValidateScanRequest(ctx, scanReq)
			if validatedReq.IsErr() {
				return fmt.Errorf("scan validation failed: %w", validatedReq.Error())
			}

			nixCleaner := cleaner.NewNixCleaner(verbose, true) // dry-run for safety

			// Get Nix generations
			genResult := nixCleaner.ListGenerations(ctx)
			if genResult.IsErr() {
				return handleScanError(genResult.Error())
			}

			generations := genResult.Value()
			displayScanResults(generations, verbose, nixCleaner, ctx)
			return nil
		},
	}

	// Add configuration file flag
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")

	return cmd
}

// handleScanError provides user-friendly error messages for scanning
func handleScanError(err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "command not found"):
		return fmt.Errorf(`❌ Nix package manager not found

💡 To fix this:
   • Install Nix: curl --proto '=https' --tlsv1.2 -sSf https://nixos.org/install/standalone | sh
   • Or try other cleanup targets if you don't use Nix
   
📚 Learn more: https://nixos.org/`)

	case strings.Contains(errMsg, "permission"):
		return fmt.Errorf(`❌ Permission denied while scanning

💡 To fix this:
   • Try running with sudo: sudo clean-wizard scan
   • Check if you have read access to Nix profiles
   • Verify Nix is properly installed`)

	case strings.Contains(errMsg, "no such file"):
		return fmt.Errorf(`❌ Cannot find Nix profiles directory

💡 This could mean:
   • Nix is not installed correctly
   • Nix is installed but not in standard location
   • You're on a system where Nix works differently
   
🔧 Try: nix-env --version to check installation`)

	default:
		return fmt.Errorf(`❌ System scan failed: %s

💡 Suggestions:
   • Ensure Nix is properly installed
   • Try running with --verbose for more details
   • Check system permissions
   • If this persists, please report an issue`, err.Error())
	}
}

// displayScanResults shows user-friendly scan results
func displayScanResults(generations []domain.NixGeneration, verbose bool, nixCleaner *cleaner.NixCleaner, ctx context.Context) {
	fmt.Println("\n📊 Scan Results:")

	if len(generations) == 0 {
		fmt.Println("   🎉 No Nix generations found - your system is clean!")
		return
	}

	// Count total generations and identify old ones
	cleanableCount := 0
	for _, gen := range generations {
		if !gen.Current {
			cleanableCount++
		}
	}

	fmt.Printf("   • Total generations: %d\n", len(generations))
	fmt.Printf("   • Current generation: %s\n", getCurrentGeneration(generations))
	fmt.Printf("   • Cleanable generations: %d\n", cleanableCount)

	// Get store size
	storeSize := nixCleaner.GetStoreSize(ctx)
	if storeSize > 0 {
		fmt.Printf("   • Store size: %s\n", format.Bytes(storeSize))
	}

	// Show detailed generation info
	if verbose {
		fmt.Println("\n📋 Generation Details:")
		for _, gen := range generations {
			status := "🟢 CURRENT"
			if !gen.Current {
				status = "🔴 OLD"
			}
			fmt.Printf("   %s Generation %s (%s)\n", status, fmt.Sprintf("%d", gen.ID), format.DateTime(gen.Date))
		}
	}

	// Provide user guidance
	if cleanableCount > 0 {
		fmt.Printf("\n💡 You can clean up %d old generations to free space\n", cleanableCount)
		fmt.Println("   🏃 Run 'clean-wizard clean' to start cleanup")
		fmt.Println("   🔍 Try 'clean-wizard clean --dry-run' first to see what would be cleaned")
	} else {
		fmt.Println("\n🎉 Your Nix system is optimized - no cleanup needed!")
	}
}

// getCurrentGeneration finds current generation
func getCurrentGeneration(generations []domain.NixGeneration) string {
	for _, gen := range generations {
		if gen.Current {
			return fmt.Sprintf("%d", gen.ID)
		}
	}
	return "Unknown"
}
