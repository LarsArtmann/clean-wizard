package main

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestScanCommandWithMockFixed(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("scan with empty results", func(t *testing.T) {
		// Create a mock scanner with empty results
		mockScanner := CreateEmptyMockScanner()
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, AssertOutputContains(output, "🎉 Your system is already clean!"))
		assert.True(t, AssertExitCode(output, 0))
	})

	// Clean up
	os.Remove(configPath)
}