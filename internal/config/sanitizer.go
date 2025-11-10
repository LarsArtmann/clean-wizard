package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigSanitizer provides configuration sanitization and normalization
type ConfigSanitizer struct {
	rules *SanitizationRules
}

// SanitizationRules defines how to sanitize different configuration aspects
type SanitizationRules struct {
	// Path sanitization
	NormalizePaths bool `json:"normalize_paths"`
	ExpandHomeDir  bool `json:"expand_home_dir"`
	ValidateExists bool `json:"validate_exists"`

	// Numeric sanitization
	ClampValues      bool `json:"clamp_values"`
	RoundPercentages bool `json:"round_percentages"`

	// String sanitization
	TrimWhitespace bool `json:"trim_whitespace"`
	NormalizeCase  bool `json:"normalize_case"`

	// Structural sanitization
	SortArrays       bool `json:"sort_arrays"`
	RemoveDuplicates bool `json:"remove_duplicates"`
	AddDefaults      bool `json:"add_defaults"`

	// Safety defaults
	DefaultSafeMode     bool          `json:"default_safe_mode"`
	DefaultMaxDiskUsage int           `json:"default_max_disk_usage"`
	DefaultBackup       time.Duration `json:"default_backup"`
}

// SanitizationResult contains sanitization outcomes
type SanitizationResult struct {
	SanitizedFields []string              `json:"sanitized_fields"`
	Warnings        []SanitizationWarning `json:"warnings"`
	Changes         map[string]any        `json:"changes"`
	Timestamp       time.Time             `json:"timestamp"`
}

// SanitizationWarning represents a sanitization warning
type SanitizationWarning struct {
	Field     string `json:"field"`
	Original  any    `json:"original"`
	Sanitized any    `json:"sanitized"`
	Reason    string `json:"reason"`
}

// NewConfigSanitizer creates a configuration sanitizer with default rules
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
func (cs *ConfigSanitizer) SanitizeConfig(cfg *domain.Config, validationResult *ValidationResult) {
	result := &SanitizationResult{
		SanitizedFields: []string{},
		Warnings:        []SanitizationWarning{},
		Changes:         make(map[string]any),
		Timestamp:       time.Now(),
	}

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

	// Update validation result with sanitization info
	if validationResult != nil {
		validationResult.Sanitized = map[string]any{
			"sanitized_fields": result.SanitizedFields,
			"warnings":         result.Warnings,
			"changes":          result.Changes,
		}
	}
}

// sanitizeBasicFields sanitizes basic configuration fields
func (cs *ConfigSanitizer) sanitizeBasicFields(cfg *domain.Config, result *SanitizationResult) {
	// Sanitize version
	if cs.rules.TrimWhitespace {
		original := cfg.Version
		cfg.Version = strings.TrimSpace(cfg.Version)
		if original != cfg.Version {
			result.addChange("version", original, cfg.Version, "trimmed whitespace")
		}
	}

	if cs.rules.NormalizeCase {
		original := cfg.Version
		cfg.Version = strings.ToLower(cfg.Version)
		if original != cfg.Version {
			result.addChange("version", original, cfg.Version, "normalized to lowercase")
		}
	}

	// Sanitize max_disk_usage
	if cs.rules.ClampValues {
		original := cfg.MaxDiskUsage
		if cfg.MaxDiskUsage < 10 {
			cfg.MaxDiskUsage = 10
			result.addWarning("max_disk_usage", original, cfg.MaxDiskUsage, "clamped to minimum value")
		} else if cfg.MaxDiskUsage > 95 {
			cfg.MaxDiskUsage = 95
			result.addWarning("max_disk_usage", original, cfg.MaxDiskUsage, "clamped to maximum value")
		}
	}

	if cs.rules.RoundPercentages {
		original := cfg.MaxDiskUsage
		cfg.MaxDiskUsage = int(float64(cfg.MaxDiskUsage+5)/10) * 10 // Round to nearest 10
		if original != cfg.MaxDiskUsage {
			result.addChange("max_disk_usage", original, cfg.MaxDiskUsage, "rounded to nearest 10%")
		}
	}

	// Ensure safe mode defaults
	if cs.rules.DefaultSafeMode && !cfg.SafeMode {
		original := cfg.SafeMode
		cfg.SafeMode = true
		result.addChange("safe_mode", original, cfg.SafeMode, "enabled safe mode for security")
	}
}

// sanitizeProtectedPaths sanitizes protected paths array
func (cs *ConfigSanitizer) sanitizeProtectedPaths(cfg *domain.Config, result *SanitizationResult) {
	sanitizedPaths := make([]string, 0, len(cfg.Protected))

	for i, path := range cfg.Protected {
		original := path

		// Trim whitespace
		if cs.rules.TrimWhitespace {
			path = strings.TrimSpace(path)
		}

		// Expand home directory
		if cs.rules.ExpandHomeDir {
			if strings.HasPrefix(path, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					path = filepath.Join(home, path[2:])
				}
			}
		}

		// Normalize path
		if cs.rules.NormalizePaths {
			path = filepath.Clean(path)
		}

		// Ensure absolute path
		if !filepath.IsAbs(path) {
			path = "/" + path
		}

		// Validate existence if enabled
		if cs.rules.ValidateExists {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				result.Warnings = append(result.Warnings, SanitizationWarning{
					Field:     fmt.Sprintf("protected[%d]", i),
					Original:  original,
					Sanitized: path,
					Reason:    "path does not exist but will be protected",
				})
			}
		}

		if original != path {
			result.addChange(fmt.Sprintf("protected[%d]", i), original, path, "path normalized")
		}

		sanitizedPaths = append(sanitizedPaths, path)
	}

	// Remove duplicates
	if cs.rules.RemoveDuplicates {
		sanitizedPaths = cs.removeDuplicates(sanitizedPaths)
	}

	// Sort paths
	if cs.rules.SortArrays {
		cs.sortStrings(sanitizedPaths)
	}

	cfg.Protected = sanitizedPaths
}

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

// sanitizeOperationSettings sanitizes operation settings
func (cs *ConfigSanitizer) sanitizeOperationSettings(fieldPrefix string, settings map[string]any, result *SanitizationResult) {
	for key, value := range settings {
		switch v := value.(type) {
		case string:
			if cs.rules.TrimWhitespace {
				original := v
				sanitized := strings.TrimSpace(v)
				if original != sanitized {
					settings[key] = sanitized
					result.addChange(fieldPrefix+"."+key, original, sanitized, "trimmed whitespace")
				}
			}

		case int:
			if cs.rules.ClampValues {
				// Example: clamp generation count to reasonable range
				if key == "generations" {
					original := v
					if v < 1 {
						settings[key] = 1
						result.addWarning(fieldPrefix+"."+key, original, 1, "clamped to minimum 1 generation")
					} else if v > 10 {
						settings[key] = 10
						result.addWarning(fieldPrefix+"."+key, original, 10, "clamped to maximum 10 generations")
					}
				}
			}

		case []any:
			if key == "excludes" {
				sanitized := cs.sanitizeStringArray(v)
				if !cs.arraysEqual(v, sanitized) {
					settings[key] = sanitized
					result.addChange(fieldPrefix+"."+key, v, sanitized, "sanitized string array")
				}
			}
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
				op.Settings = make(map[string]any)
				result.addChange(fieldPrefix+".settings", nil, op.Settings, "initialized settings map")
			}

			// Apply operation-specific defaults
			cs.applyOperationDefaults(fieldPrefix, op.Name, op.Settings, result)
		}
	}
}

// applyOperationDefaults applies operation-specific default settings
func (cs *ConfigSanitizer) applyOperationDefaults(fieldPrefix string, opName string, settings map[string]any, result *SanitizationResult) {
	switch opName {
	case "nix-generations":
		if _, exists := settings["generations"]; !exists {
			settings["generations"] = 1
			result.addChange(fieldPrefix+".settings.generations", nil, 1, "applied default generations")
		}
		if _, exists := settings["optimize"]; !exists {
			settings["optimize"] = false
			result.addChange(fieldPrefix+".settings.optimize", nil, false, "applied default optimize")
		}

	case "temp-files":
		if _, exists := settings["older_than"]; !exists {
			settings["older_than"] = "7d"
			result.addChange(fieldPrefix+".settings.older_than", nil, "7d", "applied default older_than")
		}
		if _, exists := settings["excludes"]; !exists {
			settings["excludes"] = []string{"/tmp/keep"}
			result.addChange(fieldPrefix+".settings.excludes", nil, []string{"/tmp/keep"}, "applied default excludes")
		}

	case "homebrew-cleanup":
		if _, exists := settings["unused_only"]; !exists {
			settings["unused_only"] = true
			result.addChange(fieldPrefix+".settings.unused_only", nil, true, "applied default unused_only")
		}
	}
}

// Helper methods

func (r *SanitizationResult) addChange(field string, original, sanitized any, reason string) {
	r.SanitizedFields = append(r.SanitizedFields, field)
	r.Changes[field] = map[string]any{
		"original":  original,
		"sanitized": sanitized,
		"reason":    reason,
		"timestamp": time.Now(),
	}
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

func (cs *ConfigSanitizer) sanitizeStringArray(arr []any) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		if str, ok := item.(string); ok {
			result[i] = strings.TrimSpace(str)
		} else {
			result[i] = item
		}
	}
	return result
}

func (cs *ConfigSanitizer) arraysEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// getDefaultSanitizationRules returns default sanitization configuration
func getDefaultSanitizationRules() *SanitizationRules {
	return &SanitizationRules{
		NormalizePaths:      true,
		ExpandHomeDir:       true,
		ValidateExists:      false, // Don't validate exists by default as paths may not exist yet
		ClampValues:         true,
		RoundPercentages:    true,
		TrimWhitespace:      true,
		NormalizeCase:       false, // Preserve case for paths and names
		SortArrays:          true,
		RemoveDuplicates:    true,
		AddDefaults:         true,
		DefaultSafeMode:     true,
		DefaultMaxDiskUsage: 50,
		DefaultBackup:       24 * time.Hour,
	}
}
