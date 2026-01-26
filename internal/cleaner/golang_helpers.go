package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// golangHelpers provides utility methods for Go cleaner operations.
type golangHelpers struct{}

// pathExists checks if a path exists.
func (h *golangHelpers) pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getEnv returns an environment variable.
func (h *golangHelpers) getEnv(key string) string {
	return os.Getenv(key)
}

// getGoEnv returns Go environment variable value.
func (h *golangHelpers) getGoEnv(ctx context.Context, key string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Go env %s: %w", key, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getHomeDir returns user's home directory.
func (h *golangHelpers) getHomeDir() string {
	// Try getting from HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Fallback to user profile directory
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		return userProfile
	}

	return ""
}

// getDirSize returns total size of directory recursively.
func (h *golangHelpers) getDirSize(path string) int64 {
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

// getDirModTime returns the most recent modification time in directory.
func (h *golangHelpers) getDirModTime(path string) time.Time {
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