package operations

import (
	"encoding/json/v2"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

// GitHistoryMode represents the mode of operation for git history cleaning.
//

type GitHistoryMode int

const (
	GitHistoryModeAnalyze GitHistoryMode = iota // Only report findings
	GitHistoryModeDryRun                        // Show what would be done
	GitHistoryModeExecute                       // Actually rewrite history
)

var gitHistoryModeStrings = []string{"analyze", "dry-run", "execute"} //nolint:gochecknoglobals

func (m GitHistoryMode) String() string { return enums.EnumString(m, gitHistoryModeStrings) }
func (m GitHistoryMode) IsValid() bool  { return enums.EnumIsValid(m, GitHistoryModeExecute) }
func (m GitHistoryMode) Values() []GitHistoryMode {
	return enums.EnumValues[GitHistoryMode](GitHistoryModeExecute)
}

func (m GitHistoryMode) MarshalJSON() ([]byte, error) {
	return enums.EnumMarshalJSON(m, gitHistoryModeStrings)
}

func (m *GitHistoryMode) UnmarshalJSON(data []byte) error {
	err := enums.EnumUnmarshalJSON(data, (*int)(m), gitHistoryModeStrings, "git history mode")
	if err != nil {
		var s string
		if jsonErr := json.Unmarshal(data, &s); jsonErr == nil && strings.EqualFold(s, "dryrun") {
			*m = GitHistoryModeDryRun

			return nil
		}
	}

	return err
}

// GitHistorySettings provides type-safe settings for git history cleaning.
type GitHistorySettings struct {
	// MinSizeMB is the minimum file size in MB to consider (default: 1)
	MinSizeMB int `json:"min_size_mb,omitempty" yaml:"min_size_mb,omitempty"`
	// MaxFiles limits the number of files to show (0 = unlimited)
	MaxFiles int `json:"max_files,omitempty" yaml:"max_files,omitempty"`
	// ExcludeExtensions are file extensions to exclude
	ExcludeExtensions []string `json:"exclude_extensions,omitempty" yaml:"exclude_extensions,omitempty"`
	// IncludeExtensions are file extensions to include (empty = all)
	IncludeExtensions []string `json:"include_extensions,omitempty" yaml:"include_extensions,omitempty"`
	// ExcludePaths are path patterns to exclude
	ExcludePaths []string `json:"exclude_paths,omitempty" yaml:"exclude_paths,omitempty"`
	// CreateBackup indicates if a backup should be created before rewrite
	CreateBackup bool `json:"create_backup,omitempty" yaml:"create_backup,omitempty"`
	// SkipConfirmation skips the interactive confirmation (dangerous)
	SkipConfirmation bool `json:"skip_confirmation,omitempty" yaml:"skip_confirmation,omitempty"`
}

// DefaultGitHistorySettings returns the default settings.
func DefaultGitHistorySettings() GitHistorySettings {
	return GitHistorySettings{ //nolint:exhaustruct
		MinSizeMB:         1,
		MaxFiles:          100,
		CreateBackup:      true,
		SkipConfirmation:  false,
		ExcludeExtensions: []string{".pdf", ".png", ".jpg", ".jpeg", ".gif", ".svg"},
	}
}

// DefaultBinaryExtensions are binary extensions commonly found in git history that should be cleaned.
var DefaultBinaryExtensions = []string{ //nolint:gochecknoglobals
	// Go build outputs (often extensionless)
	"", // Extensionless binaries
	".exe",
	".dll",
	".so",
	".dylib",
	".a",
	".o",
	".out",
	".app",

	// Build artifacts
	".test",
	".bench",
	".prof",

	// Archives
	".zip",
	".tar",
	".gz",
	".bz2",
	".xz",
	".7z",
	".rar",

	// Databases
	".db",
	".sqlite",
	".sqlite3",

	// Serialized data
	".bin",
	".dat",
	".data",

	// Large generated files
	".wasm",
	".class",
	".jar",
}

// ExtensionsToKeep are binary extensions that should typically NOT be removed.
var ExtensionsToKeep = []string{ //nolint:gochecknoglobals
	".pdf",
	".png",
	".jpg",
	".jpeg",
	".gif",
	".svg",
	".ico",
	".woff",
	".woff2",
	".ttf",
	".eot",
	".mp3",
	".mp4",
	".webm",
	".webp",
}
