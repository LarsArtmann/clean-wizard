package cleaner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// GitHistorySafetyChecker performs safety checks before rewriting git history.
type GitHistorySafetyChecker struct {
	repoPath string
	verbose  bool
}

// NewGitHistorySafetyChecker creates a new safety checker.
func NewGitHistorySafetyChecker(repoPath string, verbose bool) *GitHistorySafetyChecker {
	return &GitHistorySafetyChecker{
		repoPath: repoPath,
		verbose:  verbose,
	}
}

// Protected branch names that require extra caution.
var protectedBranches = map[string]bool{
	"main":       true,
	"master":     true,
	"production": true,
	"staging":    true,
	"develop":    true,
}

// SafetyCheckTimeout is the default timeout for safety checks.
const SafetyCheckTimeout = 30 * time.Second

// Check performs all safety checks and returns a report.
func (c *GitHistorySafetyChecker) Check(ctx context.Context) *domain.GitHistorySafetyReport {
	ctx, cancel := context.WithTimeout(ctx, SafetyCheckTimeout)
	defer cancel()

	report := &domain.GitHistorySafetyReport{
		Warnings: []string{},
		Blockers: []string{},
	}

	// Check if it's a git repository
	report.IsGitRepo = c.isGitRepo(ctx)
	if !report.IsGitRepo {
		report.Blockers = append(report.Blockers, "Not a git repository")

		return report
	}

	// Check for uncommitted changes
	report.HasUncommittedChanges = c.hasUncommittedChanges(ctx)
	if report.HasUncommittedChanges {
		report.Blockers = append(
			report.Blockers,
			"You have uncommitted changes. Commit or stash them first.",
		)
	}

	// Get branch info
	report.CurrentBranch = c.getCurrentBranch(ctx)

	// Check if on a protected branch
	report.IsProtectedBranch = protectedBranches[report.CurrentBranch]
	if report.IsProtectedBranch {
		report.Warnings = append(
			report.Warnings,
			"You are on a protected branch ("+report.CurrentBranch+"). Consider creating a feature branch first.",
		)
	}

	// Check remote status
	c.checkRemote(ctx, report)

	// Check for unpushed commits
	if report.HasRemote {
		report.HasUnpushedCommits = c.hasUnpushedCommits(ctx)
		if report.HasUnpushedCommits {
			report.Warnings = append(
				report.Warnings,
				"You have unpushed commits that will be rewritten.",
			)
		}
	}

	// Check if git-filter-repo is available (system or via nix)
	report.FilterRepoAvailable = c.isFilterRepoAvailable()
	report.FilterRepoProvider = DetectFilterRepoProvider().String()
	if !report.FilterRepoAvailable {
		report.Blockers = append(
			report.Blockers,
			"git-filter-repo is not installed. Install with: brew install git-filter-repo, or ensure nix is available to use it automatically",
		)
	}

	// Check backup capability
	report.DefaultBackupPath = c.getDefaultBackupPath()
	report.CanCreateBackup = c.canCreateBackup(report.DefaultBackupPath)

	// Check for Git LFS
	report.HasLFS = c.hasLFS()
	if report.HasLFS {
		report.Warnings = append(
			report.Warnings,
			"Git LFS is configured. History rewriting may affect LFS objects.",
		)
	}

	// Check for submodules
	report.HasSubmodules = c.hasSubmodules()
	if report.HasSubmodules {
		report.Warnings = append(
			report.Warnings,
			"Repository contains submodules. Ensure submodules are handled appropriately.",
		)
	}

	// Check disk space for backup
	report.HasSufficientDiskSpace = c.hasSufficientDiskSpace()

	return report
}

// isGitRepo checks if the path is a git repository.
func (c *GitHistorySafetyChecker) isGitRepo(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "rev-parse", "--git-dir")

	return cmd.Run() == nil
}

// hasUncommittedChanges checks for uncommitted changes including untracked files.
func (c *GitHistorySafetyChecker) hasUncommittedChanges(ctx context.Context) bool {
	// Check for staged changes
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "diff", "--cached", "--quiet")
	if cmd.Run() != nil {
		return true
	}

	// Check for unstaged changes
	cmd = exec.CommandContext(ctx, "git", "-C", c.repoPath, "diff", "--quiet")
	if cmd.Run() != nil {
		return true
	}

	// Check for untracked files
	cmd = exec.CommandContext(ctx, "git", "-C", c.repoPath, "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return false // If we can't check, assume clean
	}

	return len(strings.TrimSpace(string(output))) > 0
}

// getCurrentBranch returns the current branch name.
func (c *GitHistorySafetyChecker) getCurrentBranch(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "branch", "--show-current")

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}

// checkRemote checks for remote configuration.
func (c *GitHistorySafetyChecker) checkRemote(
	ctx context.Context,
	report *domain.GitHistorySafetyReport,
) {
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "remote")

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		report.HasRemote = false

		return
	}

	report.HasRemote = true

	remotes := strings.Fields(string(output))
	if len(remotes) > 0 {
		report.RemoteName = remotes[0]

		// Get remote URL
		cmd = exec.CommandContext(
			ctx,
			"git",
			"-C",
			c.repoPath,
			"remote",
			"get-url",
			report.RemoteName,
		)

		urlOutput, err := cmd.Output()
		if err == nil {
			report.RemoteURL = strings.TrimSpace(string(urlOutput))
		}
	}

	// Warn about force push requirement
	if report.HasRemote {
		report.Warnings = append(report.Warnings,
			"History rewrite will require force push to remote: "+report.RemoteName)
		report.Warnings = append(report.Warnings,
			"Team members will need to reclone or reset their local branches.")
	}
}

// hasUnpushedCommits checks if there are unpushed commits.
func (c *GitHistorySafetyChecker) hasUnpushedCommits(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath,
		"log", "@{u}..HEAD", "--oneline")

	output, err := cmd.Output()
	if err != nil {
		return false // No upstream configured
	}

	return len(strings.TrimSpace(string(output))) > 0
}

// isFilterRepoAvailable checks if git-filter-repo is installed (system or via nix).
func (c *GitHistorySafetyChecker) isFilterRepoAvailable() bool {
	provider := DetectFilterRepoProvider()

	return provider != FilterRepoNone
}

// getDefaultBackupPath returns the default backup path.
func (c *GitHistorySafetyChecker) getDefaultBackupPath() string {
	return getDefaultBackupPath(c.repoPath)
}

// canCreateBackup checks if a backup can be created at the given path.
func (c *GitHistorySafetyChecker) canCreateBackup(backupPath string) bool {
	// Check if backup path already exists
	if _, err := os.Stat(backupPath); err == nil {
		return false // Backup already exists
	}

	// Check if we can write to parent directory
	parent := filepath.Dir(backupPath)

	info, err := os.Stat(parent)
	if err != nil {
		return false
	}

	return info.IsDir() && info.Mode().Perm()&0o200 != 0
}

// hasLFS checks if Git LFS is configured in the repository.
func (c *GitHistorySafetyChecker) hasLFS() bool {
	gitattributesPath := filepath.Join(c.repoPath, ".gitattributes")
	content, err := os.ReadFile(gitattributesPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "filter=lfs")
}

// hasSubmodules checks if the repository contains submodules.
func (c *GitHistorySafetyChecker) hasSubmodules() bool {
	gitmodulesPath := filepath.Join(c.repoPath, ".gitmodules")
	_, err := os.Stat(gitmodulesPath)

	return err == nil
}

// hasSufficientDiskSpace checks if there's enough disk space for backup.
func (c *GitHistorySafetyChecker) hasSufficientDiskSpace() bool {
	return hasSufficientDiskSpace(c.repoPath)
}

// CreateBackup creates a backup of the repository.
func (c *GitHistorySafetyChecker) CreateBackup(ctx context.Context, backupPath string) error {
	return createGitMirrorBackup(ctx, c.repoPath, backupPath)
}

// hasSufficientDiskSpace checks if there's enough disk space at the given path.
// Returns true if there's at least 1GB available or if the check fails (fail-open).
func hasSufficientDiskSpace(path string) bool {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return true // Fail-open: assume sufficient space if we can't check
	}

	// Calculate available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// Require at least 1GB of free space
	const minRequiredBytes = 1 * 1024 * 1024 * 1024

	return availableBytes >= minRequiredBytes
}
