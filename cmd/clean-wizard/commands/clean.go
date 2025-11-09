package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

var (
	cleanDryRun  bool
	cleanVerbose bool
)

// NewCleanCommand creates clean command with proper domain types
func NewCleanCommand() *cobra.Command {
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		Long:  `Safely clean old files, package caches, and temporary data from your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Starting system cleanup...")
			ctx := context.Background()
			nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)
			startTime := time.Now()

			if cleanDryRun {
				fmt.Println("🔍 Running in DRY-RUN mode - no files will be deleted")
			}

			// Clean old generations (keep last 3)
			result := nixCleaner.CleanOldGenerations(ctx, 3)

			if result.IsErr() {
				return handleCleanError(result.Error(), cleanDryRun)
			}

			// Display results with user-friendly messages
			displayCleanResults(result.Value(), time.Since(startTime), cleanDryRun)
			return nil
		},
	}

	// Clean command flags
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be cleaned without doing it")
	cleanCmd.Flags().BoolVar(&cleanVerbose, "verbose", false, "Show detailed output")

	return cleanCmd
}

// handleCleanError provides user-friendly error messages
func handleCleanError(err error, isDryRun bool) error {
	if isDryRun {
		return fmt.Errorf("❌ Dry-run failed: %s\n\n💡 Suggestions:\n   • Ensure Nix is installed and accessible\n   • Check if you have permission to read Nix profiles", err.Error())
	}

	return fmt.Errorf("❌ Cleanup failed: %s\n\n💡 Suggestions:\n   • Ensure you have sufficient permissions\n   • Try running with --verbose for more details\n   • Consider using --dry-run first", err.Error())
}

// displayCleanResults shows user-friendly success messages
func displayCleanResults(operation domain.CleanResult, duration time.Duration, isDryRun bool) {
	fmt.Printf("\n✅ Cleanup completed successfully!\n\n")
	fmt.Printf("📊 Results Summary:\n")
	fmt.Printf("   • Items processed: %d\n", operation.ItemsRemoved+operation.ItemsFailed)
	fmt.Printf("   • Items cleaned: %d\n", operation.ItemsRemoved)

	if operation.FreedBytes > 0 {
		fmt.Printf("   • Space freed: %s\n", format.Bytes(operation.FreedBytes))
	}

	if operation.ItemsFailed > 0 {
		fmt.Printf("   ⚠️  Items failed: %d\n", operation.ItemsFailed)
	}

	fmt.Printf("   • Time taken: %s\n", format.Duration(duration))
	fmt.Printf("   • Strategy used: %s\n", operation.Strategy)

	if operation.ItemsFailed > 0 {
		fmt.Printf("\n⚠️  Some items failed to clean\n💡 Try:\n   • Running with --verbose for details\n   • Checking file permissions\n")
	}

	if isDryRun {
		fmt.Printf("\n🔍 This was a DRY-RUN\n💡 To actually clean, run:\n   clean-wizard clean\n")
	}

	fmt.Printf("\n💡 Next steps:\n   • Run 'clean-wizard scan' to see current system state\n   • Consider scheduling regular cleanups\n")
}
