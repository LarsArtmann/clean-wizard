package cleaner

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

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

	// Register a classifier for *os.PathError — file-operation failures.
	// The stdlib defaults already classify os.ErrNotExist and os.ErrPermission
	// as Rejection (via errors.Is sentinel matching, which runs before
	// classifiers). This classifier catches permanent errno values that would
	// otherwise fall through to the Transient default but represent conditions
	// that won't resolve on retry: a full disk (ENOSPC), a read-only filesystem
	// (EROFS), or a loop in the filesystem (ELOOP).
	errorfamily.RegisterClassifier(func(err error) (errorfamily.Family, bool) {
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			return errorfamily.Transient, false
		}

		switch {
		case errors.Is(pathErr.Err, syscall.ENOSPC),
			errors.Is(pathErr.Err, syscall.EROFS),
			errors.Is(pathErr.Err, syscall.ELOOP):
			return errorfamily.Rejection, true
		default:
			return errorfamily.Transient, false
		}
	})

	// Register user-facing message templates for key error codes. These power
	// errorfamily.HandleError's What/Why/Fix/WayOut output at the CLI boundary.
	errorfamily.RegisterTemplate("cleaner.not_available", errorfamily.MessageTemplate{
		What:   "A cleaner could not run because its prerequisite is not installed.",
		Why:    "The system does not have the required tool or runtime for this cleaner.",
		Fix:    "Install the missing tool (e.g. 'brew install nix', 'apt install docker.io').",
		WayOut: "Run 'clean-wizard scan' to see which cleaners are available on this system.",
	})

	errorfamily.RegisterTemplate("cleaner.not_found", errorfamily.MessageTemplate{
		What:   "The requested cleaner does not exist in the registry.",
		Why:    "The cleaner name was not registered during startup.",
		Fix:    "Check the cleaner name spelling or run 'clean-wizard scan' to list available cleaners.",
		WayOut: "Use 'clean-wizard clean --help' to see all available options.",
	})

	errorfamily.RegisterTemplate("validation.rejected", errorfamily.MessageTemplate{
		What:   "A configuration or input validation error occurred.",
		Why:    "The provided settings or configuration file contain invalid values.",
		Fix:    "Review the error message for the specific field and correct the value.",
		WayOut: "Run 'clean-wizard init' to generate a fresh configuration file.",
	})
}
