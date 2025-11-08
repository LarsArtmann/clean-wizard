package main

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
)

func main() {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)
	defer os.Remove(configPath)

	// Create a mock scanner with empty results
	mockScanner := CreateEmptyMockScanner()

	// Run command with mock scanner
	output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

	fmt.Printf("=== OUTPUT ===\n%s\n=== END ===\n", output.Stdout)
	fmt.Printf("Contains clean message: %t\n", AssertOutputContains(output, "🎉 Your system is already clean!"))
}
