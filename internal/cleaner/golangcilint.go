package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// golangciLintCommandTimeout is the timeout for golangci-lint operations.
const golangciLintCommandTimeout = 30 * time.Second

// Cache status parsing constants.
const (
	cacheStatusMinLines = 2

	// Binary byte conversion constants for size parsing (IEC units).
	bytesPerKiB = 1024
	bytesPerMiB = 1024 * 1024
	bytesPerGiB = 1024 * 1024 * 1024
	bytesPerTiB = 1024 * 1024 * 1024 * 1024

	// Decimal byte conversion constants (SI units).
	bytesPerKBDecimal = 1000
	bytesPerMBDecimal = 1000 * 1000
	bytesPerGBDecimal = 1000 * 1000 * 1000
	bytesPerTBDecimal = 1000 * 1000 * 1000 * 1000
)

// GolangciLintCacheCleaner handles golangci-lint cache cleanup.
type GolangciLintCacheCleaner struct {
	CleanerBase
}

// NewGolangciLintCacheCleaner creates a new golangci-lint cache cleaner.
func NewGolangciLintCacheCleaner(verbose, dryRun bool) *GolangciLintCacheCleaner {
	return &GolangciLintCacheCleaner{
		CleanerBase: NewCleanerBase(verbose, dryRun),
	}
}

// Type returns operation type.
func (glcc *GolangciLintCacheCleaner) Type() domain.OperationType {
	return domain.OperationTypeGolangciLintCache
}

// Name returns the cleaner name for result tracking.
func (glcc *GolangciLintCacheCleaner) Name() string {
	return "golangci-lint-cache"
}

// IsAvailable checks if golangci-lint is installed.
func (glcc *GolangciLintCacheCleaner) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("golangci-lint")

	return err == nil
}

// ValidateSettings validates settings.
func (glcc *GolangciLintCacheCleaner) ValidateSettings(settings *domain.OperationSettings) error {
	return nil
}

// cacheStatus holds the parsed output of golangci-lint cache status.
type cacheStatus struct {
	Dir  string
	Size int64
}

// parseCacheStatus parses the output of "golangci-lint cache status".
// Output format:
//
//	Dir: /Users/user/Library/Caches/golangci-lint
//	Size: 3.1KiB
func parseCacheStatus(output string) (*cacheStatus, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < cacheStatusMinLines {
		return nil, fmt.Errorf("unexpected output format: %s", output)
	}

	status := &cacheStatus{}

	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "Dir:"); ok {
			status.Dir = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "Size:"); ok {
			sizeStr := strings.TrimSpace(after)

			size, err := parseSize(sizeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse size %q: %w", sizeStr, err)
			}

			status.Size = size
		}
	}

	if status.Dir == "" {
		return nil, fmt.Errorf("cache directory not found in output: %s", output)
	}

	return status, nil
}

// parseSize parses a size string like "3.1KiB" or "1.5MiB" into bytes.
// Supports units: B, KiB, MiB, GiB, TiB, KB, MB, GB, TB (binary units take precedence).
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	if sizeStr == "" {
		return 0, errors.New("empty size string")
	}

	var (
		number float64
		unit   string
	)

	n, err := fmt.Sscanf(sizeStr, "%f%s", &number, &unit)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// Binary units (powers of 1024)
	switch strings.ToUpper(unit) {
	case "B", "BYTE", "BYTES":
		return int64(number), nil
	case "KIB":
		return int64(number * bytesPerKiB), nil
	case "MIB":
		return int64(number * bytesPerMiB), nil
	case "GIB":
		return int64(number * bytesPerGiB), nil
	case "TIB":
		return int64(number * bytesPerTiB), nil
	case "KB":
		return int64(number * bytesPerKBDecimal), nil
	case "MB":
		return int64(number * bytesPerMBDecimal), nil
	case "GB":
		return int64(number * bytesPerGBDecimal), nil
	case "TB":
		return int64(number * bytesPerTBDecimal), nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}

// getCacheStatus returns the cache status by running "golangci-lint cache status".
func (glcc *GolangciLintCacheCleaner) getCacheStatus(ctx context.Context) (*cacheStatus, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, golangciLintCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "golangci-lint", "cache", "status")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf(
				"golangci-lint cache status timed out after %v",
				golangciLintCommandTimeout,
			)
		}

		return nil, fmt.Errorf(
			"golangci-lint cache status failed: %w (output: %s)",
			err, string(output),
		)
	}

	return parseCacheStatus(string(output))
}

// Scan scans for golangci-lint cache.
func (glcc *GolangciLintCacheCleaner) Scan(ctx context.Context) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0, 1)

	if !glcc.IsAvailable(ctx) {
		return result.Ok(items)
	}

	status, err := glcc.getCacheStatus(ctx)
	if err != nil {
		if glcc.verbose {
			fmt.Printf("Warning: failed to get golangci-lint cache status: %v\n", err)
		}

		return result.Ok(items)
	}

	items = append(items, domain.ScanItem{
		Path:     status.Dir,
		Size:     status.Size,
		Created:  GetDirModTime(status.Dir),
		ScanType: domain.ScanTypeCache,
	})

	if glcc.verbose {
		fmt.Printf("Found golangci-lint cache: %s\n", status.Dir)
	}

	return result.Ok(items)
}

// Clean removes golangci-lint cache.
func (glcc *GolangciLintCacheCleaner) Clean(ctx context.Context) result.Result[domain.CleanResult] {
	if !glcc.IsAvailable(ctx) {
		if glcc.verbose {
			fmt.Println("  ⚠️  golangci-lint not found, skipping cache cleanup")
			fmt.Println(
				"  💡 Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
			)
			fmt.Println("  💡 Learn more: https://golangci-lint.run/usage/install/")
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			domain.StrategyConservativeType,
			0, int64(0),
			domain.SizeEstimate{Status: domain.SizeEstimateStatusUnknown},
		))
	}

	if glcc.dryRun {
		itemsResult := glcc.Scan(ctx)
		if itemsResult.IsErr() {
			return result.Err[domain.CleanResult](itemsResult.Error())
		}

		items := itemsResult.Value()

		var totalSize int64
		for _, item := range items {
			totalSize += item.Size
		}

		return result.Ok(conversions.NewCleanResultWithSizeEstimate(
			domain.StrategyDryRunType,
			len(items), totalSize,
			domain.SizeEstimate{Known: uint64(totalSize)},
		))
	}

	var bytesFreed int64

	status, err := glcc.getCacheStatus(ctx)
	if err != nil {
		if glcc.verbose {
			fmt.Printf("Warning: failed to get cache status: %v\n", err)
		}
	} else {
		bytesFreed = status.Size
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, golangciLintCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "golangci-lint", "cache", "clean")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return result.Err[domain.CleanResult](
				fmt.Errorf(
					"golangci-lint cache clean timed out after %v",
					golangciLintCommandTimeout,
				),
			)
		}

		return result.Err[domain.CleanResult](
			fmt.Errorf(
				"golangci-lint cache clean failed: %w (output: %s)",
				err, string(output),
			),
		)
	}

	if glcc.verbose {
		fmt.Println("  ✓ golangci-lint cache cleaned")
	}

	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		domain.StrategyConservativeType,
		1, bytesFreed,
		domain.SizeEstimate{Known: uint64(bytesFreed)},
	))
}

// GetVerbose returns the verbose setting.
