package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// sanitizeProfiles sanitizes profiles and their operations
func (cs *ConfigSanitizer) sanitizeProfiles(cfg *domain.Config, result *SanitizationResult) {
	for name, profile := range cfg.Profiles {
		// Sanitize profile name
		if cs.rules.TrimWhitespace {
			original := profile.Name
			profile.Name = strings.TrimSpace(profile.Name)
			if original != profile.Name {
				result.addChange(fmt.Sprintf("profiles.%s.name", name), original, profile.Name, "trimmed whitespace")
			}
		}

		// Sanitize profile description
		if cs.rules.TrimWhitespace {
			original := profile.Description
			profile.Description = strings.TrimSpace(profile.Description)
			if original != profile.Description {
				result.addChange(fmt.Sprintf("profiles.%s.description", name), original, profile.Description, "trimmed whitespace")
			}
		}

		// Sanitize operations
		cs.sanitizeOperations(name, profile.Operations, result)
	}
}

// sanitizeOperations sanitizes cleanup operations
func (cs *ConfigSanitizer) sanitizeOperations(profileName string, operations []domain.CleanupOperation, result *SanitizationResult) {
	for i := range operations {
		op := &operations[i] // Get pointer to mutate slice element in place
		fieldPrefix := fmt.Sprintf("profiles.%s.operations[%d]", profileName, i)

		// Sanitize operation name
		if cs.rules.TrimWhitespace {
			original := op.Name
			op.Name = strings.TrimSpace(op.Name)
			if original != op.Name {
				result.addChange(fieldPrefix+".name", original, op.Name, "trimmed whitespace")
			}
		}

		// Sanitize operation description
		if cs.rules.TrimWhitespace {
			original := op.Description
			op.Description = strings.TrimSpace(op.Description)
			if original != op.Description {
				result.addChange(fieldPrefix+".description", original, op.Description, "trimmed whitespace")
			}
		}

		// Sanitize settings
		if op.Settings != nil {
			cs.sanitizeOperationSettings(fieldPrefix+".settings", op.Name, op.Settings, result)
		}
	}
}

// sanitizeOperationSettings sanitizes operation settings with type safety
func (cs *ConfigSanitizer) sanitizeOperationSettings(fieldPrefix string, operationName string, settings *domain.OperationSettings, result *SanitizationResult) {
	opType := domain.GetOperationType(operationName)
	
	// Validate settings first
	if err := settings.ValidateSettings(opType); err != nil {
		// Convert validation errors to warnings since the result type doesn't have an Errors field
		if validationErr, ok := err.(*domain.ValidationError); ok {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + "." + validationErr.Field,
				Original:  validationErr.Value,
				Sanitized: validationErr.Value,
				Reason:    validationErr.Message,
			})
		} else {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix,
				Original:  "settings validation",
				Sanitized: "settings validation",
				Reason:    fmt.Sprintf("validation error: %v", err),
			})
		}
		return
	}

	// Type-aware sanitization based on operation type
	switch opType {
	case domain.OperationTypeNixGenerations:
		cs.sanitizeNixGenerationsSettings(fieldPrefix, settings.NixGenerations, result)
		
	case domain.OperationTypeTempFiles:
		cs.sanitizeTempFilesSettings(fieldPrefix, settings.TempFiles, result)
		
	case domain.OperationTypeHomebrew:
		cs.sanitizeHomebrewSettings(fieldPrefix, settings.Homebrew, result)
		
	case domain.OperationTypeSystemTemp:
		cs.sanitizeSystemTempSettings(fieldPrefix, settings.SystemTemp, result)
		
	default:
		// For custom operation types, just record that they were processed
		result.Warnings = append(result.Warnings, SanitizationWarning{
			Field:     fieldPrefix,
			Original:  "custom operation settings",
			Sanitized: "custom operation settings",
			Reason:    "custom operation type - no specific sanitization applied",
		})
	}
}

// sanitizeNixGenerationsSettings sanitizes Nix generations settings
func (cs *ConfigSanitizer) sanitizeNixGenerationsSettings(fieldPrefix string, settings *domain.NixGenerationsSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize generations range (ensure 1-10)
	original := settings.Generations
	if settings.Generations < 1 {
		settings.Generations = 1
		result.addChange(fieldPrefix+".generations", original, settings.Generations, "clamped to minimum value")
	} else if settings.Generations > 10 {
		settings.Generations = 10
		result.addChange(fieldPrefix+".generations", original, settings.Generations, "clamped to maximum value")
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".nix_generations")
}

// sanitizeTempFilesSettings sanitizes temporary files settings
func (cs *ConfigSanitizer) sanitizeTempFilesSettings(fieldPrefix string, settings *domain.TempFilesSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize older_than duration
	if cs.rules.TrimWhitespace && settings.OlderThan != "" {
		original := settings.OlderThan
		settings.OlderThan = strings.TrimSpace(settings.OlderThan)
		if original != settings.OlderThan {
			result.addChange(fieldPrefix+".older_than", original, settings.OlderThan, "trimmed whitespace")
		}

		// Validate duration format
		if _, err := time.ParseDuration(settings.OlderThan); err != nil {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + ".older_than",
				Original:  settings.OlderThan,
				Sanitized: settings.OlderThan,
				Reason:    fmt.Sprintf("invalid duration format: %v", err),
			})
		}
	}

	// Sanitize excludes array
	if len(settings.Excludes) > 0 {
		sanitizedExcludes := make([]string, 0, len(settings.Excludes))
		for i, exclude := range settings.Excludes {
			original := exclude

			if cs.rules.TrimWhitespace {
				exclude = strings.TrimSpace(exclude)
			}

			if cs.rules.NormalizePaths {
				exclude = filepath.Clean(exclude)
			}

			if original != exclude {
				result.addChange(fmt.Sprintf("%s.excludes[%d]", fieldPrefix, i), original, exclude, "sanitized exclude path")
			}

			sanitizedExcludes = append(sanitizedExcludes, exclude)
		}

		// Remove duplicates and sort
		if cs.rules.RemoveDuplicates {
			sanitizedExcludes = cs.removeDuplicates(sanitizedExcludes)
		}
		if cs.rules.SortArrays {
			cs.sortStrings(sanitizedExcludes)
		}

		settings.Excludes = sanitizedExcludes
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".temp_files")
}

// sanitizeHomebrewSettings sanitizes Homebrew settings
func (cs *ConfigSanitizer) sanitizeHomebrewSettings(fieldPrefix string, settings *domain.HomebrewSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize prune field
	if cs.rules.TrimWhitespace && settings.Prune != "" {
		original := settings.Prune
		settings.Prune = strings.TrimSpace(settings.Prune)
		if original != settings.Prune {
			result.addChange(fieldPrefix+".prune", original, settings.Prune, "trimmed whitespace")
		}
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".homebrew")
}

// sanitizeSystemTempSettings sanitizes system temp settings
func (cs *ConfigSanitizer) sanitizeSystemTempSettings(fieldPrefix string, settings *domain.SystemTempSettings, result *SanitizationResult) {
	if settings == nil {
		return
	}

	// Sanitize paths array
	if len(settings.Paths) > 0 {
		sanitizedPaths := make([]string, 0, len(settings.Paths))
		for i, path := range settings.Paths {
			original := path

			if cs.rules.TrimWhitespace {
				path = strings.TrimSpace(path)
			}

			if cs.rules.NormalizePaths {
				path = filepath.Clean(path)
			}

			// Validate absolute path requirement
			if !filepath.IsAbs(path) {
				result.Warnings = append(result.Warnings, SanitizationWarning{
					Field:     fmt.Sprintf("%s.paths[%d]", fieldPrefix, i),
					Original:  original,
					Sanitized: path,
					Reason:    fmt.Sprintf("system temp path must be absolute: %s", original),
				})
				continue // Skip processing invalid path
			}

			if original != path {
				result.addChange(fmt.Sprintf("%s.paths[%d]", fieldPrefix, i), original, path, "sanitized system temp path")
			}

			sanitizedPaths = append(sanitizedPaths, path)
		}

		// Remove duplicates and sort
		if cs.rules.RemoveDuplicates {
			sanitizedPaths = cs.removeDuplicates(sanitizedPaths)
		}
		if cs.rules.SortArrays {
			cs.sortStrings(sanitizedPaths)
		}

		settings.Paths = sanitizedPaths
	}

	// Sanitize older_than duration
	if cs.rules.TrimWhitespace && settings.OlderThan != "" {
		original := settings.OlderThan
		settings.OlderThan = strings.TrimSpace(settings.OlderThan)
		if original != settings.OlderThan {
			result.addChange(fieldPrefix+".older_than", original, settings.OlderThan, "trimmed whitespace")
		}

		// Validate duration format
		if _, err := time.ParseDuration(settings.OlderThan); err != nil {
			result.Warnings = append(result.Warnings, SanitizationWarning{
				Field:     fieldPrefix + ".older_than",
				Original:  settings.OlderThan,
				Sanitized: settings.OlderThan,
				Reason:    fmt.Sprintf("invalid duration format: %v", err),
			})
		}
	}

	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix+".system_temp")
}

// applyDefaults applies default values to missing fields
func (cs *ConfigSanitizer) applyDefaults(cfg *domain.Config, result *SanitizationResult) {
	// Set default version if empty
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
		result.addChange("version", "", cfg.Version, "applied default version")
	}

	// Set default max disk usage
	if cfg.MaxDiskUsage == 0 {
		cfg.MaxDiskUsage = cs.rules.DefaultMaxDiskUsage
		result.addChange("max_disk_usage", 0, cfg.MaxDiskUsage, "applied default max disk usage")
	}

	// Ensure default protected paths
	if len(cfg.Protected) == 0 {
		cfg.Protected = []string{"/System", "/Applications", "/Library"}
		result.addChange("protected", []string{}, cfg.Protected, "applied default protected paths")
	}

	// Apply defaults to profiles
	for name, profile := range cfg.Profiles {
		if profile.Name == "" {
			profile.Name = name
			result.addChange(fmt.Sprintf("profiles.%s.name", name), "", profile.Name, "applied default profile name")
		}

		if profile.Description == "" {
			profile.Description = "Auto-generated profile description"
			result.addChange(fmt.Sprintf("profiles.%s.description", name), "", profile.Description, "applied default description")
		}

		// Apply defaults to operations
		for i := range profile.Operations {
			op := &profile.Operations[i] // Get pointer to mutate slice element in place
			fieldPrefix := fmt.Sprintf("profiles.%s.operations[%d]", name, i)

			if op.Settings == nil {
				opType := domain.GetOperationType(op.Name)
				op.Settings = domain.DefaultSettings(opType)
				result.addChange(fieldPrefix+".settings", nil, op.Settings, "initialized type-safe settings")
			}
		}
	}
}
