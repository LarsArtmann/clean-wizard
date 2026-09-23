package domain

import (
	"runtime"
	"slices"
)

const (
	PathRoot         = "/"
	PathUser         = "/usr"
	PathEtc          = "/etc"
	PathVar          = "/var"
	PathSystem       = "/System"
	PathApplications = "/Applications"
	PathLibrary      = "/Library"
	PathNixStore     = "/nix/store"
	PathNixVar       = "/nix/var"
)

// DefaultProtectedPaths returns platform-appropriate protected system paths.
func DefaultProtectedPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{PathSystem, PathApplications, PathLibrary}
	case "linux":
		return []string{PathNixStore, PathNixVar}
	default:
		return []string{}
	}
}

// CriticalSystemPaths returns all critical system paths that should never be deleted.
func CriticalSystemPaths() []string {
	return []string{PathRoot, PathUser, PathEtc}
}

// AllProtectedSystemPaths returns all protected system paths including non-critical ones.
func AllProtectedSystemPaths() []string {
	base := []string{
		PathRoot, PathUser, PathEtc,
		PathVar, "/bin", "/sbin",
	}

	switch runtime.GOOS {
	case "darwin":
		return append(base, PathSystem, PathApplications, PathLibrary)
	case "linux":
		return append(base, PathNixStore, PathNixVar)
	default:
		return base
	}
}

// IsProtectedPath checks if a path is in the protected paths list.
func IsProtectedPath(path string, protected []string) bool {
	return slices.Contains(protected, path)
}
