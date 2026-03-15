package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCleanCommand creates a multi-cleaner command with TUI.
func NewCleanCommand() *cobra.Command {
	var (
		dryRun     bool
		verbose    bool
		jsonOutput bool
		mode       string
		profile    string
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean system caches and package managers",
		Long:  `Interactively select and clean system caches, package managers, and temporary data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanCommand(cmd, args, dryRun, verbose, jsonOutput, mode, profile, configPath)
		},
	}

	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually removing anything")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output for cleaner operations")
	cmd.Flags().
		BoolVar(&jsonOutput, "json", false, "Output results in JSON format (non-interactive)")
	cmd.Flags().StringVar(&mode, "mode", "", "Preset mode: quick, standard, or aggressive")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Use a specific configuration profile")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")

	return cmd
}

// runCleanCommand executes the clean command with multi-cleaner TUI.
func runCleanCommand(
	_ *cobra.Command,
	_ []string,
	dryRun, verbose, jsonOutput bool,
	mode, profile, configPath string,
) error {
	ctx := context.Background()

	// Load configuration if specified or use default
	cfg, err := loadConfigForClean(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println(TitleStyle.Render("🧹 Clean Wizard"))
	fmt.Println()

	if dryRun {
		fmt.Println(WarningStyle.Render("⚠️  DRY RUN MODE: No actual changes will be made"))
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

	// Determine which cleaners to run
	var selectedCleaners []CleanerType

	// Priority: profile > mode > interactive
	if profile != "" {
		// Use profile from configuration
		selectedCleaners, err = getProfileCleaners(profile, cfg, availableConfigs)
		if err != nil {
			return fmt.Errorf("profile error: %w", err)
		}
		if !jsonOutput {
			fmt.Printf("📋 Using profile: %s\n", profile)
			fmt.Println()
			for _, ct := range selectedCleaners {
				fmt.Printf("  ✓ %s\n", getCleanerName(ct))
			}
			fmt.Println()
		}
	} else if mode != "" {
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
		// In JSON mode without --mode or --profile, use all available cleaners
		for _, cfg := range availableConfigs {
			selectedCleaners = append(selectedCleaners, cfg.Type)
		}
	} else {
		// Interactive cleaner selection (TUI mode only)
		fmt.Println(InfoStyle.Render("⌨️  Keyboard Shortcuts:"))
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

		err := form.Run()
		if err != nil {
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

		err := confirmForm.Run()
		if err != nil {
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
	var (
		totalBytesFreed   uint64
		totalItemsRemoved uint
		totalItemsFailed  uint
		skippedCleaners   []string
	)

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

		// Debug: Log individual cleaner contributions
		if verbose {
			fmt.Printf(
				"  [DEBUG] %s: %d bytes (%s), %d items\n",
				name,
				result.FreedBytes,
				format.Bytes(int64(result.FreedBytes)),
				result.ItemsRemoved,
			)
		}

		totalBytesFreed += result.FreedBytes
		totalItemsRemoved += result.ItemsRemoved
		totalItemsFailed += result.ItemsFailed
		cleanerResults[name] = result
	}

	duration := time.Since(startTime)

	// Output JSON if requested
	if jsonOutput {
		jsonBytes, err := format.CleanResultsToJSON(
			cleanerResults,
			duration,
			dryRun,
			skippedErrors,
			failedErrors,
		)
		if err != nil {
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}

		fmt.Println(string(jsonBytes))

		return nil
	}

	// Show final results (TUI mode)
	fmt.Println()
	fmt.Println(TitleStyle.Render("🧹 Cleanup Results"))
	fmt.Println()

	if dryRun {
		fmt.Println(WarningStyle.Render("⚠️  DRY RUN: No actual changes were made"))
		fmt.Println()
	}

	// Print results table
	printCleanResultsTable(cleanerResults, totalBytesFreed, totalItemsRemoved, duration)

	// Add encouraging message based on space freed
	if totalBytesFreed > 1_000_000_000 { // > 1 GB
		fmt.Println(SuccessStyle.Render("🎉 Great job! You freed over 1 GB of space!"))
	} else if totalBytesFreed > 100_000_000 { // > 100 MB
		fmt.Println(SuccessStyle.Render("✅ Nice! You freed some space."))
	}

	if dryRun {
		fmt.Println()
		fmt.Println(InfoStyle.Render("💡 Tip: Remove --dry-run flag to actually clean:"))
		fmt.Println("   clean-wizard clean --mode standard")
	}

	// Show errors and warnings
	if totalItemsFailed > 0 || len(skippedCleaners) > 0 || len(failedCleaners) > 0 {
		fmt.Println()
		fmt.Println(WarningStyle.Render("⚠️  Warnings:"))
		if totalItemsFailed > 0 {
			fmt.Printf("   • %d item(s) failed to clean\n", totalItemsFailed)
		}
		if len(skippedCleaners) > 0 {
			fmt.Printf("   • %d cleaner(s) skipped (not available)\n", len(skippedCleaners))
		}
		if len(failedCleaners) > 0 {
			fmt.Printf("   • %d cleaner(s) failed\n", len(failedCleaners))
		}
	}

	return nil
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
	case CleanerTypeCompiledBinaries:
		return "Compiled Binaries"
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
	case CleanerTypeCompiledBinaries:
		return "Clean compiled binary files in project directories"
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
	case CleanerTypeCompiledBinaries:
		return "🔧"
	default:
		return ""
	}
}

// loadConfigForClean loads configuration from the specified path or returns default config.
func loadConfigForClean(configPath string) (*domain.Config, error) {
	if configPath != "" {
		// For custom config path, we'd need to modify config.Load to accept a path
		// For now, use default loading which respects CONFIG_PATH env var
		return config.Load()
	}
	return config.Load()
}

// operationTypeToCleanerType maps domain OperationType to CleanerType.
var operationTypeToCleanerType = map[domain.OperationType]CleanerType{
	domain.OperationTypeNixGenerations:               CleanerTypeNix,
	domain.OperationTypeTempFiles:                    CleanerTypeTempFiles,
	domain.OperationTypeHomebrew:                     CleanerTypeHomebrew,
	domain.OperationTypeNodePackages:                 CleanerTypeNodePackages,
	domain.OperationTypeGoPackages:                   CleanerTypeGoPackages,
	domain.OperationTypeCargoPackages:                CleanerTypeCargoPackages,
	domain.OperationTypeBuildCache:                   CleanerTypeBuildCache,
	domain.OperationTypeDocker:                       CleanerTypeDocker,
	domain.OperationTypeSystemCache:                  CleanerTypeSystemCache,
	domain.OperationTypeSystemTemp:                   CleanerTypeSystemCache,
	domain.OperationTypeProjectsManagementAutomation: CleanerTypeProjectsManagementAutomation,
	domain.OperationTypeProjectExecutables:           CleanerTypeCompiledBinaries,
	domain.OperationTypeCompiledBinaries:             CleanerTypeCompiledBinaries,
}

// getProfileCleaners returns the cleaner types for a given profile name.
func getProfileCleaners(profileName string, cfg *domain.Config, availableConfigs []CleanerConfig) ([]CleanerType, error) {
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}

	// Create a set of available cleaner types for quick lookup
	availableSet := make(map[CleanerType]bool)
	for _, ac := range availableConfigs {
		availableSet[ac.Type] = true
	}

	var cleaners []CleanerType
	for _, op := range profile.Operations {
		if op.Enabled != domain.ProfileStatusEnabled {
			continue
		}
		// Map operation name to OperationType, then to CleanerType
		opType := domain.GetOperationType(op.Name)
		cleanerType, ok := operationTypeToCleanerType[opType]
		if !ok {
			continue // Skip unknown operation types
		}
		if availableSet[cleanerType] {
			cleaners = append(cleaners, cleanerType)
		}
	}

	if len(cleaners) == 0 {
		return nil, fmt.Errorf("profile %q has no available cleaners", profileName)
	}

	return cleaners, nil
}

// printCleanResultsTable prints clean results as a formatted table.
func printCleanResultsTable(results map[string]domain.CleanResult, totalBytes uint64, totalItems uint, duration time.Duration) {
	var rows [][]string
	for name, result := range results {
		if result.FreedBytes > 0 || result.ItemsRemoved > 0 {
			rows = append(rows, []string{
				name,
				strconv.FormatUint(uint64(result.ItemsRemoved), 10),
				format.Bytes(int64(result.FreedBytes)),
			})
		}
	}

	if len(rows) == 0 {
		fmt.Println(InfoStyle.Render("No items were cleaned."))
		return
	}

	// Add summary row
	t := newResultsTable(rows...)

	fmt.Println(t)
	fmt.Println()
	fmt.Printf("📊 Total: %s freed, %s items in %s\n", format.Bytes(int64(totalBytes)), strconv.FormatUint(uint64(totalItems), 10), format.Duration(duration))
}
