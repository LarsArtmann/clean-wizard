package commands

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/progress"
)

// buildRunOptions assembles execution.RunOption slice from CLI flags.
// Shared between clean and scan commands to avoid duplicating the
// verbose/concurrency/retry flag handling. progressEmitter may be nil
// (no live progress).
func buildRunOptions(
	verbose bool,
	concurrency int,
	retries int,
	retryProfile string,
	progressEmitter execution.ProgressEmitter,
) ([]execution.RunOption, error) {
	var opts []execution.RunOption

	if progressEmitter != nil {
		opts = append(opts, execution.WithProgress(progressEmitter))
	}

	if verbose {
		opts = append(opts, execution.WithVerbose(true))
	}

	if concurrency > 0 {
		opts = append(opts, execution.WithMaxConcurrency(concurrency))
	}

	if retryProfile != "" {
		rp := execution.RetryProfile(retryProfile)
		if !rp.IsValid() {
			return nil, fmt.Errorf(
				"invalid --retry-profile %q: must be default, aggressive, conservative, or none",
				retryProfile,
			)
		}

		opts = append(opts, execution.WithRetry(rp.Apply()))
	} else if retries > 0 {
		opts = append(opts, execution.WithRetry(execution.RetryConfigFromAttempts(retries)))
	}

	return opts, nil
}

// progressRequested applies the ADR-0002 gating contract: live progress only
// when the user asked for it AND the output mode is human-facing (no
// --json/--sarif) AND the writer is an interactive terminal.
func progressRequested(flag bool, machineOutput bool) bool {
	return flag && !machineOutput && progress.Enabled(os.Stdout)
}
