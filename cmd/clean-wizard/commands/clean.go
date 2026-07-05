package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/di"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
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

// Sentinel errors for clean command.
var (
	ErrNoCleanersAvailable  = errors.New("no cleaners available on this system")
	ErrNoConfigPathProvided = errors.New("no config path provided")
)

// NewCleanCommand creates a multi-cleaner command with TUI.
func NewCleanCommand() *cobra.Command {
	validateOperationTypeMapping()

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

// printDiskUsage prints disk usage before cleanup if available.
func printDiskUsage(diskUsage *cleaner.DiskUsage, jsonOutput bool) {
	if jsonOutput || diskUsage == nil {
		return
	}
	fmt.Printf(
		"📊 Disk usage before: %s %s\n",
		cleaner.DiskUsageBar(*diskUsage, DiskUsageBarWidth),
		cleaner.FormatDiskUsage(*diskUsage),
	)
	fmt.Println()
}

// printDryRunHeader prints the header and dry-run warning.
func printDryRunHeader(dryRun bool) {
	fmt.Println(TitleStyle.Render("🧹 Clean Wizard"))
	fmt.Println()

	if dryRun {
		fmt.Println(WarningStyle.Render("⚠️  DRY RUN MODE: No actual changes will be made"))
		fmt.Println()
	}
}

// printCleanStart prints the start of cleanup.
func printCleanStart(dryRun bool) {
	fmt.Println("\n🧹 Starting cleanup...")

	if dryRun {
		fmt.Println("   (DRY RUN: Simulated only)")
	}

	fmt.Println()
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
		return fmt.Errorf(
			"failed to load configuration for mode=%v, profile=%v: %w",
			mode,
			profile,
			err,
		)
	}

	container, cleanup := di.New()
	defer cleanup()

	settings := di.RunSettings{Verbose: verbose, DryRun: dryRun}
	if err := di.RegisterAllServices(container.Injector(), cfg, settings); err != nil {
		return fmt.Errorf("failed to register DI services: %w", err)
	}

	registry, err := di.CleanerRegistry(container.Injector())
	if err != nil {
		return fmt.Errorf("failed to resolve cleaner registry from DI: %w", err)
	}

	printDryRunHeader(dryRun)

	availableConfigs := getAvailableConfigs(ctx, registry)
	if len(availableConfigs) == 0 {
		return ErrNoCleanersAvailable
	}

	fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableConfigs))

	selectedCleaners, err := selectCleaners(profile, mode, cfg, availableConfigs, jsonOutput)
	if err != nil {
		return fmt.Errorf("mode=%v, profile=%v: %w", mode, profile, err)
	}

	if selectedCleaners == nil {
		fmt.Println("❌ No cleaners selected. Nothing to clean.")

		return nil
	}

	confirmed, err := confirmExecution(skipConfirmation, dryRun)
	if err != nil {
		return fmt.Errorf("mode=%v, profile=%v: %w", mode, profile, err)
	}

	if !confirmed {
		fmt.Println("❌ Cancelled. No changes made.")

		return nil
	}

	printCleanStart(dryRun)

	diskBefore, diskErr := cleaner.GetDiskUsage("/")
	var diskBeforePtr *cleaner.DiskUsage
	if diskErr == nil {
		diskBeforePtr = &diskBefore
	}

	printDiskUsage(diskBeforePtr, jsonOutput)

	selectedNames := cleanerTypesToNames(selectedCleaners)

	wr, err := execution.RunCleaners(ctx, registry, selectedNames, execution.WithVerbose(verbose))
	if err != nil {
		return fmt.Errorf("clean workflow execution failed: %w", err)
	}

	if jsonOutput {
		return outputJSON(wr, dryRun)
	}

	displayResults(wr, dryRun, diskBeforePtr)

	return nil
}

func outputJSON(wr *execution.WorkflowResult, dryRun bool) error {
	skipped := make(map[string]error)
	for _, s := range wr.Skipped() {
		skipped[s.Name] = s.Err
	}

	failed := make(map[string]error)
	for _, s := range wr.Failed() {
		failed[s.Name] = s.Err
	}

	jsonBytes, err := format.CleanResultsToJSON(
		wr.CleanResultsMap(),
		wr.Duration,
		dryRun,
		skipped,
		failed,
	)
	if err != nil {
		return fmt.Errorf("failed to generate JSON output: %w", err)
	}

	fmt.Println(string(jsonBytes))

	return nil
}

func getAvailableConfigs(ctx context.Context, registry *cleaner.Registry) []CleanerConfig {
	cleanerConfigs := GetCleanerConfigs(ctx, registry)
	available := make([]CleanerConfig, 0, len(cleanerConfigs))

	for _, cfg := range cleanerConfigs {
		if cfg.Available == CleanerAvailabilityAvailable {
			available = append(available, cfg)
		}
	}

	return available
}

func loadConfigForClean(configPath string) (*domain.Config, error) {
	if configPath != "" {
		cfg, err := config.LoadFromPath(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		return cfg, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

// cleanerTypesToNames converts a slice of CleanerType to a slice of string
// for the execution layer, which works with cleaner names.
func cleanerTypesToNames(types []CleanerType) []string {
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = string(t)
	}
	return names
}

// operationTypeToCleanerType maps domain OperationType to CleanerType.
var operationTypeToCleanerType = map[domain.OperationType]CleanerType{ //nolint:gochecknoglobals
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

// validateOperationTypeMapping panics at package init if operationTypeToCleanerType
// references a CleanerType not present in cleanerMetadata.
func validateOperationTypeMapping() {
	for opType, cleanerType := range operationTypeToCleanerType {
		if _, ok := cleanerMetadata[cleanerType]; !ok {
			panic(
				"operationTypeToCleanerType references unknown CleanerType: " +
					string(cleanerType) + " (for " + string(opType) + ")",
			)
		}
	}
}
