package compiledbinaries

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Byte conversion constants come from the shared cleaner core
// (cleaner.BytesPerMB / cleaner.BytesPerGB).

const (
	// DefaultMinSizeMB is the default minimum file size in MB for compiled binaries.
	DefaultMinSizeMB = 10
	// DefaultOlderThan is the default age filter (0 = any age).
	DefaultOlderThan = "0"
	// DefaultCompiledBinariesTimeout is the default timeout for file operations.
	DefaultCompiledBinariesTimeout = 5 * time.Minute
	// MinAgeStringLength is the minimum length for age duration strings.
	MinAgeStringLength = 2
	// DurationUnitDays represents days in hours.
	DurationUnitDays = 24 * time.Hour
	// DurationUnitWeeks represents weeks in hours.
	DurationUnitWeeks = 7 * 24 * time.Hour
	// DurationUnitMonths represents months in hours (30 days).
	DurationUnitMonths = 30 * 24 * time.Hour
	// DurationUnitYears represents years in hours.
	DurationUnitYears = 365 * 24 * time.Hour
)

// BinaryCategory represents a category of compiled binaries to clean.
type BinaryCategory string

const (
	CategoryTmp  BinaryCategory = "tmp"
	CategoryTest BinaryCategory = "test"
	CategoryBin  BinaryCategory = "bin"
	CategoryDist BinaryCategory = "dist"
	CategoryRoot BinaryCategory = "root"
)

// DefaultExcludeDirectories are directories that should never be scanned.
var DefaultExcludeDirectories = []string{ //nolint:gochecknoglobals
	"node_modules",
	"venv",
	".venv",
	".terraform",
	"__pycache__",
	".git",
	".hg",
	".svn",
}

// DefaultExcludeBinaries are specific binaries that should not be cleaned (required by tools).
var DefaultExcludeBinaries = []string{ //nolint:gochecknoglobals
	"chromedriver",
	"geckodriver",
	"edgedriver",
}

// BinaryScanner defines the interface for scanning compiled binaries.
type BinaryScanner interface {
	ScanDirectory(
		ctx context.Context,
		dir string,
		categories []BinaryCategory,
		minSize int64,
	) ([]BinaryInfo, error)
}

// BinaryTrashOperator defines the interface for trashing binaries.
type BinaryTrashOperator interface {
	TrashBinary(ctx context.Context, path string) error
	GetFileSize(path string) int64
	GetFileModTime(path string) (time.Time, error)
}

// BinaryInfo represents information about a compiled binary.
type BinaryInfo struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Category BinaryCategory
}

// CompiledBinariesCleaner removes large compiled binary files that can be regenerated.
// It scans project directories for build outputs, test binaries, and distribution files.
type CompiledBinariesCleaner struct {
	cleaner.CleanerBase

	minSizeMB         int
	olderThan         string
	basePaths         []string
	excludePatterns   []string
	includeCategories []BinaryCategory
	scanner           BinaryScanner
	trashOperator     BinaryTrashOperator
}

// CompiledBinariesOption is a functional option for configuring the cleaner.
type CompiledBinariesOption func(*CompiledBinariesCleaner)

// WithBinaryScanner sets a custom scanner for testing.
func WithBinaryScanner(scanner BinaryScanner) CompiledBinariesOption {
	return func(c *CompiledBinariesCleaner) {
		c.scanner = scanner
	}
}

// WithBinaryTrashOperator sets a custom trash operator for testing.
func WithBinaryTrashOperator(operator BinaryTrashOperator) CompiledBinariesOption {
	return func(c *CompiledBinariesCleaner) {
		c.trashOperator = operator
	}
}

// WithBasePaths sets the base paths to scan.
func WithBasePaths(paths []string) CompiledBinariesOption {
	return func(c *CompiledBinariesCleaner) {
		c.basePaths = paths
	}
}

// WithIncludeCategories sets which binary categories to include.
func WithIncludeCategories(categories []BinaryCategory) CompiledBinariesOption {
	return func(c *CompiledBinariesCleaner) {
		c.includeCategories = categories
	}
}

// NewCompiledBinariesCleaner creates a new CompiledBinariesCleaner.
func NewCompiledBinariesCleaner(
	verbose, dryRun bool,
	minSizeMB int,
	olderThan string,
	basePaths, excludePatterns []string,
	opts ...CompiledBinariesOption,
) *CompiledBinariesCleaner {
	// Apply defaults
	if minSizeMB <= 0 {
		minSizeMB = DefaultMinSizeMB
	}

	if olderThan == "" {
		olderThan = DefaultOlderThan
	}

	if len(basePaths) == 0 {
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			basePaths = []string{filepath.Join(homeDir, "projects")}
		}
	}

	cleaner := &CompiledBinariesCleaner{ //nolint:exhaustruct_v5
		CleanerBase:     cleaner.NewCleanerBase(verbose, dryRun),
		minSizeMB:       minSizeMB,
		olderThan:       olderThan,
		basePaths:       basePaths,
		excludePatterns: excludePatterns,
		includeCategories: []BinaryCategory{
			CategoryTmp,
			CategoryTest,
			CategoryBin,
			CategoryDist,
			CategoryRoot,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(cleaner)
	}

	// Set default implementations if not provided
	if cleaner.scanner == nil {
		cleaner.scanner = &defaultBinaryScanner{
			excludePatterns:   excludePatterns,
			includeCategories: cleaner.includeCategories,
			verbose:           verbose,
		}
	}

	if cleaner.trashOperator == nil {
		cleaner.trashOperator = &defaultBinaryTrashOperator{verbose: verbose}
	}

	return cleaner
}

// Type returns operation type for Compiled Binaries cleaner.
func (c *CompiledBinariesCleaner) Type() operations.OperationType {
	return operations.OperationTypeCompiledBinaries
}

// Name returns the cleaner name for result tracking.
func (c *CompiledBinariesCleaner) Name() string {
	return "compiled-binaries"
}

// IsAvailable checks if trash command is available.
func (c *CompiledBinariesCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("trash")

	return err == nil
}

// ValidateSettings validates Compiled Binaries cleaner settings.
func (c *CompiledBinariesCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.CompiledBinariesSettings { return s.CompiledBinaries },
		validateCompiledBinariesSettings,
	)
}

// validateCompiledBinariesSettings validates a non-nil CompiledBinariesSettings struct.
func validateCompiledBinariesSettings(s *operations.CompiledBinariesSettings) error {
	// Validate MinSizeMB
	if s.MinSizeMB < 0 {
		return fmt.Errorf("min_size_mb must be >= 0, got %d", s.MinSizeMB)
	}

	// Validate OlderThan format if specified
	if s.OlderThan != "" && s.OlderThan != "0" {
		if _, err := parseAgeDuration(s.OlderThan); err != nil {
			return fmt.Errorf("invalid older_than format %q: %w", s.OlderThan, err)
		}
	}

	// Validate exclude patterns are valid globs
	for _, pattern := range s.ExcludePatterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
	}

	// Validate include patterns are valid categories
	validCategories := map[string]bool{
		string(CategoryTmp):  true,
		string(CategoryTest): true,
		string(CategoryBin):  true,
		string(CategoryDist): true,
		string(CategoryRoot): true,
	}
	for _, cat := range s.IncludePatterns {
		if !validCategories[cat] {
			return fmt.Errorf(
				"invalid include category %q, must be one of: tmp, test, bin, dist, root",
				cat,
			)
		}
	}

	return nil
}

// Scan scans for compiled binary files in configured directories.
func (c *CompiledBinariesCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	minSizeBytes := int64(c.minSizeMB) * cleaner.BytesPerMB

	var allBinaries []BinaryInfo

	for _, basePath := range c.basePaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			if c.GetVerbose() {
				fmt.Printf("Skipping non-existent path: %s\n", basePath)
			}

			continue
		}

		binaries, err := c.scanner.ScanDirectory(ctx, basePath, c.includeCategories, minSizeBytes)
		if err != nil {
			if c.GetVerbose() {
				fmt.Printf("Warning: failed to scan %s: %v\n", basePath, err)
			}

			continue
		}

		// Apply age filter if configured
		if c.olderThan != "" && c.olderThan != "0" {
			ageDuration, err := parseAgeDuration(c.olderThan)
			if err == nil {
				cutoff := time.Now().Add(-ageDuration)

				var filtered []BinaryInfo

				for _, b := range binaries {
					if b.ModTime.Before(cutoff) {
						filtered = append(filtered, b)
					}
				}

				binaries = filtered
			}
		}

		allBinaries = append(allBinaries, binaries...)
	}

	// Convert to ScanItems
	items := make([]types.ScanItem, 0, len(allBinaries))
	for _, b := range allBinaries {
		items = append(items, types.ScanItem{
			Path:     b.Path,
			Size:     b.Size,
			Created:  b.ModTime,
			ScanType: types.ScanTypeSystem,
		})
	}

	return result.Ok(items)
}

// Clean removes compiled binary files using trash.
func (c *CompiledBinariesCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	return cleaner.ExecuteTrashPipeline(
		ctx,
		c.Scan(ctx),
		c.GetDryRun(),
		c.GetVerbose(),
		"compiled binary file(s)",
		func(cleanCtx context.Context, item types.ScanItem) error {
			return c.trashOperator.TrashBinary(cleanCtx, item.Path)
		},
		func(item types.ScanItem) {
			fmt.Printf("  ✓ Trashed: %s (%.2f MB)\n", item.Path, float64(item.Size)/cleaner.BytesPerMB)
		},
	)
}

// GetStoreSize returns the total size of all compiled binaries found.
func (c *CompiledBinariesCleaner) GetStoreSize(ctx context.Context) int64 {
	return cleaner.CalculateTotalSizeFromScan(c.Scan(ctx))
}
