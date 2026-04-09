package adapters

import (
	"context"
	"os/exec"
	"time"
)

// DefaultTimeout is the default timeout for external commands.
const DefaultTimeout = 5 * time.Minute

// ExecWithTimeout creates a command with the specified timeout.
// If the context already has a deadline, respects the earlier deadline.
func ExecWithTimeout(
	ctx context.Context,
	timeout time.Duration,
	name string,
	args ...string,
) *exec.Cmd {
	// If context already has a deadline, use the existing context
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return exec.CommandContext(ctx, name, args...)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	cmd := exec.CommandContext(timeoutCtx, name, args...)
	cmd.Cancel = cancel

	return cmd
}

// ExecWithDefaultTimeout creates a command with the default 5-minute timeout.
// If the context already has a deadline, respects the existing deadline.
func ExecWithDefaultTimeout(ctx context.Context, name string, args ...string) *exec.Cmd {
	return ExecWithTimeout(ctx, DefaultTimeout, name, args...)
}

// execWithTimeout executes a command with the configured timeout.
// If adapter timeout is 0, uses defaultTimeout.
func (n *NixAdapter) execWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeout := n.timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	cmd := exec.CommandContext(timeoutCtx, name, arg...)
	cmd.Cancel = cancel

	return cmd
}
