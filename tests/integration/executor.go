package integration

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CommandExecutor provides utilities for running clean-wizard commands in integration tests
type CommandExecutor struct {
	projectRoot string
	timeout     time.Duration
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(projectRoot string) *CommandExecutor {
	return &CommandExecutor{
		projectRoot: projectRoot,
		timeout:     30 * time.Second,
	}
}

// RunCommand executes a clean-wizard command
func (e *CommandExecutor) RunCommand(t *testing.T, args []string) *CommandResult {
	ctx := NewIntegrationTestContext(t)
	defer ctx.Cleanup()

	cmd := exec.CommandContext(ctx.Context(), "go", append([]string{"run", "./cmd/clean-wizard"}, args...)...)
	cmd.Dir = e.projectRoot
	
	output, err := cmd.CombinedOutput()
	
	return &CommandResult{
		Command:    "go run ./cmd/clean-wizard " + strings.Join(args, " "),
		Args:       args,
		Output:     string(output),
		Error:      err,
		ExitCode:   getExitCode(err),
		Duration:   time.Since(time.Now()), // Placeholder - would need proper timing
		Success:    err == nil,
	}
}

// RunCommandWithTimeout executes a command with custom timeout
func (e *CommandExecutor) RunCommandWithTimeout(t *testing.T, args []string, timeout time.Duration) *CommandResult {
	testCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(testCtx, "go", append([]string{"run", "./cmd/clean-wizard"}, args...)...)
	cmd.Dir = e.projectRoot
	
	output, err := cmd.CombinedOutput()
	
	return &CommandResult{
		Command:    "go run ./cmd/clean-wizard " + strings.Join(args, " "),
		Args:       args,
		Output:     string(output),
		Error:      err,
		ExitCode:   getExitCode(err),
		Duration:   time.Since(time.Now()), // Placeholder - would need proper timing
		Success:    err == nil,
	}
}

// RunScanCommand executes a scan command
func (e *CommandExecutor) RunScanCommand(t *testing.T, args []string) *CommandResult {
	scanArgs := append([]string{"scan"}, args...)
	return e.RunCommand(t, scanArgs)
}

// RunCleanCommand executes a clean command
func (e *CommandExecutor) RunCleanCommand(t *testing.T, args []string) *CommandResult {
	cleanArgs := append([]string{"clean"}, args...)
	return e.RunCommand(t, cleanArgs)
}

// RunGenerateCommand executes a generate command
func (e *CommandExecutor) RunGenerateCommand(t *testing.T, args []string) *CommandResult {
	genArgs := append([]string{"generate"}, args...)
	return e.RunCommand(t, genArgs)
}

// CommandResult represents the result of running a command
type CommandResult struct {
	Command  string
	Args     []string
	Output   string
	Error    error
	ExitCode int
	Duration time.Duration
	Success  bool
}

// AssertSuccess asserts that the command succeeded
func (r *CommandResult) AssertSuccess(t *testing.T) {
	require.NoError(t, r.Error, "Command should succeed: %s\nOutput: %s", r.Command, r.Output)
	require.Equal(t, 0, r.ExitCode, "Command should have exit code 0, got %d", r.ExitCode)
}

// AssertFailure asserts that the command failed
func (r *CommandResult) AssertFailure(t *testing.T) {
	require.Error(t, r.Error, "Command should fail: %s", r.Command)
	require.NotEqual(t, 0, r.ExitCode, "Command should have non-zero exit code, got %d", r.ExitCode)
}

// AssertContains asserts that the output contains the expected string
func (r *CommandResult) AssertContains(t *testing.T, expected string) {
	require.Contains(t, r.Output, expected, "Command output should contain '%s'\nActual output: %s", expected, r.Output)
}

// AssertNotContains asserts that the output does not contain the unexpected string
func (r *CommandResult) AssertNotContains(t *testing.T, unexpected string) {
	require.NotContains(t, r.Output, unexpected, "Command output should not contain '%s'\nActual output: %s", unexpected, r.Output)
}

// AssertExitCode asserts that the command has the expected exit code
func (r *CommandResult) AssertExitCode(t *testing.T, expected int) {
	require.Equal(t, expected, r.ExitCode, "Command should have exit code %d, got %d", expected, r.ExitCode)
}

// getExitCode extracts the exit code from an error
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	
	return 1 // Default exit code for other errors
}

// CreateTestExecutor creates a command executor for testing
func CreateTestExecutor(ctx *IntegrationTestContext) *CommandExecutor {
	return NewCommandExecutor(ctx.ProjectRoot())
}