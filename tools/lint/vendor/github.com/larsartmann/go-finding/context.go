package finding

import "context"

// workingDirKey is the canonical context key for the working directory of a
// detection/repair run. It lives here (in go-finding, the ecosystem hub) so
// that external tools implementing Detector can read the working directory
// via WorkingDirFromContext WITHOUT importing a consumer-specific package
// (e.g. BuildFlow's runner package). Consumers set it via WithWorkingDir.
//
// BuildFlow's runner.WithWorkingDir delegates here so the key is shared: a
// value set by BuildFlow's module-fan-out is visible to any finding.Detector
// that calls WorkingDirFromContext.
type workingDirKey struct{}

// WithWorkingDir returns a context carrying dir as the working directory for a
// detection/repair/generation run. Downstream capabilities read it via
// WorkingDirFromContext. This is the canonical setter — consumer packages
// (BuildFlow runner, etc.) should delegate to it rather than define their own
// key so the value propagates across package boundaries.
func WithWorkingDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, workingDirKey{}, dir)
}

// WorkingDirFromContext extracts the working directory set by WithWorkingDir.
// Returns "" when no working directory is set (callers should then fall back
// to the process working directory or project root).
//
// This is the canonical getter for the ecosystem: any Detector implementation
// can retrieve its target directory without coupling to a specific orchestrator.
func WorkingDirFromContext(ctx context.Context) string {
	if dir, ok := ctx.Value(workingDirKey{}).(string); ok {
		return dir
	}

	return ""
}
