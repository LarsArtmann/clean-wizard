package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// golangHelpers provides utility methods for Go cleaner operations.
type golangHelpers struct{}

// goEnvTimeout is the timeout for Go environment variable queries.
// These should be fast, so we use a shorter timeout.
const goEnvTimeout = 10 * time.Second

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
	// Create a timeout context to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, goEnvTimeout)
	defer cancel()
	
	cmd := exec.CommandContext(timeoutCtx, "go", "env", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a timeout error
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("go env %s timed out after %v (command may be hanging)", key, goEnvTimeout)
		}
		return "", fmt.Errorf("failed to get Go env %s: %w", key, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getHomeDir returns user's home directory.
func (h *golangHelpers) getHomeDir() string {
	home, err := GetHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// getDirSize returns total size of directory recursively.
func (h *golangHelpers) getDirSize(path string) int64 {
	return GetDirSize(path)
}

// getDirModTime returns the most recent modification time in directory.
func (h *golangHelpers) getDirModTime(path string) time.Time {
	return GetDirModTime(path)
}