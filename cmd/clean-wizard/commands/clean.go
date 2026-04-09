package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

const (
	// DiskUsageBarWidth is the width of the disk usage bar in characters.
	DiskUsageBarWidth = 15
	// BytesThresholdGB is the threshold for displaying GB freed message.
	BytesThresholdGB = 1_000_000_000
	// BytesThresholdMB is the threshold for displaying MB freed message.
	BytesThresholdMB = 100_000_000
)

// NewCleanCommand creates a multi-cleaner command with TUI.
func NewCleanCommand() *cobra.Command {
	var (
		dryRun           bool
		verbose          bool
		jsonOutput       bool
		skipConfirmation bool
		mode             string
		profile          string
		configPath       string
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean system caches and package managers",
		Long:  `Interactively select and clean system caches, package managers, and temporary data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanCommand(
				cmd,
				args,
				dryRun,
				verbose,
				jsonOutput,
				skipConfirmation,
				mode,
				profile,
				configPath,
			)
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
	cmd.Flags().BoolVarP(&skipConfirmation, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// runCleanCommand executes the clean command with multi-cleaner TUI.
func runCleanCommand(
	_ *cobra.Command,
	_ []string,
	dryRun, verbose, jsonOutput, skipConfirmation bool,
	mode, profile, configPath string,
) error {
	ctx := context.Background()

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

	availableConfigs := getAvailableConfigs(ctx)
	if len(availableConfigs) == 0 {
		return errors.New("no cleaners available on this system")
	}

	fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableConfigs))

	selectedCleaners, err := selectCleaners(profile, mode, cfg, availableConfigs, jsonOutput)
	if err != nil {
		return err
	}

	if selectedCleaners == nil {
		fmt.Println("❌ No cleaners selected. Nothing to clean.")

		return nil
	}

	confirmed, err := confirmExecution(skipConfirmation, dryRun)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("❌ Cancelled. No changes made.")

		return nil
	}

	fmt.Println("\n🧹 Starting cleanup...")

	if dryRun {
		fmt.Println("   (DRY RUN: Simulated only)")
	}

	fmt.Println()

	diskBefore, diskErr := cleaner.GetDiskUsage("/")
	if !jsonOutput && diskErr == nil {
		fmt.Printf(
			"📊 Disk usage before: %s %s\n",
			cleaner.DiskUsageBar(diskBefore, DiskUsageBarWidth),
			cleaner.FormatDiskUsage(diskBefore),
		)
		fmt.Println()
	}

	cr := executeCleaners(ctx, selectedCleaners, dryRun, verbose)

	if jsonOutput {
		return outputJSON(cr, dryRun)
	}

	var diskBeforePtr *cleaner.DiskUsage
	if diskErr == nil {
		diskBeforePtr = &diskBefore
	}

	displayResults(cr, dryRun, diskBeforePtr)

	return nil
}

func outputJSON(cr cleanResult, dryRun bool) error {
	jsonBytes, err := format.CleanResultsToJSON(
		cr.cleanerResults,
		cr.duration,
		dryRun,
		cr.skippedErrors,
		cr.failedErrors,
	)
	if err != nil {
		return fmt.Errorf("failed to generate JSON output: %w", err)
	}

	fmt.Println(string(jsonBytes))

	return nil
}

func getAvailableConfigs(ctx context.Context) []CleanerConfig {
	cleanerConfigs := GetCleanerConfigs(ctx)
	available := make([]CleanerConfig, 0, len(cleanerConfigs))

	for _, cfg := range cleanerConfigs {
		if cfg.Available == CleanerAvailabilityAvailable {
			available = append(available, cfg)
		}
	}

	return available
}

func loadConfigForClean(configPath string) (*domain.Config, error) {
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
	domain.OperationTypeProjectExecutables:           CleanerTypeProjectExecutables,
	domain.OperationTypeCompiledBinaries:             CleanerTypeCompiledBinaries,
	domain.OperationTypeGolangciLintCache:            CleanerTypeGolangciLintCache,
}

func init() {
	for opType, cleanerType := range operationTypeToCleanerType {
		if _, ok := cleanerMetadata[cleanerType]; !ok {
			panic(
				"operationTypeToCleanerType references unknown CleanerType: " +
					string(cleanerType) + " (for " + string(opType) + ")",
			)
		}
	}
}
