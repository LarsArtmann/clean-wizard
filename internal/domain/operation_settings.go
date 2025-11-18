package domain

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
}

// NixGenerationsSettings provides type-safe settings for Nix generations cleanup
type NixGenerationsSettings struct {
	Generations int  `json:"generations"`
	Optimize    bool `json:"optimize"`
}

// TempFilesSettings provides type-safe settings for temporary files cleanup
type TempFilesSettings struct {
	OlderThan string   `json:"older_than"`
	Excludes  []string `json:"excludes,omitempty"`
}

// HomebrewSettings provides type-safe settings for Homebrew cleanup
type HomebrewSettings struct {
	UnusedOnly bool   `json:"unused_only"`
	Prune      string `json:"prune,omitempty"`
}

// SystemTempSettings provides type-safe settings for system temp cleanup
type SystemTempSettings struct {
	Paths     []string `json:"paths"`
	OlderThan string   `json:"older_than"`
}

// OperationType represents different types of cleanup operations
type OperationType string

const (
	OperationTypeNixGenerations OperationType = "nix-generations"
	OperationTypeTempFiles      OperationType = "temp-files"
	OperationTypeHomebrew       OperationType = "homebrew-cleanup"
	OperationTypeSystemTemp     OperationType = "system-temp"
)

// GetOperationType returns the operation type from operation name
func GetOperationType(name string) OperationType {
	switch name {
	case "nix-generations":
		return OperationTypeNixGenerations
	case "temp-files":
		return OperationTypeTempFiles
	case "homebrew-cleanup":
		return OperationTypeHomebrew
	case "system-temp":
		return OperationTypeSystemTemp
	default:
		return OperationType(name) // Fallback for custom types
	}
}

// DefaultSettings returns default settings for the given operation type
func DefaultSettings(opType OperationType) *OperationSettings {
	switch opType {
	case OperationTypeNixGenerations:
		return &OperationSettings{
			NixGenerations: &NixGenerationsSettings{
				Generations: 1,
				Optimize:    false,
			},
		}
	case OperationTypeTempFiles:
		return &OperationSettings{
			TempFiles: &TempFilesSettings{
				OlderThan: "7d",
				Excludes:  []string{"/tmp/keep"},
			},
		}
	case OperationTypeHomebrew:
		return &OperationSettings{
			Homebrew: &HomebrewSettings{
				UnusedOnly: true,
			},
		}
	case OperationTypeSystemTemp:
		return &OperationSettings{
			SystemTemp: &SystemTempSettings{
				Paths:     []string{"/tmp", "/var/tmp"},
				OlderThan: "30d",
			},
		}
	default:
		return &OperationSettings{} // Empty settings for custom types
	}
}

// ValidateSettings validates settings for the given operation type
func (os *OperationSettings) ValidateSettings(opType OperationType) error {
	switch opType {
	case OperationTypeNixGenerations:
		if os.NixGenerations == nil {
			return nil // Optional settings
		}
		if os.NixGenerations.Generations < 1 || os.NixGenerations.Generations > 10 {
			return &ValidationError{
				Field:   "nix_generations.generations",
				Message: "generations must be between 1 and 10",
				Value:   os.NixGenerations.Generations,
			}
		}
	case OperationTypeTempFiles:
		if os.TempFiles == nil {
			return nil
		}
		if os.TempFiles.OlderThan == "" {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than is required",
				Value:   os.TempFiles.OlderThan,
			}
		}
		if _, err := ParseCustomDuration(os.TempFiles.OlderThan); err != nil {
			return &ValidationError{
				Field:   "temp_files.older_than",
				Message: "older_than must be a valid duration (e.g., '7d', '24h')",
				Value:   os.TempFiles.OlderThan,
			}
		}
	case OperationTypeHomebrew:
		// Homebrew settings are always valid
	case OperationTypeSystemTemp:
		if os.SystemTemp == nil {
			return nil
		}
		if len(os.SystemTemp.Paths) == 0 {
			return &ValidationError{
				Field:   "system_temp.paths",
				Message: "paths are required",
				Value:   os.SystemTemp.Paths,
			}
		}
	}
	return nil
}

// ValidationError represents a settings validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value"`
}

func (e *ValidationError) Error() string {
	return e.Message
}