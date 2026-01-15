package commands

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCleanCommand creates a simplified clean command with TUI.
func NewCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean old Nix generations",
		Long:  `Interactively select and clean old Nix generations.`,
		RunE:  runCleanCommand,
	}

	return cmd
}

// runCleanCommand executes the clean command with TUI.
func runCleanCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("🔍 Scanning for Nix generations...")

	// List all generations using adapter
	nixAdapter := adapters.NewNixAdapter(0, 0)
	result := nixAdapter.ListGenerations(ctx)

	if result.IsErr() {
		return fmt.Errorf("failed to list generations: %w", result.Error())
	}

	generations := result.Value()

	if len(generations) == 0 {
		fmt.Println("✅ No generations found to clean.")
		return nil
	}

	// Separate current and historical generations
	var currentGen domain.NixGeneration
	var historicalGens []domain.NixGeneration
	for _, gen := range generations {
		if gen.Current.IsCurrent() {
			currentGen = gen
		} else {
			historicalGens = append(historicalGens, gen)
		}
	}

	if len(historicalGens) == 0 {
		fmt.Println("✅ No old generations to clean.")
		return nil
	}

	// Show current generation info
	fmt.Printf("✓ Current generation: %d (from %s)\n", currentGen.ID, formatRelativeTime(currentGen.Date))
	fmt.Printf("✓ Found %d old generations\n\n", len(historicalGens))

	// Create variable to store selected IDs
	var selectedIDs []int

	// Build Huh form with multi-select
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select generations to clean").
				Description("Choose which old generations to delete (Space to select, Enter to confirm)").
				Options(
					func() []huh.Option[int] {
						opts := make([]huh.Option[int], len(historicalGens))
						for i, gen := range historicalGens {
							sizeEstimate := format.Bytes(gen.EstimateSize())
							desc := fmt.Sprintf("%s ago • ~%s", formatRelativeTime(gen.Date), sizeEstimate)
							opts[i] = huh.NewOption(desc, gen.ID)
						}
						return opts
					}()...,
				).
				Value(&selectedIDs),
		),
	)

	// Run the TUI form
	if err := form.Run(); err != nil {
		return fmt.Errorf("form error: %w", err)
	}

	// If user cancelled or selected nothing
	if len(selectedIDs) == 0 {
		fmt.Println("❌ No generations selected. Nothing to clean.")
		return nil
	}

	fmt.Printf("\n🗑️  Cleaning %d generation(s)...\n", len(selectedIDs))

	// Filter selected generations
	var toClean []domain.NixGeneration
	for _, gen := range historicalGens {
		if slices.Contains(selectedIDs, gen.ID) {
			toClean = append(toClean, gen)
		}
	}

	// Show what will be deleted
	fmt.Println("\nWill delete:")
	totalBytes := int64(0)
	for _, gen := range toClean {
		size := gen.EstimateSize()
		totalBytes += size
		fmt.Printf("  • Generation %d (from %s) ~ %s\n", gen.ID, formatRelativeTime(gen.Date), format.Bytes(size))
	}

	fmt.Printf("\nTotal space to free: %s\n", format.Bytes(totalBytes))

	// Confirm deletion
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Proceed with deletion?").
				Affirmative("Yes, delete them").
				Negative("No, cancel").
				Value(&confirm),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("confirmation error: %w", err)
	}

	if !confirm {
		fmt.Println("❌ Cancelled. No changes made.")
		return nil
	}

	// Perform the cleanup
	fmt.Println("\n🧹 Cleaning...")

	startTime := time.Now()

	// Remove selected generations
	itemsRemoved := 0
	for _, gen := range toClean {
		result := nixAdapter.RemoveGeneration(ctx, gen.ID)
		if result.IsErr() {
			fmt.Printf("  ⚠️  Failed to remove generation %d: %v\n", gen.ID, result.Error())
		} else {
			fmt.Printf("  ✓ Removed generation %d\n", gen.ID)
			itemsRemoved++
		}
	}

	// Run garbage collection
	fmt.Println("  🔄 Running garbage collection...")
	nixAdapter.CollectGarbage(ctx)

	duration := time.Since(startTime)

	fmt.Printf("\n✅ Cleanup completed in %s\n", duration.String())
	fmt.Printf("   • Removed %d generation(s)\n", itemsRemoved)
	fmt.Printf("   • Freed approximately %s\n", format.Bytes(totalBytes))

	return nil
}

// formatRelativeTime formats a time as a relative string (e.g., "2 days ago").
func formatRelativeTime(t time.Time) string {
	dur := time.Since(t)
	hours := int(dur.Hours())
	days := hours / 24

	switch {
	case days > 30:
		months := days / 30
		return fmt.Sprintf("%d month%s ago", months, pluralize(months))
	case days > 0:
		return fmt.Sprintf("%d day%s ago", days, pluralize(days))
	case hours > 0:
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	default:
		minutes := int(dur.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	}
}

// pluralize returns "s" if n != 1, otherwise "".
func pluralize(n int) string {
	if n != 1 {
		return "s"
	}
	return ""
}
