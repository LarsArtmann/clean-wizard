package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DryRunBytesPerItem is the estimated bytes freed per item in dry run mode.
const DryRunBytesPerItem = 300 * 1024 * 1024 // 300MB per item

// CleanItemFunc is a function that cleans a single item of type T.
type CleanItemFunc[T any] func(ctx context.Context, item T, homeDir string) result.Result[domain.CleanResult]

// AvailableCheckFunc is a function that checks if the cleaner is available.
type AvailableCheckFunc func(ctx context.Context) bool

// cleanWithIterator is a shared helper function that performs the common clean pattern.
// It iterates over items, calls the cleanFunc for each, and aggregates results.
func cleanWithIterator[T any](
	ctx context.Context,
	cleanerName string,
	availableCheck AvailableCheckFunc,
	items []T,
	cleanFunc CleanItemFunc[T],
	verbose bool,
	dryRun bool,
) result.Result[domain.CleanResult] {
	if !availableCheck(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("%s not available", cleanerName))
	}

	if dryRun {
		totalBytes := int64(len(items)) * DryRunBytesPerItem
		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, len(items), totalBytes)
		return result.Ok(cleanResult)
	}

	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	homeDir, err := getHomeDir()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get home directory: %w", err))
	}

	for _, item := range items {
		result := cleanFunc(ctx, item, homeDir)
		if result.IsErr() {
			itemsFailed++
			if verbose {
				fmt.Printf("Warning: failed to clean %v: %v\n", item, result.Error())
			}
			continue
		}

		cleanResult := result.Value()
		itemsRemoved++
		bytesFreed += int64(cleanResult.FreedBytes)
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanResult)
}

func getHomeDir() (string, error) {
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

func getDirSize(path string) int64 {
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

func getDirModTime(path string) time.Time {
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