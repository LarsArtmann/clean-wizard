package cleaner

import (
	"context"
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
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.HomeDir, nil
	}

	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile, nil
	}

	return "", fmt.Errorf("unable to determine home directory")
}

// GetDirSize returns total size of directory recursively.
func GetDirSize(path string) int64 {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}

	return size
}

// GetDirModTime returns most recent modification time in directory.
func GetDirModTime(path string) time.Time {
	var modTime time.Time

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.ModTime().After(modTime) {
			modTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
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
		items = append(items, domain.ScanItem{
			Path:     match,
			Size:     GetDirSize(match),
			Created:  GetDirModTime(match),
			ScanType: domain.ScanTypeTemp,
		})

		if verbose {
			fmt.Printf("Found %s version: %s\n", managerName, filepath.Base(match))
		}
	}

	return result.Ok(items)
}