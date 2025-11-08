package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCommand(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	// Test config show
	t.Run("config show", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"config", "show"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Configuration file:")
		assert.Contains(t, output, "Version:")
		assert.Contains(t, output, "Safe mode:")
		assert.Contains(t, output, "Dry run:")
		assert.Contains(t, output, "Profiles:")
	})

	// Test config path
	t.Run("config path", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"config", "path"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Equal(t, configPath, strings.TrimSpace(output))
	})

	// Clean up
	os.Remove(configPath)
}

func TestScanCommandWithMock(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("scan with dry run", func(t *testing.T) {
		// Create a mock scanner
		mockScanner := CreateDefaultMockScanner()
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, AssertOutputContains(output, "📊 Total cleanable:"))
		assert.True(t, AssertExitCode(output, 0))
	})

	t.Run("scan with verbose", func(t *testing.T) {
		// Create a mock scanner
		mockScanner := CreateDefaultMockScanner()
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run", "--verbose"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, AssertExitCode(output, 0))
	})

	t.Run("scan with profile", func(t *testing.T) {
		// Create a mock scanner
		mockScanner := CreateDefaultMockScanner()
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run", "--profile", "daily"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, AssertExitCode(output, 0))
	})

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

	t.Run("scan with error", func(t *testing.T) {
		// Create a mock scanner that returns an error
		mockScanner := CreateErrorMockScanner(assert.AnError)
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertExitCode(output, 1))
	})

	t.Run("scan with cancellable mock", func(t *testing.T) {
		// Create a mock scanner that checks for cancellation
		mockScanner := CreateCancellableMockScanner()
		
		// Run command with mock scanner
		output := RunCommandWithMock([]string{"scan", "--dry-run"}, mockScanner)
		
		// Verify results
		assert.True(t, AssertOutputContains(output, "🔍 Scanning system..."))
		assert.True(t, AssertOutputContains(output, "✅ Scan completed!"))
		assert.True(t, AssertExitCode(output, 0))
	})

	// Clean up
	os.Remove(configPath)
}

func TestCleanCommand(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("clean with dry run", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "🧹 Cleanup Summary")
		assert.Contains(t, output, "Profile:")
		assert.Contains(t, output, "Operations:")
		assert.Contains(t, output, "Dry run: true")
		assert.Contains(t, output, "🧹 Starting cleanup...")
		assert.Contains(t, output, "✅ Cleanup completed successfully!")
	})

	t.Run("clean with daily profile", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--profile", "daily"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Profile: daily")
		assert.Contains(t, output, "Description: Quick daily cleanup")
		assert.Contains(t, output, "Operations: 2")
	})

	t.Run("clean with comprehensive profile", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--profile", "comprehensive"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Profile: comprehensive")
		assert.Contains(t, output, "Description: Complete system cleanup")
		assert.Contains(t, output, "Operations: 3")
	})

	t.Run("clean with aggressive profile", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--profile", "aggressive"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Profile: aggressive")
		assert.Contains(t, output, "Description: Nuclear option - everything")
		assert.Contains(t, output, "Operations: 3")
		assert.Contains(t, output, "⚠️") // High risk operations
		assert.Contains(t, output, "⚡")  // Medium risk operations
	})

	t.Run("clean with verbose", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--verbose"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "[DRY RUN] Would execute")
	})

	t.Run("clean with force", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--force"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "🧹 Cleanup Summary")
	})

	// Clean up
	os.Remove(configPath)
}

func TestRootCommand(t *testing.T) {
	t.Run("root help", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"--help"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "A comprehensive CLI/TUI tool for system cleanup")
		assert.Contains(t, output, "Available Commands:")
		assert.Contains(t, output, "clean")
		assert.Contains(t, output, "scan")
		assert.Contains(t, output, "config")
		assert.Contains(t, output, "init")
		assert.Contains(t, output, "completion")
		assert.Contains(t, output, "help")
	})

	t.Run("root version", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"--version"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "clean-wizard version")
	})

	t.Run("root with invalid profile", func(t *testing.T) {
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--dry-run", "--profile", "nonexistent"})

		err := rootCmd.Execute()
		assert.Error(t, err)
	})

	t.Run("root with invalid flags", func(t *testing.T) {
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"--invalid-flag"})

		err := rootCmd.Execute()
		assert.Error(t, err)
	})
}

func TestInitCommand(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("init command", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"init"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Interactive setup wizard")
		assert.Contains(t, output, "This wizard will help you create a configuration file")
	})

	// Clean up
	os.Remove(configPath)
}

func TestCompletionCommand(t *testing.T) {
	t.Run("completion command", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"completion", "--help"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Generate the autocompletion script for the specified shell")
	})
}

func TestHelpCommand(t *testing.T) {
	t.Run("help command", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"help"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "A comprehensive CLI/TUI tool for system cleanup")
		assert.Contains(t, output, "Available Commands:")
	})
}

func TestCommandHelp(t *testing.T) {
	commands := []string{"config", "scan", "clean", "init", "completion"}

	for _, cmd := range commands {
		t.Run(cmd+" help", func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			rootCmd := NewRootCmd()
			rootCmd.SetArgs([]string{cmd, "--help"})

			err := rootCmd.Execute()
			require.NoError(t, err)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			assert.NotEmpty(t, output)
		})
	}
}

func TestConfigFileHandling(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	// Test with no config file (should use defaults)
	t.Run("no config file", func(t *testing.T) {
		assert.NoFileExists(t, configPath)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"config", "show"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Version: dev")
		assert.Contains(t, output, "Safe mode: true")
		assert.Contains(t, output, "Dry run: true")
	})

	// Test with existing config file
	t.Run("existing config file", func(t *testing.T) {
		// Create a config file by loading defaults and modifying
		cfg, err := config.Load()
		require.NoError(t, err)

		cfg.Version = "test-version"
		cfg.SafeMode = false
		cfg.DryRun = false
		cfg.Verbose = true

		err = config.Save(cfg)
		require.NoError(t, err)
		assert.FileExists(t, configPath)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"config", "show"})

		err = rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Version: test-version")
		assert.Contains(t, output, "Safe mode: false")
		assert.Contains(t, output, "Dry run: false")
		assert.Contains(t, output, "Verbose: true")
	})

	// Clean up
	os.Remove(configPath)
}

func TestGlobalFlags(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("global flags work", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"--verbose", "--dry-run", "--profile", "daily", "clean"})

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Profile: daily")
		assert.Contains(t, output, "Dry run: true")
	})

	// Clean up
	os.Remove(configPath)
}

func TestErrorHandling(t *testing.T) {
	// Remove any existing config file
	configPath := config.GetConfigPath()
	os.Remove(configPath)

	t.Run("invalid profile error", func(t *testing.T) {
		rootCmd := NewRootCmd()
		rootCmd.SetArgs([]string{"clean", "--profile", "nonexistent"})

		err := rootCmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Profile not found")
	})

	// Clean up
	os.Remove(configPath)
}