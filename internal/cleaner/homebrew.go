package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
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
		return hbc.handleDryRun()
	}

	startTime := time.Now()
	cacheDir := hbc.getCacheDir(ctx)
	commands := hbc.buildCleanupCommands()

	itemsRemoved, itemsFailed, bytesFreed := hbc.executeCleanup(ctx, commands, cacheDir)

	return result.Ok(conversions.NewCleanResultWithFailures(
		domain.CleanStrategyType(domain.StrategyConservativeType),
		itemsRemoved,
		itemsFailed,
		bytesFreed,
		time.Since(startTime),
	))
}

// handleDryRun returns result for dry-run mode.
func (hbc *HomebrewCleaner) handleDryRun() result.Result[domain.CleanResult] {
	fmt.Println("⚠️  Dry-run mode is not yet supported for Homebrew cleanup.")
	fmt.Println("   Homebrew does not provide a native dry-run feature.")
	fmt.Println("   To see what would be cleaned, use: brew cleanup -n (manual check)")
	return result.Ok(conversions.NewCleanResult(domain.CleanStrategyType(domain.StrategyDryRunType), 0, 0))
}

// getCacheDir returns the Homebrew cache directory.
func (hbc *HomebrewCleaner) getCacheDir(ctx context.Context) string {
	if cacheOutput, err := hbc.execWithTimeout(ctx, "brew", "--cache").Output(); err == nil {
		return strings.TrimSpace(string(cacheOutput))
	}
	return ""
}

// buildCleanupCommands returns the list of cleanup commands to run.
func (hbc *HomebrewCleaner) buildCleanupCommands() []string {
	commands := []string{"cleanup"}
	if hbc.unusedOnly == domain.HomebrewModeUnusedOnly {
		commands = append(commands, "prune")
	}
	return commands
}

// executeCleanup runs cleanup commands and returns results.
func (hbc *HomebrewCleaner) executeCleanup(ctx context.Context, commands []string, cacheDir string) (itemsRemoved, itemsFailed int, bytesFreed int64) {
	if cacheDir != "" {
		bytesFreed, _, _ = CalculateBytesFreed(cacheDir, func() error {
			itemsRemoved, itemsFailed = hbc.runCleanupCommands(ctx, commands)
			return nil
		}, hbc.verbose, "Homebrew Cache")
	} else {
		itemsRemoved, itemsFailed = hbc.runCleanupCommands(ctx, commands)
	}
	return itemsRemoved, itemsFailed, bytesFreed
}

// runCleanupCommands executes brew commands and counts removed/failed items.
func (hbc *HomebrewCleaner) runCleanupCommands(ctx context.Context, commands []string) (itemsRemoved, itemsFailed int) {
	for _, cmd := range commands {
		cleanCmd := hbc.execWithTimeout(ctx, "brew", cmd)
		hbc.logCommandStart(cmd)

		output, err := cleanCmd.CombinedOutput()
		if err != nil {
			itemsFailed++
			hbc.logCommandError(cmd, string(output))
			continue
		}

		itemsRemoved += hbc.countRemovedItems(string(output))
		hbc.logCommandSuccess(cmd)
	}
	return itemsRemoved, itemsFailed
}

// logCommandStart logs command start if verbose.
func (hbc *HomebrewCleaner) logCommandStart(cmd string) {
	if hbc.verbose {
		fmt.Printf("🔧 Running 'brew %s'\n", cmd)
	}
}

// logCommandError logs command error if verbose.
func (hbc *HomebrewCleaner) logCommandError(cmd, output string) {
	if hbc.verbose {
		fmt.Printf("Warning: 'brew %s' failed: %s\n", cmd, output)
	}
}

// logCommandSuccess logs command success if verbose.
func (hbc *HomebrewCleaner) logCommandSuccess(cmd string) {
	if hbc.verbose {
		fmt.Printf("✅ 'brew %s' completed\n", cmd)
	}
}

// countRemovedItems counts items marked as removed in output.
func (hbc *HomebrewCleaner) countRemovedItems(output string) int {
	count := 0
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if strings.Contains(line, "removed") || strings.Contains(line, "deleted") {
			count++
		}
	}
	return count
}
