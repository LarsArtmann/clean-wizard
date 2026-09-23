package githistory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// GitHistoryCleaner removes binary files from git history.
// This is a destructive operation that rewrites history.
const (
	// GitHistoryDefaultMinSizeMB is the default minimum file size in MB to consider.
	GitHistoryDefaultMinSizeMB = 1
	// GitHistoryDefaultMaxFiles is the default maximum number of files to show.
	GitHistoryDefaultMaxFiles = 100
	// GitHistoryDefaultMaxSearchDepth is the default maximum depth for repository search.
	GitHistoryDefaultMaxSearchDepth = 3
)

type GitHistoryCleaner struct {
	cleaner.CleanerBase

	repoPath      string
	minSizeMB     int
	excludeExts   []string
	includeExts   []string
	excludePaths  []string
	maxFiles      int
	createBackup  bool
	selectedFiles []types.GitHistoryFile

	scanner       *GitHistoryScanner
	safetyChecker *GitHistorySafetyChecker
	executor      *GitHistoryExecutor
}

// GitHistoryCleanerOption is a functional option for the cleaner.
type GitHistoryCleanerOption func(*GitHistoryCleaner)

// NewGitHistoryCleaner creates a new GitHistoryCleaner.
func NewGitHistoryCleaner(opts ...GitHistoryCleanerOption) *GitHistoryCleaner {
	c := &GitHistoryCleaner{ //nolint:exhaustruct
		repoPath:     ".",
		minSizeMB:    GitHistoryDefaultMinSizeMB,
		maxFiles:     GitHistoryDefaultMaxFiles,
		createBackup: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize components
	c.scanner = NewGitHistoryScanner(
		c.repoPath,
		WithMinSizeMB(c.minSizeMB),
		WithExcludeExtensions(c.excludeExts),
		WithIncludeExtensions(c.includeExts),
		WithExcludePaths(c.excludePaths),
		WithMaxFiles(c.maxFiles),
		WithVerbose(c.GetVerbose()),
	)

	c.safetyChecker = NewGitHistorySafetyChecker(c.repoPath, c.GetVerbose())
	c.executor = NewGitHistoryExecutor(c.repoPath, c.GetVerbose(), c.GetDryRun())

	return c
}

// WithGitHistoryRepoPath sets the repository path.
func WithGitHistoryRepoPath(path string) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.repoPath = path
	}
}

// WithGitHistoryMinSizeMB sets the minimum file size in MB.
func WithGitHistoryMinSizeMB(mb int) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.minSizeMB = mb
	}
}

// WithGitHistoryExcludeExtensions sets extensions to exclude.
func WithGitHistoryExcludeExtensions(exts []string) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.excludeExts = exts
	}
}

// WithGitHistoryIncludeExtensions sets extensions to include.
func WithGitHistoryIncludeExtensions(exts []string) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.includeExts = exts
	}
}

// WithGitHistoryExcludePaths sets path patterns to exclude.
func WithGitHistoryExcludePaths(paths []string) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.excludePaths = paths
	}
}

// WithGitHistoryMaxFiles sets the maximum number of files to show.
func WithGitHistoryMaxFiles(maxFiles int) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.maxFiles = maxFiles
	}
}

// WithGitHistoryCreateBackup sets whether to create a backup.
func WithGitHistoryCreateBackup(create bool) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.createBackup = create
	}
}

// WithGitHistoryVerbose sets verbose mode.
func WithGitHistoryVerbose(verbose bool) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.SetVerbose(verbose)
	}
}

// WithGitHistoryDryRun sets dry run mode.
func WithGitHistoryDryRun(dryRun bool) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.SetDryRun(dryRun)
	}
}

// WithGitHistorySelectedFiles sets the files selected for removal.
func WithGitHistorySelectedFiles(files []types.GitHistoryFile) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.selectedFiles = files
	}
}

// Type returns the operation type.
func (c *GitHistoryCleaner) Type() operations.OperationType {
	return operations.OperationTypeGitHistory
}

// Name returns the cleaner name.
func (c *GitHistoryCleaner) Name() string {
	return "git-history"
}

// IsAvailable checks if git and git-filter-repo are available.
func (c *GitHistoryCleaner) IsAvailable(ctx context.Context) bool {
	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}

	// Check if we're in a git repo
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "rev-parse", "--git-dir")

	return cmd.Run() == nil
}

// ValidateSettings validates the cleaner settings.
func (c *GitHistoryCleaner) ValidateSettings(settings *operations.OperationSettings) error {
	return cleaner.ValidateOptionalSettings(
		settings,
		func(s *operations.OperationSettings) *operations.GitHistorySettings { return s.GitHistory },
		func(s *operations.GitHistorySettings) error {
			if s.MinSizeMB < 0 {
				return fmt.Errorf("min_size_mb must be >= 0, got %d", s.MinSizeMB)
			}

			if s.MaxFiles < 0 {
				return fmt.Errorf("max_files must be >= 0, got %d", s.MaxFiles)
			}

			return nil
		},
	)
}

// Scan scans git history for large binary files.
func (c *GitHistoryCleaner) Scan(ctx context.Context) result.Result[[]types.ScanItem] {
	scanResult, err := c.scanner.Scan(ctx)
	if err != nil {
		return result.Err[[]types.ScanItem](err)
	}

	items := make([]types.ScanItem, len(scanResult.Files))
	for i, f := range scanResult.Files {
		items[i] = types.ScanItem{
			Path:     f.Path,
			Size:     f.SizeBytes,
			Created:  f.CommitDate,
			ScanType: types.ScanTypeSystem,
		}
	}

	return result.Ok(items)
}

// Clean removes selected files from git history.
func (c *GitHistoryCleaner) Clean(ctx context.Context) result.Result[types.CleanResult] {
	// Ensure files are selected
	err := c.ensureSelectedFiles(ctx)
	if err != nil {
		return result.Err[types.CleanResult](err)
	}

	if len(c.selectedFiles) == 0 {
		return c.emptyResult()
	}

	// Run safety checks
	safetyReport := c.safetyChecker.Check(ctx)
	if !safetyReport.CanProceed() {
		return result.Err[types.CleanResult](
			fmt.Errorf("safety checks failed: %s", safetyReport.Blockers),
		)
	}

	totalBytes := c.calculateTotalBytes()

	if c.GetDryRun() {
		return c.executeDryRun(totalBytes)
	}

	return c.executeClean(ctx, totalBytes)
}

// ensureSelectedFiles scans for files if none are pre-selected.
func (c *GitHistoryCleaner) ensureSelectedFiles(ctx context.Context) error {
	if len(c.selectedFiles) == 0 {
		scanResult, err := c.scanner.Scan(ctx)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		c.selectedFiles = scanResult.Files
	}

	return nil
}

// emptyResult returns a result for when there are no files to clean.
func (c *GitHistoryCleaner) emptyResult() result.Result[types.CleanResult] {
	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		0, 0,
		types.SizeEstimate{Known: 0, Status: enums.SizeEstimateStatusKnown},
	))
}

// calculateTotalBytes calculates the total size of selected files.
func (c *GitHistoryCleaner) calculateTotalBytes() int64 {
	var total int64
	for _, f := range c.selectedFiles {
		total += f.SizeBytes
	}

	return total
}

// executeDryRun returns a result for dry run mode.
func (c *GitHistoryCleaner) executeDryRun(totalBytes int64) result.Result[types.CleanResult] {
	if c.GetVerbose() {
		fmt.Printf("Would remove %d binary file(s) from git history (%.2f MB)\n",
			len(c.selectedFiles), float64(totalBytes)/float64(BytesPerMB))
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyDryRunType,
		len(c.selectedFiles), totalBytes,
		types.SizeEstimate{Known: uint64(totalBytes), Status: enums.SizeEstimateStatusKnown},
	))
}

// executeClean performs the actual history rewrite.
func (c *GitHistoryCleaner) executeClean(
	ctx context.Context,
	_ int64,
) result.Result[types.CleanResult] {
	execResult, err := c.executor.Execute(ctx, ExecuteOptions{ //nolint:exhaustruct
		FilesToRemove: c.selectedFiles,
		CreateBackup:  c.createBackup,
	})
	if err != nil {
		return result.Err[types.CleanResult](fmt.Errorf("execution failed: %w", err))
	}

	if c.GetVerbose() {
		fmt.Printf("Removed %d file(s) from history, reclaimed %.2f MB\n",
			len(execResult.FilesRemoved), float64(execResult.BytesReclaimed)/float64(BytesPerMB))

		if execResult.BackupCreated {
			fmt.Printf("Backup created at: %s\n", execResult.BackupPath)
		}
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyAggressiveType,
		len(execResult.FilesRemoved),
		execResult.BytesRemoved,
		types.SizeEstimate{
			Known:  uint64(execResult.BytesReclaimed),
			Status: enums.SizeEstimateStatusKnown,
		},
	))
}

// GetStoreSize returns the size of the .git directory.
func (c *GitHistoryCleaner) GetStoreSize(ctx context.Context) int64 {
	size, err := c.scanner.GetRepoSize()
	if err != nil {
		return 0
	}

	return size
}

// GetSafetyReport returns the safety check report for the repository.
func (c *GitHistoryCleaner) GetSafetyReport(ctx context.Context) *types.GitHistorySafetyReport {
	return c.safetyChecker.Check(ctx)
}

// GetScanResult performs a full scan and returns the result.
func (c *GitHistoryCleaner) GetScanResult(
	ctx context.Context,
) (*types.GitHistoryScanResult, error) {
	return c.scanner.Scan(ctx)
}

// EstimateImpact estimates the impact of removing the selected files.
func (c *GitHistoryCleaner) EstimateImpact(ctx context.Context) (*ImpactEstimate, error) {
	if len(c.selectedFiles) == 0 {
		scanResult, err := c.scanner.Scan(ctx)
		if err != nil {
			return nil, err
		}

		c.selectedFiles = scanResult.Files
	}

	return c.executor.EstimateImpact(ctx, c.selectedFiles)
}

// SetSelectedFiles sets the files to remove.
func (c *GitHistoryCleaner) SetSelectedFiles(files []types.GitHistoryFile) {
	c.selectedFiles = files
}

// CreateBackup creates a backup of the repository.
func (c *GitHistoryCleaner) CreateBackup(ctx context.Context, backupPath string) error {
	return c.safetyChecker.CreateBackup(ctx, backupPath)
}

// StripLargeBlobs removes all blobs larger than the specified size.
func (c *GitHistoryCleaner) StripLargeBlobs(ctx context.Context, sizeMB int) error {
	return c.executor.StripLargeBlobs(ctx, sizeMB)
}

// getDefaultBackupPath returns the default backup path for a repository.
func getDefaultBackupPath(repoPath string) string {
	absPath, _ := filepath.Abs(repoPath)
	parent := filepath.Dir(absPath)
	base := filepath.Base(absPath)

	return filepath.Join(parent, base+"-backup.git")
}

// FindGitRepositories finds all git repositories under the given base path.
func FindGitRepositories(basePath string, maxDepth int) ([]string, error) {
	var repos []string

	// Limit search depth
	if maxDepth <= 0 {
		maxDepth = GitHistoryDefaultMaxSearchDepth
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := basePath + "/" + entry.Name()

		// Check if this is a git repo
		gitDir := fullPath + "/.git"
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			repos = append(repos, fullPath)

			continue // Don't search inside git repos
		}

		// Recurse if we haven't reached max depth
		if maxDepth > 1 {
			subRepos, err := FindGitRepositories(fullPath, maxDepth-1)
			if err != nil {
				continue
			}

			repos = append(repos, subRepos...)
		}
	}

	return repos, nil
}
