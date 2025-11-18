package config

import (
	"fmt"
	"strings"

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
			cs.sanitizeOperationSettings(fieldPrefix+".settings", op.Settings, result)
		}
	}
}

// sanitizeOperationSettings sanitizes operation settings with type safety
func (cs *ConfigSanitizer) sanitizeOperationSettings(fieldPrefix string, settings *domain.OperationSettings, result *SanitizationResult) {
	// TODO: Implement type-safe operation settings sanitization
	// For now, just record that settings were processed
	result.SanitizedFields = append(result.SanitizedFields, fieldPrefix)
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
