package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCleanCommand creates a multi-cleaner command with TUI.
func NewCleanCommand() *cobra.Command {
	var dryRun bool
	var verbose bool
	var jsonOutput bool
	var mode string

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean system caches and package managers",
		Long:  `Interactively select and clean system caches, package managers, and temporary data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanCommand(cmd, args, dryRun, verbose, jsonOutput, mode)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually removing anything")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output for cleaner operations")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format (non-interactive)")
	cmd.Flags().StringVar(&mode, "mode", "", "Preset mode: quick, standard, or aggressive")

	return cmd
}

// runCleanCommand executes the clean command with multi-cleaner TUI.
func runCleanCommand(cmd *cobra.Command, args []string, dryRun, verbose, jsonOutput bool, mode string) error {
	ctx := context.Background()

	fmt.Println("🔍 Detecting available cleaners...")
	if dryRun {
		fmt.Println("⚠️  DRY RUN MODE: No actual changes will be made")
		fmt.Println()
	}

	// Get cleaner configurations with availability status
	cleanerConfigs := GetCleanerConfigs(ctx)

	// Filter to available cleaners only
	availableConfigs := make([]CleanerConfig, 0, len(cleanerConfigs))
	for _, cfg := range cleanerConfigs {
		if cfg.Available == CleanerAvailabilityAvailable {
			availableConfigs = append(availableConfigs, cfg)
		}
	}

	if len(availableConfigs) == 0 {
		return errors.New("no cleaners available on this system")
	}

	fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableConfigs))

	// If mode is specified, use preset selection
	var selectedCleaners []CleanerType
	if mode != "" {
		selectedCleaners = getPresetSelection(mode, availableConfigs)
		if !jsonOutput {
			fmt.Printf("🎯 Using preset mode: %s\n", mode)
			fmt.Println()
			for _, ct := range selectedCleaners {
				fmt.Printf("  ✓ %s\n", getCleanerName(ct))
			}
			fmt.Println()
		}
	} else if jsonOutput {
		// In JSON mode without --mode, use all available cleaners
		for _, cfg := range availableConfigs {
			selectedCleaners = append(selectedCleaners, cfg.Type)
		}
	} else {
		// Interactive cleaner selection (TUI mode only)
		fmt.Println("⌨️  Keyboard Shortcuts:")
		fmt.Println("   ↑↓ : Navigate  |  Space : Select  |  Enter : Confirm  |  Esc : Cancel")
		fmt.Println()
		// Interactive cleaner selection
		var selectedTypes []CleanerType
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[CleanerType]().
					Title("Select cleaners to run").
					Description("Choose which cleaners to execute (Space to select, Enter to confirm)").
					Options(
						func() []huh.Option[CleanerType] {
							opts := make([]huh.Option[CleanerType], len(availableConfigs))
							for i, cfg := range availableConfigs {
								desc := fmt.Sprintf("%s %s", cfg.Description, cfg.Icon)
								opts[i] = huh.NewOption(desc, cfg.Type)
							}
							return opts
						}()...,
					).
					Value(&selectedTypes),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form error: %w", err)
		}

		if len(selectedTypes) == 0 {
			fmt.Println("❌ No cleaners selected. Nothing to clean.")
			return nil
		}

		selectedCleaners = selectedTypes
	}

	// Confirm before running
	if !dryRun {
		var confirm bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Run selected cleaners?").
					Affirmative("Yes, run them").
					Negative("No, cancel").
					Value(&confirm),
			),
		)

		if err := confirmForm.Run(); err != nil {
			return fmt.Errorf("confirmation error: %w", err)
		}

		if !confirm {
			fmt.Println("❌ Cancelled. No changes made.")
			return nil
		}
	}

	// Run selected cleaners
	fmt.Println("\n🧹 Starting cleanup...")
	if dryRun {
		fmt.Println("   (DRY RUN: Simulated only)")
	}
	fmt.Println()

	startTime := time.Now()

	// Aggregate results from all cleaners
	var totalBytesFreed uint64
	var totalItemsRemoved uint
	var totalItemsFailed uint
	var skippedCleaners []string
	skippedErrors := make(map[string]error)
	var failedCleaners []struct {
		name  string
		error string
	}
	failedErrors := make(map[string]error)
	cleanerResults := make(map[string]domain.CleanResult)

	for _, cleanerType := range selectedCleaners {
		result, err := runCleaner(ctx, cleanerType, dryRun, verbose)
		name := getCleanerName(cleanerType)

		if err != nil {
			errMsg := err.Error()

			// Check if this is a "not available" error vs actual failure
			if isNotAvailableError(errMsg) {
				skippedCleaners = append(skippedCleaners, name)
				skippedErrors[name] = err
				if !jsonOutput {
					fmt.Printf("  ℹ️  Skipped %s: %s\n", name, errMsg)
				}
			} else {
				failedCleaners = append(failedCleaners, struct {
					name  string
					error string
				}{name: name, error: errMsg})
				failedErrors[name] = err
				if !jsonOutput {
					fmt.Printf("  ❌ Cleaner %s failed: %s\n", name, errMsg)
				}
			}
			continue
		}

		totalBytesFreed += result.FreedBytes
		totalItemsRemoved += result.ItemsRemoved
		totalItemsFailed += result.ItemsFailed
		cleanerResults[name] = result
	}

	duration := time.Since(startTime)

	// Output JSON if requested
	if jsonOutput {
		jsonBytes, err := format.CleanResultsToJSON(cleanerResults, duration, dryRun, skippedErrors, failedErrors)
		if err != nil {
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Show final results (TUI mode)
	fmt.Printf("\n✅ Cleanup completed in %s\n", duration.String())
	if dryRun {
		fmt.Println("   (DRY RUN: No actual changes were made)")
	}
	fmt.Printf("   • Cleaned %d item(s)\n", totalItemsRemoved)
	fmt.Printf("   • Freed %s\n", format.Bytes(int64(totalBytesFreed)))

	// Add encouraging message based on space freed
	if totalBytesFreed > 1_000_000_000 { // > 1 GB
		fmt.Println("\n🎉 Great job! You freed over 1 GB of space!")
	} else if totalBytesFreed > 100_000_000 { // > 100 MB
		fmt.Println("\n✅ Nice! You freed some space.")
	}

	if dryRun {
		fmt.Println("\n💡 Tip: Remove --dry-run flag to actually clean:")
		fmt.Println("   clean-wizard clean --mode standard")
	}

	// Show errors and warnings
	if totalItemsFailed > 0 {
		fmt.Printf("   • %d item(s) failed to clean\n", totalItemsFailed)
	}
	if len(skippedCleaners) > 0 {
		fmt.Printf("   • %d cleaner(s) skipped (not available)\n", len(skippedCleaners))
	}
	if len(failedCleaners) > 0 {
		fmt.Printf("   • %d cleaner(s) failed\n", len(failedCleaners))
	}

	return nil
}

// printCleanerResult displays detailed results for a cleaner.
func printCleanerResult(name string, result domain.CleanResult, dryRun bool) {
	details := ""
	if result.ItemsRemoved > 0 {
		if dryRun {
			details = fmt.Sprintf("would clean %d item(s)", result.ItemsRemoved)
		} else {
			details = fmt.Sprintf("cleaned %d item(s), freed %s", result.ItemsRemoved, format.Bytes(int64(result.FreedBytes)))
		}
	} else if result.FreedBytes > 0 {
		details = "freed " + format.Bytes(int64(result.FreedBytes))
	} else {
		details = "no items to clean"
	}

	fmt.Printf("  ✓ %s cleaner: %s\n", name, details)
}

// isNotAvailableError checks if an error indicates a cleaner is not available.
func isNotAvailableError(errMsg string) bool {
	lowerMsg := strings.ToLower(errMsg)
	unavailableKeywords := []string{
		"not available",
		"not found",
		"not installed",
		"command not found",
		"no such file or directory",
	}

	for _, keyword := range unavailableKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

// getPresetSelection returns cleaner selection based on preset mode.
func getPresetSelection(mode string, configs []CleanerConfig) []CleanerType {
	switch mode {
	case "quick":
		// Quick mode: Homebrew + Node + Go + TempFiles + BuildCache (fast, no Nix/Docker/System changes)
		return []CleanerType{
			CleanerTypeHomebrew,
			CleanerTypeNodePackages,
			CleanerTypeGoPackages,
			CleanerTypeTempFiles,
			CleanerTypeBuildCache,
		}
	case "aggressive":
		// Aggressive mode: All cleaners including Docker and System Cache
		var allTypes []CleanerType
		for _, cfg := range configs {
			allTypes = append(allTypes, cfg.Type)
		}
		return allTypes
	case "standard":
		fallthrough
	default:
		// Standard mode: All available cleaners
		var allTypes []CleanerType
		for _, cfg := range configs {
			allTypes = append(allTypes, cfg.Type)
		}
		return allTypes
	}
}

// getCleanerName returns the display name for a cleaner type.
func getCleanerName(cleanerType CleanerType) string {
	switch cleanerType {
	case CleanerTypeNix:
		return "Nix"
	case CleanerTypeHomebrew:
		return "Homebrew"
	case CleanerTypeTempFiles:
		return "Temp Files"
	case CleanerTypeNodePackages:
		return "Node.js Packages"
	case CleanerTypeGoPackages:
		return "Go Packages"
	case CleanerTypeCargoPackages:
		return "Cargo Packages"
	case CleanerTypeBuildCache:
		return "Build Cache"
	case CleanerTypeDocker:
		return "Docker"
	case CleanerTypeSystemCache:
		return "System Cache"
	case CleanerTypeLangVersionMgr:
		return "Language Version Managers"
	case CleanerTypeProjectsManagementAutomation:
		return "Projects Management Automation"
	default:
		return string(cleanerType)
	}
}

// getCleanerDescription returns the description for a cleaner type.
func getCleanerDescription(cleanerType CleanerType) string {
	switch cleanerType {
	case CleanerTypeNix:
		return "Clean old Nix store generations and optimize store"
	case CleanerTypeHomebrew:
		return "Clean Homebrew cache and unused packages"
	case CleanerTypeTempFiles:
		return "Clean /tmp files (not dirs) older than 7 days"
	case CleanerTypeNodePackages:
		return "Clean npm, pnpm, yarn, bun caches"
	case CleanerTypeGoPackages:
		return "Clean Go module, test, and build caches"
	case CleanerTypeCargoPackages:
		return "Clean Rust/Cargo registry and source caches"
	case CleanerTypeBuildCache:
		return "Clean Gradle, Maven, and SBT caches"
	case CleanerTypeDocker:
		return "Clean Docker images, containers, and volumes"
	case CleanerTypeSystemCache:
		return "Clean macOS Spotlight, Xcode, CocoaPods caches"
	case CleanerTypeLangVersionMgr:
		return "Clean NVM, Pyenv, and Rbenv versions (WARNING: Destructive)"
	case CleanerTypeProjectsManagementAutomation:
		return "Clear projects-management-automation cache"
	default:
		return ""
	}
}

// getCleanerIcon returns the icon for a cleaner type.
func getCleanerIcon(cleanerType CleanerType) string {
	switch cleanerType {
	case CleanerTypeNix:
		return "❄️"
	case CleanerTypeHomebrew:
		return "🍺"
	case CleanerTypeTempFiles:
		return "🗂️"
	case CleanerTypeNodePackages:
		return "📦"
	case CleanerTypeGoPackages:
		return "🐹"
	case CleanerTypeCargoPackages:
		return "🦀"
	case CleanerTypeBuildCache:
		return "🔨"
	case CleanerTypeDocker:
		return "🐳"
	case CleanerTypeSystemCache:
		return "⚙️"
	case CleanerTypeLangVersionMgr:
		return "🗑️"
	case CleanerTypeProjectsManagementAutomation:
		return "⚙️"
	default:
		return ""
	}
}
