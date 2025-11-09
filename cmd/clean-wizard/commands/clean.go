package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
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
			fmt.Printf("🗑️  Space freed: %s\n", formatSize(operation.FreedBytes))
			fmt.Printf("⏱️  Duration: %s\n", duration)
			fmt.Println()

			if operation.ErrorMessage != "" {
				fmt.Printf("ℹ️  Note: %s\n", operation.ErrorMessage)
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

// formatSize formats bytes for human reading
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
