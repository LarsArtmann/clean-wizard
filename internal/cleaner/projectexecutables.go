package cleaner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DefaultProjectExecutablesTimeout is the default timeout for project listing commands.
const DefaultProjectExecutablesTimeout = 2 * time.Minute

// ProjectLister defines the interface for listing projects.
// This interface enables mocking in tests.
type ProjectLister interface {
	ListProjects(ctx context.Context) ([]ProjectInfo, error)
}

// FileOperator defines the interface for file operations.
// This interface enables mocking in tests.
type FileOperator interface {
	FindExecutableFiles(dir string) ([]string, error)
	TrashFile(ctx context.Context, path string) error
	GetFileSize(path string) int64
}

// ProjectInfo represents a project from projects-management-automation list.
type ProjectInfo struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ProjectExecutablesCleaner removes executable files (not shell scripts) from project directories.
// It integrates with projects-management-automation to discover projects and uses trash for safe deletion.
type ProjectExecutablesCleaner struct {
	verbose           bool
	dryRun            bool
	excludeExtensions []string
	excludePatterns   []string
	projectLister     ProjectLister
	fileOperator      FileOperator
}

// ProjectExecutablesOption is a functional option for configuring the cleaner.
type ProjectExecutablesOption func(*ProjectExecutablesCleaner)

// WithProjectLister sets a custom project lister for testing.
func WithProjectLister(lister ProjectLister) ProjectExecutablesOption {
	return func(c *ProjectExecutablesCleaner) {
		c.projectLister = lister
	}
}

// WithFileOperator sets a custom file operator for testing.
func WithFileOperator(operator FileOperator) ProjectExecutablesOption {
	return func(c *ProjectExecutablesCleaner) {
		c.fileOperator = operator
	}
}

// NewProjectExecutablesCleaner creates a new ProjectExecutablesCleaner.
func NewProjectExecutablesCleaner(verbose, dryRun bool, excludeExtensions, excludePatterns []string, opts ...ProjectExecutablesOption) *ProjectExecutablesCleaner {
	// Default to excluding .sh files if not specified
	if len(excludeExtensions) == 0 {
		excludeExtensions = []string{".sh"}
	}

	cleaner := &ProjectExecutablesCleaner{
		verbose:           verbose,
		dryRun:            dryRun,
		excludeExtensions: excludeExtensions,
		excludePatterns:   excludePatterns,
	}

	// Apply options
	for _, opt := range opts {
		opt(cleaner)
	}

	// Set default implementations if not provided
	if cleaner.projectLister == nil {
		cleaner.projectLister = &defaultProjectLister{}
	}
	if cleaner.fileOperator == nil {
		cleaner.fileOperator = &defaultFileOperator{
			excludeExtensions: excludeExtensions,
			excludePatterns:   excludePatterns,
			verbose:           verbose,
		}
	}

	return cleaner
}

// Type returns operation type for Project Executables cleaner.
func (p *ProjectExecutablesCleaner) Type() domain.OperationType {
	return domain.OperationTypeProjectExecutables
}

// Name returns the cleaner name for result tracking.
func (p *ProjectExecutablesCleaner) Name() string {
	return "project-executables"
}

// IsAvailable checks if projects-management-automation and trash are available.
func (p *ProjectExecutablesCleaner) IsAvailable(ctx context.Context) bool {
	_, errPMA := exec.LookPath("projects-management-automation")
	_, errTrash := exec.LookPath("trash")
	return errPMA == nil && errTrash == nil
}

// ValidateSettings validates Project Executables cleaner settings.
func (p *ProjectExecutablesCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.ProjectExecutables == nil {
		return nil // Settings are optional
	}

	// Validate exclude patterns are valid globs
	for _, pattern := range settings.ProjectExecutables.ExcludePatterns {
		_, err := filepath.Match(pattern, "test")
		if err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
	}

	return nil
}

// Scan scans for executable files in project directories.
func (p *ProjectExecutablesCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	projects, err := p.projectLister.ListProjects(ctx)
	if err != nil {
		return result.Err[[]domain.ScanItem](err)
	}

	items := make([]domain.ScanItem, 0)

	for _, project := range projects {
		executables, err := p.fileOperator.FindExecutableFiles(project.Path)
		if err != nil {
			if p.verbose {
				fmt.Printf("Warning: %v\n", err)
			}
			continue
		}

		for _, execPath := range executables {
			items = append(items, domain.ScanItem{
				Path:     execPath,
				Size:     p.fileOperator.GetFileSize(execPath),
				Created:  time.Now(),
				ScanType: domain.ScanTypeSystem,
			})
		}

		if p.verbose && len(executables) > 0 {
			fmt.Printf("Found %d executable(s) in %s\n", len(executables), project.Name)
		}
	}

	return result.Ok(items)
}

// Clean removes executable files from project directories using trash.
func (p *ProjectExecutablesCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	scanResult := p.Scan(ctx)
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

	// Calculate total size for dry-run preview
	var totalBytes int64
	for _, item := range items {
		totalBytes += item.Size
	}

	if p.dryRun {
		if p.verbose {
			fmt.Printf("Would trash %d executable file(s) (%d bytes)\n", len(items), totalBytes)
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
		err := p.fileOperator.TrashFile(ctx, item.Path)
		if err != nil {
			itemsFailed++
			if p.verbose {
				fmt.Printf("Warning: %v\n", err)
			}
			continue
		}

		itemsRemoved++
		bytesFreed += item.Size

		if p.verbose {
			fmt.Printf("  ✓ Trashed: %s\n", item.Path)
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

// GetStoreSize returns the total size of all executable files found.
func (p *ProjectExecutablesCleaner) GetStoreSize(ctx context.Context) int64 {
	scanResult := p.Scan(ctx)
	if scanResult.IsErr() {
		return 0
	}

	var total int64
	for _, item := range scanResult.Value() {
		total += item.Size
	}
	return total
}

// IsExcludedByExtension checks if the file should be excluded based on its extension.
func (p *ProjectExecutablesCleaner) IsExcludedByExtension(filename string) bool {
	return isExcludedByExtension(filename, p.excludeExtensions)
}

// IsExcludedByPattern checks if the file should be excluded based on glob patterns.
func (p *ProjectExecutablesCleaner) IsExcludedByPattern(filename string) bool {
	return isExcludedByPattern(filename, p.excludePatterns)
}

// isExcludedByExtension checks if the file should be excluded based on its extension.
func isExcludedByExtension(filename string, excludeExtensions []string) bool {
	for _, ext := range excludeExtensions {
		if strings.HasSuffix(strings.ToLower(filename), strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

// isExcludedByPattern checks if the file should be excluded based on glob patterns.
func isExcludedByPattern(filename string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		matched, err := filepath.Match(pattern, filename)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// Default implementations

type defaultProjectLister struct{}

func (d *defaultProjectLister) ListProjects(ctx context.Context) ([]ProjectInfo, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, DefaultProjectExecutablesTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "projects-management-automation", "list", "--output", "json", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projects []ProjectInfo
	if err := json.Unmarshal(output, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse project list: %w", err)
	}

	return projects, nil
}

type defaultFileOperator struct {
	excludeExtensions []string
	excludePatterns   []string
	verbose           bool
}

func (d *defaultFileOperator) FindExecutableFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var executables []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is executable
		if info.Mode()&0111 == 0 {
			continue
		}

		filename := entry.Name()
		fullPath := filepath.Join(dir, filename)

		// Check exclusion by extension
		if isExcludedByExtension(filename, d.excludeExtensions) {
			if d.verbose {
				fmt.Printf("  Skipping (extension excluded): %s\n", fullPath)
			}
			continue
		}

		// Check exclusion by pattern
		if isExcludedByPattern(filename, d.excludePatterns) {
			if d.verbose {
				fmt.Printf("  Skipping (pattern excluded): %s\n", fullPath)
			}
			continue
		}

		executables = append(executables, fullPath)
	}

	return executables, nil
}

func (d *defaultFileOperator) TrashFile(ctx context.Context, path string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "trash", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("trash failed for %s: %w (output: %s)", path, err, string(output))
	}
	return nil
}

func (d *defaultFileOperator) GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
