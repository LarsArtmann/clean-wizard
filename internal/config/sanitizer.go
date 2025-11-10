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
		Original:  make(map[string]any),
		Sanitized: make(map[string]any),
		Changes:   []SanitizationChange{},
		Timestamp: time.Now(),
	}

	// Store original
	result.Original["version"] = cfg.Version
	result.Original["safe_mode"] = cfg.SafeMode
	result.Original["max_disk_usage"] = cfg.MaxDiskUsage
	result.Original["protected"] = cfg.Protected
	result.Original["profiles"] = cfg.Profiles

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
		for i, op := range profile.Operations {
			if cs.rules.TrimWhitespace {
				trimmed := strings.TrimSpace(op.Name)
				if trimmed != op.Name {
					cs.addChange(result, fmt.Sprintf("profiles.%s.operations[%d].name", name, i), op.Name, trimmed, "trimmed whitespace")
					op.Name = trimmed
				}

				trimmedDesc := strings.TrimSpace(op.Description)
				if trimmedDesc != op.Description {
					cs.addChange(result, fmt.Sprintf("profiles.%s.operations[%d].description", name, i), op.Description, trimmedDesc, "trimmed whitespace")
					op.Description = trimmedDesc
				}
			}
		}
	}
}

// applyDefaults applies default values
func (cs *ConfigSanitizer) applyDefaults(cfg *domain.Config, result *SanitizationResult) {
	// Apply default safe mode
	if !cfg.SafeMode && cs.rules.DefaultBackup {
		cs.addChange(result, "safe_mode", false, true, "applied default safe mode")
		cfg.SafeMode = true
	}

	// Apply default max disk usage
	if cfg.MaxDiskUsage == 0 {
		cs.addChange(result, "max_disk_usage", 0, 50, "applied default max disk usage")
		cfg.MaxDiskUsage = 50
	}

	// Clamp max disk usage to reasonable range (5-95)
	if cfg.MaxDiskUsage > 95 {
		cs.addChange(result, "max_disk_usage", cfg.MaxDiskUsage, 95, "clamped to maximum allowed")
		cfg.MaxDiskUsage = 95
	} else if cfg.MaxDiskUsage < 5 && cfg.MaxDiskUsage > 0 {
		cs.addChange(result, "max_disk_usage", cfg.MaxDiskUsage, 5, "clamped to minimum allowed")
		cfg.MaxDiskUsage = 5
	}
}

// addChange records a sanitization change
func (cs *ConfigSanitizer) addChange(result *SanitizationResult, field string, oldValue, newValue any, reason string) {
	change := SanitizationChange{
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
		Reason:   reason,
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
