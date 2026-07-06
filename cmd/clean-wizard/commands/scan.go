package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/di"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

// NewScanCommand creates a command that scans for cleanable items.
func NewScanCommand() *cobra.Command {
	var (
		verbose     bool
		profile     string
		jsonOut     bool
		configPath  string
		retries     int
		concurrency int
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for cleanable items",
		Long:  `Scan your system for cleanable items and show size estimates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanCommand(verbose, profile, jsonOut, configPath, retries, concurrency)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed scan information")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Filter results by profile")
	cmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output in JSON format")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")
	cmd.Flags().IntVar(&retries, "retries", 3, "Number of retry attempts per scanner (0=disabled)")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "C", 0, "Max scanners running concurrently (0=unlimited)")

	return cmd
}

// printNixStoreSize prints the Nix store size if available.
func printNixStoreSize(ctx context.Context, registry *cleaner.Registry) {
	nixCleaner, ok := registry.Get("nix")
	if !ok || !nixCleaner.IsAvailable(ctx) {
		return
	}

	nixSizer, ok := nixCleaner.(cleaner.NixStoreSizer)
	if !ok {
		return
	}

	storeSize := nixSizer.GetStoreSize(ctx)
	if storeSize > 0 {
		fmt.Printf("❄️  Nix store size: %s\n", format.Bytes(storeSize))
	}
}

// runScanCommand executes the scan command.
func runScanCommand(verbose bool, profile string, jsonOutput bool, configPath string, retries, concurrency int) error {
	ctx := context.Background()

	if profile != "" {
		fmt.Printf("⚠️  Warning: --profile %q is not yet supported for scan; showing all available cleaners\n", profile)
	}

	cfg, err := loadConfigFromPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	container, cleanup := di.New()
	defer cleanup()

	settings := di.RunSettings{Verbose: verbose, DryRun: false, MaxConcurrency: concurrency}
	if err := di.RegisterAllServices(container.Injector(), cfg, settings); err != nil {
		return fmt.Errorf("failed to register DI services: %w", err)
	}

	registry, err := di.CleanerRegistry(container.Injector())
	if err != nil {
		return fmt.Errorf("failed to resolve cleaner registry from DI: %w", err)
	}

	if !jsonOutput {
		fmt.Println(TitleStyle.Render("🔍 Scanning system for cleanable items..."))
		fmt.Println()
	}

	// Get cleaner configurations
	cleanerConfigs := GetCleanerConfigs(ctx, registry)

	// Filter to available cleaners only
	var availableCleaners []CleanerConfig

	for _, c := range cleanerConfigs {
		if c.Available == CleanerAvailabilityAvailable {
			availableCleaners = append(availableCleaners, c)
		}
	}

	if len(availableCleaners) == 0 {
		fmt.Println("ℹ️  No cleanable items found on this system.")
		fmt.Println(
			"   Install package managers (Nix, Homebrew, Docker, etc.) to see cleaning options.",
		)

		return nil
	}

	if !jsonOutput {
		fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableCleaners))
	}

	// Run scans via the workflow engine for parallel execution
	selectedNames := make([]string, len(availableCleaners))
	for i, c := range availableCleaners {
		selectedNames[i] = getRegistryName(c.Type)
	}

	var runOpts []execution.RunOption
	if verbose {
		runOpts = append(runOpts, execution.WithVerbose(true))
	}
	if concurrency > 0 {
		runOpts = append(runOpts, execution.WithMaxConcurrency(concurrency))
	}
	if retries > 0 {
		runOpts = append(runOpts, execution.WithRetry(&execution.RetryConfig{
			MaxAttempts:    retries,
			InitialBackoff: 2 * time.Second,
			MaxBackoff:     30 * time.Second,
		}))
	}

	wr, err := execution.RunScans(ctx, registry, selectedNames, runOpts...)
	if err != nil {
		return fmt.Errorf("scan workflow execution failed: %w", err)
	}

	// Build display results from the workflow result
	scanResults := buildScanResults(wr, availableCleaners)

	var (
		totalCleanable uint64
		totalItems     uint
	)
	for _, r := range scanResults {
		if r.BytesCleanable > 0 {
			totalCleanable += r.BytesCleanable
			totalItems += r.ItemsCount
		}
	}

	// Output JSON if requested
	if jsonOutput {
		outputScanJSON(scanResults, totalCleanable, totalItems)

		return nil
	}

	// Print table output
	printScanTable(scanResults, verbose)

	// Print summary
	fmt.Println()
	fmt.Println(HeaderStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf(
		"💡 Total cleanable: %s (%d items)\n",
		format.Bytes(int64(totalCleanable)),
		totalItems,
	)

	// Show Nix store size if available
	printNixStoreSize(ctx, registry)

	fmt.Println(HeaderStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	if totalCleanable > 0 {
		fmt.Println()
		fmt.Println("💡 Tip: Run 'clean-wizard clean' to remove these items")
		fmt.Println("   Or use 'clean-wizard clean --dry-run' to preview first")
	}

	return nil
}

// buildScanResults converts a WorkflowResult into display-ready ScanResult structs.
func buildScanResults(wr *execution.WorkflowResult, available []CleanerConfig) []ScanResult {
	stepByName := make(map[string]execution.StepResult, len(wr.Steps))
	for _, s := range wr.Steps {
		stepByName[s.Name] = s
	}

	results := make([]ScanResult, 0, len(available))
	for _, cfg := range available {
		sr := ScanResult{ //nolint:exhaustruct
			Name:        cfg.Name,
			Description: cfg.Description,
			Icon:        cfg.Icon,
			Available:   cfg.Available,
		}

		regName := getRegistryName(cfg.Type)
		if step, ok := stepByName[regName]; ok && step.Err == nil {
			sr.ItemsCount = step.Clean.ItemsRemoved
			sr.BytesCleanable = step.Clean.FreedBytes
		}

		results = append(results, sr)
	}

	return results
}

// ScanResult holds the scan result for a cleaner.
type ScanResult struct {
	Name           string
	Available      CleanerAvailability
	ItemsCount     uint
	BytesCleanable uint64
	Description    string
	Icon           string
}

func getRegistryName(cleanerType CleanerType) string {
	if m, ok := cleanerMetadata[cleanerType]; ok {
		return m.RegistryName
	}

	return ""
}

// printScanTable prints scan results as a formatted table.
func printScanTable(results []ScanResult, _ bool) {
	// Filter to only available cleaners with items
	var availableResults []ScanResult

	for _, r := range results {
		if r.Available == CleanerAvailabilityAvailable && r.BytesCleanable > 0 {
			availableResults = append(availableResults, r)
		}
	}

	if len(availableResults) == 0 {
		fmt.Println(MutedStyle.Render("No cleanable items found."))

		return
	}

	// Build table rows
	var rows [][]string

	for _, r := range availableResults {
		sizeStr := format.Bytes(int64(r.BytesCleanable))
		rows = append(rows, []string{
			r.Icon + " " + r.Name,
			strconv.FormatUint(uint64(r.ItemsCount), 10),
			sizeStr,
		})
	}

	t := newResultsTable(rows...)

	fmt.Println(t)
}

// outputScanJSON outputs scan results in JSON format.
func outputScanJSON(results []ScanResult, totalBytes uint64, totalItems uint) {
	type scanJSONResult struct {
		Name      string `json:"name"`
		Items     uint   `json:"items"`
		Bytes     uint64 `json:"bytes"`
		Available bool   `json:"available"`
	}

	type scanJSONSummary struct {
		TotalBytes uint64 `json:"totalBytes"`
		TotalItems uint   `json:"totalItems"`
	}

	type scanJSONOutput struct {
		Results []scanJSONResult `json:"results"`
		Summary scanJSONSummary  `json:"summary"`
	}

	jsonResults := make([]scanJSONResult, 0, len(results))
	for _, r := range results {
		jsonResults = append(jsonResults, scanJSONResult{
			Name:      r.Name,
			Items:     r.ItemsCount,
			Bytes:     r.BytesCleanable,
			Available: r.Available == CleanerAvailabilityAvailable,
		})
	}

	output := scanJSONOutput{
		Results: jsonResults,
		Summary: scanJSONSummary{
			TotalBytes: totalBytes,
			TotalItems: totalItems,
		},
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("{\"error\": %q}\n", err.Error())

		return
	}

	fmt.Println(string(jsonBytes))
}
