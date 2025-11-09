package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/spf13/cobra"
)

// NewScanCommand creates scan command with proper domain types
func NewScanCommand(verbose bool) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan system for cleanable items",
		Long:  `Analyze your system to identify old files, package caches, and temporary data that can be safely cleaned.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔍 Analyzing system state...")
			ctx := context.Background()
			nixCleaner := cleaner.NewNixCleaner(verbose, true) // dry-run for safety

			// Get Nix generations
			genResult := nixCleaner.ListGenerations(ctx)
			if genResult.IsErr() {
				return handleScanError(genResult.Error())
			}

			generations := genResult.Value()
			displayScanResults(generations, verbose, nixCleaner, ctx)
			return nil
		},
	}
}

// handleScanError provides user-friendly error messages for scanning
func handleScanError(err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "command not found"):
		return fmt.Errorf(`❌ Nix package manager not found

💡 To fix this:
   • Install Nix: curl --proto '=https' --tlsv1.2 -sSf https://nixos.org/install/standalone | sh
   • Or try other cleanup targets if you don't use Nix
   
📚 Learn more: https://nixos.org/`)

	case strings.Contains(errMsg, "permission"):
		return fmt.Errorf(`❌ Permission denied while scanning

💡 To fix this:
   • Try running with sudo: sudo clean-wizard scan
   • Check if you have read access to Nix profiles
   • Verify Nix is properly installed`)

	case strings.Contains(errMsg, "no such file"):
		return fmt.Errorf(`❌ Cannot find Nix profiles directory

💡 This could mean:
   • Nix is not installed correctly
   • Nix is installed but not in standard location
   • You're on a system where Nix works differently
   
🔧 Try: nix-env --version to check installation`)

	default:
		return fmt.Errorf(`❌ System scan failed: %s

💡 Suggestions:
   • Ensure Nix is properly installed
   • Try running with --verbose for more details
   • Check system permissions
   • If this persists, please report an issue`, err.Error())
	}
}

// displayScanResults shows user-friendly scan results
func displayScanResults(generations []domain.NixGeneration, verbose bool, nixCleaner *cleaner.NixCleaner, ctx context.Context) {
	fmt.Println("\n📊 Scan Results:")

	if len(generations) == 0 {
		fmt.Println("   🎉 No Nix generations found - your system is clean!")
		return
	}

	// Count total generations and identify old ones
	cleanableCount := 0
	for _, gen := range generations {
		if !gen.Current {
			cleanableCount++
		}
	}

	fmt.Printf("   • Total generations: %d\n", len(generations))
	fmt.Printf("   • Current generation: %s\n", getCurrentGeneration(generations))
	fmt.Printf("   • Cleanable generations: %d\n", cleanableCount)

	// Get store size
	storeResult := nixCleaner.GetStoreSize(ctx)
	if storeResult.IsOk() {
		fmt.Printf("   • Store size: %s\n", format.Bytes(storeResult.Value()))
	}

	// Show detailed generation info
	if verbose {
		fmt.Println("\n📋 Generation Details:")
		for _, gen := range generations {
			status := "🟢 CURRENT"
			if !gen.Current {
				status = "🔴 OLD"
			}
			fmt.Printf("   %s Generation %s (%s)\n", status, fmt.Sprintf("%d", gen.ID), format.DateTime(gen.Date))
		}
	}

	// Provide user guidance
	if cleanableCount > 0 {
		fmt.Printf("\n💡 You can clean up %d old generations to free space\n", cleanableCount)
		fmt.Println("   🏃 Run 'clean-wizard clean' to start cleanup")
		fmt.Println("   🔍 Try 'clean-wizard clean --dry-run' first to see what would be cleaned")
	} else {
		fmt.Println("\n🎉 Your Nix system is optimized - no cleanup needed!")
	}
}

// getCurrentGeneration finds current generation
func getCurrentGeneration(generations []domain.NixGeneration) string {
	for _, gen := range generations {
		if gen.Current {
			return fmt.Sprintf("%d", gen.ID)
		}
	}
	return "Unknown"
}
