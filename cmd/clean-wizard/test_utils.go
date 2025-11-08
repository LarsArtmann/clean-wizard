package main

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/scan"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

// TestOutput captures stdout and stderr for testing
type TestOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// TestRunner runs commands and captures output for testing
type TestRunner struct {
	originalStdout *os.File
	originalStderr *os.File
	stdoutPipe     *os.File
	stderrPipe     *os.File
	stdoutBuffer   *bytes.Buffer
	stderrBuffer   *bytes.Buffer
	output         TestOutput
	mu             sync.Mutex
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		stdoutBuffer: &bytes.Buffer{},
		stderrBuffer: &bytes.Buffer{},
	}
}

// Setup redirects stdout and stderr to buffers
func (tr *TestRunner) Setup() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Save original file descriptors
	tr.originalStdout = os.Stdout
	tr.originalStderr = os.Stderr

	// Create pipes
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return err
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		return err
	}

	// Redirect stdout and stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	tr.stdoutPipe = stdoutWriter
	tr.stderrPipe = stderrWriter

	// Start goroutines to read from pipes
	go func() {
		io.Copy(tr.stdoutBuffer, stdoutReader)
	}()

	go func() {
		io.Copy(tr.stderrBuffer, stderrReader)
	}()

	return nil
}

// Restore restores original stdout and stderr
func (tr *TestRunner) Restore() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.originalStdout != nil {
		os.Stdout = tr.originalStdout
	}
	if tr.originalStderr != nil {
		os.Stderr = tr.originalStderr
	}

	if tr.stdoutPipe != nil {
		tr.stdoutPipe.Close()
	}
	if tr.stderrPipe != nil {
		tr.stderrPipe.Close()
	}
}

// GetOutput returns the captured output
func (tr *TestRunner) GetOutput() TestOutput {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	return TestOutput{
		Stdout:   tr.stdoutBuffer.String(),
		Stderr:   tr.stderrBuffer.String(),
		ExitCode: tr.output.ExitCode,
	}
}

// SetExitCode sets the exit code for the test output
func (tr *TestRunner) SetExitCode(code int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.output.ExitCode = code
}

// Reset resets the test runner
func (tr *TestRunner) Reset() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.stdoutBuffer.Reset()
	tr.stderrBuffer.Reset()
	tr.output = TestOutput{}
}

// SetupTestWithMockScanner sets up a test with a mock scanner
func SetupTestWithMockScanner(mockScanner *scan.MockScanner) *TestRunner {
	// Set up mock scanner
	setMockScannerForTesting(mockScanner)

	// Create test runner
	tr := NewTestRunner()
	if err := tr.Setup(); err != nil {
		panic(err)
	}

	return tr
}

// CleanupTest cleans up after a test
func CleanupTest(tr *TestRunner) {
	// Restore original scanner
	resetToRealScannerForTesting()

	// Restore stdout/stderr
	tr.Restore()

	// Reset the test runner
	tr.Reset()
}

// RunCommandWithMock runs a command with a mock scanner and returns output
func RunCommandWithMock(args []string, mockScanner *scan.MockScanner) TestOutput {
	tr := SetupTestWithMockScanner(mockScanner)
	defer CleanupTest(tr)

	// Create root command and set arguments
	rootCmd := NewRootCmd()
	rootCmd.SetArgs(args)

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		tr.SetExitCode(1)
	} else {
		tr.SetExitCode(0)
	}

	return tr.GetOutput()
}

// CreateDefaultMockScanner creates a mock scanner with default test data
func CreateDefaultMockScanner() *scan.MockScanner {
	return scan.NewMockScanner()
}

// CreateEmptyMockScanner creates a mock scanner with empty results
func CreateEmptyMockScanner() *scan.MockScanner {
	return scan.NewEmptyMockScanner()
}

// CreateErrorMockScanner creates a mock scanner that returns an error
func CreateErrorMockScanner(err error) *scan.MockScanner {
	return scan.NewErrorMockScanner(err)
}

// CreateSlowMockScanner creates a mock scanner that simulates slow scanning
func CreateSlowMockScanner(delay string) *scan.MockScanner {
	// For testing, we'll use a very short delay
	mockScanner := scan.NewMockScanner()
	return mockScanner.WithScanDelay(1) // 1 nanosecond
}

// CreateCancellableMockScanner creates a mock scanner that checks for cancellation
func CreateCancellableMockScanner() *scan.MockScanner {
	mockScanner := scan.NewMockScanner()
	return mockScanner.WithCancelCheck(true)
}

// AssertOutputContains checks if the output contains expected text
func AssertOutputContains(output TestOutput, expected string) bool {
	return contains(output.Stdout, expected) || contains(output.Stderr, expected)
}

// AssertOutputNotContains checks if the output does not contain unexpected text
func AssertOutputNotContains(output TestOutput, unexpected string) bool {
	return !contains(output.Stdout, unexpected) && !contains(output.Stderr, unexpected)
}

// AssertExitCode checks if the exit code matches the expected code
func AssertExitCode(output TestOutput, expected int) bool {
	return output.ExitCode == expected
}

// contains is a helper function to check if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (
			(s[:len(substr)] == substr) ||
			(s[len(s)-len(substr):] == substr) ||
			(len(s) > len(substr)+1 && findSubstring(s, substr))))
}

// findSubstring is a simple substring search helper
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}