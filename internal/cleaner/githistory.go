package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// GitHistoryCleaner removes binary files from git history.
// This is a destructive operation that rewrites history.
type GitHistoryCleaner struct {
	verbose       bool
	dryRun        bool
	repoPath      string
	minSizeMB     int
	excludeExts   []string
	includeExts   []string
	excludePaths  []string
	maxFiles      int
	createBackup  bool
	selectedFiles []domain.GitHistoryFile // Files selected by user for removal

	scanner       *GitHistoryScanner
	safetyChecker *GitHistorySafetyChecker
	executor      *GitHistoryExecutor
}

// GitHistoryCleanerOption is a functional option for the cleaner.
type GitHistoryCleanerOption func(*GitHistoryCleaner)

// NewGitHistoryCleaner creates a new GitHistoryCleaner.
func NewGitHistoryCleaner(opts ...GitHistoryCleanerOption) *GitHistoryCleaner {
	c := &GitHistoryCleaner{
		repoPath:     ".",
		minSizeMB:    1,
		maxFiles:     100,
		createBackup: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize components
	c.scanner = NewGitHistoryScanner(c.repoPath,
		WithMinSizeMB(c.minSizeMB),
		WithExcludeExtensions(c.excludeExts),
		WithIncludeExtensions(c.includeExts),
		WithExcludePaths(c.excludePaths),
		WithMaxFiles(c.maxFiles),
		WithVerbose(c.verbose),
	)

	c.safetyChecker = NewGitHistorySafetyChecker(c.repoPath, c.verbose)
	c.executor = NewGitHistoryExecutor(c.repoPath, c.verbose, c.dryRun)

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
func WithGitHistoryMaxFiles(max int) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.maxFiles = max
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
		c.verbose = verbose
	}
}

// WithGitHistoryDryRun sets dry run mode.
func WithGitHistoryDryRun(dryRun bool) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.dryRun = dryRun
	}
}

// WithGitHistorySelectedFiles sets the files selected for removal.
func WithGitHistorySelectedFiles(files []domain.GitHistoryFile) GitHistoryCleanerOption {
	return func(c *GitHistoryCleaner) {
		c.selectedFiles = files
	}
}

// Type returns the operation type.
func (c *GitHistoryCleaner) Type() domain.OperationType {
	return domain.OperationTypeGitHistory
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
func (c *GitHistoryCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	if settings == nil || settings.GitHistory == nil {
		return nil // Settings are optional
	}

	s := settings.GitHistory

	if s.MinSizeMB < 0 {
		return fmt.Errorf("min_size_mb must be >= 0, got %d", s.MinSizeMB)
	}

	if s.MaxFiles < 0 {
		return fmt.Errorf("max_files must be >= 0, got %d", s.MaxFiles)
	}

	return nil
}

// Scan scans git history for large binary files.
func (c *GitHistoryCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	scanResult, err := c.scanner.Scan(ctx)
	if err != nil {
		return result.Err[[]domain.ScanItem](err)
	}

	items := make([]domain.ScanItem, len(scanResult.Files))
	for i, f := range scanResult.Files {
		items[i] = domain.ScanItem{
			Path:     f.Path,
			Size:     f.SizeBytes,
			Created:  f.CommitDate,
			ScanType: domain.ScanTypeSystem,
		}
	}

	return result.Ok(items)
}

// Clean removes selected files from git history.
func (c *GitHistoryCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	// If no files pre-selected, scan first
	if len(c.selectedFiles) == 0 {
		scanResult, err := c.scanner.Scan(ctx)
		if err != nil {
			return result.Err[domain.CleanResult](fmt.Errorf("scan failed: %w", err))
		}

		c.selectedFiles = scanResult.Files
	}

	if len(c.selectedFiles) == 0 {
		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			domain.CleanStrategyType(domain.StrategyConservativeType),
			0, 0,
			domain.SizeEstimate{Known: 0, Status: domain.SizeEstimateStatusKnown},
		))
	}

	// Run safety checks
	safetyReport := c.safetyChecker.Check(ctx)
	if !safetyReport.CanProceed() {
		return result.Err[domain.CleanResult](
			fmt.Errorf("safety checks failed: %s", safetyReport.Blockers),
		)
	}

	// Calculate total size
	var totalBytes int64
	for _, f := range c.selectedFiles {
		totalBytes += f.SizeBytes
	}

	if c.dryRun {
		if c.verbose {
			fmt.Printf("Would remove %d binary file(s) from git history (%.2f MB)\n",
				len(c.selectedFiles), float64(totalBytes)/(1024*1024))
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			domain.CleanStrategyType(domain.StrategyDryRunType),
			len(c.selectedFiles), totalBytes,
			domain.SizeEstimate{Known: uint64(totalBytes), Status: domain.SizeEstimateStatusKnown},
		))
	}

	// Execute the history rewrite
	execResult, err := c.executor.Execute(ctx, ExecuteOptions{
		FilesToRemove: c.selectedFiles,
		CreateBackup:  c.createBackup,
	})
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("execution failed: %w", err))
	}

	if c.verbose {
		fmt.Printf("Removed %d file(s) from history, reclaimed %.2f MB\n",
			len(execResult.FilesRemoved), float64(execResult.BytesReclaimed)/(1024*1024))

		if execResult.BackupCreated {
			fmt.Printf("Backup created at: %s\n", execResult.BackupPath)
		}
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		domain.CleanStrategyType(domain.StrategyAggressiveType),
		len(execResult.FilesRemoved),
		execResult.BytesRemoved,
		domain.SizeEstimate{
			Known:  uint64(execResult.BytesReclaimed),
			Status: domain.SizeEstimateStatusKnown,
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
func (c *GitHistoryCleaner) GetSafetyReport(ctx context.Context) *domain.GitHistorySafetyReport {
	return c.safetyChecker.Check(ctx)
}

// GetScanResult performs a full scan and returns the result.
func (c *GitHistoryCleaner) GetScanResult(
	ctx context.Context,
) (*domain.GitHistoryScanResult, error) {
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
func (c *GitHistoryCleaner) SetSelectedFiles(files []domain.GitHistoryFile) {
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
		maxDepth = 3
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
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
