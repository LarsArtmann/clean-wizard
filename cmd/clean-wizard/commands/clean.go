package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

var (
	cleanDryRun  bool
	cleanVerbose bool
)

// NewCleanCommand creates clean command
func NewCleanCommand() *cobra.Command {
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Perform system cleanup",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Starting cleanup...")
			ctx := context.Background()
			nixCleaner := cleaner.NewNixCleaner(cleanVerbose, cleanDryRun)
			startTime := time.Now()

			// Clean old generations (keep last 3)
			result := nixCleaner.CleanOldGenerations(ctx, 3)
			duration := time.Since(startTime)

			if !result.IsOk() {
				return fmt.Errorf("cleanup failed: %w", result.Error())
			}

			operation := result.Value()

			// Display results
			fmt.Println("✅ Cleanup completed!")
			fmt.Printf("🗑️  Items removed: %d\n", operation.ItemsRemoved)
			fmt.Printf("🗑️  Space freed: %s\n", format.Size(operation.FreedBytes))
			fmt.Printf("🗑️  Items failed: %d\n", operation.ItemsFailed)
			fmt.Printf("⏱️  Duration: %s\n", format.Duration(duration))
			fmt.Printf("🗑️  Strategy: %s\n", operation.Strategy)
			fmt.Println()

			if operation.ItemsFailed > 0 {
				fmt.Printf("⚠️  %d items failed to clean\n", operation.ItemsFailed)
			}

			fmt.Println("💡 Run 'clean-wizard scan' to see current system state")
			return nil
		},
	}

	// Clean command flags
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be cleaned without doing it")
	cleanCmd.Flags().BoolVar(&cleanVerbose, "verbose", false, "Show detailed output")

	return cleanCmd
}
