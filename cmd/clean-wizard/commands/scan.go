package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/scan"
	"github.com/spf13/cobra"
)

// NewScanCommand creates scan command
func NewScanCommand(verbose bool) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Scanning system...")

			scanner := scan.NewMockScanner()
			result := scanner.Scan(context.Background())

			if !result.IsOk() {
				return fmt.Errorf("scan failed: %w", result.Error())
			}

			scanResults := result.Value()
			duration := time.Duration(0)

			fmt.Println("✅ Scan completed!")
			fmt.Printf("📊 Total cleanable: %d items\n", scanResults.CleanableItems)
			fmt.Printf("📦 Found: %d items\n", len(scanResults.Items))
			fmt.Printf("⏱️  Scan time: %s\n", duration)
			fmt.Println()

			for _, item := range scanResults.Items {
				fmt.Printf("📦 %s: %s\n", item.Type, formatSize(item.Size))
			}

			if scanResults.CleanableItems > 0 {
				fmt.Println()
				fmt.Println("💡 Run 'clean-wizard clean' to start cleaning")
			} else {
				fmt.Println("🎉 Your system is already clean!")
			}

			return nil
		},
	}
}

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
