package bdd

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/stretchr/testify/assert"
)

// TestCleaningScenarios tests end-to-end cleaning behavior
func TestCleaningScenarios(t *testing.T) {
	// Feature: System Cleaning
	// As a user
	// I want to scan and clean my system safely
	// So that I can recover disk space and maintain system performance

	t.Run("Scenario: Safe system cleaning", func(t *testing.T) {
		// Given I have a cleanable system
		mockScanner := scan.NewMockScanner() // With some cleanable items

		// When I run scan command
		scanOutput := main.RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		// Then I see cleanable items
		assert.True(t, main.AssertOutputContains(scanOutput, "🔍 Scanning system..."))
		assert.True(t, main.AssertOutputContains(scanOutput, "✅ Scan completed!"))
		assert.True(t, main.AssertOutputContains(scanOutput, "📊 Total cleanable:"))
		assert.True(t, main.AssertExitCode(scanOutput, 0))

		// And I see suggestions for cleanup
		assert.True(t, main.AssertOutputContains(scanOutput, "💡 Run 'clean-wizard clean'"))
	})

	t.Run("Scenario: System already clean", func(t *testing.T) {
		// Given I have a system with nothing to clean
		mockScanner := scan.NewEmptyMockScanner()

		// When I run scan command
		scanOutput := main.RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		// Then I see that system is already clean
		assert.True(t, main.AssertOutputContains(scanOutput, "🔍 Scanning system..."))
		assert.True(t, main.AssertOutputContains(scanOutput, "✅ Scan completed!"))
		assert.True(t, main.AssertOutputContains(scanOutput, "🎉 Your system is already clean!"))
		assert.True(t, main.AssertExitCode(scanOutput, 0))
	})

	t.Run("Scenario: Dry-run cleaning shows what would be done", func(t *testing.T) {
		// Given I have a cleanable system
		// When I run clean command with dry-run
		output := main.RunCommand([]string{"clean", "--dry-run"})

		// Then I see what would be cleaned without doing it
		assert.True(t, main.AssertOutputContains(output, "🧹 Cleanup Summary"))
		assert.True(t, main.AssertOutputContains(output, "Dry run: true"))
		assert.True(t, main.AssertOutputContains(output, "🧹 Starting cleanup..."))
		assert.True(t, main.AssertOutputContains(output, "[DRY RUN] Would execute"))
		assert.True(t, main.AssertOutputContains(output, "✅ Cleanup completed successfully!"))
		assert.True(t, main.AssertExitCode(output, 0))
	})

	t.Run("Scenario: Profile-based cleaning", func(t *testing.T) {
		// Given I have different cleaning profiles
		// When I run clean command with daily profile
		output := main.RunCommand([]string{"clean", "--dry-run", "--profile", "daily"})

		// Then I see daily profile operations
		assert.True(t, main.AssertOutputContains(output, "Profile: daily"))
		assert.True(t, main.AssertOutputContains(output, "Description: Quick daily cleanup"))
		assert.True(t, main.AssertOutputContains(output, "Dry run: true"))
		assert.True(t, main.AssertExitCode(output, 0))
	})

	t.Run("Scenario: Configuration management", func(t *testing.T) {
		// Given I want to see current configuration
		// When I run config show command
		output := main.RunCommand([]string{"config", "show"})

		// Then I see configuration details
		assert.True(t, main.AssertOutputContains(output, "Configuration file:"))
		assert.True(t, main.AssertOutputContains(output, "Version:"))
		assert.True(t, main.AssertOutputContains(output, "Safe mode:"))
		assert.True(t, main.AssertOutputContains(output, "Dry run:"))
		assert.True(t, main.AssertOutputContains(output, "Profiles:"))
		assert.True(t, main.AssertExitCode(output, 0))
	})
}

// TestErrorScenarios tests error handling behavior
func TestErrorScenarios(t *testing.T) {
	// Feature: Error Handling
	// As a user
	// I want clear error messages when things go wrong
	// So that I can understand and resolve issues

	t.Run("Scenario: Scanner error shows clear message", func(t *testing.T) {
		// Given scanner encounters an error
		mockScanner := scan.NewErrorMockScanner(assert.AnError)

		// When I run scan command
		output := main.RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)

		// Then I see clear error information
		assert.True(t, main.AssertExitCode(output, 1))
		// TODO: Add proper error message checking once error handling is implemented
	})

	t.Run("Scenario: Invalid profile shows helpful error", func(t *testing.T) {
		// Given I try to use non-existent profile
		// When I run clean command with invalid profile
		output := main.RunCommand([]string{"clean", "--dry-run", "--profile", "nonexistent"})

		// Then I see helpful error message
		assert.False(t, main.AssertExitCode(output, 0))
		// TODO: Add proper error message checking once error handling is implemented
	})
}

// TestSafetyScenarios tests safety features
func TestSafetyScenarios(t *testing.T) {
	// Feature: Safety Features
	// As a user
	// I want protection from accidental data loss
	// So that I can use the tool confidently

	t.Run("Scenario: Default dry-run prevents accidental changes", func(t *testing.T) {
		// Given I run clean command without explicit dry-run
		// When I check default settings
		// Then dry-run should be enabled by default for safety
		// TODO: Test default safety behavior once implemented
	})

	t.Run("Scenario: Protected paths prevent accidental deletion", func(t *testing.T) {
		// Given I have protected paths configured
		// When I run clean command
		// Then protected paths should not be cleaned
		// TODO: Test path protection once implemented
	})
}
