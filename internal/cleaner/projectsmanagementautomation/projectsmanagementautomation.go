package projectsmanagementautomation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

const bytesPerKB = 1024

const (
	// DefaultProjectsAutomationTimeout is the default timeout for automation commands.
	DefaultProjectsAutomationTimeout = 2 * time.Minute
	// DefaultProjectsAutomationCacheSizeMB is the default cache size estimate in MB.
	DefaultProjectsAutomationCacheSizeMB = 100
)

// ProjectsManagementAutomationCleaner handles projects-management-automation cache cleanup.
//
// Deprecated: This cleaner requires the external 'projects-management-automation' tool
// which is not commonly available. It will be removed in a future version.
// Use ProjectExecutablesCleaner or CompiledBinariesCleaner instead.
type ProjectsManagementAutomationCleaner struct {
	cleaner.CleanerBase
}

// NewProjectsManagementAutomationCleaner creates Projects Management Automation cleaner.
func NewProjectsManagementAutomationCleaner(
	verbose, dryRun bool,
) *ProjectsManagementAutomationCleaner {
	return &ProjectsManagementAutomationCleaner{
		CleanerBase: cleaner.NewCleanerBase(verbose, dryRun),
	}
}

// Type returns operation type for Projects Management Automation cleaner.
func (pc *ProjectsManagementAutomationCleaner) Type() operations.OperationType {
	return operations.OperationTypeProjectsManagementAutomation
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
// All settings are valid by default; the field is optional.
func (pc *ProjectsManagementAutomationCleaner) ValidateSettings(
	settings *operations.OperationSettings,
) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.ProjectsManagementAutomationSettings {
			return s.ProjectsManagementAutomation
		},
		func(*operations.ProjectsManagementAutomationSettings) error { return nil },
	)
}

// Scan scans for Projects Management Automation cache.
func (pc *ProjectsManagementAutomationCleaner) Scan(
	ctx context.Context,
) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0, 1)

	if !pc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	// Add cache item
	items = append(items, types.ScanItem{
		Path:     "~/.config/projects-management-automation/cache",
		Size:     pc.estimateCacheSize(),
		Created:  time.Now(),
		ScanType: types.ScanTypeSystem,
	})

	if pc.GetVerbose() {
		fmt.Printf("Found Projects Management Automation cache\n")
	}

	return result.Ok(items)
}

// Clean removes Projects Management Automation cache.
func (pc *ProjectsManagementAutomationCleaner) Clean(
	ctx context.Context,
) result.Result[types.CleanResult] {
	if !pc.IsAvailable(ctx) {
		return result.Err[types.CleanResult](
			cleaner.NewNotAvailableError("projects-management-automation", ""),
		)
	}

	startTime := time.Now()
	itemsRemoved := 0
	bytesFreed := int64(0)

	if pc.GetDryRun() {
		// Estimate cache sizes
		totalBytes := pc.estimateCacheSize()
		itemsRemoved := 1
		duration := time.Since(startTime)

		return result.Ok(
			conversions.NewCleanResultWithTiming(
				enums.StrategyDryRunType,
				itemsRemoved,
				totalBytes,
				duration,
			),
		)
	}

	// Execute projects-management-automation --clear-cache command
	cmd := adapters.ExecWithTimeout(ctx, DefaultProjectsAutomationTimeout,
		"projects-management-automation", "--clear-cache")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return result.Err[types.CleanResult](
			fmt.Errorf("projects-management-automation --clear-cache failed: %w (output: %s)",
				err, string(output)),
		)
	}

	itemsRemoved++
	bytesFreed += pc.estimateCacheSize()

	if pc.GetVerbose() {
		fmt.Println("  ✓ Projects Management Automation cache cleared")
	}

	duration := time.Since(startTime)

	return result.Ok(
		conversions.NewCleanResultWithTiming(
			enums.StrategyConservativeType,
			itemsRemoved,
			bytesFreed,
			duration,
		),
	)
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

	return cleaner.GetDirSize(cachePath)
}

// estimateCacheSize estimates the size of the cache.
func (pc *ProjectsManagementAutomationCleaner) estimateCacheSize() int64 {
	// Try to get actual size first
	if actualSize := pc.getActualCacheSize(); actualSize > 0 {
		return actualSize
	}
	// Fallback estimate: typical cache size
	return int64(DefaultProjectsAutomationCacheSizeMB * bytesPerKB * bytesPerKB)
}
