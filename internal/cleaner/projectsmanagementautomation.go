package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DefaultProjectsAutomationTimeout is the default timeout for automation commands.
const DefaultProjectsAutomationTimeout = 2 * time.Minute

// execWithTimeout creates a command with timeout protection.
func (pc *ProjectsManagementAutomationCleaner) execWithTimeout(ctx context.Context, name string, args ...string) *exec.Cmd {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return exec.CommandContext(ctx, name, args...)
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultProjectsAutomationTimeout)
	_ = cancel
	return exec.CommandContext(timeoutCtx, name, args...)
}

// ProjectsManagementAutomationCleaner handles projects-management-automation cache cleanup.
type ProjectsManagementAutomationCleaner struct {
	verbose bool
	dryRun  bool
}

// NewProjectsManagementAutomationCleaner creates Projects Management Automation cleaner.
func NewProjectsManagementAutomationCleaner(verbose, dryRun bool) *ProjectsManagementAutomationCleaner {
	return &ProjectsManagementAutomationCleaner{
		verbose: verbose,
		dryRun:  dryRun,
	}
}

// Type returns operation type for Projects Management Automation cleaner.
func (pc *ProjectsManagementAutomationCleaner) Type() domain.OperationType {
	return domain.OperationTypeProjectsManagementAutomation
}

// Name returns the cleaner name for result tracking.
func (pc *ProjectsManagementAutomationCleaner) Name() string {
	return "projects"
}

// IsAvailable checks if projects-management-automation is available.
func (pc *ProjectsManagementAutomationCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("projects-management-automation")
	return err == nil
}

// ValidateSettings validates Projects Management Automation cleaner settings.
func (pc *ProjectsManagementAutomationCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.ProjectsManagementAutomation == nil {
		return nil // Settings are optional
	}

	// All settings are valid by default
	return nil
}

// Scan scans for Projects Management Automation cache.
func (pc *ProjectsManagementAutomationCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	if !pc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	// Add cache item
	items = append(items, domain.ScanItem{
		Path:     "~/.config/projects-management-automation/cache",
		Size:     pc.estimateCacheSize(),
		Created:  time.Now(),
		ScanType: domain.ScanTypeSystem,
	})

	if pc.verbose {
		fmt.Printf("Found Projects Management Automation cache\n")
	}

	return result.Ok(items)
}

// Clean removes Projects Management Automation cache.
func (pc *ProjectsManagementAutomationCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !pc.IsAvailable(ctx) {
		return result.Err[domain.CleanResult](errors.New("projects-management-automation not available"))
	}

	startTime := time.Now()
	itemsRemoved := 0
	bytesFreed := int64(0)

	if pc.dryRun {
		// Estimate cache sizes
		totalBytes := pc.estimateCacheSize()
		itemsRemoved := 1
		duration := time.Since(startTime)

		return result.Ok(conversions.NewCleanResultWithTiming(domain.CleanStrategyType(domain.StrategyDryRunType), itemsRemoved, totalBytes, duration))
	}

	// Execute projects-management-automation --clear-cache command
	cmd := pc.execWithTimeout(ctx, "projects-management-automation", "--clear-cache")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("projects-management-automation --clear-cache failed: %w (output: %s)", err, string(output)))
	}

	itemsRemoved++
	bytesFreed += pc.estimateCacheSize()

	if pc.verbose {
		fmt.Println("  ✓ Projects Management Automation cache cleared")
	}

	duration := time.Since(startTime)
	return result.Ok(conversions.NewCleanResultWithTiming(domain.CleanStrategyType(domain.StrategyConservativeType), itemsRemoved, bytesFreed, duration))
}

// getCachePath returns the expanded cache directory path.
func (pc *ProjectsManagementAutomationCleaner) getCachePath() string {
	cachePath := "~/.config/projects-management-automation/cache"
	if strings.HasPrefix(cachePath, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, cachePath[2:])
		}
	}
	return cachePath
}

// getActualCacheSize returns the actual size of the cache directory.
func (pc *ProjectsManagementAutomationCleaner) getActualCacheSize() int64 {
	cachePath := pc.getCachePath()
	return GetDirSize(cachePath)
}

// estimateCacheSize estimates the size of the cache.
func (pc *ProjectsManagementAutomationCleaner) estimateCacheSize() int64 {
	// Try to get actual size first
	if actualSize := pc.getActualCacheSize(); actualSize > 0 {
		return actualSize
	}
	// Fallback estimate: 100MB for typical cache size
	return int64(100 * 1024 * 1024)
}
