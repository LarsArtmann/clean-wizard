package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/spf13/cobra"
)

// NewScanCommand creates scan command
func NewScanCommand(verbose bool) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Scanning system...")
			ctx := context.Background()
			nixCleaner := cleaner.NewNixCleaner(verbose, true) // dry-run
			startTime := time.Now()

			// Get generations
			genResult := nixCleaner.ListGenerations(ctx)
			if genResult.IsErr() {
				return fmt.Errorf("scan failed: %w", genResult.Error())
			}

			generations := genResult.Value()
			duration := time.Since(startTime)

			// Calculate stats
			var cleanableCount int
			for _, gen := range generations {
				if !gen.Current {
					cleanableCount++
				}
			}

			// Display results
			fmt.Println("✅ Scan completed!")
			fmt.Printf("📊 Total generations: %d\n", len(generations))
			fmt.Printf("📦 Cleanable: %d\n", cleanableCount)
			fmt.Printf("⏱️  Scan time: %s\n", duration)
			fmt.Println()

			for _, gen := range generations {
				status := "✅ Current"
				if !gen.Current {
					status = "🗑️  Cleanable"
				}
				fmt.Printf("%s Generation %d (%s)\n", status, gen.ID, gen.Date.Format("2006-01-02"))
			}

			if cleanableCount > 0 {
				fmt.Println()
				fmt.Println("💡 Run 'clean-wizard clean' to start cleaning")
			} else {
				fmt.Println("🎉 Your system is already clean!")
			}

			return nil
		},
	}
}
