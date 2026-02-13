package domain

import (
	stdctx "context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/shared/context"
)

// SanitizeConfigResult contains the result of sanitizing a configuration.
type SanitizeConfigResult struct {
	Config          *Config
	Changes         int
	Warnings        []SanitizationWarning
	Duration        time.Duration
	Timestamp       time.Time
	ValidationLevel ValidationLevelType
}

// SanitizationWarning represents a warning during sanitization.
type SanitizationWarning struct {
	Field    string
	Message  string
	OldValue any
	NewValue any
}

// Sanitize applies sanitization rules to the configuration using the generic Context system.
// Returns a SanitizeConfigResult with details about changes made.
func (c *Config) Sanitize(ctx stdctx.Context, level ValidationLevelType) *SanitizeConfigResult {
	start := time.Now()

	// Create validation config using the generic Context system
	validationConfig := context.ValidationConfig{
		ValidationLevel: level.String(),
		Profile:         c.CurrentProfile,
		Section:         "config",
		Constraints:     make(map[string]string),
		Metadata: map[string]string{
			"sanitize_operation": "true",
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	}

	// Create the generic context - keep the stdlib context separate
	_ = context.NewContext[context.ValidationConfig](ctx, validationConfig)

	result := &SanitizeConfigResult{
		Config:          c,
		Warnings:        []SanitizationWarning{},
		Timestamp:       time.Now(),
		ValidationLevel: level,
	}

	// Apply sanitization rules based on validation level
	switch level {
	case ValidationLevelNoneType:
		// No sanitization
	case ValidationLevelBasicType:
		result.applyBasicSanitization()
	case ValidationLevelComprehensiveType:
		result.applyComprehensiveSanitization(c)
	case ValidationLevelStrictType:
		result.applyStrictSanitization(c)
	}

	result.Duration = time.Since(start)
	return result
}

// applyBasicSanitization performs basic sanitization without profile context.
func (r *SanitizeConfigResult) applyBasicSanitization() {
	changes := 0

	// Basic field sanitization is handled inline since we're in the domain package
	// and don't have access to the config package's sanitizer

	r.Changes = changes
}

// applyComprehensiveSanitization performs comprehensive sanitization.
func (r *SanitizeConfigResult) applyComprehensiveSanitization(c *Config) {
	changes := 0

	// Sanitize version string
	if c.Version != "" {
		original := c.Version
		sanitized := sanitizeString(original)
		if sanitized != original {
			c.Version = sanitized
			changes++
		}
	}

	// Sanitize protected paths
	for i, path := range c.Protected {
		if sanitized := sanitizePath(path); sanitized != path {
			c.Protected[i] = sanitized
			changes++
		}
	}

	// Remove duplicate protected paths
	seen := make(map[string]bool)
	uniquePaths := make([]string, 0, len(c.Protected))
	for _, path := range c.Protected {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}
	if len(uniquePaths) != len(c.Protected) {
		c.Protected = uniquePaths
		changes++
	}

	r.Changes = changes
}

// applyStrictSanitization performs strict sanitization with all validations.
func (r *SanitizeConfigResult) applyStrictSanitization(c *Config) {
	changes := 0

	// Apply comprehensive sanitization first
	r.applyComprehensiveSanitization(c)
	changes = r.Changes

	// Clamp MaxDiskUsage to valid range
	if c.MaxDiskUsage < 0 {
		r.Warnings = append(r.Warnings, SanitizationWarning{
			Field:    "max_disk_usage",
			Message:  "MaxDiskUsage was negative, clamping to 0",
			OldValue: c.MaxDiskUsage,
			NewValue: 0,
		})
		c.MaxDiskUsage = 0
		changes++
	} else if c.MaxDiskUsage > 100 {
		r.Warnings = append(r.Warnings, SanitizationWarning{
			Field:    "max_disk_usage",
			Message:  "MaxDiskUsage exceeded 100, clamping to 100",
			OldValue: c.MaxDiskUsage,
			NewValue: 100,
		})
		c.MaxDiskUsage = 100
		changes++
	}

	// Ensure at least one default protected path exists
	defaultProtected := []string{"/System", "/Applications", "/Library"}
	hasSystemPath := false
	for _, path := range c.Protected {
		if path == "/System" || path == "/System/" {
			hasSystemPath = true
			break
		}
	}
	if !hasSystemPath && len(c.Protected) == 0 {
		c.Protected = append([]string{"/System"}, defaultProtected[1:]...)
		changes++
		r.Warnings = append(r.Warnings, SanitizationWarning{
			Field:    "protected",
			Message:  "Added default protected paths",
			OldValue: nil,
			NewValue: c.Protected,
		})
	}

	// Sanitize profile names and descriptions
	for _, profile := range c.Profiles {
		if sanitizeString(profile.Name) != profile.Name {
			profile.Name = sanitizeString(profile.Name)
			changes++
		}
		if sanitizeString(profile.Description) != profile.Description {
			profile.Description = sanitizeString(profile.Description)
			changes++
		}
	}

	r.Changes = changes
}

// sanitizeString performs basic string sanitization.
func sanitizeString(s string) string {
	// Trim whitespace
	s = trimAllWhitespace(s)
	return s
}

// sanitizePath performs path sanitization.
func sanitizePath(path string) string {
	// Trim whitespace
	sanitized := trimAllWhitespace(path)
	return sanitized
}

// trimAllWhitespace removes all whitespace characters from a string.
func trimAllWhitespace(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		b := s[i]
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			result = append(result, b)
		}
	}
	return string(result)
}

// ApplyProfileResult contains the result of applying a profile.
type ApplyProfileResult struct {
	Applied     bool
	ProfileName string
	Operations  []CleanupOperation
	Changes     int
	Duration    time.Duration
	Timestamp   time.Time
}

// ApplyProfile applies a profile to the configuration.
// Returns an ApplyProfileResult with details about the application.
func (c *Config) ApplyProfile(ctx stdctx.Context, profileName string) *ApplyProfileResult {
	start := time.Now()

	result := &ApplyProfileResult{
		ProfileName: profileName,
		Timestamp:   time.Now(),
	}

	// Check if profile exists
	profile, exists := c.Profiles[profileName]
	if !exists {
		result.Applied = false
		result.Duration = time.Since(start)
		return result
	}

	// Check if profile is enabled
	if !profile.Enabled.IsEnabled() {
		result.Applied = false
		result.Duration = time.Since(start)
		return result
	}

	result.Applied = true
	result.Operations = profile.Operations

	// Update current profile if different
	if c.CurrentProfile != profileName {
		c.CurrentProfile = profileName
		result.Changes++
	}

	result.Duration = time.Since(start)
	return result
}

// ApplyProfileWithContext applies a profile using the generic Context system.
func (c *Config) ApplyProfileWithContext(ctx stdctx.Context, profileName string) *ApplyProfileResult {
	// Create validation config using the generic Context system
	validationConfig := context.ValidationConfig{
		Profile:     profileName,
		Section:     "profile",
		Constraints: make(map[string]string),
		Metadata: map[string]string{
			"operation": "apply_profile",
		},
	}

	// Create the generic context - keep the stdlib context separate
	_ = context.NewContext[context.ValidationConfig](ctx, validationConfig)

	// Apply the profile
	result := c.ApplyProfile(ctx, profileName)

	return result
}

// EstimateImpactResult contains the result of estimating cleanup impact.
type EstimateImpactResult struct {
	TotalSize       SizeEstimate
	OperationsCount int
	HighRiskCount   int
	MediumRiskCount int
	LowRiskCount    int
	ByRiskLevel     map[RiskLevelType]SizeEstimate
	Duration        time.Duration
	Timestamp       time.Time
	Details         []OperationImpactDetail
}

// OperationImpactDetail contains impact details for a single operation.
type OperationImpactDetail struct {
	Operation   string
	RiskLevel   RiskLevelType
	Size        SizeEstimate
	Description string
}

// EstimateImpact estimates the potential impact of cleaning operations in a profile.
// Returns an EstimateImpactResult with size estimates and risk analysis.
func (c *Config) EstimateImpact(ctx stdctx.Context, profileName string) *EstimateImpactResult {
	start := time.Now()

	result := &EstimateImpactResult{
		ByRiskLevel: make(map[RiskLevelType]SizeEstimate),
		Details:     []OperationImpactDetail{},
		Timestamp:   time.Now(),
	}

	// Get profile
	profile, exists := c.Profiles[profileName]
	if !exists {
		result.TotalSize = SizeEstimate{
			Status: SizeEstimateStatusUnknown,
		}
		result.Duration = time.Since(start)
		return result
	}

	result.OperationsCount = len(profile.Operations)

	// Estimate impact for each operation
	for _, op := range profile.Operations {
		detail := OperationImpactDetail{
			Operation: op.Name,
			RiskLevel: op.RiskLevel,
		}

		// Estimate size based on operation type
		size := estimateOperationSize(op)
		detail.Size = size
		detail.Description = getOperationDescription(op)

		result.Details = append(result.Details, detail)

		// Aggregate by risk level
		if existing, ok := result.ByRiskLevel[op.RiskLevel]; ok {
			result.ByRiskLevel[op.RiskLevel] = SizeEstimate{
				Known:  existing.Known + size.Known,
				Status: size.Status,
			}
		} else {
			result.ByRiskLevel[op.RiskLevel] = size
		}

		// Count by risk level
		switch op.RiskLevel {
		case RiskLevelCriticalType:
			result.HighRiskCount++
		case RiskLevelHighType:
			result.HighRiskCount++
		case RiskLevelMediumType:
			result.MediumRiskCount++
		case RiskLevelLowType:
			result.LowRiskCount++
		}
	}

	// Calculate total size
	totalKnown := uint64(0)
	allUnknown := true
	for _, size := range result.ByRiskLevel {
		totalKnown += size.Known
		if size.Status == SizeEstimateStatusKnown {
			allUnknown = false
		}
	}

	status := SizeEstimateStatusKnown
	if allUnknown && totalKnown == 0 {
		status = SizeEstimateStatusUnknown
	}

	result.TotalSize = SizeEstimate{
		Known:  totalKnown,
		Status: status,
	}

	result.Duration = time.Since(start)
	return result
}

// estimateOperationSize estimates the size impact of a cleanup operation.
func estimateOperationSize(op CleanupOperation) SizeEstimate {
	// Base estimates in bytes
	var baseSize uint64

	switch op.RiskLevel {
	case RiskLevelCriticalType:
		baseSize = 10 * 1024 * 1024 // 10MB - potentially large system impact
	case RiskLevelHighType:
		baseSize = 5 * 1024 * 1024 // 5MB - significant but manageable
	case RiskLevelMediumType:
		baseSize = 1 * 1024 * 1024 // 1MB - moderate impact
	case RiskLevelLowType:
		baseSize = 100 * 1024 // 100KB - minimal impact
	default:
		baseSize = 512 * 1024 // 512KB - default moderate
	}

	// Add variation based on enabled status
	if !op.Enabled.IsEnabled() {
		baseSize = 0
	}

	return SizeEstimate{
		Known:  baseSize,
		Status: SizeEstimateStatusUnknown, // Estimates are inherently unknown
	}
}

// getOperationDescription returns a human-readable description of an operation.
func getOperationDescription(op CleanupOperation) string {
	switch op.RiskLevel {
	case RiskLevelCriticalType:
		return "High-risk operation: " + op.Description
	case RiskLevelHighType:
		return "Significant cleanup: " + op.Description
	case RiskLevelMediumType:
		return "Standard cleanup: " + op.Description
	case RiskLevelLowType:
		return "Low-risk cleanup: " + op.Description
	default:
		return op.Description
	}
}

// SanitizeWithContext sanitizes configuration using the generic Context system.
func (c *Config) SanitizeWithContext(ctx stdctx.Context, level ValidationLevelType) *SanitizeConfigResult {
	// Create validation config using the generic Context system
	validationConfig := context.ValidationConfig{
		ValidationLevel: level.String(),
		Profile:         c.CurrentProfile,
		Section:         "config",
		Constraints:     make(map[string]string),
		Metadata: map[string]string{
			"sanitize_operation": "true",
			"context_based":      "true",
		},
	}

	// Create the generic context - keep the stdlib context separate
	_ = context.NewContext[context.ValidationConfig](ctx, validationConfig)

	// Delegate to the standard Sanitize method
	return c.Sanitize(ctx, level)
}
