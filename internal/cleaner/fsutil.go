package cleaner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"golang.org/x/sys/unix"
)

// Byte conversion constants for disk size formatting.
const (
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024
)

// Disk usage and formatting constants.
const (
	// PercentConversionFactor converts fraction to percentage.
	PercentConversionFactor = 100
	// DiskUsageHighThreshold is the threshold for high disk usage warning (in percent).
	DiskUsageHighThreshold = 90
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
			return nil //nolint:nilerr // Skip files/dirs we can't access
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
	Items []types.ScanItem
	Found bool
}

// ScanDirectory scans a directory and returns scan items if it exists and is a directory.
// This helper consolidates the common pattern of checking if a path exists and is a directory,
// then creating scan items for it.
func ScanDirectory(path string, scanType types.ScanType, verbose bool) ScanDirectoryResult {
	result := ScanDirectoryResult{
		Items: make([]types.ScanItem, 0),
		Found: false,
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		result.Found = true
		result.Items = append(result.Items, types.ScanItem{
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
func appendScanItem(
	items []types.ScanItem, path, displayName string, scanType types.ScanType, verbose bool,
) []types.ScanItem {
	item := types.ScanItem{
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
func ScanVersionDirectory(
	ctx context.Context,
	versionsDir, managerName string,
	verbose bool,
) result.Result[[]types.ScanItem] {
	items := make([]types.ScanItem, 0)

	info, err := os.Stat(versionsDir)
	if err != nil || !info.IsDir() {
		return result.Ok(items)
	}

	matches, err := filepath.Glob(filepath.Join(versionsDir, "*"))
	if err != nil {
		return result.Err[[]types.ScanItem](
			fmt.Errorf(
				"failed to find %s versions at versionsDir=%v: %w",
				managerName,
				versionsDir,
				err,
			),
		)
	}

	for _, match := range matches {
		items = appendScanItem(items, match, managerName, types.ScanTypeTemp, verbose)
	}

	return result.Ok(items)
}

// ScanPath scans a directory path constructed from components and returns scan items.
// This is a generic helper that consolidates the common pattern of:
// 1. Constructing a path from components (homeDir + pathComponents)
// 2. Checking if the path exists and is a directory
// 3. Creating a scan item for the directory
// If homeDir is empty and pathComponents contains a complete path, it uses that directly.
// If pattern is provided, it walks the directory to find matching entries instead of scanning.
func ScanPath(
	homeDir string, scanType types.ScanType, displayName string,
	verbose bool, pattern string, pathComponents ...string,
) ScanDirectoryResult {
	result := ScanDirectoryResult{
		Items: make([]types.ScanItem, 0),
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
func CalculateBytesFreed(
	path string, cleanup func() error, verbose bool, cacheName string,
) (bytesFreed, beforeSize, afterSize int64) {
	beforeSize = GetDirSize(path)

	err := cleanup()
	if err != nil {
		// Return 0 bytes freed if cleanup failed, but still calculate size
		afterSize = GetDirSize(path)
		bytesFreed = max(beforeSize-afterSize, 0)

		return bytesFreed, beforeSize, afterSize
	}

	afterSize = GetDirSize(path)
	bytesFreed = max(beforeSize-afterSize, 0)

	if verbose {
		fmt.Printf("  %s size before: %d bytes\n", cacheName, beforeSize)
		fmt.Printf("  %s size after: %d bytes\n", cacheName, afterSize)
		fmt.Printf("  Bytes freed: %d bytes\n", bytesFreed)
	}

	return bytesFreed, beforeSize, afterSize
}

// DiskUsage represents disk usage information for a filesystem.
type DiskUsage struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// GetDiskUsage returns disk usage information for the filesystem containing path.
// Uses golang.org/x/sys/unix for cross-platform support.
func GetDiskUsage(path string) (DiskUsage, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		return DiskUsage{}, fmt.Errorf("failed to get disk usage for %s: %w", path, err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	usedPercent := 0.0
	if total > 0 {
		usedPercent = float64(used) / float64(total) * PercentConversionFactor
	}

	return DiskUsage{
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
	}, nil
}

// FormatDiskUsage returns a formatted string showing disk usage like "229G 224G 4.8G 98%"
// Format: "total used free percent" all in human-readable format.
func FormatDiskUsage(du DiskUsage) string {
	return fmt.Sprintf(
		"%.0fG %.0fG %.1fG %.0f%%",
		float64(du.Total)/bytesPerGB,
		float64(du.Used)/bytesPerGB,
		float64(du.Free)/bytesPerGB,
		du.UsedPercent,
	)
}

// DiskUsageBar returns a visual bar representation of disk usage.
func DiskUsageBar(du DiskUsage, width int) string {
	if width <= 0 {
		width = 20
	}

	filled := int(float64(width) * du.UsedPercent / PercentConversionFactor)
	empty := width - filled

	var bar strings.Builder

	for i := 0; i < filled && i < width; i++ {
		if du.UsedPercent >= DiskUsageHighThreshold {
			bar.WriteString("█") // Red for high usage
		} else {
			bar.WriteString("▓") // Yellow for medium-high or lower
		}
	}

	for i := 0; i < empty && filled+i < width; i++ {
		bar.WriteString("░")
	}

	return fmt.Sprintf("[%s] %.0f%%", bar.String(), du.UsedPercent)
}

// NewEmptyCleanResult returns a conservative result for when there are no items to clean.
func NewEmptyCleanResult() result.Result[types.CleanResult] {
	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyConservativeType,
		0, int64(0),
		types.SizeEstimate{Known: 0, Status: enums.SizeEstimateStatusKnown},
	))
}

// NewDryRunCleanResult returns a dry-run result with the given item count and total bytes.
func NewDryRunCleanResult(itemCount int, totalBytes int64) result.Result[types.CleanResult] {
	return result.Ok(conversions.NewCleanResultWithSizeEstimate(
		enums.StrategyDryRunType,
		itemCount, totalBytes,
		types.SizeEstimate{Known: uint64(totalBytes), Status: enums.SizeEstimateStatusKnown},
	))
}

// NewCleanResultWithMetrics returns a result with the given metrics.
func NewCleanResultWithMetrics(
	itemsRemoved, itemsFailed int,
	bytesFreed int64,
	duration time.Duration,
) result.Result[types.CleanResult] {
	return result.Ok(conversions.NewCleanResultWithTimingAndSize(
		enums.StrategyAggressiveType,
		itemsRemoved, itemsFailed, bytesFreed, duration,
		types.SizeEstimate{Known: uint64(bytesFreed), Status: enums.SizeEstimateStatusKnown},
	))
}

// CleanItemTrashFn trashes a single scan item and returns an error if it fails.
type CleanItemTrashFn func(ctx context.Context, item types.ScanItem) error

// CleanItemLogFn logs a successfully trashed item (typically verbose mode).
// Pass nil to disable per-item logging.
type CleanItemLogFn func(item types.ScanItem)

// LockedMapLookup reads mu.RLock, fetches the value for key from m, and returns it.
// Centralizes the idiomatic "RWMutex + map lookup" pattern used by every registry-style
// container in this package (Registry.Get, MetricsCollector.GetMetrics, etc.).
func LockedMapLookup[K comparable, V any](mu *sync.RWMutex, m map[K]V, key K) (V, bool) {
	mu.RLock()
	defer mu.RUnlock()

	v, ok := m[key]

	return v, ok
}

// ExecuteTrashPipeline runs the common trash flow shared by item-based cleaners:
// scan-result short-circuit, dry-run preview, counter init, per-item trash loop,
// and metrics aggregation. Callers supply the per-item trash operation and an
// optional verbose per-item log callback.
func ExecuteTrashPipeline(
	ctx context.Context,
	scanResult result.Result[[]types.ScanItem],
	dryRun, verbose bool,
	dryRunLabel string,
	trash CleanItemTrashFn,
	logItem CleanItemLogFn,
) result.Result[types.CleanResult] {
	if scanResult.IsErr() {
		return result.Err[types.CleanResult](scanResult.Error())
	}

	items := scanResult.Value()

	if len(items) == 0 {
		return NewEmptyCleanResult()
	}

	var totalBytes int64
	for _, item := range items {
		totalBytes += item.Size
	}

	if dryRun {
		if verbose {
			fmt.Printf("Would trash %d %s (%.2f MB)\n", len(items), dryRunLabel, float64(totalBytes)/bytesPerMB)
		}

		return NewDryRunCleanResult(len(items), totalBytes)
	}

	counters := NewCleanCounters()

	for _, item := range items {
		if err := trash(ctx, item); err != nil {
			counters.RecordFailure(verbose, item.Path, err)

			continue
		}

		counters.RecordSuccess(item.Size)

		if verbose && logItem != nil {
			logItem(item)
		}
	}

	return NewCleanResultWithMetrics(
		counters.ItemsRemoved,
		counters.ItemsFailed,
		counters.BytesFreed,
		counters.Duration(),
	)
}

// CleanCounters aggregates the running totals tracked during a Clean pipeline.
// Callers update it inline as each item is processed, then read the public fields
// to build the result. Centralizing the counter set eliminates the repeated
// `startTime/itemsRemoved/itemsFailed/bytesFreed` boilerplate shared by all
// item-based cleaners.
type CleanCounters struct {
	startTime time.Time

	ItemsRemoved int
	ItemsFailed  int
	BytesFreed   int64
}

// NewCleanCounters returns counters with startTime set to time.Now().
// Counters are populated incrementally by Record* methods.
func NewCleanCounters() CleanCounters {
	return CleanCounters{startTime: time.Now()} //nolint:exhaustruct
}

// Duration returns the elapsed time since the counters were created.
func (c *CleanCounters) Duration() time.Duration {
	return time.Since(c.startTime)
}

// RecordSuccess adds the freed bytes to the success total and increments ItemsRemoved.
func (c *CleanCounters) RecordSuccess(freedBytes int64) {
	c.ItemsRemoved++
	c.BytesFreed += freedBytes
}

// RecordFailure increments the failure counter and prints a verbose warning.
func (c *CleanCounters) RecordFailure(verbose bool, label any, err error) {
	c.ItemsFailed++

	if verbose {
		fmt.Printf("Warning: failed to clean %v: %v\n", label, err)
	}
}

// NormalizePaths normalizes a slice of paths using filepath.Clean.
// This consolidates the common path normalization pattern used in multiple cleaners.
func NormalizePaths(paths []string) []string {
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		normalized = append(normalized, filepath.Clean(path))
	}

	return normalized
}
