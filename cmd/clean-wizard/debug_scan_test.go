package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
)

func TestDebugScannerResults(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("debug empty scanner", func(t *testing.T) {
		// Create a mock scanner with empty results
		mockScanner := CreateEmptyMockScanner()

		fmt.Printf("=== MOCK SCANNER DEBUG ===\n")
		if mockScanner != nil {
			fmt.Printf("Mock scanner created successfully\n")
		}

		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		// Debug output
		fmt.Printf("=== DEBUG OUTPUT ===\n%s\n=== END DEBUG ===\n", output.Stdout)
		fmt.Printf("=== DEBUG STDERR ===\n%s\n=== END DEBUG ===\n", output.Stderr)
		fmt.Printf("=== DEBUG EXIT CODE ===\n%d\n=== END DEBUG ===\n", output.ExitCode)

		// Check if scanner was called
		if mockScanner.WasCalled() {
			fmt.Printf("Mock scanner was called\n")
		} else {
			fmt.Printf("Mock scanner was NOT called\n")
		}
	})

	// Clean up
	os.Remove(configPath)
}
