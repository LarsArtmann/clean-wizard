package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// homebrewCommandTimeout is the timeout for Homebrew operations.
const homebrewCommandTimeout = 5 * time.Minute

// HomebrewCleaner handles Homebrew package manager cleanup with proper type safety.
type HomebrewCleaner struct {
	verbose    bool
	dryRun     bool
	unusedOnly domain.HomebrewMode
}

// NewHomebrewCleaner creates Homebrew cleaner with proper configuration.
func NewHomebrewCleaner(verbose, dryRun bool, unusedOnly domain.HomebrewMode) *HomebrewCleaner {
	return &HomebrewCleaner{
		verbose:    verbose,
		dryRun:     dryRun,
		unusedOnly: unusedOnly,
	}
}

// Type returns operation type for Homebrew cleaner.
func (hbc *HomebrewCleaner) Type() domain.OperationType {
	return domain.OperationTypeHomebrew
}

// Name returns the unique identifier for this cleaner.
func (hbc *HomebrewCleaner) Name() string {
	return "homebrew"
}

// execWithTimeout executes a Homebrew command with timeout.
func (hbc *HomebrewCleaner) execWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeoutCtx, cancel := context.WithTimeout(ctx, homebrewCommandTimeout)
	_ = cancel // will be called by cmd.Wait() or context usage
	return exec.CommandContext(timeoutCtx, name, arg...)
}

// IsAvailable checks if Homebrew cleaner is available.
func (hbc *HomebrewCleaner) IsAvailable(ctx context.Context) bool {
	// Check if brew command exists
	_, err := exec.LookPath("brew")
	return err == nil
}

// ValidateSettings validates Homebrew cleaner settings with type safety.
func (hbc *HomebrewCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.Homebrew == nil {
		return nil // Settings are optional
	}

	// Validate unused_only mode
	if settings.Homebrew.UnusedOnly != domain.HomebrewModeUnusedOnly &&
		settings.Homebrew.UnusedOnly != domain.HomebrewModeAll {
		return errors.New("invalid unused_only mode: must be either 'unused_only' or 'all'")
	}

	return nil
}

// Scan scans for Homebrew packages that can be cleaned.
func (hbc *HomebrewCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	if !hbc.IsAvailable(ctx) {
		return result.Err[[]domain.ScanItem](errors.New("homebrew not available"))
	}

	items := make([]domain.ScanItem, 0)

	// Get list of outdated packages
	outdatedCmd := hbc.execWithTimeout(ctx, "brew", "outdated")
	output, err := outdatedCmd.CombinedOutput()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to check for outdated packages: %w", err))
	}

	// Parse output
	lines := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for line := range lines {
		if line == "" {
			continue
		}

		// Parse package name and versions
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			packageName := fields[0]
			currentVersion := fields[1]

			items = append(items, domain.ScanItem{
				Path:     "homebrew://" + packageName,
				Size:     0, // Size unknown without checking
				Created:  time.Time{},
				ScanType: domain.ScanTypeHomebrew,
			})

			if hbc.verbose {
				fmt.Printf("Found outdated package: %s (current: %s)\n", packageName, currentVersion)
			}
		}
	}

	return result.Ok(items)
}

// Clean removes old Homebrew packages with proper type safety.
func (hbc *HomebrewCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !hbc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](errors.New("homebrew not available"))
	}

	if hbc.dryRun {
		// Dry-run not supported for Homebrew - print message and return
		fmt.Println("⚠️  Dry-run mode is not yet supported for Homebrew cleanup.")
		fmt.Println("   Homebrew does not provide a native dry-run feature.")
		fmt.Println("   To see what would be cleaned, use: brew cleanup -n (manual check)")
		cleanResult := domain.CleanResult{
			FreedBytes:   0,
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyDryRunType),
		}
		return result.Ok(cleanResult)
	}

	// Real cleaning implementation
	startTime := time.Now()

	// Determine which cleanup commands to run
	commands := []string{}

	// Always run cleanup
	commands = append(commands, "cleanup")

	// Prune based on settings
	if hbc.unusedOnly == domain.HomebrewModeUnusedOnly {
		commands = append(commands, "prune")
	}

	itemsRemoved := 0
	bytesFreed := int64(0)
	itemsFailed := 0

	for _, cmd := range commands {
		var cleanCmd *exec.Cmd
		switch cmd {
		case "cleanup":
			cleanCmd = hbc.execWithTimeout(ctx, "brew", "cleanup")
			if hbc.verbose {
				fmt.Println("🔧 Running 'brew cleanup'")
			}
		case "prune":
			cleanCmd = hbc.execWithTimeout(ctx, "brew", "prune")
			if hbc.verbose {
				fmt.Println("🔧 Running 'brew prune'")
			}
		}

		output, err := cleanCmd.CombinedOutput()
		if err != nil {
			itemsFailed++
			if hbc.verbose {
				fmt.Printf("Warning: 'brew %s' failed: %v\n", cmd, string(output))
			}
			continue
		}

		// Count items removed from output
		lines := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
		for line := range lines {
			if strings.Contains(line, "removed") || strings.Contains(line, "deleted") {
				itemsRemoved++
			}
		}

		if hbc.verbose {
			fmt.Printf("✅ 'brew %s' completed\n", cmd)
		}
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	}

	return result.Ok(cleanResult)
}
