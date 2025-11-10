package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// OperationType represents cleanup operation type
type OperationType string

const (
	OperationTypeNixStore       OperationType = "nix-store"
	OperationTypeHomebrew      OperationType = "homebrew"
	OperationTypePackageCache   OperationType = "package-cache"
	OperationTypeTempFiles     OperationType = "temp-files"
)

// NixStoreSettings represents Nix store operation settings
type NixStoreSettings struct {
	KeepGenerations int           `json:"keep_generations" yaml:"keep_generations"`
	MinAge         time.Duration `json:"min_age" yaml:"min_age"`
	IncludeProfiles bool          `json:"include_profiles" yaml:"include_profiles"`
	DryRun         bool          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// HomebrewSettings represents Homebrew operation settings
type HomebrewSettings struct {
	KeepLatest     int           `json:"keep_latest" yaml:"keep_latest"`
	CleanupCaches  bool          `json:"cleanup_caches" yaml:"cleanup_caches"`
	MinAge         time.Duration `json:"min_age" yaml:"min_age"`
	DryRun         bool          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// PackageCacheSettings represents package cache operation settings
type PackageCacheSettings struct {
	MaxAge      time.Duration `json:"max_age" yaml:"max_age"`
	MaxSize      int64        `json:"max_size,omitempty" yaml:"max_size,omitempty"`
	IncludeTypes []string     `json:"include_types" yaml:"include_types"`
	DryRun       bool          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// TempFilesSettings represents temporary files operation settings
type TempFilesSettings struct {
	MaxAge      time.Duration `json:"max_age" yaml:"max_age"`
	Paths        []string     `json:"paths,omitempty" yaml:"paths,omitempty"`
	Patterns     []string     `json:"patterns,omitempty" yaml:"patterns,omitempty"`
	ExcludePaths []string     `json:"exclude_paths,omitempty" yaml:"exclude_paths,omitempty"`
	DryRun       bool          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// OperationSettings represents strongly typed operation settings
type OperationSettings struct {
	Type      OperationType       `json:"type" yaml:"type"`
	NixStore  *NixStoreSettings  `json:"nix_store,omitempty" yaml:"nix_store,omitempty"`
	Homebrew   *HomebrewSettings   `json:"homebrew,omitempty" yaml:"homebrew,omitempty"`
	Package    *PackageCacheSettings `json:"package,omitempty" yaml:"package,omitempty"`
	TempFiles  *TempFilesSettings   `json:"temp_files,omitempty" yaml:"temp_files,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for backward compatibility
func (os *OperationSettings) MarshalJSON() ([]byte, error) {
	switch os.Type {
	case OperationTypeNixStore:
		return json.Marshal(os.NixStore)
	case OperationTypeHomebrew:
		return json.Marshal(os.Homebrew)
	case OperationTypePackageCache:
		return json.Marshal(os.Package)
	case OperationTypeTempFiles:
		return json.Marshal(os.TempFiles)
	default:
		return json.Marshal(map[string]any{})
	}
}

// UnmarshalJSON implements custom JSON unmarshaling
func (os *OperationSettings) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as different types
	var nixStore NixStoreSettings
	if err := json.Unmarshal(data, &nixStore); err == nil {
		os.Type = OperationTypeNixStore
		os.NixStore = &nixStore
		return nil
	}

	var homebrew HomebrewSettings
	if err := json.Unmarshal(data, &homebrew); err == nil {
		os.Type = OperationTypeHomebrew
		os.Homebrew = &homebrew
		return nil
	}

	var packageCache PackageCacheSettings
	if err := json.Unmarshal(data, &packageCache); err == nil {
		os.Type = OperationTypePackageCache
		os.Package = &packageCache
		return nil
	}

	var tempFiles TempFilesSettings
	if err := json.Unmarshal(data, &tempFiles); err == nil {
		os.Type = OperationTypeTempFiles
		os.TempFiles = &tempFiles
		return nil
	}

	return fmt.Errorf("unknown operation settings format")
}

// IsValid validates operation settings
func (os *OperationSettings) IsValid() bool {
	switch os.Type {
	case OperationTypeNixStore:
		return os.NixStore != nil && os.NixStore.IsValid()
	case OperationTypeHomebrew:
		return os.Homebrew != nil && os.Homebrew.IsValid()
	case OperationTypePackageCache:
		return os.Package != nil && os.Package.IsValid()
	case OperationTypeTempFiles:
		return os.TempFiles != nil && os.TempFiles.IsValid()
	default:
		return false
	}
}

// Validation methods for each settings type
func (ns *NixStoreSettings) IsValid() bool {
	return ns.KeepGenerations >= 0 && ns.MinAge >= 0
}

func (hs *HomebrewSettings) IsValid() bool {
	return hs.KeepLatest >= 0 && hs.MinAge >= 0
}

func (ps *PackageCacheSettings) IsValid() bool {
	return ps.MaxAge > 0
}

func (ts *TempFilesSettings) IsValid() bool {
	return ts.MaxAge > 0 && len(ts.Paths) > 0
}