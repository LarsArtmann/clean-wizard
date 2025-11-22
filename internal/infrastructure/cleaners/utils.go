package cleaner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// checkCommandExists checks if a command exists in PATH
func checkCommandExists(ctx context.Context, command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// runCommand executes a command and returns result
func runCommand(ctx context.Context, name string, args ...string) result.Result[string] {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return result.Err[string](fmt.Errorf("command '%s %s' failed: %w\noutput: %s",
			name, strings.Join(args, " "), err, string(output)))
	}

	return result.Ok(string(output))
}
