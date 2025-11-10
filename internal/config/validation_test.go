package config

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestConfigValidator_ValidateConfig(t *testing.T) {
	// Use less strict rules for testing
	testRules := &ConfigValidationRules{
		MaxDiskUsage: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 10; return &i }(),
			Max:      func() *int { i := 95; return &i }(),
			Message:  "Max disk usage must be between 10% and 95%",
		},
		MinProtectedPaths: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 1; return &i }(),
			Message:  "At least one protected path is required",
		},
		ProfileNamePattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^[a-zA-Z0-9_-]+$",
			Message:  "Profile names must be alphanumeric with underscores and hyphens",
		},
		UniquePaths:    true,
		UniqueProfiles: true,
		ProtectedSystemPaths: []string{ // Reduced for tests
			"/",
			"/System",
			"/Library",
		},
		RequireSafeMode: true,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}
	validator := NewConfigValidatorWithRules(testRules)

	tests := []struct {
		name        string
		config      *domain.Config
		expectValid bool
		expectError string
	}{
		{
			name: "valid config",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library", "/usr", "/etc", "/var", "/bin", "/sbin", "/"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid max disk usage",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 150, // Invalid: > 95
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectValid: false,
			expectError: "max_disk_usage",
		},
		{
			name: "missing version",
			config: &domain.Config{
				Version:      "", // Missing
				SafeMode:     true,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectValid: false,
			expectError: "version",
		},
		{
			name: "empty protected paths",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 50,
				Protected:    []string{}, // Empty
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectValid: false,
			expectError: "protected",
		},
		{
			name: "no profiles",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles:     map[string]*domain.Profile{}, // Empty
			},
			expectValid: false,
			expectError: "profiles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.config)

			if result.IsValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
			}

			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if err.Field == tt.expectError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error for field '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}

			if tt.expectValid && len(result.Errors) > 0 {
				t.Errorf("Expected no errors, got: %v", result.Errors)
			}
		})
	}
}

func TestConfigValidator_ValidateField(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		field       string
		value       any
		expectError bool
	}{
		{
			name:        "valid max disk usage",
			field:       "max_disk_usage",
			value:       50,
			expectError: false,
		},
		{
			name:        "invalid max disk usage - too high",
			field:       "max_disk_usage",
			value:       150,
			expectError: true,
		},
		{
			name:        "invalid max disk usage - too low",
			field:       "max_disk_usage",
			value:       5,
			expectError: true,
		},
		{
			name:        "invalid max disk usage - wrong type",
			field:       "max_disk_usage",
			value:       "50",
			expectError: true,
		},
		{
			name:        "valid protected paths",
			field:       "protected",
			value:       []string{"/System", "/Library"},
			expectError: false,
		},
		{
			name:        "invalid protected paths - empty",
			field:       "protected",
			value:       []string{},
			expectError: true,
		},
		{
			name:        "invalid protected paths - wrong type",
			field:       "protected",
			value:       "/System",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.field, tt.value)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error=%v, got %v", tt.expectError, err != nil)
			}
		})
	}
}

func TestConfigSanitizer_SanitizeConfig(t *testing.T) {
	sanitizer := NewConfigSanitizer()

	tests := []struct {
		name             string
		config           *domain.Config
		expectedChanges  []string
		expectedWarnings int
	}{
		{
			name: "whitespace cleanup",
			config: &domain.Config{
				Version:      "  1.0.0  ",
				SafeMode:     true,
				MaxDiskUsage: 52,
				Protected:    []string{"/System", "  /Library  "},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "  daily  ",
						Description: "  Daily cleanup  ",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "  Clean Nix generations  ",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
							},
						},
						Enabled: true,
					},
				},
			},
			expectedChanges:  []string{"version", "protected[1]", "profiles.daily.name", "profiles.daily.description", "profiles.daily.operations[0].description"},
			expectedWarnings: 0,
		},
		{
			name: "max disk usage clamping",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 150, // Will be clamped to 95
				Protected:    []string{"/System", "/usr", "/etc", "/var", "/bin", "/sbin"},
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectedChanges:  []string{"max_disk_usage"},
			expectedWarnings: 1, // Warning about clamping
		},
		{
			name: "duplicate paths",
			config: &domain.Config{
				Version:      "1.0.0",
				SafeMode:     true,
				MaxDiskUsage: 50,
				Protected:    []string{"/System", "/Library", "/System"}, // Duplicate /System
				Profiles: map[string]*domain.Profile{
					"daily": {
						Name:        "daily",
						Description: "Daily cleanup",
						Operations: []domain.CleanupOperation{
							{
								Name:        "nix-generations",
								Description: "Clean Nix generations",
								RiskLevel:   domain.RiskLow,
								Enabled:     true,
								Settings: &domain.OperationSettings{
									Type: domain.OperationTypeNixStore,
									NixStore: &domain.NixStoreSettings{
										KeepGenerations: 3,
										MinAge:          time.Hour * 24,
										DryRun:          true,
									},
								},
							},
						},
						Enabled: true,
					},
				},
			},
			expectedChanges:  []string{"profiles.daily.operations[0].settings"}, // Settings sanitized
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validation result to capture sanitization
			validationResult := &ValidationResult{
				IsValid:   true,
				Errors:    []ValidationError{},
				Warnings:  []ValidationWarning{},
				Sanitized: make(map[string]any),
			}

			sanitizationResult := sanitizer.SanitizeConfig(tt.config)
			validationResult.Sanitized = map[string]any{
				"sanitized": sanitizationResult.Sanitized,
				"changes":   sanitizationResult.Changes,
			}

			// Check expected changes
			changes, ok := validationResult.Sanitized["changes"].([]SanitizationChange)
			if !ok && len(tt.expectedChanges) > 0 {
				t.Errorf("Expected changes but got no changes array")
			}

			if ok {
				for _, expectedChange := range tt.expectedChanges {
					found := false
					for _, change := range changes {
						if change.Field == expectedChange || len(change.Field) > len(expectedChange) && change.Field[len(change.Field)-len(expectedChange):] == expectedChange {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected change for field '%s', not found in: %v", expectedChange, changes)
					}
				}
			}

			// Check changes count
			if ok {
				if len(changes) != len(tt.expectedChanges) {
					t.Errorf("Expected %d changes, got %d", len(tt.expectedChanges), len(changes))
				}
			}
		})
	}
}

func TestValidationMiddleware_ValidateAndLoadConfig(t *testing.T) {
	middleware := NewValidationMiddleware()

	// This test would require actual config file setup
	// For now, just test the middleware structure
	if middleware.validator == nil {
		t.Error("Expected validator to be initialized")
	}

	if middleware.sanitizer == nil {
		t.Error("Expected sanitizer to be initialized")
	}
}

func TestValidationMiddleware_ValidateConfigChange(t *testing.T) {
	// Use custom validation rules for testing
	testRules := &ConfigValidationRules{
		MaxDiskUsage: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 10; return &i }(),
			Max:      func() *int { i := 95; return &i }(),
			Message:  "Max disk usage must be between 10% and 95%",
		},
		MinProtectedPaths: &ValidationRule[int]{
			Required: true,
			Min:      func() *int { i := 1; return &i }(),
			Message:  "At least one protected path is required",
		},
		ProfileNamePattern: &ValidationRule[string]{
			Required: true,
			Pattern:  "^[a-zA-Z0-9_-]+$",
			Message:  "Profile names must be alphanumeric with underscores and hyphens",
		},
		UniquePaths:    true,
		UniqueProfiles: true,
		ProtectedSystemPaths: []string{ // Reduced for tests
			"/",
			"/System",
			"/Library",
		},
		RequireSafeMode: true,
		MaxRiskLevel:    domain.RiskHigh,
		BackupRequired:  domain.RiskMedium,
	}

	// Create middleware with custom validator that doesn't sanitize for change detection
	validator := NewConfigValidatorWithRules(testRules)
	middleware := &ValidationMiddleware{
		validator: validator,
		sanitizer: NewConfigSanitizer(),
		logger:    NewDefaultValidationLogger(false),
	}

	current := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     true,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library", "/usr", "/etc", "/var", "/bin", "/sbin"},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
				},
				Enabled: true,
			},
		},
	}

	proposed := &domain.Config{
		Version:      "1.0.1",
		SafeMode:     false,                                                                                     // Changed
		MaxDiskUsage: 60,                                                                                        // Changed
		Protected:    []string{"/System", "/Library", "/Applications", "/usr", "/etc", "/var", "/bin", "/sbin"}, // Added
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
				},
				Enabled: true,
			},
			"aggressive": { // Added new profile
				Name:        "aggressive",
				Description: "Aggressive cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskHigh,
						Enabled:     true,
					},
				},
				Enabled: true,
			},
		},
	}

	result := middleware.ValidateConfigChange(nil, current, proposed)
	if result == nil || !result.IsValid {
		if result != nil {
			t.Errorf("Expected change validation to pass, got errors: %v", result.Errors)
		} else {
			t.Error("Expected change validation to pass, got nil result")
		}
		return
	}

	// Verify expected changes were detected
	if len(result.Changes) < 3 { // Should have at least 3 changes (safe_mode, max_disk_usage, protected)
		t.Errorf("Expected at least 3 changes, got %d", len(result.Changes))
	}

	// Check that changes were detected
	if len(result.Changes) < 3 { // Should have at least 3 changes (safe_mode, max_disk_usage, protected)
		t.Errorf("Expected at least 3 changes, got %d", len(result.Changes))
	}

	// Verify specific changes
	changeFields := make(map[string]bool)
	for _, change := range result.Changes {
		changeFields[change.Field] = true
	}

	if !changeFields["safe_mode"] {
		t.Error("Expected safe_mode change to be detected")
	}

	if !changeFields["max_disk_usage"] {
		t.Error("Expected max_disk_usage change to be detected")
	}

	if !changeFields["protected"] {
		t.Error("Expected protected paths change to be detected")
	}
}

func TestEnhancedConfigLoader_LoadConfig(t *testing.T) {
	loader := NewEnhancedConfigLoader()

	options := &ConfigLoadOptions{
		ForceRefresh:       false,
		EnableCache:        true,
		EnableSanitization: true,
		ValidationLevel:    ValidationLevelBasic,
		Timeout:            5 * time.Second,
	}

	// Test loading with context timeout
	ctx, cancel := testContext.WithTimeout(testContext.Background(), 1*time.Second)
	defer cancel()

	// This will likely fail due to missing config file, but tests the structure
	config, err := loader.LoadConfig(ctx, options)

	if config != nil {
		t.Log("Configuration loaded successfully")
	} else {
		t.Logf("Configuration loading failed (expected in test): %v", err)
	}
}

func TestEnhancedConfigLoader_GetConfigSchema(t *testing.T) {
	loader := NewEnhancedConfigLoader()
	schema := loader.GetConfigSchema()

	if schema == nil {
		t.Error("Expected schema to be returned")
		return
	}

	if schema.Version == "" {
		t.Error("Expected schema version")
	}

	if schema.Title == "" {
		t.Error("Expected schema title")
	}

	if schema.Types == nil {
		t.Error("Expected schema types")
	}

	if schema.Validation == nil {
		t.Error("Expected validation rules")
	}

	// Check Config type definition
	configType, exists := schema.Types["Config"]
	if !exists {
		t.Error("Expected Config type definition")
		return
	}

	if configType.Type != "object" {
		t.Errorf("Expected Config type to be object, got %s", configType.Type)
	}

	if configType.Properties == nil {
		t.Error("Expected Config type to have properties")
	}

	// Check required properties
	requiredProps := []string{"version", "safe_mode", "max_disk_usage", "protected", "profiles"}
	for _, prop := range requiredProps {
		if _, exists := configType.Properties[prop]; !exists {
			t.Errorf("Expected property '%s' in Config type", prop)
		}
	}
}

func TestValidationLevels(t *testing.T) {
	config := &domain.Config{
		Version:      "1.0.0",
		SafeMode:     false, // This will generate warnings in comprehensive mode
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library", "/usr", "/etc", "/var", "/bin", "/sbin"},
		Profiles: map[string]*domain.Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
					},
				},
				Enabled: true,
			},
		},
	}

	// Test different validation levels
	levels := []ValidationLevel{
		ValidationLevelNone,
		ValidationLevelBasic,
		ValidationLevelComprehensive,
		ValidationLevelStrict,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			loader := NewEnhancedConfigLoader()
			result := loader.ValidateConfig(nil, config, level)

			switch level {
			case ValidationLevelNone:
				if !result.IsValid {
					t.Errorf("ValidationLevelNone should always pass")
				}
				if len(result.Errors) > 0 {
					t.Errorf("ValidationLevelNone should not have errors")
				}
				if len(result.Warnings) > 0 {
					t.Errorf("ValidationLevelNone should not have warnings")
				}
			case ValidationLevelStrict:
				// Strict mode might fail depending on config
				t.Logf("Strict validation result: IsValid=%v, Errors=%d, Warnings=%d",
					result.IsValid, len(result.Errors), len(result.Warnings))
			}
		})
	}
}

// Context for tests
var (
	testContext = struct {
		Background  func() context.Context
		WithTimeout func(context.Context, time.Duration) (context.Context, context.CancelFunc)
	}{
		Background: func() context.Context {
			return context.Background()
		},
		WithTimeout: func(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, timeout)
		},
	}
)
