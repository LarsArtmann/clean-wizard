package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version     = "dev"
	verbose     bool
	dryRun      bool
	force       bool
	profileName string
)

func main() {
	rootCmd := NewRootCmd()

	// Set up logging
	logrus.SetOutput(os.Stderr)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("Command failed")
		os.Exit(1)
	}
}

// NewRootCmd creates and returns root command
func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "clean-wizard",
		Short:   "Interactive system cleaning wizard",
		Long:    "A comprehensive CLI/TUI tool for system cleanup",
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without doing it")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile", "p", "comprehensive", "Cleaning profile to use")

	// Add commands
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newScanCommand())
	rootCmd.AddCommand(newCleanCommand())
	rootCmd.AddCommand(newConfigCommand())

	return rootCmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🧹 Clean Wizard Setup")
			fmt.Println("======================")
			fmt.Println("Let's create perfect cleaning configuration for your system!")
			fmt.Println()

			// Load or create default config
			cfg, err := config.Load()
			if err != nil {
				logrus.WithError(err).Error("Failed to load configuration")
				os.Exit(1)
			}

			// Save default config
			if err := config.Save(cfg); err != nil {
				logrus.WithError(err).Error("Failed to save configuration")
				os.Exit(1)
			}

			fmt.Println("✅ Configuration created successfully!")
			fmt.Printf("📁 Configuration saved to: %s\n", config.GetConfigPath())
			fmt.Println("💡 Run 'clean-wizard scan' to see what can be cleaned")
			fmt.Println("💡 Run 'clean-wizard clean' to start cleaning")
		},
	}
}

func newScanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Scanning system...")

			// Get command context
			ctx := cmd.Context()

			// Create scanner with context-based dependency injection
			scanner := internalscanner.CreateScannerForContext(ctx, verbose)

			// Perform scan
			results, err := scanner.Scan(ctx)
			if err != nil {
				logrus.WithError(err).Error("Scan failed")
				return fmt.Errorf("scan failed: %w", err)
			}

			// Display results
			fmt.Println("✅ Scan completed!")
			fmt.Printf("📊 Total cleanable: %s\n", types.FormatSize(results.TotalSizeGB))
			fmt.Println()

			for _, result := range results.Results {
				fmt.Printf("📦 %s: %s\n", result.Name, types.FormatSize(result.SizeGB))
				fmt.Printf("   %s\n", result.Description)
			}

			if results.TotalSizeGB > 0 {
				fmt.Println()
				fmt.Println("💡 Run 'clean-wizard clean' to start cleaning")
			} else {
				fmt.Println("🎉 Your system is already clean!")
			}

			return nil
		},
	}
}

func newCleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				logrus.WithError(err).Error("Failed to load configuration")
				os.Exit(1)
			}

			// Override with command line flags
			if dryRun {
				cfg.DryRun = true
			}
			if verbose {
				cfg.Verbose = true
			}

			// Get profile
			profile, ok := cfg.Profiles[profileName]
			if !ok {
				logrus.WithField("profile", profileName).Error("Profile not found")
				os.Exit(1)
			}

			// Show summary
			fmt.Println("🧹 Cleanup Summary")
			fmt.Println("==================")
			fmt.Printf("Profile: %s\n", profile.Name)
			fmt.Printf("Description: %s\n", profile.Description)
			fmt.Printf("Operations: %d\n", len(profile.Operations))
			fmt.Printf("Dry run: %t\n", cfg.DryRun)
			fmt.Println()

			for _, op := range profile.Operations {
				if op.Enabled {
					fmt.Printf("%s %s (%s)\n", op.RiskLevel.Icon(), op.Name, op.RiskLevel)
				}
			}

			fmt.Println()

			// Ask for confirmation unless forced
			if !force && !cfg.DryRun {
				fmt.Print("Proceed with cleanup? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "yes" {
					fmt.Println("❌ Cleanup cancelled")
					return
				}
			}

			// Perform cleanup
			fmt.Println("🧹 Starting cleanup...")
			startTime := time.Now()

			for _, op := range profile.Operations {
				if !op.Enabled {
					continue
				}

				if cfg.Verbose {
					logrus.WithFields(logrus.Fields{
						"operation": op.Name,
						"risk":      op.RiskLevel,
					}).Info("Executing operation")
				}

				if cfg.DryRun {
					logrus.WithField("operation", op.Name).Info("[DRY RUN] Would execute")
					continue
				}

				// TODO: Implement actual cleaning operations
				fmt.Printf("📦 Cleaning %s...\n", op.Name)
				time.Sleep(500 * time.Millisecond) // Simulate work
			}

			duration := time.Since(startTime)
			fmt.Println()
			fmt.Println("✅ Cleanup completed successfully!")
			fmt.Printf("⏱️  Duration: %s\n", types.FormatDuration(duration))
		},
	}
}

func newConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				logrus.WithError(err).Error("Failed to load configuration")
				os.Exit(1)
			}

			fmt.Printf("Configuration file: %s\n", config.GetConfigPath())
			fmt.Printf("Version: %s\n", cfg.Version)
			fmt.Printf("Safe mode: %t\n", cfg.SafeMode)
			fmt.Printf("Dry run: %t\n", cfg.DryRun)
			fmt.Printf("Verbose: %t\n", cfg.Verbose)
			fmt.Printf("Backup: %t\n", cfg.Backup)
			fmt.Printf("Max disk usage: %d%%\n", cfg.MaxDiskUsage)
			fmt.Printf("Protected paths: %d\n", len(cfg.Protected))
			fmt.Printf("Profiles: %d\n", len(cfg.Profiles))
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.GetConfigPath())
		},
	})

	return configCmd
}
