package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCLICommands tests end-to-end CLI functionality
func TestCLICommands(t *testing.T) {
	t.Run("scan command shows help", func(t *testing.T) {
		// Remove any existing config file
		configPath := "config.yaml"
		os.Remove(configPath)
		defer os.Remove(configPath)

		// Create test runner
		tr := main.NewTestRunner()
		defer main.CleanupTest(tr)

		// Run scan help
		output := main.RunCommand([]string{"scan", "--help"})

		// Should show help information
		assert.True(t, main.AssertOutputContains(output, "scan system for cleanable items"))
		assert.True(t, main.AssertExitCode(output, 0))
	})

	t.Run("clean command shows help", func(t *testing.T) {
		// Remove any existing config file
		configPath := "config.yaml"
		os.Remove(configPath)
		defer os.Remove(configPath)

		// Create test runner
		tr := main.NewTestRunner()
		defer main.CleanupTest(tr)

		// Run clean help
		output := main.RunCommand([]string{"clean", "--help"})

		// Should show help information
		assert.True(t, main.AssertOutputContains(output, "perform system cleanup"))
		assert.True(t, main.AssertExitCode(output, 0))
	})

	t.Run("root command shows help", func(t *testing.T) {
		// Remove any existing config file
		configPath := "config.yaml"
		os.Remove(configPath)
		defer os.Remove(configPath)

		// Create test runner
		tr := main.NewTestRunner()
		defer main.CleanupTest(tr)

		// Run root help
		output := main.RunCommand([]string{"--help"})

		// Should show help information
		assert.True(t, main.AssertOutputContains(output, "Interactive system cleaning wizard"))
		assert.True(t, main.AssertOutputContains(output, "Available Commands:"))
		assert.True(t, main.AssertExitCode(output, 0))
	})
}

// TODO: Add context propagation test once fixed
// TestScannerContextPropagation tests that context properly propagates from ExecuteContext to commands
/*
func TestScannerContextPropagation(t *testing.T) {
	t.Run("mock scanner works end-to-end", func(t *testing.T) {
		// Create mock scanner with empty results
		mockScanner := scan.NewEmptyMockScanner()

		// Run command with context-based dependency injection
		output := main.RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		// Should show empty results message
		assert.True(t, main.AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, main.AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, main.AssertOutputContains(output, "🎉 Your system is already clean!"))
		assert.True(t, main.AssertExitCode(output, 0))
	})
}
*/
