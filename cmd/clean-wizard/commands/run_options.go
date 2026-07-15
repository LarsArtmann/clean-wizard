package commands

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
)

// buildRunOptions assembles execution.RunOption slice from CLI flags.
// Shared between clean and scan commands to avoid duplicating the
// verbose/concurrency/retry flag handling.
func buildRunOptions(verbose bool, concurrency int, retries int, retryProfile string) ([]execution.RunOption, error) {
	var opts []execution.RunOption

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
