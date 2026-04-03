package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

// cleanResult holds aggregated results from executing multiple cleaners.
type cleanResult struct {
	totalBytesFreed   uint64
	totalItemsRemoved uint
	totalItemsFailed  uint
	cleanerResults    map[string]domain.CleanResult
	skippedCleaners   []string
	failedCleaners    []cleanerFailure
	skippedErrors     map[string]error
	failedErrors      map[string]error
	duration          time.Duration
}

type cleanerFailure struct {
	name  string
	error string
}

// executeCleaners runs all selected cleaners and aggregates their results.
func executeCleaners(
	ctx context.Context,
	selectedCleaners []CleanerType,
	dryRun, verbose bool,
) cleanResult {
	cr := cleanResult{
		cleanerResults: make(map[string]domain.CleanResult),
		skippedErrors:  make(map[string]error),
		failedErrors:   make(map[string]error),
	}

	startTime := time.Now()

	for _, cleanerType := range selectedCleaners {
		result, err := runCleaner(ctx, cleanerType, dryRun, verbose)
		name := getCleanerName(cleanerType)

		if err != nil {
			cr.handleCleanerError(name, err)
			continue
		}

		if verbose {
			fmt.Printf(
				"  [DEBUG] %s: %d bytes (%s), %d items\n",
				name,
				result.FreedBytes,
				format.Bytes(int64(result.FreedBytes)),
				result.ItemsRemoved,
			)
		}

		cr.totalBytesFreed += result.FreedBytes
		cr.totalItemsRemoved += result.ItemsRemoved
		cr.totalItemsFailed += result.ItemsFailed
		cr.cleanerResults[name] = result
	}

	cr.duration = time.Since(startTime)

	return cr
}

func (cr *cleanResult) handleCleanerError(name string, err error) {
	errMsg := err.Error()

	if isNotAvailableError(errMsg) {
		cr.skippedCleaners = append(cr.skippedCleaners, name)
		cr.skippedErrors[name] = err
		fmt.Printf("  ℹ️  Skipped %s: %s\n", name, errMsg)
	} else {
		cr.failedCleaners = append(cr.failedCleaners, cleanerFailure{
			name:  name,
			error: errMsg,
		})
		cr.failedErrors[name] = err
		fmt.Printf("  ❌ Cleaner %s failed: %s\n", name, errMsg)
	}
}

// displayResults renders the final cleanup results to the terminal.
func displayResults(
	cr cleanResult,
	dryRun bool,
	diskBefore *cleaner.DiskUsage,
) {
	fmt.Println()
	fmt.Println(TitleStyle.Render("🧹 Cleanup Results"))
	fmt.Println()

	if dryRun {
		fmt.Println(WarningStyle.Render("⚠️  DRY RUN: No actual changes were made"))
		fmt.Println()
	}

	printCleanResultsTable(cr.cleanerResults, cr.totalBytesFreed, cr.totalItemsRemoved, cr.duration)
	printEncouragement(cr.totalBytesFreed)
	displayDiskUsageAfter(cr, dryRun, diskBefore)
	displayDryRunTip(dryRun)
	displayWarnings(cr)
}

func printEncouragement(totalBytes uint64) {
	if totalBytes > BytesThresholdGB {
		fmt.Println(SuccessStyle.Render("🎉 Great job! You freed over 1 GB of space!"))
	} else if totalBytes > BytesThresholdMB {
		fmt.Println(SuccessStyle.Render("✅ Nice! You freed some space."))
	}
}

func displayDiskUsageAfter(cr cleanResult, dryRun bool, diskBefore *cleaner.DiskUsage) {
	if dryRun || diskBefore == nil {
		return
	}

	diskAfter, err := cleaner.GetDiskUsage("/")
	if err != nil {
		return
	}

	fmt.Println()

	if diskAfter.UsedPercent < diskBefore.UsedPercent {
		freedPercent := diskBefore.UsedPercent - diskAfter.UsedPercent
		fmt.Printf(
			"📊 Disk usage after:  %s %s (-%.1f%%)\n",
			cleaner.DiskUsageBar(diskAfter, 15),
			cleaner.FormatDiskUsage(diskAfter),
			freedPercent,
		)
	} else {
		fmt.Printf(
			"📊 Disk usage after:  %s %s\n",
			cleaner.DiskUsageBar(diskAfter, 15),
			cleaner.FormatDiskUsage(diskAfter),
		)
	}
}

func displayDryRunTip(dryRun bool) {
	if !dryRun {
		return
	}

	fmt.Println()
	fmt.Println(InfoStyle.Render("💡 Tip: Remove --dry-run flag to actually clean:"))
	fmt.Println("   clean-wizard clean --mode standard")
}

func displayWarnings(cr cleanResult) {
	if cr.totalItemsFailed == 0 && len(cr.skippedCleaners) == 0 && len(cr.failedCleaners) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(WarningStyle.Render("⚠️  Warnings:"))

	if cr.totalItemsFailed > 0 {
		fmt.Printf("   • %d item(s) failed to clean\n", cr.totalItemsFailed)
	}

	if len(cr.skippedCleaners) > 0 {
		fmt.Printf("   • %d cleaner(s) skipped (not available)\n", len(cr.skippedCleaners))
	}

	if len(cr.failedCleaners) > 0 {
		fmt.Printf("   • %d cleaner(s) failed\n", len(cr.failedCleaners))
	}
}

// printCleanResultsTable prints clean results as a formatted table.
func printCleanResultsTable(
	results map[string]domain.CleanResult,
	totalBytes uint64,
	totalItems uint,
	duration time.Duration,
) {
	var rows [][]string

	for name, result := range results {
		if result.FreedBytes > 0 || result.ItemsRemoved > 0 {
			rows = append(rows, []string{
				name,
				strconv.FormatUint(uint64(result.ItemsRemoved), 10),
				format.Bytes(int64(result.FreedBytes)),
			})
		}
	}

	if len(rows) == 0 {
		fmt.Println(InfoStyle.Render("No items were cleaned."))

		return
	}

	t := newResultsTable(rows...)

	fmt.Println(t)
	fmt.Println()
	fmt.Printf(
		"📊 Total: %s freed, %s items in %s\n",
		format.Bytes(int64(totalBytes)),
		strconv.FormatUint(uint64(totalItems), 10),
		format.Duration(duration),
	)
}

// isNotAvailableError checks if an error indicates a cleaner is not available.
func isNotAvailableError(errMsg string) bool {
	lowerMsg := strings.ToLower(errMsg)
	unavailableKeywords := []string{
		"not available",
		"not found",
		"not installed",
		"command not found",
		"no such file or directory",
	}

	for _, keyword := range unavailableKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}

	return false
}
