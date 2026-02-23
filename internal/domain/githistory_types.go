package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// GitHistoryFile represents a binary file found in git history.
type GitHistoryFile struct {
	// Path is the file path in the repository
	Path string `json:"path"`
	// SizeBytes is the file size in bytes
	SizeBytes int64 `json:"size_bytes"`
	// BlobHash is the git blob hash
	BlobHash string `json:"blob_hash"`
	// CommitHash is the commit that added this file
	CommitHash string `json:"commit_hash"`
	// CommitDate is when the commit was made
	CommitDate time.Time `json:"commit_date"`
	// Author is who committed the file
	Author string `json:"author"`
	// IsDeleted indicates if the file exists in HEAD
	IsDeleted bool `json:"is_deleted"`
	// Extension is the file extension (empty for extensionless binaries)
	Extension string `json:"extension"`
	// Selected indicates if this file is selected for removal (for TUI)
	Selected bool `json:"selected,omitempty"`
}

// IsValid validates the GitHistoryFile.
func (f GitHistoryFile) IsValid() bool {
	return f.Path != "" && f.SizeBytes > 0 && f.BlobHash != ""
}

// Validate returns errors for invalid GitHistoryFile.
func (f GitHistoryFile) Validate() error {
	if f.Path == "" {
		return errors.New("path cannot be empty")
	}
	if f.SizeBytes <= 0 {
		return fmt.Errorf("size must be positive, got: %d", f.SizeBytes)
	}
	if f.BlobHash == "" {
		return errors.New("blob hash cannot be empty")
	}
	return nil
}

// SizeMB returns the file size in megabytes.
func (f GitHistoryFile) SizeMB() float64 {
	return float64(f.SizeBytes) / (1024 * 1024)
}

// GitHistoryScanResult contains the results of scanning git history.
type GitHistoryScanResult struct {
	// Files is the list of binary files found
	Files []GitHistoryFile `json:"files"`
	// TotalBytes is the total size of all files
	TotalBytes int64 `json:"total_bytes"`
	// TotalFiles is the count of files found
	TotalFiles int `json:"total_files"`
	// RepoPath is the path to the repository
	RepoPath string `json:"repo_path"`
	// ScannedAt is when the scan was performed
	ScannedAt time.Time `json:"scanned_at"`
	// Duration is how long the scan took
	Duration time.Duration `json:"duration"`
}

// IsValid validates the GitHistoryScanResult.
func (r GitHistoryScanResult) IsValid() bool {
	return r.RepoPath != "" && !r.ScannedAt.IsZero()
}

// GitHistoryRewriteResult contains the results of rewriting git history.
type GitHistoryRewriteResult struct {
	// FilesRemoved is the list of files that were removed
	FilesRemoved []GitHistoryFile `json:"files_removed"`
	// BytesRemoved is the total bytes removed
	BytesRemoved int64 `json:"bytes_removed"`
	// CommitsAffected is the number of commits that were modified
	CommitsAffected int `json:"commits_affected"`
	// OldRepoSize is the repository size before rewrite
	OldRepoSize int64 `json:"old_repo_size"`
	// NewRepoSize is the repository size after rewrite
	NewRepoSize int64 `json:"new_repo_size"`
	// BytesReclaimed is the actual space reclaimed
	BytesReclaimed int64 `json:"bytes_reclaimed"`
	// BackupCreated indicates if a backup was made
	BackupCreated bool `json:"backup_created"`
	// BackupPath is the path to the backup
	BackupPath string `json:"backup_path,omitempty"`
	// ExecutedAt is when the rewrite was performed
	ExecutedAt time.Time `json:"executed_at"`
	// Duration is how long the rewrite took
	Duration time.Duration `json:"duration"`
}

// IsValid validates the GitHistoryRewriteResult.
func (r GitHistoryRewriteResult) IsValid() bool {
	return !r.ExecutedAt.IsZero()
}

// GitHistorySafetyReport contains safety check results.
type GitHistorySafetyReport struct {
	// IsGitRepo indicates if the path is a git repository
	IsGitRepo bool `json:"is_git_repo"`
	// HasUncommittedChanges indicates if there are uncommitted changes
	HasUncommittedChanges bool `json:"has_uncommitted_changes"`
	// HasRemote indicates if the repo has a remote
	HasRemote bool `json:"has_remote"`
	// RemoteURL is the remote URL (if any)
	RemoteURL string `json:"remote_url,omitempty"`
	// RemoteName is the remote name (usually "origin")
	RemoteName string `json:"remote_name,omitempty"`
	// CurrentBranch is the current branch name
	CurrentBranch string `json:"current_branch"`
	// HasUnpushedCommits indicates if there are unpushed commits
	HasUnpushedCommits bool `json:"has_unpushed_commits"`
	// CanCreateBackup indicates if a backup can be created
	CanCreateBackup bool `json:"can_create_backup"`
	// DefaultBackupPath is the suggested backup path
	DefaultBackupPath string `json:"default_backup_path,omitempty"`
	// FilterRepoAvailable indicates if git-filter-repo is installed
	FilterRepoAvailable bool `json:"filter_repo_available"`
	// Warnings are non-blocking issues
	Warnings []string `json:"warnings,omitempty"`
	// Blockers are issues that must be resolved before proceeding
	Blockers []string `json:"blockers,omitempty"`
}

// CanProceed returns true if there are no blockers.
func (r GitHistorySafetyReport) CanProceed() bool {
	return r.IsGitRepo && !r.HasUncommittedChanges && r.FilterRepoAvailable && len(r.Blockers) == 0
}

// GitHistoryMode represents the mode of operation for git history cleaning.
type GitHistoryMode int

const (
	GitHistoryModeAnalyze GitHistoryMode = iota // Only report findings
	GitHistoryModeDryRun                        // Show what would be done
	GitHistoryModeExecute                       // Actually rewrite history
)

// String returns the string representation.
func (m GitHistoryMode) String() string {
	switch m {
	case GitHistoryModeAnalyze:
		return "analyze"
	case GitHistoryModeDryRun:
		return "dry-run"
	case GitHistoryModeExecute:
		return "execute"
	default:
		return "unknown"
	}
}

// IsValid checks if the mode is valid.
func (m GitHistoryMode) IsValid() bool {
	return m >= GitHistoryModeAnalyze && m <= GitHistoryModeExecute
}

// Values returns all possible values.
func (m GitHistoryMode) Values() []GitHistoryMode {
	return []GitHistoryMode{
		GitHistoryModeAnalyze,
		GitHistoryModeDryRun,
		GitHistoryModeExecute,
	}
}

// MarshalJSON implements json.Marshaler.
func (m GitHistoryMode) MarshalJSON() ([]byte, error) {
	if !m.IsValid() {
		return nil, fmt.Errorf("invalid git history mode: %d", m)
	}
	return json.Marshal(m.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *GitHistoryMode) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONEnum(data, m, map[string]GitHistoryMode{
		"analyze": GitHistoryModeAnalyze,
		"dry-run": GitHistoryModeDryRun,
		"dryrun":  GitHistoryModeDryRun,
		"execute": GitHistoryModeExecute,
	}, "invalid git history mode")
}

// GitHistorySettings provides type-safe settings for git history cleaning.
type GitHistorySettings struct {
	// MinSizeMB is the minimum file size in MB to consider (default: 1)
	MinSizeMB int `json:"min_size_mb,omitempty" yaml:"min_size_mb,omitempty"`
	// MaxFiles limits the number of files to show (0 = unlimited)
	MaxFiles int `json:"max_files,omitempty" yaml:"max_files,omitempty"`
	// ExcludeExtensions are file extensions to exclude
	ExcludeExtensions []string `json:"exclude_extensions,omitempty" yaml:"exclude_extensions,omitempty"`
	// IncludeExtensions are file extensions to include (empty = all)
	IncludeExtensions []string `json:"include_extensions,omitempty" yaml:"include_extensions,omitempty"`
	// ExcludePaths are path patterns to exclude
	ExcludePaths []string `json:"exclude_paths,omitempty" yaml:"exclude_paths,omitempty"`
	// CreateBackup indicates if a backup should be created before rewrite
	CreateBackup bool `json:"create_backup,omitempty" yaml:"create_backup,omitempty"`
	// SkipConfirmation skips the interactive confirmation (dangerous)
	SkipConfirmation bool `json:"skip_confirmation,omitempty" yaml:"skip_confirmation,omitempty"`
}

// DefaultGitHistorySettings returns the default settings.
func DefaultGitHistorySettings() GitHistorySettings {
	return GitHistorySettings{
		MinSizeMB:         1,
		MaxFiles:          100,
		CreateBackup:      true,
		SkipConfirmation:  false,
		ExcludeExtensions: []string{".pdf", ".png", ".jpg", ".jpeg", ".gif", ".svg"},
	}
}

// Binary extensions commonly found in git history that should be cleaned.
var DefaultBinaryExtensions = []string{
	// Go build outputs (often extensionless)
	"", // Extensionless binaries
	".exe",
	".dll",
	".so",
	".dylib",
	".a",
	".o",
	".out",
	".app",

	// Build artifacts
	".test",
	".bench",
	".prof",

	// Archives
	".zip",
	".tar",
	".gz",
	".bz2",
	".xz",
	".7z",
	".rar",

	// Databases
	".db",
	".sqlite",
	".sqlite3",

	// Serialized data
	".bin",
	".dat",
	".data",

	// Large generated files
	".wasm",
	".class",
	".jar",
}

// ExtensionsToKeep are binary extensions that should typically NOT be removed.
var ExtensionsToKeep = []string{
	".pdf",
	".png",
	".jpg",
	".jpeg",
	".gif",
	".svg",
	".ico",
	".woff",
	".woff2",
	".ttf",
	".eot",
	".mp3",
	".mp4",
	".webm",
	".webp",
}
