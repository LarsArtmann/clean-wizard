package errorfamily

import (
	"context"
	"database/sql"
	"os"
)

// RegisterStdlibDefaults registers classification mappings for common Go
// standard-library errors onto the given Registry. Pass [DefaultRegistry] to
// affect the package-level [Classify], or a custom Registry for scoped handling.
//
// Ambiguous cases and the rationale for their assigned Family:
//
//   - context.DeadlineExceeded  → Transient (retryable). Usually means an
//     upstream was slow; a fresh attempt may succeed. If the deadline was set
//     too short by the caller, the fix is at the call site, but the error
//     itself is still transient in nature.
//   - context.Canceled          → Rejection (not retryable). The caller
//     abandoned the operation; retrying without a new request is wrong.
//   - sql.ErrNoRows / os.ErrNotExist → Rejection. A missing resource is the
//     caller's concern (wrong key / path), not a system fault.
//   - sql.ErrConnDone           → Transient. The connection pool closed a
//     connection; a retry on a fresh connection typically succeeds.
//   - os.ErrPermission          → Rejection. Permission is a caller/environment
//     state, not a transient failure.
//
// This is advisory: applications whose semantics differ should register their
// own mappings after calling this.
func RegisterStdlibDefaults(reg *Registry) {
	reg.RegisterClassifications(map[error]Family{
		// context package — see rationale above.
		context.DeadlineExceeded: Transient,
		context.Canceled:         Rejection,

		// database/sql — connection issues are transient; missing rows are not.
		sql.ErrNoRows:   Rejection,
		sql.ErrConnDone: Transient,

		// os — missing/forbidden resources are the caller's concern.
		os.ErrNotExist:   Rejection,
		os.ErrPermission: Rejection,
	})
}
