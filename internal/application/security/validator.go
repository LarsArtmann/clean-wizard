package security

import (
	"fmt"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationResult represents validation operation result
type ValidationResult struct {
	IsValid bool                       `json:"is_valid"`
	Errors  []ValidationError          `json:"errors"`
	Level   shared.ValidationLevelType `json:"level"`
}

// SecurityValidator provides comprehensive input validation and security
type SecurityValidator struct {
	validator *validator.Validate
	config    *SecurityConfig
}

// SecurityConfig defines security validation configuration
type SecurityConfig struct {
	MaxPathLength         int                        `json:"max_path_length"`
	AllowedFileExtensions []string                   `json:"allowed_file_extensions"`
	BlockedPatterns       []string                   `json:"blocked_patterns"`
	MaxUsernameLength     int                        `json:"max_username_length"`
	ValidationLevel       shared.ValidationLevelType `json:"validation_level"`
	StrictPathChecking    bool                       `json:"strict_path_checking"`
}

// NewSecurityValidator creates new security validator
func NewSecurityValidator(config *SecurityConfig) *SecurityValidator {
	v := validator.New()

	// Register custom validations
	v.RegisterValidation("safe-path", validateSafePath)
	v.RegisterValidation("clean-operation", validateCleanOperation)
	v.RegisterValidation("profile-name", validateProfileName)

	return &SecurityValidator{
		validator: v,
		config:    config,
	}
}

// ValidateInput validates user input with comprehensive security checks
func (sv *SecurityValidator) ValidateInput(field, value, validationTag string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Perform validation based on tag
	err := sv.validator.Var(value, validationTag)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: err.Error(),
			Code:    "VALIDATION_FAILED",
		})
	}

	// Additional security checks
	sv.performSecurityChecks(field, value, result)

	return result
}

// ValidateConfig validates configuration with security checks
func (sv *SecurityValidator) ValidateConfig(config interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Perform comprehensive validation
	err := sv.validator.Struct(config)
	if err != nil {
		result.IsValid = false
		for _, fieldErr := range err.(validator.ValidationErrors) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldErr.Field(),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
				Message: fieldErr.Tag(),
				Code:    fieldErr.Error(),
			})
		}
	}

	return result
}

// ValidatePath validates file path with security checks
func (sv *SecurityValidator) ValidatePath(path string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Path length check
	if len(path) > sv.config.MaxPathLength {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "path",
			Value:   path,
			Message: "Path too long",
			Code:    "PATH_TOO_LONG",
		})
	}

	// Path traversal protection
	if sv.config.StrictPathChecking {
		cleanPath := filepath.Clean(path)
		if cleanPath != path {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "path",
				Value:   path,
				Message: "Path contains traversal attempts",
				Code:    "PATH_TRAVERSAL",
			})
		}
	}

	// Secure path joining
	if strings.Contains(path, "..") {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "path",
			Value:   path,
			Message: "Path contains parent directory traversal",
			Code:    "PARENT_TRAVERSAL",
		})
	}

	// Blocked patterns check
	for _, pattern := range sv.config.BlockedPatterns {
		if matched, _ := regexp.MatchString(pattern, path); matched {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "path",
				Value:   path,
				Message: "Path matches blocked pattern",
				Code:    "BLOCKED_PATTERN",
			})
			break
		}
	}

	return result
}

// ValidateOperation validates cleaning operation parameters
func (sv *SecurityValidator) ValidateOperation(operation string, settings map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Validate operation name
	if !sv.isValidOperation(operation) {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "operation",
			Value:   operation,
			Message: "Invalid operation name",
			Code:    "INVALID_OPERATION",
		})
	}

	// Validate settings based on operation level
	switch operation {
	case "nix-generations":
		sv.validateNixOperation(settings, result)
	case "homebrew":
		sv.validateHomebrewOperation(settings, result)
	case "npm-cache":
		sv.validateNpmOperation(settings, result)
	case "pnpm-store":
		sv.validatePnpmOperation(settings, result)
	case "temp-files":
		sv.validateTempFilesOperation(settings, result)
	default:
		// Unknown operation - higher security level
		if sv.config.ValidationLevel == shared.ValidationLevelStrictType {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "operation",
				Value:   operation,
				Message: "Unknown operation not allowed",
				Code:    "UNKNOWN_OPERATION",
			})
		}
	}

	return result
}

// SanitizeInput sanitizes user input
func (sv *SecurityValidator) SanitizeInput(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	sanitized = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	// Normalize whitespace
	sanitized = strings.Join(strings.Fields(sanitized), " ")

	return sanitized
}

// SecureJoinPath securely joins path components
func (sv *SecurityValidator) SecureJoinPath(base, path string) (string, *ValidationResult) {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Use secure join library
	securePath, err := securejoin.SecureJoin(base, path)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "path",
			Value:   path,
			Message: "Path joining failed",
			Code:    "SECURE_JOIN_FAILED",
		})
		return securePath, result
	}

	// Validate the joined path
	pathResult := sv.ValidatePath(securePath)
	if !pathResult.IsValid {
		return "", pathResult
	}

	return securePath, result
}

// ValidateEmail validates email address with security checks
func (sv *SecurityValidator) ValidateEmail(email string) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Basic email validation
	if _, err := mail.ParseAddress(email); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "email",
			Value:   email,
			Message: "Invalid email format",
			Code:    "INVALID_EMAIL",
		})
	}

	// Additional security checks
	if len(email) > sv.config.MaxUsernameLength {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "email",
			Value:   email,
			Message: "Email too long",
			Code:    "EMAIL_TOO_LONG",
		})
	}

	return result
}

// ValidateTimestamp validates timestamp with reasonable bounds
func (sv *SecurityValidator) ValidateTimestamp(timestamp time.Time) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   sv.config.ValidationLevel,
	}

	// Skip validation if level is NONE
	if sv.config.ValidationLevel == shared.ValidationLevelNoneType {
		return result
	}

	// Check reasonable time bounds
	now := time.Now()
	minTime := now.Add(-365 * 24 * time.Hour) // 1 year ago
	maxTime := now.Add(365 * 24 * time.Hour)  // 1 year in future

	if timestamp.Before(minTime) || timestamp.After(maxTime) {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "timestamp",
			Value:   timestamp.String(),
			Message: "Timestamp outside reasonable bounds",
			Code:    "INVALID_TIMESTAMP",
		})
	}

	return result
}

// performSecurityChecks performs additional security checks
func (sv *SecurityValidator) performSecurityChecks(field, value string, result *ValidationResult) {
	// XSS protection
	if sv.hasXSSPatterns(value) {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: "Input contains XSS patterns",
			Code:    "XSS_PATTERN",
		})
	}

	// SQL injection protection
	if sv.hasSQLInjectionPatterns(value) {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: "Input contains SQL injection patterns",
			Code:    "SQL_INJECTION",
		})
	}
}

// hasXSSPatterns checks for XSS patterns
func (sv *SecurityValidator) hasXSSPatterns(input string) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"alert(",
		"document.cookie",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// hasSQLInjectionPatterns checks for SQL injection patterns
func (sv *SecurityValidator) hasSQLInjectionPatterns(input string) bool {
	sqlPatterns := []string{
		"'",
		"\"",
		";",
		"--",
		"/*",
		"*/",
		"xp_",
		"union",
		"select",
		"insert",
		"update",
		"delete",
		"drop",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// isValidOperation checks if operation is valid
func (sv *SecurityValidator) isValidOperation(operation string) bool {
	validOperations := []string{
		"nix-generations",
		"homebrew",
		"npm-cache",
		"pnpm-store",
		"temp-files",
	}

	for _, validOp := range validOperations {
		if operation == validOp {
			return true
		}
	}

	return false
}

// validateNixOperation validates Nix operation settings
func (sv *SecurityValidator) validateNixOperation(settings map[string]interface{}, result *ValidationResult) {
	if generations, ok := settings["generations"].(int); ok {
		if generations < 1 || generations > 100 {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "generations",
				Value:   fmt.Sprintf("%d", generations),
				Message: "Generations must be between 1 and 100",
				Code:    "INVALID_GENERATIONS",
			})
		}
	}
}

// validateHomebrewOperation validates Homebrew operation settings
func (sv *SecurityValidator) validateHomebrewOperation(settings map[string]interface{}, result *ValidationResult) {
	// Homebrew validation logic would go here
	// For now, assume all settings are valid
}

// validateNpmOperation validates npm operation settings
func (sv *SecurityValidator) validateNpmOperation(settings map[string]interface{}, result *ValidationResult) {
	// npm validation logic would go here
	// For now, assume all settings are valid
}

// validatePnpmOperation validates pnpm operation settings
func (sv *SecurityValidator) validatePnpmOperation(settings map[string]interface{}, result *ValidationResult) {
	// pnpm validation logic would go here
	// For now, assume all settings are valid
}

// validateTempFilesOperation validates temp files operation settings
func (sv *SecurityValidator) validateTempFilesOperation(settings map[string]interface{}, result *ValidationResult) {
	if olderThan, ok := settings["olderThan"].(string); ok {
		_, err := time.ParseDuration(olderThan)
		if err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "olderThan",
				Value:   olderThan,
				Message: "Invalid duration format",
				Code:    "INVALID_DURATION",
			})
		}
	}
}

// Custom validation functions

// validateSafePath validates safe path
func validateSafePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return !strings.Contains(path, "..") && !strings.Contains(path, "\x00")
}

// validateCleanOperation validates clean operation name
func validateCleanOperation(fl validator.FieldLevel) bool {
	operation := fl.Field().String()
	validOps := []string{
		"nix-generations",
		"homebrew",
		"npm-cache",
		"pnpm-store",
		"temp-files",
	}

	for _, validOp := range validOps {
		if operation == validOp {
			return true
		}
	}

	return false
}

// validateProfileName validates profile name
func validateProfileName(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	// Check for valid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name); !matched {
		return false
	}

	// Check length
	return len(name) >= 1 && len(name) <= 50
}

// GetDefaultSecurityConfig returns default security configuration
func GetDefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxPathLength:         4096,
		AllowedFileExtensions: []string{"", ".log", ".tmp", ".cache"},
		BlockedPatterns:       []string{`\.\.`, `\.\.\.`, `[<>]`},
		MaxUsernameLength:     255,
		ValidationLevel:       shared.ValidationLevelComprehensiveType,
		StrictPathChecking:    true,
	}
}
