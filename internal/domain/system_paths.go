package domain

// SystemPaths contains protected system path constants to avoid magic strings.
const (
	// PathSystem is the macOS system directory
	PathSystem = "/System"
	// PathApplications is the macOS applications directory
	PathApplications = "/Applications"
	// PathLibrary is the macOS library directory
	PathLibrary = "/Library"
	// PathRoot is the root directory
	PathRoot = "/"
	// PathUser is the user directory
	PathUser = "/usr"
	// PathEtc is the etc directory
	PathEtc = "/etc"
	// PathVar is the var directory
	PathVar = "/var"
)

// DefaultProtectedPaths returns the default protected system paths.
func DefaultProtectedPaths() []string {
	return []string{PathSystem, PathApplications, PathLibrary}
}

// CriticalSystemPaths returns all critical system paths that should never be deleted.
func CriticalSystemPaths() []string {
	return []string{PathRoot, PathSystem, PathUser, PathEtc}
}

// AllProtectedSystemPaths returns all protected system paths including non-critical ones.
func AllProtectedSystemPaths() []string {
	return []string{
		PathRoot, PathSystem, PathUser, PathEtc,
		PathApplications, PathLibrary, PathVar, "/bin", "/sbin",
	}
}

// IsProtectedPath checks if a path is in the protected paths list.
func IsProtectedPath(path string, protected []string) bool {
	for _, p := range protected {
		if p == path {
			return true
		}
	}

	return false
}
