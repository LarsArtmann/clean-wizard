package security

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// SecurityService provides comprehensive security service
type SecurityService struct {
	validator *SecurityValidator
	middleware *SecurityMiddleware
	config    *SecurityConfig
}

// NewSecurityService creates new security service
func NewSecurityService() *SecurityService {
	// Get default security config
	securityConfig := getDefaultSecurityConfig()
	
	// Override with environment variables if present
	securityConfig = applyEnvironmentOverrides(securityConfig)
	
	// Create validator and middleware
	validator := NewSecurityValidator(securityConfig)
	middleware := NewSecurityMiddleware(securityConfig)
	
	return &SecurityService{
		validator:  validator,
		middleware: middleware,
		config:     securityConfig,
	}
}

// ValidateCommandInput validates command-line input with security checks
func (ss *SecurityService) ValidateCommandInput(
	ctx context.Context,
	command string,
	args map[string]interface{},
	user string,
) (*ValidationResult, map[string]interface{}) {
	
	// Create security context
	secCtx := ss.middleware.CreateSecureContext(
		user, 
		command, 
		ss.config.ValidationLevel,
	)
	
	// Validate and sanitize input
	return ss.middleware.ValidateAndSanitizeInput(secCtx, args)
}

// ValidateConfiguration validates configuration with security checks
func (ss *SecurityService) ValidateConfiguration(
	ctx context.Context,
	configuration interface{},
	user string,
) *ValidationResult {
	
	// Create security context
	secCtx := ss.middleware.CreateSecureContext(
		user,
		"validate_config",
		ss.config.ValidationLevel,
	)
	
	// Validate configuration
	return ss.middleware.ValidateConfiguration(secCtx, configuration)
}

// ValidateCleaningOperation validates cleaning operation with security checks
func (ss *SecurityService) ValidateCleaningOperation(
	ctx context.Context,
	operation string,
	settings map[string]interface{},
	user string,
) (*ValidationResult, map[string]interface{}) {
	
	// Create security context
	secCtx := ss.middleware.CreateSecureContext(
		user,
		operation,
		ss.config.ValidationLevel,
	)
	
	// Validate operation
	validationResult := ss.middleware.ValidateOperation(secCtx, operation, settings)
	
	// Return validation result and sanitized settings
	return validationResult, settings
}

// ValidateAndSecurePath validates and secures file paths
func (ss *SecurityService) ValidateAndSecurePath(
	ctx context.Context,
	path string,
	baseDir string,
	user string,
) (string, *ValidationResult) {
	
	// Create security context
	secCtx := ss.middleware.CreateSecureContext(
		user,
		"validate_path",
		ss.config.ValidationLevel,
	)
	
	// Validate and secure path
	return ss.middleware.ValidatePathWithSecurity(secCtx, path, baseDir)
}

// ValidateUserInput validates general user input with security checks
func (ss *SecurityService) ValidateUserInput(
	ctx context.Context,
	field, value, validationTag string,
	user string,
) *ValidationResult {
	
	// Validate input
	return ss.validator.ValidateInput(field, value, validationTag)
}

// SanitizeInput sanitizes user input
func (ss *SecurityService) SanitizeInput(input string) string {
	return ss.validator.SanitizeInput(input)
}

// GetSecurityReport returns comprehensive security report
func (ss *SecurityService) GetSecurityReport() *SecurityReport {
	metrics := ss.middleware.GetSecurityMetrics()
	auditLog := ss.middleware.GetAuditLog()
	
	return &SecurityReport{
		GeneratedAt:    time.Now(),
		ValidationLevel: ss.config.ValidationLevel,
		Config:         ss.config,
		Metrics:        metrics,
		AuditLog:       auditLog,
		Summary:        ss.generateSecuritySummary(metrics, auditLog),
	}
}

// IsSecurityEnabled checks if security validation is enabled
func (ss *SecurityService) IsSecurityEnabled() bool {
	return ss.config.ValidationLevel != shared.ValidationLevelNoneType
}

// IsStrictMode checks if strict security mode is enabled
func (ss *SecurityService) IsStrictMode() bool {
	return ss.config.ValidationLevel == shared.ValidationLevelStrictType
}

// GetValidationLevel returns current validation level
func (ss *SecurityService) GetValidationLevel() shared.ValidationLevelType {
	return ss.config.ValidationLevel
}

// UpdateSecurityConfig updates security configuration
func (ss *SecurityService) UpdateSecurityConfig(newConfig *SecurityConfig) error {
	// Validate new configuration
	if err := ss.validateSecurityConfig(newConfig); err != nil {
		return fmt.Errorf("invalid security config: %w", err)
	}
	
	// Update configuration
	ss.config = newConfig
	ss.validator = NewSecurityValidator(newConfig)
	ss.middleware = NewSecurityMiddleware(newConfig)
	
	return nil
}

// SecureWorkingDirectory ensures working directory is secure
func (ss *SecurityService) SecureWorkingDirectory() error {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// Validate working directory
	validationResult := ss.validator.ValidatePath(wd)
	if !validationResult.IsValid {
		return fmt.Errorf("working directory is not secure: %v", validationResult.Errors)
	}
	
	// Check if working directory is in user home
	userHome, err := os.UserHomeDir()
	if err == nil {
		if !strings.HasPrefix(wd, userHome) {
			return fmt.Errorf("working directory should be under user home directory")
		}
	}
	
	return nil
}

// ValidateOperationPermissions validates user permissions for operations
func (ss *SecurityService) ValidateOperationPermissions(
	ctx context.Context,
	operation string,
	user string,
) *ValidationResult {
	
	// Check operation permissions
	if !ss.hasOperationPermission(operation, user) {
		return &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{
					Field:   "operation",
					Value:   operation,
					Message: "User does not have permission for this operation",
					Code:    "PERMISSION_DENIED",
				},
			},
			Level: ss.config.ValidationLevel,
		}
	}
	
	return &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
		Level:   ss.config.ValidationLevel,
	}
}

// CreateSecureTempDir creates secure temporary directory
func (ss *SecurityService) CreateSecureTempDir(baseDir, pattern string) (string, error) {
	// Validate base directory
	validationResult := ss.validator.ValidatePath(baseDir)
	if !validationResult.IsValid {
		return "", fmt.Errorf("base directory is not secure: %v", validationResult.Errors)
	}
	
	// Create secure temporary directory
	tempDir, err := os.MkdirTemp(baseDir, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Validate created directory
	validationResult = ss.validator.ValidatePath(tempDir)
	if !validationResult.IsValid {
		// Clean up on validation failure
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("created temp directory is not secure: %v", validationResult.Errors)
	}
	
	return tempDir, nil
}

// SecureJoinPaths securely joins multiple path components
func (ss *SecurityService) SecureJoinPaths(base string, components ...string) (string, error) {
	// Start with base path
	securePath := base
	
	// Join each component using middleware
	for _, component := range components {
		secCtx := ss.middleware.CreateSecureContext(
			"system",
			"validate_path",
			ss.config.ValidationLevel,
		)
		
		var validationResult *ValidationResult
		securePath, validationResult = ss.middleware.ValidatePathWithSecurity(secCtx, securePath, component)
		if !validationResult.IsValid {
			return "", fmt.Errorf("failed to join path components: %v", validationResult.Errors)
		}
	}
	
	// Validate final path
	validationResult := ss.validator.ValidatePath(securePath)
	if !validationResult.IsValid {
		return "", fmt.Errorf("joined path is not secure: %v", validationResult.Errors)
	}
	
	return securePath, nil
}

// GetSecurityStatistics returns security operation statistics
func (ss *SecurityService) GetSecurityStatistics() map[string]interface{} {
	metrics := ss.middleware.GetSecurityMetrics()
	
	statistics := map[string]interface{}{
		"total_validations":    metrics.TotalValidations,
		"total_failures":       metrics.TotalFailures,
		"total_sanitizations":  metrics.TotalSanitizations,
		"total_blocks":         metrics.TotalBlocks,
		"last_validation":      metrics.LastValidation,
		"most_blocked_field":   metrics.MostBlockedField,
		"validation_level":     ss.config.ValidationLevel.String(),
		"strict_path_checking": ss.config.StrictPathChecking,
		"max_path_length":      ss.config.MaxPathLength,
	}
	
	// Calculate success rate
	if metrics.TotalValidations > 0 {
		successRate := float64(metrics.TotalValidations-metrics.TotalFailures) / float64(metrics.TotalValidations) * 100
		statistics["success_rate"] = successRate
	}
	
	// Calculate block rate
	if metrics.TotalValidations > 0 {
		blockRate := float64(metrics.TotalBlocks) / float64(metrics.TotalValidations) * 100
		statistics["block_rate"] = blockRate
	}
	
	return statistics
}

// SecurityReport represents comprehensive security report
type SecurityReport struct {
	GeneratedAt    time.Time         `json:"generated_at"`
	ValidationLevel shared.ValidationLevelType `json:"validation_level"`
	Config         *SecurityConfig   `json:"config"`
	Metrics        SecurityMetrics   `json:"metrics"`
	AuditLog       []AuditEntry     `json:"audit_log"`
	Summary        SecuritySummary   `json:"summary"`
}

// SecuritySummary represents security summary
type SecuritySummary struct {
	OverallSecurity     string             `json:"overall_security"`
	Recommendations    []string           `json:"recommendations"`
	SecurityScore      float64            `json:"security_score"`
	CriticalFindings   int                `json:"critical_findings"`
	HighRiskFindings   int                `json:"high_risk_findings"`
	MediumRiskFindings int                `json:"medium_risk_findings"`
}

// Helper functions

// getDefaultSecurityConfig returns default security configuration
func getDefaultSecurityConfig() *SecurityConfig {
	config := GetDefaultSecurityConfig()
	
	// Override with application-specific settings
	config.AllowedFileExtensions = append(config.AllowedFileExtensions,
		".bak", ".old", ".orig", ".tmp",
	)
	
	config.BlockedPatterns = append(config.BlockedPatterns,
		`\.\.passwd`, `\.\.shadow`, `\.\.key`,
	)
	
	config.MaxPathLength = 8192 // 8KB max path length
	config.ValidationLevel = shared.ValidationLevelComprehensiveType
	config.StrictPathChecking = true
	
	return config
}

// applyEnvironmentOverrides applies environment variable overrides
func applyEnvironmentOverrides(config *SecurityConfig) *SecurityConfig {
	// Override validation level
	if level := os.Getenv("CLEAN_WIZARD_VALIDATION_LEVEL"); level != "" {
		if parsedLevel, err := shared.ParseValidationLevel(level); err == nil {
			config.ValidationLevel = parsedLevel
		}
	}
	
	// Override strict path checking
	if strict := os.Getenv("CLEAN_WIZARD_STRICT_PATH"); strict != "" {
		if strict == "true" || strict == "1" {
			config.StrictPathChecking = true
		} else {
			config.StrictPathChecking = false
		}
	}
	
	// Override max path length
	if maxPath := os.Getenv("CLEAN_WIZARD_MAX_PATH"); maxPath != "" {
		if parsed, err := strconv.Atoi(maxPath); err == nil && parsed > 0 {
			config.MaxPathLength = parsed
		}
	}
	
	return config
}

// validateSecurityConfig validates security configuration
func (ss *SecurityService) validateSecurityConfig(config *SecurityConfig) error {
	// Validate validation level
	if !config.ValidationLevel.IsValid() {
		return fmt.Errorf("invalid validation level: %v", config.ValidationLevel)
	}
	
	// Validate max path length
	if config.MaxPathLength <= 0 || config.MaxPathLength > 65536 {
		return fmt.Errorf("max path length must be between 1 and 65536")
	}
	
	// Validate blocked patterns
	for _, pattern := range config.BlockedPatterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid blocked pattern: %s", pattern)
		}
	}
	
	return nil
}

// hasOperationPermission checks if user has permission for operation
func (ss *SecurityService) hasOperationPermission(operation, user string) bool {
	// For now, all operations are allowed for all users
	// This could be extended with role-based access control
	return true
}

// generateSecuritySummary generates security summary
func (ss *SecurityService) generateSecuritySummary(metrics SecurityMetrics, auditLog []AuditEntry) SecuritySummary {
	summary := SecuritySummary{
		Recommendations:    []string{},
		CriticalFindings:   0,
		HighRiskFindings:   0,
		MediumRiskFindings: 0,
	}
	
	// Analyze audit log for findings
	for _, entry := range auditLog {
		for _, error := range entry.Validation.Errors {
			switch error.Code {
			case "PATH_TRAVERSAL", "XSS_PATTERN", "SQL_INJECTION":
				summary.CriticalFindings++
				summary.Recommendations = append(summary.Recommendations,
					fmt.Sprintf("Critical security issue detected: %s", error.Message))
			case "BLOCKED_PATTERN", "INVALID_OPERATION":
				summary.HighRiskFindings++
				summary.Recommendations = append(summary.Recommendations,
					fmt.Sprintf("High risk issue detected: %s", error.Message))
			default:
				summary.MediumRiskFindings++
			}
		}
	}
	
	// Calculate security score
	summary.SecurityScore = ss.calculateSecurityScore(metrics, summary)
	
	// Determine overall security
	if summary.SecurityScore >= 90 {
		summary.OverallSecurity = "EXCELLENT"
	} else if summary.SecurityScore >= 80 {
		summary.OverallSecurity = "GOOD"
	} else if summary.SecurityScore >= 70 {
		summary.OverallSecurity = "FAIR"
	} else {
		summary.OverallSecurity = "POOR"
	}
	
	// Add general recommendations
	if summary.CriticalFindings > 0 {
		summary.Recommendations = append(summary.Recommendations,
			"Enable strict validation mode for enhanced security")
	}
	
	if summary.SecurityScore < 80 {
		summary.Recommendations = append(summary.Recommendations,
			"Review and update security configuration")
	}
	
	return summary
}

// calculateSecurityScore calculates security score
func (ss *SecurityService) calculateSecurityScore(metrics SecurityMetrics, summary SecuritySummary) float64 {
	// Base score from validation level
	levelScores := map[shared.ValidationLevelType]float64{
		shared.ValidationLevelNoneType:           0,
		shared.ValidationLevelBasicType:          60,
		shared.ValidationLevelComprehensiveType: 80,
		shared.ValidationLevelStrictType:         95,
	}
	
	score := levelScores[ss.config.ValidationLevel]
	
	// Penalize for security issues
	if summary.CriticalFindings > 0 {
		score -= 20 * float64(summary.CriticalFindings)
	}
	
	if summary.HighRiskFindings > 0 {
		score -= 10 * float64(summary.HighRiskFindings)
	}
	
	if summary.MediumRiskFindings > 0 {
		score -= 5 * float64(summary.MediumRiskFindings)
	}
	
	// Bonus for strict path checking
	if ss.config.StrictPathChecking {
		score += 5
	}
	
	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}