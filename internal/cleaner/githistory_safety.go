package cleaner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// Check performs all safety checks and returns a report.
func (c *GitHistorySafetyChecker) Check(ctx context.Context) *domain.GitHistorySafetyReport {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
		report.Blockers = append(report.Blockers, "You have uncommitted changes. Commit or stash them first.")
	}

	// Get branch info
	report.CurrentBranch = c.getCurrentBranch(ctx)

	// Check remote status
	c.checkRemote(ctx, report)

	// Check for unpushed commits
	if report.HasRemote {
		report.HasUnpushedCommits = c.hasUnpushedCommits(ctx)
		if report.HasUnpushedCommits {
			report.Warnings = append(report.Warnings, "You have unpushed commits that will be rewritten.")
		}
	}

	// Check if git-filter-repo is available
	report.FilterRepoAvailable = c.isFilterRepoAvailable(ctx)
	if !report.FilterRepoAvailable {
		report.Blockers = append(report.Blockers, "git-filter-repo is not installed. Install it with: pip install git-filter-repo or brew install git-filter-repo")
	}

	// Check backup capability
	report.DefaultBackupPath = c.getDefaultBackupPath()
	report.CanCreateBackup = c.canCreateBackup(report.DefaultBackupPath)

	return report
}

// isGitRepo checks if the path is a git repository.
func (c *GitHistorySafetyChecker) isGitRepo(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// hasUncommittedChanges checks for uncommitted changes.
func (c *GitHistorySafetyChecker) hasUncommittedChanges(ctx context.Context) bool {
	// Check for staged changes
	cmd := exec.CommandContext(ctx, "git", "-C", c.repoPath, "diff", "--cached", "--quiet")
	if cmd.Run() != nil {
		return true
	}

	// Check for unstaged changes
	cmd = exec.CommandContext(ctx, "git", "-C", c.repoPath, "diff", "--quiet")
	return cmd.Run() != nil
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
func (c *GitHistorySafetyChecker) checkRemote(ctx context.Context, report *domain.GitHistorySafetyReport) {
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
		cmd = exec.CommandContext(ctx, "git", "-C", c.repoPath, "remote", "get-url", report.RemoteName)
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

// isFilterRepoAvailable checks if git-filter-repo is installed.
func (c *GitHistorySafetyChecker) isFilterRepoAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "filter-repo", "--version")
	return cmd.Run() == nil
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

// CreateBackup creates a backup of the repository.
func (c *GitHistorySafetyChecker) CreateBackup(ctx context.Context, backupPath string) error {
	return createGitMirrorBackup(ctx, c.repoPath, backupPath)
}
