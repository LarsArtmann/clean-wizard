package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// GitHistoryScanner scans git history for large binary files.
type GitHistoryScanner struct {
	repoPath     string
	verbose      bool
	minSizeBytes int64
	excludeExts  map[string]bool
	includeExts  map[string]bool
	excludePaths []string
	maxFiles     int
}

// GitHistoryScannerOption is a functional option for the scanner.
type GitHistoryScannerOption func(*GitHistoryScanner)

// NewGitHistoryScanner creates a new git history scanner.
func NewGitHistoryScanner(repoPath string, opts ...GitHistoryScannerOption) *GitHistoryScanner {
	s := &GitHistoryScanner{
		repoPath:     repoPath,
		minSizeBytes: 1024 * 1024, // Default 1MB
		excludeExts:  make(map[string]bool),
		includeExts:  make(map[string]bool),
		maxFiles:     100,
	}

	// Add default extensions to keep (not remove)
	for _, ext := range domain.ExtensionsToKeep {
		s.excludeExts[strings.ToLower(ext)] = true
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithMinSizeMB sets the minimum file size in MB.
func WithMinSizeMB(mb int) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		if mb > 0 {
			s.minSizeBytes = int64(mb) * 1024 * 1024
		}
	}
}

// WithExcludeExtensions sets extensions to exclude.
func WithExcludeExtensions(exts []string) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		for _, ext := range exts {
			s.excludeExts[strings.ToLower(ext)] = true
		}
	}
}

// WithIncludeExtensions sets extensions to include (only these will be found).
func WithIncludeExtensions(exts []string) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		for _, ext := range exts {
			s.includeExts[strings.ToLower(ext)] = true
		}
	}
}

// WithExcludePaths sets path patterns to exclude.
func WithExcludePaths(paths []string) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		s.excludePaths = paths
	}
}

// WithMaxFiles sets the maximum number of files to return.
func WithMaxFiles(max int) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		if max > 0 {
			s.maxFiles = max
		}
	}
}

// WithVerbose sets verbose mode.
func WithVerbose(verbose bool) GitHistoryScannerOption {
	return func(s *GitHistoryScanner) {
		s.verbose = verbose
	}
}

// Scan scans git history for large binary files.
func (s *GitHistoryScanner) Scan(ctx context.Context) (*domain.GitHistoryScanResult, error) {
	start := time.Now()

	// Check if this is a git repo
	if !s.isGitRepo(ctx) {
		return nil, fmt.Errorf("not a git repository: %s", s.repoPath)
	}

	// Get all large blobs from history
	files, err := s.findLargeBlobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scan git history: %w", err)
	}

	// Filter and sort files
	files = s.filterFiles(files)
	s.sortBySize(files)

	// Limit results
	if len(files) > s.maxFiles {
		files = files[:s.maxFiles]
	}

	// Calculate total size
	var totalBytes int64
	for _, f := range files {
		totalBytes += f.SizeBytes
	}

	return &domain.GitHistoryScanResult{
		Files:      files,
		TotalBytes: totalBytes,
		TotalFiles: len(files),
		RepoPath:   s.repoPath,
		ScannedAt:  time.Now(),
		Duration:   time.Since(start),
	}, nil
}

// isGitRepo checks if the path is a git repository.
func (s *GitHistoryScanner) isGitRepo(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Dir = s.repoPath
	return cmd.Run() == nil
}

// findLargeBlobs finds all large blobs in git history.
func (s *GitHistoryScanner) findLargeBlobs(ctx context.Context) ([]domain.GitHistoryFile, error) {
	// Use git rev-list to get all objects, then cat-file to get sizes
	// Command: git rev-list --objects --all | git cat-file --batch-check

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Get all objects with their types and sizes
	cmd := exec.CommandContext(ctx, "git", "-C", s.repoPath,
		"rev-list", "--objects", "--all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w", err)
	}

	// Parse object list and get blob info
	lines := strings.Split(string(output), "\n")
	var files []domain.GitHistoryFile
	seenPaths := make(map[string]bool) // Avoid duplicates

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}

		blobHash := parts[0]
		var path string
		if len(parts) > 1 {
			path = parts[1]
		}

		// Skip if no path (commit/tree objects)
		if path == "" {
			continue
		}

		// Skip if already seen
		if seenPaths[path] {
			continue
		}
		seenPaths[path] = true

		// Get blob info
		file, err := s.getBlobInfo(ctx, blobHash, path)
		if err != nil {
			if s.verbose {
				fmt.Printf("Warning: failed to get blob info for %s: %v\n", blobHash[:8], err)
			}
			continue
		}

		// Skip if below size threshold
		if file.SizeBytes < s.minSizeBytes {
			continue
		}

		files = append(files, file)
	}

	// Enrich with commit information
	files, err = s.enrichWithCommitInfo(ctx, files)
	if err != nil && s.verbose {
		fmt.Printf("Warning: failed to enrich with commit info: %v\n", err)
	}

	// Check which files still exist in HEAD
	s.markDeletedFiles(ctx, files)

	return files, nil
}

// getBlobInfo gets information about a blob.
func (s *GitHistoryScanner) getBlobInfo(ctx context.Context, blobHash, path string) (domain.GitHistoryFile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Use cat-file to get type and size
	cmd := exec.CommandContext(ctx, "git", "-C", s.repoPath,
		"cat-file", "-s", blobHash)
	output, err := cmd.Output()
	if err != nil {
		return domain.GitHistoryFile{}, fmt.Errorf("git cat-file failed: %w", err)
	}

	sizeStr := strings.TrimSpace(string(output))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return domain.GitHistoryFile{}, fmt.Errorf("invalid size: %s", sizeStr)
	}

	// Check if it's actually a blob (binary-like)
	cmd = exec.CommandContext(ctx, "git", "-C", s.repoPath,
		"cat-file", "-t", blobHash)
	typeOutput, err := cmd.Output()
	if err != nil {
		return domain.GitHistoryFile{}, fmt.Errorf("git cat-file -t failed: %w", err)
	}

	objType := strings.TrimSpace(string(typeOutput))
	if objType != "blob" {
		return domain.GitHistoryFile{}, fmt.Errorf("not a blob: %s", objType)
	}

	// Get extension
	ext := strings.ToLower(filepath.Ext(path))

	return domain.GitHistoryFile{
		Path:      path,
		SizeBytes: size,
		BlobHash:  blobHash,
		Extension: ext,
	}, nil
}

// enrichWithCommitInfo adds commit hash, date, and author information.
func (s *GitHistoryScanner) enrichWithCommitInfo(ctx context.Context, files []domain.GitHistoryFile) ([]domain.GitHistoryFile, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for i := range files {
		// Find the commit that added this file
		cmd := exec.CommandContext(ctx, "git", "-C", s.repoPath,
			"log", "--diff-filter=A", "--format=%H %ct %an", "-1", "--", files[i].Path)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		line := strings.TrimSpace(string(output))
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 1 {
			files[i].CommitHash = parts[0]
		}
		if len(parts) >= 2 {
			if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				files[i].CommitDate = time.Unix(ts, 0)
			}
		}
		if len(parts) >= 3 {
			files[i].Author = parts[2]
		}
	}

	return files, nil
}

// markDeletedFiles checks which files no longer exist in HEAD.
func (s *GitHistoryScanner) markDeletedFiles(_ context.Context, files []domain.GitHistoryFile) {
	for i := range files {
		fullPath := filepath.Join(s.repoPath, files[i].Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			files[i].IsDeleted = true
		}
	}
}

// filterFiles filters files based on extension and path rules.
func (s *GitHistoryScanner) filterFiles(files []domain.GitHistoryFile) []domain.GitHistoryFile {
	var result []domain.GitHistoryFile

	for _, f := range files {
		ext := strings.ToLower(f.Extension)

		// If include extensions are set, only include those
		if len(s.includeExts) > 0 {
			if !s.includeExts[ext] {
				continue
			}
		} else {
			// Exclude extensions we want to keep
			if s.excludeExts[ext] {
				continue
			}
		}

		// Check path exclusions
		excluded := false
		for _, excludePath := range s.excludePaths {
			if strings.HasPrefix(f.Path, excludePath) || strings.Contains(f.Path, excludePath) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check if it's likely a binary (has no extension or binary extension)
		if !s.isLikelyBinary(f) {
			continue
		}

		result = append(result, f)
	}

	return result
}

// isLikelyBinary checks if a file is likely a binary based on extension and path.
func (s *GitHistoryScanner) isLikelyBinary(f domain.GitHistoryFile) bool {
	ext := strings.ToLower(f.Extension)

	// Check against known binary extensions
	if slices.Contains(domain.DefaultBinaryExtensions, ext) {
		return true
	}

	// Extensionless files in common binary directories
	if ext == "" {
		binaryDirs := []string{"bin/", "dist/", "build/", "out/", "target/"}
		for _, dir := range binaryDirs {
			if strings.HasPrefix(f.Path, dir) {
				return true
			}
		}

		// Check for common binary names
		binaryNames := []string{"main", "app", "server", "cli", "cmd", "run", "start", "stop"}
		base := filepath.Base(f.Path)
		if slices.Contains(binaryNames, base) {
			return true
		}

		// Files with .test suffix (Go test binaries)
		if strings.HasSuffix(f.Path, ".test") {
			return true
		}
	}

	return false
}

// sortBySize sorts files by size in descending order.
func (s *GitHistoryScanner) sortBySize(files []domain.GitHistoryFile) {
	slices.SortFunc(files, func(a, b domain.GitHistoryFile) int {
		if a.SizeBytes > b.SizeBytes {
			return -1
		} else if a.SizeBytes < b.SizeBytes {
			return 1
		}
		return 0
	})
}

// GetRepoSize returns the size of the .git directory.
func (s *GitHistoryScanner) GetRepoSize() (int64, error) {
	gitDir := filepath.Join(s.repoPath, ".git")
	var size int64

	err := filepath.Walk(gitDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // Skip files we can't access
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
