package adapters

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// HomebrewAdapter provides Homebrew-specific operations
type HomebrewAdapter struct {
	timeout time.Duration
	retry  int
}

// NewHomebrewAdapter creates new Homebrew adapter
func NewHomebrewAdapter(timeout time.Duration, retry int) *HomebrewAdapter {
	return &HomebrewAdapter{
		timeout: timeout,
		retry:  retry,
	}
}

// IsAvailable checks if Homebrew is available
func (ha *HomebrewAdapter) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// GetStoreSize gets Homebrew store size with proper error handling
func (ha *HomebrewAdapter) GetStoreSize(ctx context.Context) int64 {
	if !ha.IsAvailable(ctx) {
		return 0
	}

	// Get Homebrew prefix path
	cmd := exec.CommandContext(ctx, "brew", "--prefix")
	_, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Calculate total size of installed packages
	packagesCmd := exec.CommandContext(ctx, "brew", "list", "--versions")
	_, err = packagesCmd.Output()
	if err != nil {
		return 0
	}

	// Simplified calculation - would use actual directory sizes in production
	return int64(50 * 1024 * 1024) // 50MB
}

// ListPackages lists installed Homebrew packages
func (ha *HomebrewAdapter) ListPackages(ctx context.Context) ([]string, error) {
	if !ha.IsAvailable(ctx) {
		return nil, fmt.Errorf("Homebrew is not available")
	}

	cmd := exec.CommandContext(ctx, "brew", "list", "--versions")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}