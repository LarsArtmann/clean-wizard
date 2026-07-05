package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/format"
)

// displayResults renders the final cleanup results to the terminal.
func displayResults(
	wr *execution.WorkflowResult,
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

	printCleanResultsTable(wr.CleanResultsMap(), wr.TotalBytesFreed, wr.TotalItemsRemoved, wr.Duration)
	printEncouragement(wr.TotalBytesFreed)
	displayDiskUsageAfter(dryRun, diskBefore)
	displayDryRunTip(dryRun)
	displayWarnings(wr)
}

func printEncouragement(totalBytes uint64) {
	if totalBytes > BytesThresholdGB {
		fmt.Println(SuccessStyle.Render("🎉 Great job! You freed over 1 GB of space!"))
	} else if totalBytes > BytesThresholdMB {
		fmt.Println(SuccessStyle.Render("✅ Nice! You freed some space."))
	}
}

func displayDiskUsageAfter(dryRun bool, diskBefore *cleaner.DiskUsage) {
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
			cleaner.DiskUsageBar(diskAfter, DiskUsageBarWidth),
			cleaner.FormatDiskUsage(diskAfter),
			freedPercent,
		)
	} else {
		fmt.Printf(
			"📊 Disk usage after:  %s %s\n",
			cleaner.DiskUsageBar(diskAfter, DiskUsageBarWidth),
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

func displayWarnings(wr *execution.WorkflowResult) {
	skipped := wr.Skipped()
	failed := wr.Failed()

	if wr.TotalItemsFailed == 0 && len(skipped) == 0 && len(failed) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(WarningStyle.Render("⚠️  Warnings:"))

	if wr.TotalItemsFailed > 0 {
		fmt.Printf("   • %d item(s) failed to clean\n", wr.TotalItemsFailed)
	}

	if len(skipped) > 0 {
		fmt.Printf("   • %d cleaner(s) skipped (not available)\n", len(skipped))
		for _, s := range skipped {
			fmt.Printf("     ℹ️  Skipped %s: %s\n", s.Name, s.Err.Error())
		}
	}

	if len(failed) > 0 {
		fmt.Printf("   • %d cleaner(s) failed\n", len(failed))
		for _, f := range failed {
			fmt.Printf("     ❌ %s failed: %s\n", f.Name, f.Err.Error())
		}
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
