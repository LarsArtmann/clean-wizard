package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/di"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

// NewScanCommand creates a command that scans for cleanable items.
func NewScanCommand() *cobra.Command {
	var (
		verbose    bool
		profile    string
		jsonOut    bool
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for cleanable items",
		Long:  `Scan your system for cleanable items and show size estimates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanCommand(verbose, profile, jsonOut, configPath)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed scan information")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Filter results by profile")
	cmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output in JSON format")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")

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
func runScanCommand(verbose bool, _ string, jsonOutput bool, configPath string) error {
	ctx := context.Background()

	cfg, err := loadConfigForScan(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	container, cleanup := di.New()
	defer cleanup()

	settings := di.RunSettings{Verbose: verbose, DryRun: false}
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

	wr, err := execution.RunScans(ctx, registry, selectedNames, execution.WithVerbose(verbose))
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

// loadConfigForScan loads config from path or default.
func loadConfigForScan(configPath string) (*domain.Config, error) {
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

// ScanResult holds the scan result for a cleaner.
type ScanResult struct {
	Name           string
	Available      CleanerAvailability
	ItemsCount     uint
	BytesCleanable uint64
	Description    string
	Icon           string
}

// scanCleanerReal scans a cleaner using the real Scan method from the cleaner interface.
func scanCleanerReal(
	ctx context.Context,
	registry *cleaner.Registry,
	cleanerType CleanerType,
	verbose bool,
) ScanResult {
	result := ScanResult{ //nolint:exhaustruct
		Name:        getCleanerName(cleanerType),
		Description: getCleanerDescription(cleanerType),
		Icon:        getCleanerIcon(cleanerType),
	}

	name := getRegistryName(cleanerType)

	c, ok := registry.Get(name)
	if !ok {
		result.Available = CleanerAvailabilityUnavailable

		return result
	}

	result.Available = toCleanerAvailability(c.IsAvailable(ctx))

	if result.Available != CleanerAvailabilityAvailable {
		return result
	}

	// Use the real Scan method
	scanRes := c.Scan(ctx)

	items, err := scanRes.Unwrap()
	if err != nil {
		if verbose {
			fmt.Printf("  ⚠️  Scan error for %s: %v\n", result.Name, err)
		}

		return result
	}

	result.ItemsCount = uint(len(items))

	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}

	result.BytesCleanable = uint64(totalSize)

	return result
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
