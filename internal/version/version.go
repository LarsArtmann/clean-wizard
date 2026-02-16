package version

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	version = ""      // Set via -ldflags at build time
	commit  = ""      // Set via -ldflags at build time
	date    = ""      // Set via -ldflags at build time
	builtBy = "goreleaser" // Set via -ldflags at build time
)

// Info contains version information.
type Info struct {
	Version   string
	Commit    string
	Date      string
	BuiltBy   string
	IsDirty   bool
	GitTag    string
}

// Get returns the current version info.
func Get() Info {
	info := Info{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}

	// If version not set via ldflags, generate it
	if info.Version == "" {
		info.Version = generateVersion()
	}

	// If commit not set via ldflags, try to get it from git
	if info.Commit == "" {
		info.Commit = getGitCommit()
	}

	// Check if repo is dirty
	info.IsDirty = isGitDirty()

	// Get git tag if available
	info.GitTag = getGitTag()

	// If we have a git tag, use it as version (with -dirty suffix if needed)
	if info.GitTag != "" {
		info.Version = info.GitTag
		if info.IsDirty {
			info.Version += "-dirty"
		}
	} else if info.IsDirty {
		// Add -dirty suffix to date-based version
		info.Version += "-dirty"
	}

	// Set date if not provided
	if info.Date == "" {
		info.Date = time.Now().Format("2006-01-02")
	}

	return info
}

// generateVersion creates a date-based version string.
func generateVersion() string {
	return time.Now().Format("2006.01.02")
}

// getGitCommit returns the current git commit hash.
func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getGitTag returns the current git tag if available.
func getGitTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// isGitDirty returns true if there are uncommitted changes.
func isGitDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// String returns a formatted version string.
func (i Info) String() string {
	var sb strings.Builder

	sb.WriteString(i.Version)

	if i.Commit != "" && i.Commit != "unknown" {
		sb.WriteString(fmt.Sprintf(" (commit: %s)", i.Commit))
	}

	if i.Date != "" {
		sb.WriteString(fmt.Sprintf(" built on %s", i.Date))
	}

	if i.BuiltBy != "" && i.BuiltBy != "goreleaser" {
		sb.WriteString(fmt.Sprintf(" by %s", i.BuiltBy))
	}

	return sb.String()
}

// Short returns just the version string.
func (i Info) Short() string {
	return i.Version
}

// Version returns the version string for fang.
func Version() string {
	return Get().Version
}

// Commit returns the commit string for fang.
func Commit() string {
	return Get().Commit
}
