package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigSanitizer provides comprehensive configuration sanitization
type ConfigSanitizer struct {
	rules *SanitizationRules
}

// SanitizationRules defines sanitization constraints
type SanitizationRules struct {
	TrimWhitespace bool `json:"trim_whitespace"`
	NormalizePaths bool `json:"normalize_paths"`
	AddDefaults    bool `json:"add_defaults"`
	DefaultBackup  bool `json:"default_backup"`
}

// SanitizationResult represents the result of configuration sanitization
type SanitizationResult struct {
	Original        *ValidationSanitizedData `json:"original"`
	Sanitized       *ValidationSanitizedData `json:"sanitized"`
	SanitizedFields []string                `json:"sanitized_fields"`
	Warnings        []SanitizationWarning   `json:"warnings"`
	Changes         []SanitizationChange   `json:"changes"`
	Timestamp       time.Time               `json:"timestamp"`
}

// SanitizationChange represents a single sanitization change
type SanitizationChange struct {
	Field    string    `json:"field"`
	OldValue any       `json:"old_value"`
	NewValue any       `json:"new_value"`
	Reason   string    `json:"reason"`
	Time     time.Time `json:"time"`
}

// SanitizationWarning represents a sanitization warning
type SanitizationWarning struct {
	Field     string `json:"field"`
	Message   string `json:"message"`
	Original  any    `json:"original"`
	Sanitized any    `json:"sanitized"`
	Reason    string `json:"reason"`
}

// NewConfigSanitizer creates a configuration sanitizer
func NewConfigSanitizer() *ConfigSanitizer {
	return &ConfigSanitizer{
		rules: getDefaultSanitizationRules(),
	}
}

// NewConfigSanitizerWithRules creates a sanitizer with custom rules
func NewConfigSanitizerWithRules(rules *SanitizationRules) *ConfigSanitizer {
	return &ConfigSanitizer{
		rules: rules,
	}
}

// SanitizeConfig performs comprehensive configuration sanitization
func (cs *ConfigSanitizer) SanitizeConfig(cfg *domain.Config) *SanitizationResult {
	result := &SanitizationResult{
		Original:  &ValidationSanitizedData{Data: make(map[string]any)},
		Sanitized: &ValidationSanitizedData{Data: make(map[string]any)},
		Changes:   []SanitizationChange{},
		Timestamp: time.Now(),
	}

	// Store original
	result.Original.Data["version"] = cfg.Version
	result.Original.Data["safe_mode"] = cfg.SafeMode
	result.Original.Data["max_disk_usage"] = cfg.MaxDiskUsage
	result.Original.Data["protected"] = cfg.Protected
	result.Original.Data["profiles"] = cfg.Profiles

	// Sanitize basic fields
	cs.sanitizeBasicFields(cfg, result)

	// Sanitize protected paths
	cs.sanitizeProtectedPaths(cfg, result)

	// Sanitize profiles
	cs.sanitizeProfiles(cfg, result)

	// Apply defaults
	if cs.rules.AddDefaults {
		cs.applyDefaults(cfg, result)
	}

	return result
}

// sanitizeBasicFields sanitizes basic configuration fields
func (cs *ConfigSanitizer) sanitizeBasicFields(cfg *domain.Config, result *SanitizationResult) {
	// Sanitize version
	if cs.rules.TrimWhitespace {
		trimmed := strings.TrimSpace(cfg.Version)
		if trimmed != cfg.Version {
			cs.addChange(result, "version", cfg.Version, trimmed, "trimmed whitespace")
			cfg.Version = trimmed
		}
	}

	// Ensure version is not empty
	if cfg.Version == "" {
		cs.addChange(result, "version", cfg.Version, "1.0.0", "applied default version")
		cfg.Version = "1.0.0"
	}
}

// sanitizeProtectedPaths sanitizes protected paths
func (cs *ConfigSanitizer) sanitizeProtectedPaths(cfg *domain.Config, result *SanitizationResult) {
	for i, path := range cfg.Protected {
		if cs.rules.NormalizePaths {
			cleaned := filepath.Clean(path)
			if cleaned != path {
				cs.addChange(result, fmt.Sprintf("protected[%d]", i), path, cleaned, "normalized path")
				cfg.Protected[i] = cleaned
			}
		}

		if cs.rules.TrimWhitespace {
			trimmed := strings.TrimSpace(cfg.Protected[i])
			if trimmed != cfg.Protected[i] {
				cs.addChange(result, fmt.Sprintf("protected[%d]", i), cfg.Protected[i], trimmed, "trimmed whitespace")
				cfg.Protected[i] = trimmed
			}
		}
	}

	// Remove duplicate paths
	if len(cfg.Protected) > 1 {
		seen := make(map[string]bool)
		unique := []string{}
		for _, path := range cfg.Protected {
			if !seen[path] {
				seen[path] = true
				unique = append(unique, path)
			}
		}
		if len(unique) != len(cfg.Protected) {
			cs.addChange(result, "protected", cfg.Protected, unique, "removed duplicates")
			cfg.Protected = unique
		}
	}
}

// sanitizeProfiles sanitizes profile configurations
func (cs *ConfigSanitizer) sanitizeProfiles(cfg *domain.Config, result *SanitizationResult) {
	for name, profile := range cfg.Profiles {
		// Sanitize profile name
		if cs.rules.TrimWhitespace {
			trimmed := strings.TrimSpace(profile.Name)
			if trimmed != profile.Name {
				cs.addChange(result, fmt.Sprintf("profiles.%s.name", name), profile.Name, trimmed, "trimmed whitespace")
				profile.Name = trimmed
			}
		}

		// Sanitize profile description
		if cs.rules.TrimWhitespace {
			trimmed := strings.TrimSpace(profile.Description)
			if trimmed != profile.Description {
				cs.addChange(result, fmt.Sprintf("profiles.%s.description", name), profile.Description, trimmed, "trimmed whitespace")
				profile.Description = trimmed
			}
		}

		// Sanitize operations
		cs.sanitizeOperations(name, profile.Operations, result)
	}
}

// sanitizeOperations sanitizes cleanup operations
func (cs *ConfigSanitizer) sanitizeOperations(profileName string, operations []domain.CleanupOperation, result *SanitizationResult) {
	for i, op := range operations {
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
		cfg.MaxDiskUsage = 50
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
		for i, op := range profile.Operations {
			fieldPrefix := fmt.Sprintf("profiles.%s.operations[%d]", name, i)
			if op.Settings == nil {
				opType := domain.GetOperationType(op.Name)
				op.Settings = domain.DefaultSettings(opType)
				result.addChange(fieldPrefix+".settings", nil, op.Settings, "initialized type-safe settings")
			}
		}
	}
}

// Helper methods

func (r *SanitizationResult) addChange(field string, original, sanitized any, reason string) {
	r.SanitizedFields = append(r.SanitizedFields, field)
	change := SanitizationChange{
		Field:    field,
		OldValue: original,
		NewValue: sanitized,
		Reason:    reason,
		Time:     time.Now(),
	}
	r.Changes = append(r.Changes, change)
}

func (r *SanitizationResult) addWarning(field string, original, sanitized any, reason string) {
	r.Warnings = append(r.Warnings, SanitizationWarning{
		Field:     field,
		Original:  original,
		Sanitized: sanitized,
		Reason:    reason,
	})
}

func (cs *ConfigSanitizer) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (cs *ConfigSanitizer) sortStrings(slice []string) {
	// Simple implementation - in production, use more efficient sorting
	for i := 0; i < len(slice)-1; i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] > slice[j] {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}

// addChange records a sanitization change
func (cs *ConfigSanitizer) addChange(result *SanitizationResult, field string, oldValue, newValue any, reason string) {
	change := SanitizationChange{
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
		Reason:   reason,
		Time:     time.Now(),
	}
	result.Changes = append(result.Changes, change)
}

// getDefaultSanitizationRules returns default sanitization rules
func getDefaultSanitizationRules() *SanitizationRules {
	return &SanitizationRules{
		TrimWhitespace: true,
		NormalizePaths: true,
		AddDefaults:    true,
		DefaultBackup:  true,
	}
}