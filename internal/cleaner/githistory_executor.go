package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// GitHistoryExecutor executes git history rewrites using git-filter-repo.
type GitHistoryExecutor struct {
	repoPath string
	verbose  bool
	dryRun   bool
}

// NewGitHistoryExecutor creates a new executor.
func NewGitHistoryExecutor(repoPath string, verbose, dryRun bool) *GitHistoryExecutor {
	return &GitHistoryExecutor{
		repoPath: repoPath,
		verbose:  verbose,
		dryRun:   dryRun,
	}
}

// ExecuteOptions configures the history rewrite.
type ExecuteOptions struct {
	FilesToRemove []domain.GitHistoryFile
	CreateBackup  bool
	BackupPath    string
	SkipGC        bool // Skip garbage collection (for testing)
}

// Execute runs the history rewrite.
func (e *GitHistoryExecutor) Execute(ctx context.Context, opts ExecuteOptions) (*domain.GitHistoryRewriteResult, error) {
	start := time.Now()

	if len(opts.FilesToRemove) == 0 {
		return nil, errors.New("no files to remove")
	}

	// Get old repo size
	oldSize, _ := e.getRepoSize()

	// Create backup if requested
	var backupCreated bool
	var backupPath string
	if opts.CreateBackup && !e.dryRun {
		backupPath = opts.BackupPath
		if backupPath == "" {
			backupPath = e.getDefaultBackupPath()
		}
		if err := e.createBackup(ctx, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
		backupCreated = true
	}

	if e.dryRun {
		return &domain.GitHistoryRewriteResult{
			FilesRemoved:    opts.FilesToRemove,
			BytesRemoved:    e.calculateTotalSize(opts.FilesToRemove),
			CommitsAffected: 0,
			OldRepoSize:     oldSize,
			BackupCreated:   false,
			ExecutedAt:      time.Now(),
			Duration:        time.Since(start),
		}, nil
	}

	// Execute git-filter-repo
	commitsAffected, err := e.runFilterRepo(ctx, opts.FilesToRemove)
	if err != nil {
		return nil, fmt.Errorf("git-filter-repo failed: %w", err)
	}

	// Run garbage collection
	if !opts.SkipGC {
		if err := e.runGC(ctx); err != nil && e.verbose {
			fmt.Printf("Warning: garbage collection failed: %v\n", err)
		}
	}

	// Get new repo size
	newSize, _ := e.getRepoSize()

	// Calculate bytes reclaimed
	bytesReclaimed := max(oldSize-newSize, 0)

	return &domain.GitHistoryRewriteResult{
		FilesRemoved:    opts.FilesToRemove,
		BytesRemoved:    e.calculateTotalSize(opts.FilesToRemove),
		CommitsAffected: commitsAffected,
		OldRepoSize:     oldSize,
		NewRepoSize:     newSize,
		BytesReclaimed:  bytesReclaimed,
		BackupCreated:   backupCreated,
		BackupPath:      backupPath,
		ExecutedAt:      time.Now(),
		Duration:        time.Since(start),
	}, nil
}

// runFilterRepo executes git-filter-repo to remove the specified files.
func (e *GitHistoryExecutor) runFilterRepo(ctx context.Context, files []domain.GitHistoryFile) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Build arguments for git-filter-repo
	args := []string{
		"filter-repo",
		"--force", // Required for non-fresh clones
	}

	// Add paths to remove
	for _, file := range files {
		args = append(args, "--path", file.Path)
	}

	// Invert paths (remove instead of keep)
	args = append(args, "--invert-paths")

	// Don't strip leading directories
	args = append(args, "--protect-blobs-from", "HEAD")

	if e.verbose {
		fmt.Printf("Running: git %s\n", strings.Join(args, " "))
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%w\nOutput: %s", err, string(output))
	}

	// Parse output to count affected commits
	commitsAffected := e.parseCommitCount(string(output))

	return commitsAffected, nil
}

// runGC runs git garbage collection.
func (e *GitHistoryExecutor) runGC(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Run reflog expire first
	cmd := exec.CommandContext(ctx, "git", "-C", e.repoPath,
		"reflog", "expire", "--expire=now", "--all")
	if output, err := cmd.CombinedOutput(); err != nil && e.verbose {
		fmt.Printf("Warning: reflog expire failed: %v\n%s\n", err, string(output))
	}

	// Run aggressive garbage collection
	cmd = exec.CommandContext(ctx, "git", "-C", e.repoPath,
		"gc", "--prune=now", "--aggressive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}

	return nil
}

// getRepoSize returns the size of the .git directory.
func (e *GitHistoryExecutor) getRepoSize() (int64, error) {
	gitDir := filepath.Join(e.repoPath, ".git")
	var size int64

	err := filepath.Walk(gitDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// createBackup creates a mirror backup of the repository.
func (e *GitHistoryExecutor) createBackup(ctx context.Context, backupPath string) error {
	return createGitMirrorBackup(ctx, e.repoPath, backupPath)
}

// createGitMirrorBackup creates a mirror backup of a git repository.
// This is a shared helper used by both GitHistoryExecutor and GitHistorySafetyChecker.
func createGitMirrorBackup(ctx context.Context, repoPath, backupPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", repoPath, backupPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("backup failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// getDefaultBackupPath returns the default backup path.
func (e *GitHistoryExecutor) getDefaultBackupPath() string {
	return getDefaultBackupPath(e.repoPath)
}

// calculateTotalSize calculates the total size of files to remove.
func (e *GitHistoryExecutor) calculateTotalSize(files []domain.GitHistoryFile) int64 {
	var total int64
	for _, f := range files {
		total += f.SizeBytes
	}
	return total
}

// parseCommitCount parses git-filter-repo output to count affected commits.
func (e *GitHistoryExecutor) parseCommitCount(output string) int {
	// Look for patterns like "Processed 123 commits"
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if strings.Contains(line, "Processed") && strings.Contains(line, "commit") {
			var count int
			if _, err := fmt.Sscanf(line, "Processed %d commit", &count); err == nil {
				return count
			}
		}
	}
	return 0
}

// EstimateImpact estimates the impact of removing files without executing.
func (e *GitHistoryExecutor) EstimateImpact(ctx context.Context, files []domain.GitHistoryFile) (*ImpactEstimate, error) {
	// Get current repo size
	oldSize, err := e.getRepoSize()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo size: %w", err)
	}

	// Calculate total size of files to remove
	totalFileBytes := e.calculateTotalSize(files)

	// Estimate commits affected (rough estimate based on file count)
	commitsEstimate := len(files) * 2 // Assume each file touches ~2 commits on average

	// Estimated new size (files represent roughly 70% of their size in pack files)
	packRatio := 0.7
	estimatedNewSize := max(oldSize-int64(float64(totalFileBytes)*packRatio), 0)

	return &ImpactEstimate{
		CurrentRepoSizeMB:  float64(oldSize) / (1024 * 1024),
		EstimatedNewSizeMB: float64(estimatedNewSize) / (1024 * 1024),
		SpaceReclaimedMB:   float64(oldSize-estimatedNewSize) / (1024 * 1024),
		FilesToRemove:      len(files),
		EstimatedCommits:   commitsEstimate,
		EstimatedDuration:  time.Duration(len(files)) * time.Second,
	}, nil
}

// ImpactEstimate contains impact estimates for a history rewrite.
type ImpactEstimate struct {
	CurrentRepoSizeMB  float64
	EstimatedNewSizeMB float64
	SpaceReclaimedMB   float64
	FilesToRemove      int
	EstimatedCommits   int
	EstimatedDuration  time.Duration
}

// RemoveFilesFromHistory is a convenience method that combines scanning and removal.
func (e *GitHistoryExecutor) RemoveFilesFromHistory(ctx context.Context, paths []string) error {
	files := make([]domain.GitHistoryFile, len(paths))
	for i, path := range paths {
		files[i] = domain.GitHistoryFile{Path: path}
	}

	_, err := e.Execute(ctx, ExecuteOptions{
		FilesToRemove: files,
		CreateBackup:  true,
	})
	return err
}

// StripLargeBlobs removes all blobs larger than the specified size.
func (e *GitHistoryExecutor) StripLargeBlobs(ctx context.Context, sizeMB int) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	args := []string{
		"filter-repo",
		"--force",
		"--strip-blobs-bigger-than", fmt.Sprintf("%dM", sizeMB),
	}

	if e.verbose {
		fmt.Printf("Running: git %s\n", strings.Join(args, " "))
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}

	// Run GC after stripping
	return e.runGC(ctx)
}

// GetFilesToRemoveFromSelection filters scan results by user selection.
func GetFilesToRemoveFromSelection(allFiles []domain.GitHistoryFile, selectedIndices []int) []domain.GitHistoryFile {
	result := make([]domain.GitHistoryFile, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		if idx >= 0 && idx < len(allFiles) {
			result = append(result, allFiles[idx])
		}
	}
	return result
}

// GetUniquePaths returns unique paths from files.
func GetUniquePaths(files []domain.GitHistoryFile) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, f := range files {
		if !seen[f.Path] {
			seen[f.Path] = true
			paths = append(paths, f.Path)
		}
	}
	slices.Sort(paths)
	return paths
}
