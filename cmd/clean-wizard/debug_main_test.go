package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
)

func TestDebugScannerReturn(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("debug scanner return value", func(t *testing.T) {
		// Create a mock scanner with empty results
		mockScanner := CreateEmptyMockScanner()

		// Test scanner directly
		fmt.Printf("=== DIRECT SCANNER TEST ===\n")
		ctx := context.Background()
		results, err := mockScanner.Scan(ctx)

		if err != nil {
			fmt.Printf("Scanner error: %v\n", err)
		} else {
			fmt.Printf("Scanner returned results:\n")
			fmt.Printf("TotalSizeGB: %f\n", results.TotalSizeGB)
			fmt.Printf("Results count: %d\n", len(results.Results))
			for i, result := range results.Results {
				fmt.Printf("Result[%d]: %s - %f GB\n", i, result.Name, result.SizeGB)
			}
		}
		fmt.Printf("=== END DIRECT TEST ===\n")

		// Now test with command
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		fmt.Printf("=== COMMAND OUTPUT ===\n%s\n=== END COMMAND ===\n", output.Stdout)
	})

	// Clean up
	os.Remove(configPath)
}
