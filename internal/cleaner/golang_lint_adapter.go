package cleaner

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// lintCommandTimeout is the timeout for lint cache cleaning operations.
const lintCommandTimeout = 30 * time.Second

// LintCleaner defines an interface for lint cache cleaning operations.
type LintCleaner interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
	CacheDir() string
}

// GolangciLintCleaner handles golangci-lint cache cleaning.
type GolangciLintCleaner struct {
	verbose bool
	helper  *golangHelpers
}

// NewGolangciLintCleaner creates a new golangci-lint cleaner.
func NewGolangciLintCleaner(verbose bool) *GolangciLintCleaner {
	return &GolangciLintCleaner{
		verbose: verbose,
		helper:  &golangHelpers{},
	}
}

// IsAvailable checks if golangci-lint is installed.
func (glc *GolangciLintCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("golangci-lint")
	return err == nil
}

// Clean removes golangci-lint cache.
func (glc *GolangciLintCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	verbose := glc.verbose
	if !glc.IsAvailable(ctx) {
		if verbose {
			fmt.Println("  ⚠️  golangci-lint not found, skipping cache cleanup")
			fmt.Println("  💡 Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest")
			fmt.Println("  💡 Learn more: https://golangci-lint.run/usage/install/")
		}
		return result.Ok(domain.CleanResult{
			SizeEstimate: domain.SizeEstimate{Status: domain.SizeEstimateStatusUnknown}, // Honest: we don't know the size
			FreedBytes:   0, // Deprecated field
			ItemsRemoved: 0,
			ItemsFailed:  0,
			CleanTime:    0,
			CleanedAt:    time.Now(),
			Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
		})
	}

	// Create a timeout context to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, lintCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "golangci-lint", "cache", "clean")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a timeout error
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return result.Err[domain.CleanResult](fmt.Errorf("golangci-lint cache clean timed out after %v (command may be hanging)", lintCommandTimeout))
		}
		return result.Err[domain.CleanResult](fmt.Errorf("golangci-lint cache clean failed: %w (output: %s)", err, string(output)))
	}

	if verbose {
		fmt.Println("  ✓ golangci-lint cache cleaned")
	}

	return result.Ok(domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{Status: domain.SizeEstimateStatusUnknown}, // Honest: we don't know the size
		FreedBytes:   0,                                  // Deprecated field
		ItemsRemoved: 1,
		ItemsFailed:  0,
		CleanTime:    0,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyConservativeType),
	})
}

// CacheDir returns the lint cache directory path.
func (glc *GolangciLintCleaner) CacheDir() string {
	// Try XDG_CACHE_HOME first
	if xdgCache := glc.helper.getEnv("XDG_CACHE_HOME"); xdgCache != "" {
		cacheDir := xdgCache + "/golangci-lint"
		if glc.helper.pathExists(cacheDir) {
			return cacheDir
		}
	}

	// Fallback to ~/.cache
	if home := glc.helper.getEnv("HOME"); home != "" {
		cacheDir := home + "/.cache/golangci-lint"
		if glc.helper.pathExists(cacheDir) {
			return cacheDir
		}
	}

	return ""
}
