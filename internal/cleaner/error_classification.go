package cleaner

import (
	"errors"
	"os/exec"

	errorfamily "github.com/larsartmann/go-error-family"
)

// init registers error classifications so that errorfamily.Classify can
// correctly categorize errors from the standard library and exec subsystem
// without relying on fragile keyword matching.
//
//nolint:gochecknoinits // registration of error classifications at startup
func init() {
	errorfamily.RegisterStdlibDefaults(errorfamily.DefaultRegistry)

	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		// exec.ErrNotFound is returned by exec.LookPath when a binary is not
		// in PATH. This is Infrastructure (system can't serve), not a user
		// error — the binary simply isn't installed.
		exec.ErrNotFound: errorfamily.Infrastructure,
	})

	// Register cleaner-specific sentinels so errorfamily.Classify returns
	// the correct family without keyword matching.
	errorfamily.RegisterClassifications(map[error]errorfamily.Family{
		ErrNoCacheTypeSpecified:    errorfamily.Rejection, // user must specify a cache type
		ErrLintCacheNotImplemented: errorfamily.Rejection, // feature not supported
		ErrGoProcessesRunning:      errorfamily.Conflict,  // state conflict — user must close processes
	})

	// Register a classifier for *exec.ExitError — a non-zero exit from a
	// subprocess. Most cleaner failures come from commands that exit non-zero
	// (e.g. "nix store gc" fails). These are Transient: retrying may succeed
	// if the failure was caused by a lock contention or temporary I/O issue.
	errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return errorfamily.Transient, true
		}
		return errorfamily.Transient, false
	})
}
