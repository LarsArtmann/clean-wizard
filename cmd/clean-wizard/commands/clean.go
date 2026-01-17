package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// CleanerType represents available cleaner types for TUI selection.
type CleanerType string

const (
	CleanerTypeNix        CleanerType = "nix"
	CleanerTypeHomebrew   CleanerType = "homebrew"
	CleanerTypeTempFiles  CleanerType = "tempfiles"
	CleanerTypeNodePackages CleanerType = "nodepackages"
	CleanerTypeGoPackages CleanerType = "gopackages"
	CleanerTypeCargoPackages CleanerType = "cargopackages"
)

// AvailableCleaners returns all available cleaner types.
func AvailableCleaners() []CleanerType {
	return []CleanerType{
		CleanerTypeNix,
		CleanerTypeHomebrew,
		CleanerTypeTempFiles,
		CleanerTypeNodePackages,
		CleanerTypeGoPackages,
		CleanerTypeCargoPackages,
	}
}

// CleanerConfig holds configuration for each cleaner type.
type CleanerConfig struct {
	Type       CleanerType
	Name       string
	Description string
	Icon       string
	Available  bool
}

// GetCleanerConfigs returns all cleaner configurations with availability status.
func GetCleanerConfigs(ctx context.Context) []CleanerConfig {
	configs := []CleanerConfig{
		{
			Type:       CleanerTypeNix,
			Name:       "Nix Generations",
			Description: "Clean old Nix store generations and optimize store",
			Icon:       "❄️",
			Available:  true, // Nix cleaner always available (uses mock data if not)
		},
		{
			Type:       CleanerTypeHomebrew,
			Name:       "Homebrew",
			Description: "Clean Homebrew cache and unused packages",
			Icon:       "🍺",
			Available:  cleaner.NewHomebrewCleaner(false, false, domain.HomebrewModeAll).IsAvailable(ctx),
		},
		{
			Type:       CleanerTypeTempFiles,
			Name:       "Temporary Files",
			Description: "Clean system temporary files older than 7 days",
			Icon:       "🗂️",
			Available:  true, // Temp files cleaner always available
		},
		{
			Type:       CleanerTypeNodePackages,
			Name:       "Node.js Packages",
			Description: "Clean npm, pnpm, yarn, bun caches",
			Icon:       "📦",
			Available:  cleaner.NewNodePackageManagerCleaner(false, false, cleaner.AvailableNodePackageManagers()).IsAvailable(ctx),
		},
		{
			Type:       CleanerTypeGoPackages,
			Name:       "Go Packages",
			Description: "Clean Go module, test, and build caches",
			Icon:       "🐹",
			Available:  cleaner.NewGoCleaner(false, false, true, true, true, true).IsAvailable(ctx),
		},
		{
			Type:       CleanerTypeCargoPackages,
			Name:       "Cargo Packages",
			Description: "Clean Rust/Cargo registry and source caches",
			Icon:       "🦀",
			Available:  cleaner.NewCargoCleaner(false, false).IsAvailable(ctx),
		},
	}
	return configs
}

// NewCleanCommand creates a multi-cleaner command with TUI.
func NewCleanCommand() *cobra.Command {
	var dryRun bool
	var mode string

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean system caches and package managers",
		Long:  `Interactively select and clean system caches, package managers, and temporary data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanCommand(cmd, args, dryRun, mode)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate deletion without actually removing anything")
	cmd.Flags().StringVar(&mode, "mode", "", "Preset mode: quick, standard, or aggressive")

	return cmd
}

// runCleanCommand executes the clean command with multi-cleaner TUI.
func runCleanCommand(cmd *cobra.Command, args []string, dryRun bool, mode string) error {
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
		if cfg.Available {
			availableConfigs = append(availableConfigs, cfg)
		}
	}

	if len(availableConfigs) == 0 {
		return fmt.Errorf("no cleaners available on this system")
	}

	fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableConfigs))

	// If mode is specified, use preset selection
	var selectedCleaners []CleanerType
	if mode != "" {
		selectedCleaners = getPresetSelection(mode, availableConfigs)
		fmt.Printf("🎯 Using preset mode: %s\n", mode)
		fmt.Println()
		for _, ct := range selectedCleaners {
			fmt.Printf("  ✓ %s\n", getCleanerName(ct))
		}
		fmt.Println()
	} else {
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

	for _, cleanerType := range selectedCleaners {
		result, err := runCleaner(ctx, cleanerType, dryRun)
		if err != nil {
			fmt.Printf("  ⚠️  Cleaner %s failed: %v\n", getCleanerName(cleanerType), err)
			continue
		}

		totalBytesFreed += result.FreedBytes
		totalItemsRemoved += result.ItemsRemoved
		totalItemsFailed += result.ItemsFailed
	}

	duration := time.Since(startTime)

	// Show final results
	fmt.Printf("\n✅ Cleanup completed in %s\n", duration.String())
	if dryRun {
		fmt.Println("   (DRY RUN: No actual changes were made)")
	}
	fmt.Printf("   • Cleaned %d item(s)\n", totalItemsRemoved)
	fmt.Printf("   • Freed %s\n", format.Bytes(int64(totalBytesFreed)))
	if totalItemsFailed > 0 {
		fmt.Printf("   • %d item(s) failed to clean\n", totalItemsFailed)
	}

	return nil
}

// runCleaner runs a specific cleaner and returns the result.
func runCleaner(ctx context.Context, cleanerType CleanerType, dryRun bool) (domain.CleanResult, error) {
	fmt.Printf("🔧 Running %s cleaner...\n", getCleanerName(cleanerType))

	switch cleanerType {
	case CleanerTypeNix:
		return runNixCleaner(ctx, dryRun)
	case CleanerTypeHomebrew:
		return runHomebrewCleaner(ctx, dryRun)
	case CleanerTypeTempFiles:
		return runTempFilesCleaner(ctx, dryRun)
	case CleanerTypeNodePackages:
		return runNodePackageManagerCleaner(ctx, dryRun)
	case CleanerTypeGoPackages:
		return runGoCleaner(ctx, dryRun)
	case CleanerTypeCargoPackages:
		return runCargoCleaner(ctx, dryRun)
	default:
		return domain.CleanResult{}, fmt.Errorf("unknown cleaner type: %s", cleanerType)
	}
}

// runNixCleaner executes the Nix cleaner.
func runNixCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	// Import adapter for Nix operations
	nixAdapter := cleaner.NewNixCleaner(false, dryRun)

	if !nixAdapter.IsAvailable(ctx) {
		fmt.Println("  ℹ️  Nix not available - using mock data")
	}

	// Clean old generations (keep last 5)
	keepCount := 5
	result := nixAdapter.CleanOldGenerations(ctx, keepCount)

	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Nix cleaner completed\n")
	return cleanResult, nil
}

// runHomebrewCleaner executes the Homebrew cleaner.
func runHomebrewCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	homebrewCleaner := cleaner.NewHomebrewCleaner(false, dryRun, domain.HomebrewModeAll)

	result := homebrewCleaner.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Homebrew cleaner completed\n")
	return cleanResult, nil
}

// runTempFilesCleaner executes the TempFiles cleaner.
func runTempFilesCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	// Default temp paths and excludes
	defaultTempPaths := []string{filepath.Join("/", "tmp")}
	defaultExcludes := []string{}

	tempFilesCleaner, err := cleaner.NewTempFilesCleaner(false, dryRun, "7d", defaultExcludes, defaultTempPaths)
	if err != nil {
		return domain.CleanResult{}, err
	}

	result := tempFilesCleaner.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Temp files cleaner completed\n")
	return cleanResult, nil
}

// runNodePackageManagerCleaner executes the Node package manager cleaner.
func runNodePackageManagerCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	nodeCleaner := cleaner.NewNodePackageManagerCleaner(false, dryRun, cleaner.AvailableNodePackageManagers())

	result := nodeCleaner.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Node package manager cleaner completed\n")
	return cleanResult, nil
}

// runGoCleaner executes the Go cleaner.
func runGoCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	goCleaner := cleaner.NewGoCleaner(false, dryRun, true, true, true, true)

	result := goCleaner.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Go cleaner completed\n")
	return cleanResult, nil
}

// runCargoCleaner executes the Cargo cleaner.
func runCargoCleaner(ctx context.Context, dryRun bool) (domain.CleanResult, error) {
	cargoCleaner := cleaner.NewCargoCleaner(false, dryRun)

	result := cargoCleaner.Clean(ctx)
	if result.IsErr() {
		return domain.CleanResult{}, result.Error()
	}

	cleanResult := result.Value()
	fmt.Printf("  ✓ Cargo cleaner completed\n")
	return cleanResult, nil
}

// getPresetSelection returns cleaner selection based on preset mode.
func getPresetSelection(mode string, configs []CleanerConfig) []CleanerType {
	switch mode {
	case "quick":
		// Quick mode: Homebrew + Node + Go + TempFiles (fast, no Nix store changes)
		return []CleanerType{
			CleanerTypeHomebrew,
			CleanerTypeNodePackages,
			CleanerTypeGoPackages,
			CleanerTypeTempFiles,
		}
	case "aggressive":
		// Aggressive mode: All cleaners
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
	default:
		return string(cleanerType)
	}
}
