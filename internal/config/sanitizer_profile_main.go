package config

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	stringsutil "github.com/LarsArtmann/clean-wizard/internal/shared/utils/strings"
)

// SanitizationResultAdapter adapts local SanitizationResult to strings utility interface
type SanitizationResultAdapter struct {
	result *SanitizationResult
}

func (a *SanitizationResultAdapter) AddChange(path string, original, newValue any, reason string) {
	a.result.addChange(path, original, newValue, reason)
}

// sanitizeProfiles sanitizes profiles and their operations
func (cs *ConfigSanitizer) sanitizeProfiles(cfg *domain.Config, result *SanitizationResult) {
	for name, profile := range cfg.Profiles {
		// Sanitize profile name and description using utility
		adapter := &SanitizationResultAdapter{result: result}
		profileFields := stringsutil.NewTrimmableFieldsBuilder().
			AddField("name", fmt.Sprintf("profiles.%s.name", name), &profile.Name).
			AddField("description", fmt.Sprintf("profiles.%s.description", name), &profile.Description).
			Build()

		stringsutil.TrimMultipleFields(cs.rules.TrimWhitespace, profileFields, adapter)

		// Sanitize operations
		cs.sanitizeOperations(name, profile.Operations, result)
	}
}

// sanitizeOperations sanitizes cleanup operations
func (cs *ConfigSanitizer) sanitizeOperations(profileName string, operations []domain.CleanupOperation, result *SanitizationResult) {
	for i := range operations {
		op := &operations[i] // Get pointer to mutate slice element in place
		fieldPrefix := fmt.Sprintf("profiles.%s.operations[%d]", profileName, i)

		// Sanitize operation name and description using utility
		adapter := &SanitizationResultAdapter{result: result}
		opFields := stringsutil.NewTrimmableFieldsBuilder().
			AddField("name", fieldPrefix+".name", &op.Name).
			AddField("description", fieldPrefix+".description", &op.Description).
			Build()

		stringsutil.TrimMultipleFields(cs.rules.TrimWhitespace, opFields, adapter)

		// Sanitize settings
		if op.Settings != nil {
			cs.sanitizeOperationSettings(fieldPrefix+".settings", op.Name, op.Settings, result)
		}
	}
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

	// Ensure default protected paths using centralized source
	if len(cfg.Protected) == 0 {
		cfg.Protected = cs.rules.DefaultProtectedPaths
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
