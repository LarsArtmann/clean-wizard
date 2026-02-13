package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

const (
	// DefaultMinSizeMB is the default minimum file size in MB for compiled binaries.
	DefaultMinSizeMB = 10
	// DefaultOlderThan is the default age filter (0 = any age).
	DefaultOlderThan = "0"
	// DefaultCompiledBinariesTimeout is the default timeout for file operations.
	DefaultCompiledBinariesTimeout = 5 * time.Minute
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
var DefaultExcludeDirectories = []string{
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
var DefaultExcludeBinaries = []string{
	"chromedriver",
	"geckodriver",
	"edgedriver",
}

// BinaryScanner defines the interface for scanning compiled binaries.
type BinaryScanner interface {
	ScanDirectory(ctx context.Context, dir string, categories []BinaryCategory, minSize int64) ([]BinaryInfo, error)
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
	verbose           bool
	dryRun            bool
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
func NewCompiledBinariesCleaner(verbose, dryRun bool, minSizeMB int, olderThan string, basePaths, excludePatterns []string, opts ...CompiledBinariesOption) *CompiledBinariesCleaner {
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

	cleaner := &CompiledBinariesCleaner{
		verbose:           verbose,
		dryRun:            dryRun,
		minSizeMB:         minSizeMB,
		olderThan:         olderThan,
		basePaths:         basePaths,
		excludePatterns:   excludePatterns,
		includeCategories: []BinaryCategory{CategoryTmp, CategoryTest, CategoryBin, CategoryDist, CategoryRoot},
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
func (c *CompiledBinariesCleaner) Type() domain.OperationType {
	return domain.OperationTypeCompiledBinaries
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
func (c *CompiledBinariesCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.CompiledBinaries == nil {
		return nil // Settings are optional
	}

	s := settings.CompiledBinaries

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
			return fmt.Errorf("invalid include category %q, must be one of: tmp, test, bin, dist, root", cat)
		}
	}

	return nil
}

// Scan scans for compiled binary files in configured directories.
func (c *CompiledBinariesCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	minSizeBytes := int64(c.minSizeMB) * 1024 * 1024

	var allBinaries []BinaryInfo
	for _, basePath := range c.basePaths {
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			if c.verbose {
				fmt.Printf("Skipping non-existent path: %s\n", basePath)
			}
			continue
		}

		binaries, err := c.scanner.ScanDirectory(ctx, basePath, c.includeCategories, minSizeBytes)
		if err != nil {
			if c.verbose {
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
	items := make([]domain.ScanItem, 0, len(allBinaries))
	for _, b := range allBinaries {
		items = append(items, domain.ScanItem{
			Path:     b.Path,
			Size:     b.Size,
			Created:  b.ModTime,
			ScanType: domain.ScanTypeSystem,
		})
	}

	return result.Ok(items)
}

// Clean removes compiled binary files using trash.
func (c *CompiledBinariesCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	scanResult := c.Scan(ctx)
	if scanResult.IsErr() {
		return result.Err[domain.CleanResult](scanResult.Error())
	}

	items := scanResult.Value()

	if len(items) == 0 {
		cleanResult := domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: 0, Status: domain.SizeEstimateStatusKnown},
			FreedBytes:   0,
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
		}
		return result.Ok(cleanResult)
	}

	// Calculate total size
	var totalBytes int64
	for _, item := range items {
		totalBytes += item.Size
	}

	if c.dryRun {
		if c.verbose {
			fmt.Printf("Would trash %d compiled binary file(s) (%.2f MB)\n", len(items), float64(totalBytes)/(1024*1024))
		}
		cleanResult := domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Known: uint64(totalBytes), Status: domain.SizeEstimateStatusKnown},
			FreedBytes:   uint64(totalBytes),
			ItemsRemoved: uint(len(items)),
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyDryRunType),
		}
		return result.Ok(cleanResult)
	}

	// Actual cleaning
	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	for _, item := range items {
		err := c.trashOperator.TrashBinary(ctx, item.Path)
		if err != nil {
			itemsFailed++
			if c.verbose {
				fmt.Printf("Warning: %v\n", err)
			}
			continue
		}

		itemsRemoved++
		bytesFreed += item.Size

		if c.verbose {
			fmt.Printf("  ✓ Trashed: %s (%.2f MB)\n", item.Path, float64(item.Size)/(1024*1024))
		}
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{Known: uint64(bytesFreed), Status: domain.SizeEstimateStatusKnown},
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyAggressiveType),
	}

	return result.Ok(cleanResult)
}

// GetStoreSize returns the total size of all compiled binaries found.
func (c *CompiledBinariesCleaner) GetStoreSize(ctx context.Context) int64 {
	scanResult := c.Scan(ctx)
	if scanResult.IsErr() {
		return 0
	}

	var total int64
	for _, item := range scanResult.Value() {
		total += item.Size
	}
	return total
}

// parseAgeDuration parses a duration string like "7d", "30d", "1h", etc.
func parseAgeDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	var multiplier time.Duration
	unit := s[len(s)-1:]

	switch unit {
	case "d":
		multiplier = 24 * time.Hour
	case "h":
		multiplier = time.Hour
	case "w":
		multiplier = 7 * 24 * time.Hour
	case "m":
		multiplier = 30 * 24 * time.Hour
	case "y":
		multiplier = 365 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}

	var value int
	if _, err := fmt.Sscanf(s[:len(s)-1], "%d", &value); err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", s)
	}

	return time.Duration(value) * multiplier, nil
}

// Default implementations

type defaultBinaryScanner struct {
	excludePatterns   []string
	includeCategories []BinaryCategory
	verbose           bool
}

func (s *defaultBinaryScanner) ScanDirectory(ctx context.Context, dir string, categories []BinaryCategory, minSize int64) ([]BinaryInfo, error) {
	var binaries []BinaryInfo

	categorySet := make(map[BinaryCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories (but continue walking)
		if info.IsDir() {
			// Skip excluded directories
			if s.shouldSkipDirectory(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file size
		if info.Size() < minSize {
			return nil
		}

		// Check if it's an executable
		if info.Mode()&0o111 == 0 {
			return nil
		}

		// Determine category and check if included
		category := s.categorizeBinary(path, dir)
		if !categorySet[category] {
			return nil
		}

		// Check exclude patterns
		if s.isExcludedByPattern(path) {
			return nil
		}

		// Check default exclude binaries
		filename := filepath.Base(path)
		for _, exclude := range DefaultExcludeBinaries {
			if strings.Contains(strings.ToLower(filename), strings.ToLower(exclude)) {
				return nil
			}
		}

		binaries = append(binaries, BinaryInfo{
			Path:     path,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Category: category,
		})

		return nil
	})

	return binaries, err
}

func (s *defaultBinaryScanner) shouldSkipDirectory(path string) bool {
	base := filepath.Base(path)

	// Skip excluded directories
	if slices.Contains(DefaultExcludeDirectories, base) {
		if s.verbose {
			fmt.Printf("Skipping excluded directory: %s\n", path)
		}
		return true
	}

	return false
}

func (s *defaultBinaryScanner) categorizeBinary(path, baseDir string) BinaryCategory {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return CategoryRoot
	}

	// Check for test binaries (*.test)
	if strings.HasSuffix(path, ".test") {
		return CategoryTest
	}

	// Split path to check directory structure
	parts := strings.Split(filepath.Dir(relPath), string(filepath.Separator))

	for _, part := range parts {
		switch part {
		case "tmp":
			return CategoryTmp
		case "bin":
			return CategoryBin
		case "dist":
			return CategoryDist
		}
	}

	// If in root of project (no subdirectory or only one level)
	if len(parts) <= 1 || (len(parts) == 1 && parts[0] == ".") {
		return CategoryRoot
	}

	return CategoryRoot
}

func (s *defaultBinaryScanner) isExcludedByPattern(path string) bool {
	for _, pattern := range s.excludePatterns {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
		// Also try matching just the filename
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}

type defaultBinaryTrashOperator struct {
	verbose bool
}

func (t *defaultBinaryTrashOperator) TrashBinary(ctx context.Context, path string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "trash", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("trash failed for %s: %w (output: %s)", path, err, string(output))
	}
	return nil
}

func (t *defaultBinaryTrashOperator) GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (t *defaultBinaryTrashOperator) GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
