package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// GetHomeDir returns user's home directory.
func GetHomeDir() (string, error) {
	// Check environment variables first (allows testing and overrides)
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile, nil
	}

	// Fall back to system user
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.HomeDir, nil
	}

	return "", errors.New("unable to determine home directory")
}

// walkDirectory walks the directory tree starting at path, collecting size and modTime.
// This consolidates the common directory walking pattern to avoid duplication.
func walkDirectory(path string) (size int64, modTime time.Time, ok bool) {
	err := filepath.Walk(path, func(_ string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		if info.ModTime().After(modTime) {
			modTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return 0, time.Time{}, false
	}
	return size, modTime, true
}

// GetDirSize returns total size of directory recursively.
func GetDirSize(path string) int64 {
	size, _, ok := walkDirectory(path)
	if !ok {
		return 0
	}
	return size
}

// GetDirModTime returns the most recent modification time in directory.
func GetDirModTime(path string) time.Time {
	_, modTime, ok := walkDirectory(path)
	if !ok {
		return time.Time{}
	}
	return modTime
}

// ScanDirectoryResult represents the result of scanning a directory.
type ScanDirectoryResult struct {
	Items []domain.ScanItem
	Found bool
}

// ScanDirectory scans a directory and returns scan items if it exists and is a directory.
// This helper consolidates the common pattern of checking if a path exists and is a directory,
// then creating scan items for it.
func ScanDirectory(path string, scanType domain.ScanType, verbose bool) ScanDirectoryResult {
	result := ScanDirectoryResult{
		Items: make([]domain.ScanItem, 0),
		Found: false,
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		result.Found = true
		result.Items = append(result.Items, domain.ScanItem{
			Path:     path,
			Size:     GetDirSize(path),
			Created:  GetDirModTime(path),
			ScanType: scanType,
		})

		if verbose {
			fmt.Printf("Found: %s\n", filepath.Base(path))
		}
	}

	return result
}

// appendScanItem appends a scan item for a directory to the items slice with verbose output.
func appendScanItem(items []domain.ScanItem, path, displayName string, scanType domain.ScanType, verbose bool) []domain.ScanItem {
	item := domain.ScanItem{
		Path:     path,
		Size:     GetDirSize(path),
		Created:  GetDirModTime(path),
		ScanType: scanType,
	}
	items = append(items, item)

	if verbose {
		fmt.Printf("Found %s: %s\n", displayName, filepath.Base(path))
	}

	return items
}

// ScanVersionDirectory scans a version directory for a language version manager.
// It returns scan items for each version subdirectory found.
func ScanVersionDirectory(ctx context.Context, versionsDir, managerName string, verbose bool) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	info, err := os.Stat(versionsDir)
	if err != nil || !info.IsDir() {
		return result.Ok(items)
	}

	matches, err := filepath.Glob(filepath.Join(versionsDir, "*"))
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to find %s versions: %w", managerName, err))
	}

	for _, match := range matches {
		items = appendScanItem(items, match, managerName, domain.ScanTypeTemp, verbose)
	}

	return result.Ok(items)
}

// ScanPath scans a directory path constructed from components and returns scan items.
// This is a generic helper that consolidates the common pattern of:
// 1. Constructing a path from components (homeDir + pathComponents)
// 2. Checking if the path exists and is a directory
// 3. Creating a scan item for the directory
// If homeDir is empty and pathComponents contains a complete path, it uses that directly.
// If pattern is provided, it walks the directory to find matching entries instead of scanning the directory itself.
func ScanPath(homeDir string, scanType domain.ScanType, displayName string, verbose bool, pattern string, pathComponents ...string) ScanDirectoryResult {
	result := ScanDirectoryResult{
		Items: make([]domain.ScanItem, 0),
		Found: false,
	}

	var fullPath string
	if homeDir == "" {
		fullPath = filepath.Join(pathComponents...)
	} else {
		fullPath = filepath.Join(append([]string{homeDir}, pathComponents...)...)
	}

	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		result.Found = true

		if pattern != "" {
			// Walk the directory to find matching entries
			walkPattern := filepath.Join(fullPath, pattern)
			matches, err := filepath.Glob(walkPattern)
			if err != nil {
				return result
			}

			for _, match := range matches {
				result.Items = appendScanItem(result.Items, match, displayName, scanType, verbose)
			}
		} else {
			// Scan the directory itself
			result.Items = appendScanItem(result.Items, fullPath, displayName, scanType, verbose)
		}
	}

	return result
}

// CalculateBytesFreed calculates the bytes freed from a directory after a cleanup operation.
// This consolidates the common pattern of:
// 1. Getting directory size before cleanup
// 2. Executing the cleanup function
// 3. Getting directory size after cleanup
// 4. Calculating the difference (bytes freed)
// 5. Logging verbose output if requested
// Returns the bytes freed (always non-negative), beforeSize, and afterSize for logging.
func CalculateBytesFreed(path string, cleanup func() error, verbose bool, cacheName string) (bytesFreed int64, beforeSize int64, afterSize int64) {
	beforeSize = GetDirSize(path)

	err := cleanup()
	if err != nil {
		// Return 0 bytes freed if cleanup failed, but still calculate size
		afterSize = GetDirSize(path)
		bytesFreed = beforeSize - afterSize
		if bytesFreed < 0 {
			bytesFreed = 0
		}
		return bytesFreed, beforeSize, afterSize
	}

	afterSize = GetDirSize(path)
	bytesFreed = beforeSize - afterSize
	if bytesFreed < 0 {
		bytesFreed = 0
	}

	if verbose {
		fmt.Printf("  %s size before: %d bytes\n", cacheName, beforeSize)
		fmt.Printf("  %s size after: %d bytes\n", cacheName, afterSize)
		fmt.Printf("  Bytes freed: %d bytes\n", bytesFreed)
	}

	return bytesFreed, beforeSize, afterSize
}
