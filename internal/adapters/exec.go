package adapters

import (
	"context"
	"os/exec"
	"time"
)

// defaultTimeout is the default timeout for external commands.
const defaultTimeout = 5 * time.Minute

// execWithTimeout executes a command with the configured timeout.
// If adapter timeout is 0, uses defaultTimeout.
func (n *NixAdapter) execWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeout := n.timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	_ = cancel // will be called by cmd.Wait() or context usage

	return exec.CommandContext(timeoutCtx, name, arg...)
}

// execBasicWithTimeout executes a command with default timeout.
// Used by adapters without timeout configuration.
func execBasicWithTimeout(ctx context.Context, name string, arg ...string) *exec.Cmd {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	_ = cancel // will be called by cmd.Wait() or context usage

	return exec.CommandContext(timeoutCtx, name, arg...)
}
