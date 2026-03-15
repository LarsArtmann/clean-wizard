package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var (
	scanTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			MarginBottom(1)
	scanHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Bold(true)
	scanMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// NewScanCommand creates a command that scans for cleanable items.
func NewScanCommand() *cobra.Command {
	var (
		verbose bool
		profile string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for cleanable items",
		Long:  `Scan your system for cleanable items and show size estimates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanCommand(verbose, profile, jsonOut)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed scan information")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Filter results by profile")
	cmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output in JSON format")

	return cmd
}

// runScanCommand executes the scan command.
func runScanCommand(verbose bool, _ string, jsonOutput bool) error {
	ctx := context.Background()

	if !jsonOutput {
		fmt.Println(scanTitleStyle.Render("🔍 Scanning system for cleanable items..."))
		fmt.Println()
	}

	// Get cleaner configurations
	cleanerConfigs := GetCleanerConfigs(ctx)

	// Filter to available cleaners only
	var availableCleaners []CleanerConfig

	for _, cfg := range cleanerConfigs {
		if cfg.Available == CleanerAvailabilityAvailable {
			availableCleaners = append(availableCleaners, cfg)
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

	// Scan each cleaner and collect results using real Scan method
	var (
		totalCleanable uint64
		totalItems     uint
		scanResults    []ScanResult
	)

	for _, cfg := range availableCleaners {
		result := scanCleanerReal(ctx, cfg.Type, verbose)
		scanResults = append(scanResults, result)
		if result.BytesCleanable > 0 {
			totalCleanable += result.BytesCleanable
			totalItems += result.ItemsCount
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
	fmt.Println(scanHeaderStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("💡 Total cleanable: %s (%d items)\n", format.Bytes(int64(totalCleanable)), totalItems)
	fmt.Println(scanHeaderStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	if totalCleanable > 0 {
		fmt.Println()
		fmt.Println("💡 Tip: Run 'clean-wizard clean' to remove these items")
		fmt.Println("   Or use 'clean-wizard clean --dry-run' to preview first")
	}

	return nil
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
func scanCleanerReal(ctx context.Context, cleanerType CleanerType, verbose bool) ScanResult {
	result := ScanResult{
		Name:        getCleanerName(cleanerType),
		Description: getCleanerDescription(cleanerType),
		Icon:        getCleanerIcon(cleanerType),
	}

	// Get cleaner from registry
	registry := cleaner.DefaultRegistry()
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

// getRegistryName returns the registry name for a cleaner type.
func getRegistryName(cleanerType CleanerType) string {
	switch cleanerType {
	case CleanerTypeNix:
		return "nix"
	case CleanerTypeHomebrew:
		return "homebrew"
	case CleanerTypeTempFiles:
		return "tempfiles"
	case CleanerTypeNodePackages:
		return "node"
	case CleanerTypeGoPackages:
		return "go"
	case CleanerTypeCargoPackages:
		return "cargo"
	case CleanerTypeBuildCache:
		return "buildcache"
	case CleanerTypeDocker:
		return "docker"
	case CleanerTypeSystemCache:
		return "systemcache"
	case CleanerTypeLangVersionMgr:
		return "langversion"
	case CleanerTypeProjectsManagementAutomation:
		return "projects"
	case CleanerTypeCompiledBinaries:
		return "compiled-binaries"
	default:
		return ""
	}
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
		fmt.Println(scanMutedStyle.Render("No cleanable items found."))
		return
	}

	// Build table rows
	var rows [][]string
	for _, r := range availableResults {
		sizeStr := format.Bytes(int64(r.BytesCleanable))
		rows = append(rows, []string{
			r.Icon + " " + r.Name,
			fmt.Sprintf("%d", r.ItemsCount),
			sizeStr,
		})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("212")).
					Bold(true).
					Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		}).
		Headers("Cleaner", "Items", "Size").
		Rows(rows...)

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
		TotalBytes uint64 `json:"total_bytes"`
		TotalItems uint   `json:"total_items"`
	}
	type scanJSONOutput struct {
		Results []scanJSONResult `json:"results"`
		Summary scanJSONSummary  `json:"summary"`
	}

	var jsonResults []scanJSONResult
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
