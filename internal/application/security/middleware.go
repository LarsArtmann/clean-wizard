package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// SecurityMiddleware provides security validation for all operations
type SecurityMiddleware struct {
	validator *SecurityValidator
	auditLog  []AuditEntry
}

// AuditEntry represents security audit log entry
type AuditEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Operation  string                 `json:"operation"`
	Input      map[string]interface{} `json:"input"`
	Validation *ValidationResult      `json:"validation"`
	User       string                 `json:"user"`
	Context    string                 `json:"context"`
	Success    bool                   `json:"success"`
}

// SecurityContext provides security context for operations
type SecurityContext struct {
	User            string                     `json:"user"`
	Operation       string                     `json:"operation"`
	ValidationLevel shared.ValidationLevelType `json:"validation_level"`
	StrictMode      bool                       `json:"strict_mode"`
	StartTime       time.Time                  `json:"start_time"`
	RequestID       string                     `json:"request_id"`
}

// NewSecurityMiddleware creates new security middleware
func NewSecurityMiddleware(securityConfig *SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		validator: NewSecurityValidator(securityConfig),
		auditLog:  []AuditEntry{},
	}
}

// ValidateAndSanitizeInput validates and sanitizes input with comprehensive security checks
func (sm *SecurityMiddleware) ValidateAndSanitizeInput(
	ctx *SecurityContext,
	inputs map[string]interface{},
) (*ValidationResult, map[string]interface{}) {
	// Create validation result
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   ctx.ValidationLevel,
	}

	// Create sanitized inputs map
	sanitized := make(map[string]interface{})

	// Validate and sanitize each input
	for field, value := range inputs {
		// Convert value to string for validation
		valueStr := fmt.Sprintf("%v", value)

		// Validate input
		validationResult := sm.validator.ValidateInput(field, valueStr, "required")
		if !validationResult.IsValid {
			result.IsValid = false
			result.Errors = append(result.Errors, validationResult.Errors...)
		}

		// Sanitize input
		sanitizedValue := sm.validator.SanitizeInput(valueStr)
		sanitized[field] = sanitizedValue
	}

	// Add audit entry
	sm.addAuditEntry(ctx, inputs, result, result.IsValid)

	return result, sanitized
}

// ValidateOperation validates operation parameters with security checks
func (sm *SecurityMiddleware) ValidateOperation(
	ctx *SecurityContext,
	operation string,
	settings map[string]interface{},
) *ValidationResult {
	// Validate operation
	result := sm.validator.ValidateOperation(operation, settings)

	// Add additional context validation in strict mode
	if ctx.StrictMode {
		sm.performStrictValidation(ctx, settings, result)
	}

	// Add audit entry
	auditData := map[string]interface{}{
		"operation": operation,
		"settings":  settings,
	}
	sm.addAuditEntry(ctx, auditData, result, result.IsValid)

	return result
}

// ValidateConfiguration validates configuration with security checks
func (sm *SecurityMiddleware) ValidateConfiguration(
	ctx *SecurityContext,
	configuration interface{},
) *ValidationResult {
	// Validate overall configuration
	result := sm.validator.ValidateConfig(configuration)

	// Add audit entry
	auditData := map[string]interface{}{
		"validation_type": "configuration",
	}
	sm.addAuditEntry(ctx, auditData, result, result.IsValid)

	return result
}

// ValidatePathWithSecurity validates file paths with security checks
func (sm *SecurityMiddleware) ValidatePathWithSecurity(
	ctx *SecurityContext,
	path string,
	baseDir string,
) (string, *ValidationResult) {
	// Validate original path
	result := sm.validator.ValidatePath(path)
	if !result.IsValid {
		return "", result
	}

	// Sanitize path
	sanitizedPath := sm.validator.SanitizeInput(path)

	// Secure join with base directory
	securePath, pathResult := sm.validator.SecureJoinPath(baseDir, sanitizedPath)
	if !pathResult.IsValid {
		return "", pathResult
	}

	// Final validation
	finalResult := sm.validator.ValidatePath(securePath)
	if !finalResult.IsValid {
		return "", finalResult
	}

	// Add audit entry
	auditData := map[string]interface{}{
		"original_path": path,
		"secure_path":   securePath,
		"base_dir":      baseDir,
	}
	sm.addAuditEntry(ctx, auditData, finalResult, finalResult.IsValid)

	return securePath, finalResult
}

// CreateSecureContext creates secure context for operations
func (sm *SecurityMiddleware) CreateSecureContext(
	user, operation string,
	validationLevel shared.ValidationLevelType,
) *SecurityContext {
	return &SecurityContext{
		User:            user,
		Operation:       operation,
		ValidationLevel: validationLevel,
		StrictMode:      validationLevel == shared.ValidationLevelStrictType,
		StartTime:       time.Now(),
		RequestID:       fmt.Sprintf("req_%d", time.Now().UnixNano()),
	}
}

// GetAuditLog returns security audit log
func (sm *SecurityMiddleware) GetAuditLog() []AuditEntry {
	return sm.auditLog
}

// GetAuditLogForTimeRange returns audit log entries for time range
func (sm *SecurityMiddleware) GetAuditLogForTimeRange(
	start, end time.Time,
) []AuditEntry {
	var filtered []AuditEntry
	for _, entry := range sm.auditLog {
		if entry.Timestamp.After(start) && entry.Timestamp.Before(end) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// ClearAuditLog clears security audit log
func (sm *SecurityMiddleware) ClearAuditLog() {
	sm.auditLog = []AuditEntry{}
}

// GetSecurityMetrics returns security metrics
func (sm *SecurityMiddleware) GetSecurityMetrics() SecurityMetrics {
	metrics := SecurityMetrics{
		TotalValidations:   0,
		TotalFailures:      0,
		TotalSanitizations: 0,
		TotalBlocks:        0,
		LastValidation:     time.Time{},
		MostBlockedField:   "",
	}

	for _, entry := range sm.auditLog {
		metrics.TotalValidations++

		if !entry.Validation.IsValid {
			metrics.TotalFailures++
		}

		if entry.Timestamp.After(metrics.LastValidation) {
			metrics.LastValidation = entry.Timestamp
		}

		// Count blocked operations
		for _, error := range entry.Validation.Errors {
			if error.Code == "PATH_TRAVERSAL" ||
				error.Code == "BLOCKED_PATTERN" ||
				error.Code == "XSS_PATTERN" ||
				error.Code == "SQL_INJECTION" {
				metrics.TotalBlocks++

				if metrics.MostBlockedField == "" ||
					len(error.Field) > len(metrics.MostBlockedField) {
					metrics.MostBlockedField = error.Field
				}
			}
		}
	}

	return metrics
}

// SecurityMetrics represents security operation metrics
type SecurityMetrics struct {
	TotalValidations   int       `json:"total_validations"`
	TotalFailures      int       `json:"total_failures"`
	TotalSanitizations int       `json:"total_sanitizations"`
	TotalBlocks        int       `json:"total_blocks"`
	LastValidation     time.Time `json:"last_validation"`
	MostBlockedField   string    `json:"most_blocked_field"`
}

// performStrictValidation performs additional validation in strict mode
func (sm *SecurityMiddleware) performStrictValidation(
	ctx *SecurityContext,
	settings map[string]interface{},
	result *ValidationResult,
) {
	// Check for suspicious patterns in settings
	for key, value := range settings {
		valueStr := fmt.Sprintf("%v", value)

		// Check for suspicious patterns
		if sm.hasSuspiciousPatterns(valueStr) {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   key,
				Value:   valueStr,
				Message: "Suspicious pattern detected",
				Code:    "SUSPICIOUS_PATTERN",
			})
		}

		// Check for excessive values
		if sm.hasExcessiveValues(key, valueStr) {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   key,
				Value:   valueStr,
				Message: "Excessive value detected",
				Code:    "EXCESSIVE_VALUE",
			})
		}
	}
}

// hasSuspiciousPatterns checks for suspicious patterns
func (sm *SecurityMiddleware) hasSuspiciousPatterns(value string) bool {
	suspiciousPatterns := []string{
		"${jndi:",    // JNDI injection
		"<%@=",       // Code injection
		"eval(",      // Code execution
		"system(",    // System command execution
		"exec(",      // Command execution
		"shell_exec", // Shell execution
		"cmd.exe",    // Windows command
		"/bin/sh",    // Unix shell
		"\\x\\.",     // Hex encoded patterns
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}

	return false
}

// hasExcessiveValues checks for excessive values
func (sm *SecurityMiddleware) hasExcessiveValues(key, value string) bool {
	// Check for excessive length
	if len(value) > 10000 { // 10KB limit
		return true
	}

	// Check for excessive numbers
	if strings.Contains(key, "count") || strings.Contains(key, "limit") {
		var num int
		_, err := fmt.Sscanf(value, "%d", &num)
		if err == nil && num > 1000000 { // 1M limit
			return true
		}
	}

	// Check for excessive sizes
	if strings.Contains(key, "size") || strings.Contains(key, "bytes") {
		var num int64
		_, err := fmt.Sscanf(value, "%d", &num)
		if err == nil && num > 1024*1024*1024 { // 1GB limit
			return true
		}
	}

	return false
}

// addAuditEntry adds entry to audit log
func (sm *SecurityMiddleware) addAuditEntry(
	ctx *SecurityContext,
	input map[string]interface{},
	validation *ValidationResult,
	success bool,
) {
	entry := AuditEntry{
		Timestamp:  time.Now(),
		Operation:  ctx.Operation,
		Input:      input,
		Validation: validation,
		User:       ctx.User,
		Context:    ctx.RequestID,
		Success:    success,
	}

	// Add to audit log
	sm.auditLog = append(sm.auditLog, entry)

	// Limit audit log size to prevent memory issues
	if len(sm.auditLog) > 10000 {
		sm.auditLog = sm.auditLog[5000:] // Keep last 5000 entries
	}
}
