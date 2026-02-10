package commands

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/cobra"
)

// NewScanCommand creates a command that scans for cleanable items.
func NewScanCommand() *cobra.Command {
	var verbose bool
	var profile string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for cleanable items",
		Long:  `Scan your system for cleanable items and show size estimates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScanCommand(cmd, args, verbose, profile)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed scan information")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Filter results by profile")

	return cmd
}

// runScanCommand executes the scan command.
func runScanCommand(cmd *cobra.Command, args []string, verbose bool, profile string) error {
	ctx := context.Background()

	fmt.Println("🔍 Scanning system for cleanable items...")
	fmt.Println()

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
		fmt.Println("   Install package managers (Nix, Homebrew, Docker, etc.) to see cleaning options.")
		return nil
	}

	fmt.Printf("✅ Found %d available cleaner(s)\n\n", len(availableCleaners))

	// Scan each cleaner and collect results
	var totalCleanable uint64
	var scanResults []ScanResult

	for _, cfg := range availableCleaners {
		result := scanCleaner(ctx, cfg.Type, verbose)
		scanResults = append(scanResults, result)
		if result.BytesCleanable > 0 {
			totalCleanable += result.BytesCleanable
		}
	}

	// Print summary
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Scan Results")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, result := range scanResults {
		printScanResult(result, verbose)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("💡 Total cleanable: %s\n", formatBytes(int64(totalCleanable)))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

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
}

// scanCleaner scans a single cleaner for cleanable items.
func scanCleaner(ctx context.Context, cleanerType CleanerType, verbose bool) ScanResult {
	result := ScanResult{
		Name:        getCleanerName(cleanerType),
		Description: getCleanerDescription(cleanerType),
	}

	// Get cleaner from registry
	registry := cleaner.DefaultRegistry()
	name := getRegistryName(cleanerType)

	c, ok := registry.Get(name)
	if !ok {
		return result
	}

	result.Available = toCleanerAvailability(c.IsAvailable(ctx))

	if result.Available != CleanerAvailabilityAvailable {
		return result
	}

	// Run scan by calling the cleaner's Scan method if available
	// For now, we'll use the existing patterns
	switch cleanerType {
	case CleanerTypeNix:
		result = scanNixCleaner(ctx, verbose)
	case CleanerTypeDocker:
		result = scanDockerCleaner(ctx, verbose)
	default:
		// Generic estimation for other cleaners
		result.ItemsCount = 1
		result.BytesCleanable = estimateCleanerSize(cleanerType)
	}

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
	default:
		return ""
	}
}

// scanNixCleaner scans the Nix cleaner.
func scanNixCleaner(ctx context.Context, verbose bool) ScanResult {
	result := ScanResult{
		Name:        "Nix",
		Description: "Clean old Nix store generations and optimize store",
	}

	nixAdapter := cleaner.NewNixCleaner(verbose, false)

	if !nixAdapter.IsAvailable(ctx) {
		result.Available = CleanerAvailabilityUnavailable
		return result
	}

	result.Available = CleanerAvailabilityAvailable
	result.ItemsCount = 5                        // Estimate
	result.BytesCleanable = 50 * 1024 * 1024 * 5 // 50MB per generation, 5 generations

	return result
}

// scanDockerCleaner scans the Docker cleaner.
func scanDockerCleaner(ctx context.Context, verbose bool) ScanResult {
	result := ScanResult{
		Name:        "Docker",
		Description: "Clean Docker images, containers, and volumes",
	}

	dockerCleaner := cleaner.NewDockerCleaner(false, false, domain.DockerPruneAll)

	if !dockerCleaner.IsAvailable(ctx) {
		result.Available = CleanerAvailabilityUnavailable
		return result
	}

	result.Available = CleanerAvailabilityAvailable
	result.ItemsCount = 3                     // Estimate: images, containers, volumes
	result.BytesCleanable = 500 * 1024 * 1024 // Estimate 500MB

	return result
}

// estimateCleanerSize returns a size estimate for a cleaner type.
func estimateCleanerSize(cleanerType CleanerType) uint64 {
	switch cleanerType {
	case CleanerTypeNix:
		return 250 * 1024 * 1024 // 250MB
	case CleanerTypeHomebrew:
		return 150 * 1024 * 1024 // 150MB
	case CleanerTypeNodePackages:
		return 100 * 1024 * 1024 // 100MB
	case CleanerTypeGoPackages:
		return 200 * 1024 * 1024 // 200MB
	case CleanerTypeCargoPackages:
		return 500 * 1024 * 1024 // 500MB
	case CleanerTypeBuildCache:
		return 300 * 1024 * 1024 // 300MB
	case CleanerTypeSystemCache:
		return 400 * 1024 * 1024 // 400MB
	case CleanerTypeLangVersionMgr:
		return 50 * 1024 * 1024 // 50MB
	case CleanerTypeProjectsManagementAutomation:
		return 100 * 1024 * 1024 // 100MB
	default:
		return 0
	}
}

// printScanResult prints a scan result.
func printScanResult(result ScanResult, verbose bool) {
	if result.Available != CleanerAvailabilityAvailable {
		if verbose {
			fmt.Printf("  ⚪ %s: Not available\n", result.Name)
		}
		return
	}

	fmt.Printf("  📦 %s\n", result.Name)
	fmt.Printf("     %s\n", result.Description)

	if result.BytesCleanable > 0 {
		fmt.Printf("     Cleanable: %s", formatBytes(int64(result.BytesCleanable)))
		if result.ItemsCount > 0 {
			fmt.Printf(" (%d item(s))\n", result.ItemsCount)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("     Items: %d\n", result.ItemsCount)
	}
}

// formatBytes formats bytes into human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
