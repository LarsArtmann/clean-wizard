package config

import (
	"fmt"
	"time"
	
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TypedContext provides type-safe alternative to map[string]any for validation contexts
type TypedContext struct {
	Field       string        `json:"field"`
	Operation   string        `json:"operation"`
	Timestamp   time.Time     `json:"timestamp"`
	UserID      *string       `json:"user_id,omitempty"`
	Environment string        `json:"environment"`
	Metadata    ContextValues `json:"metadata,omitempty"`
}

// ContextValues represents typed context values
type ContextValues struct {
	ProfileName  *string              `json:"profile_name,omitempty"`
	RiskLevel    *domain.RiskLevelType `json:"risk_level,omitempty"`
	CleanerType  *domain.CleanType     `json:"cleaner_type,omitempty"`
	SettingType  *string               `json:"setting_type,omitempty"`
	Original     any                   `json:"original,omitempty"`
	Requested    any                   `json:"requested,omitempty"`
	Validation   *ValidationMetadata    `json:"validation,omitempty"`
}

// ValidationMetadata contains validation-specific context
type ValidationMetadata struct {
	Level     string    `json:"level"`
	Rule      string    `json:"rule"`
	CheckTime time.Time `json:"check_time"`
}

// NewTypedContext creates a new typed context with required fields
func NewTypedContext(field, operation string) *TypedContext {
	return &TypedContext{
		Field:       field,
		Operation:   operation,
		Timestamp:   time.Now(),
		Environment: "development",
		Metadata:    ContextValues{},
	}
}

// WithEnvironment sets the environment in the context
func (tc *TypedContext) WithEnvironment(env string) *TypedContext {
	tc.Environment = env
	return tc
}

// WithUser sets the user ID in the context
func (tc *TypedContext) WithUser(userID string) *TypedContext {
	tc.UserID = &userID
	return tc
}

// WithProfile sets the profile name in context
func (tc *TypedContext) WithProfile(profileName string) *TypedContext {
	tc.Metadata.ProfileName = &profileName
	return tc
}

// WithRiskLevel sets the risk level in context
func (tc *TypedContext) WithRiskLevel(risk domain.RiskLevelType) *TypedContext {
	tc.Metadata.RiskLevel = &risk
	return tc
}

// WithCleanerType sets the cleaner type in context
func (tc *TypedContext) WithCleanerType(cleanerType domain.CleanType) *TypedContext {
	tc.Metadata.CleanerType = &cleanerType
	return tc
}

// WithOriginal sets the original value in context
func (tc *TypedContext) WithOriginal(original any) *TypedContext {
	tc.Metadata.Original = original
	return tc
}

// WithRequested sets the requested value in context
func (tc *TypedContext) WithRequested(requested any) *TypedContext {
	tc.Metadata.Requested = requested
	return tc
}

// WithValidation adds validation metadata to context
func (tc *TypedContext) WithValidation(level, rule string) *TypedContext {
	tc.Metadata.Validation = &ValidationMetadata{
		Level:     level,
		Rule:      rule,
		CheckTime: time.Now(),
	}
	return tc
}

// ToMap converts typed context to map[string]any for legacy compatibility
func (tc *TypedContext) ToMap() map[string]any {
	result := map[string]any{
		"field":       tc.Field,
		"operation":   tc.Operation,
		"timestamp":   tc.Timestamp,
		"environment": tc.Environment,
	}
	
	if tc.UserID != nil {
		result["user_id"] = *tc.UserID
	}
	
	if tc.Metadata.ProfileName != nil {
		result["profile_name"] = *tc.Metadata.ProfileName
	}
	
	if tc.Metadata.RiskLevel != nil {
		result["risk_level"] = tc.Metadata.RiskLevel.String()
	}
	
	if tc.Metadata.CleanerType != nil {
		result["cleaner_type"] = string(*tc.Metadata.CleanerType)
	}
	
	if tc.Metadata.Original != nil {
		result["original"] = tc.Metadata.Original
	}
	
	if tc.Metadata.Requested != nil {
		result["requested"] = tc.Metadata.Requested
	}
	
	if tc.Metadata.Validation != nil {
		result["validation"] = map[string]any{
			"level":      tc.Metadata.Validation.Level,
			"rule":       tc.Metadata.Validation.Rule,
			"check_time": tc.Metadata.Validation.CheckTime,
		}
	}
	
	return result
}

// FromMap creates typed context from map[string]any for legacy compatibility
func FromMap(m map[string]any) (*TypedContext, error) {
	tc := &TypedContext{
		Timestamp:   time.Now(),
		Environment: "development",
		Metadata:    ContextValues{},
	}
	
	if field, ok := m["field"].(string); ok {
		tc.Field = field
	}
	
	if operation, ok := m["operation"].(string); ok {
		tc.Operation = operation
	}
	
	if env, ok := m["environment"].(string); ok {
		tc.Environment = env
	}
	
	if userID, ok := m["user_id"].(string); ok {
		tc.UserID = &userID
	}
	
	if profileName, ok := m["profile_name"].(string); ok {
		tc.Metadata.ProfileName = &profileName
	}
	
	if riskLevel, ok := m["risk_level"].(string); ok {
		// Convert string back to RiskLevelType (simplified for example)
		tc.Metadata.RiskLevel = nil // Would need proper string->enum conversion
	}
	
	return tc, nil
}

// String provides string representation of typed context
func (tc *TypedContext) String() string {
	return fmt.Sprintf("TypedContext{field=%s, operation=%s, env=%s}", tc.Field, tc.Operation, tc.Environment)
}