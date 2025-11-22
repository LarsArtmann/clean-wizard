// TYPE-SAFE-EXEMPT: Documentation mentioning map[string]any for explanatory purposes
package shared

// OperationSettings provides type-safe configuration for different operation types
// This eliminates map[string]any violations while maintaining flexibility
type OperationSettings struct {
	// Nix Generations Settings
	NixGenerations *NixGenerationsSettings `json:"nix_generations,omitempty"`

	// Temp Files Settings
	TempFiles *TempFilesSettings `json:"temp_files,omitempty"`

	// Homebrew Settings
	Homebrew *HomebrewSettings `json:"homebrew,omitempty"`

	// System Temp Settings
	SystemTemp *SystemTempSettings `json:"system_temp,omitempty"`

	// General settings
	ExecutionMode       ExecutionModeType `json:"executionMode,omitempty"`
	Verbose             bool              `json:"verbose,omitempty"`
	TimeoutSeconds      int32             `json:"timeoutSeconds,omitempty"`
	ConfirmBeforeDelete bool              `json:"confirmBeforeDelete,omitempty"`
}

// NixGenerationsSettings provides configuration for Nix generations cleanup
type NixGenerationsSettings struct {
	Generations int `json:"generations,omitempty"`
}

// TempFilesSettings provides configuration for temporary files cleanup
type TempFilesSettings struct {
	OlderThan string   `json:"older_than,omitempty"` // Duration string (e.g., "7d", "24h")
	Excludes  []string `json:"excludes,omitempty"`   // Paths to exclude
}

// HomebrewSettings provides configuration for Homebrew cleanup
type HomebrewSettings struct {
	Prune bool `json:"prune,omitempty"` // Whether to prune old versions
}

// SystemTempSettings provides configuration for system temp cleanup
type SystemTempSettings struct {
	OlderThan string   `json:"older_than,omitempty"` // Duration string (e.g., "7d", "24h")
	Excludes  []string `json:"excludes,omitempty"`   // Paths to exclude
	Paths     []string `json:"paths,omitempty"`      // Additional paths to clean
}
