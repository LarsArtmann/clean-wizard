package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Builder helpers for BDD test scenarios

// newBaseNixConfig creates a common config skeleton for Nix tests
func newBaseNixConfig(t *testing.T, safeMode bool) *domain.Config {
	t.Helper()

	var safetyLevel domain.SafetyLevelType
	if safeMode {
		safetyLevel = domain.SafetyLevelEnabled
	} else {
		safetyLevel = domain.SafetyLevelDisabled
	}

	return &domain.Config{
		Version:      "1.0.0",
		SafetyLevel:  safetyLevel,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Applications"},
		Profiles: map[string]*domain.Profile{
			"nix-cleanup": {
				Name:        "Nix Cleanup",
				Description: "Clean Nix generations",
				Status:      domain.StatusEnabled,
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskLow,
						Status:      domain.StatusEnabled,
						Settings: &domain.OperationSettings{
							NixGenerations: &domain.NixGenerationsSettings{
								Generations:  3,
								Optimization: domain.OptimizationLevelConservative,
							},
						},
					},
				},
			},
		},
	}
}

// withGenerations sets/updates the nix-generations operation Generations value
func withGenerations(t *testing.T, cfg *domain.Config, generations int) *domain.Config {
	t.Helper()

	// Find the nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		t.Fatalf("FAILED: 'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				t.Fatalf("FAILED: nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				t.Fatalf("FAILED: nix-generations operation has nil NixGenerations")
			}
			profile.Operations[i].Settings.NixGenerations.Generations = generations
			operationFound = true
			break
		}
	}

	if !operationFound {
		t.Fatalf("FAILED: 'nix-generations' operation not found in nix-cleanup profile")
	}
	return cfg
}

// withRiskLevel adjusts the operation RiskLevel and Enabled flags
func withRiskLevel(t *testing.T, cfg *domain.Config, level domain.RiskLevelType) *domain.Config {
	t.Helper()

	// Find the nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		t.Fatalf("FAILED: 'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			profile.Operations[i].RiskLevel = level
			operationFound = true
			break
		}
	}

	if !operationFound {
		t.Fatalf("FAILED: 'nix-generations' operation not found in nix-cleanup profile")
	}
	return cfg
}

// withOptimize sets the Optimize flag for nix-generations
// withOptimize sets Optimization level for nix-generations
func withOptimize(t *testing.T, cfg *domain.Config, optimize bool) *domain.Config {
	t.Helper()

	var optimizationLevel domain.OptimizationLevelType
	if optimize {
		optimizationLevel = domain.OptimizationLevelConservative
	} else {
		optimizationLevel = domain.OptimizationLevelNone
	}

	// Find the nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		t.Fatalf("FAILED: 'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			if op.Settings == nil {
				t.Fatalf("FAILED: nix-generations operation has nil Settings")
			}
			if op.Settings.NixGenerations == nil {
				t.Fatalf("FAILED: nix-generations operation has nil NixGenerations")
			}
			profile.Operations[i].Settings.NixGenerations.Optimization = optimizationLevel
			operationFound = true
			break
		}
	}

	if !operationFound {
		t.Fatalf("FAILED: 'nix-generations' operation not found in nix-cleanup profile")
	}
	return cfg
}

// withEnabled sets the Status for nix-generations operation
func withEnabled(t *testing.T, cfg *domain.Config, enabled bool) *domain.Config {
	t.Helper()

	var status domain.StatusType
	if enabled {
		status = domain.StatusEnabled
	} else {
		status = domain.StatusDisabled
	}

	// Find the nix-cleanup profile and its nix-generations operation
	profile, profileExists := cfg.Profiles["nix-cleanup"]
	if !profileExists {
		t.Fatalf("FAILED: 'nix-cleanup' profile not found in config")
	}

	operationFound := false
	for i, op := range profile.Operations {
		if op.Name == "nix-generations" {
			profile.Operations[i].Status = status
			operationFound = true
			break
		}
	}

	if !operationFound {
		t.Fatalf("FAILED: 'nix-generations' operation not found in nix-cleanup profile")
	}
	return cfg
}

// TestBDD_NixGenerationsValidation provides comprehensive BDD tests for Nix generations
func TestBDD_NixGenerationsValidation(t *testing.T) {
	feature := BDDFeature{
		Name:        "Nix Generations Configuration Validation",
		Description: "As a system administrator, I want to configure Nix generations cleanup with proper validation and constraints",
		Background:  "The system should validate all Nix generations settings against business rules and safety constraints",
		Scenarios: []BDDScenario{
			{
				Name:        "Valid Nix generations within acceptable range",
				Description: "Should accept valid Nix generations configuration",
				Given: []BDDGiven{
					{
						Description: "a configuration with valid Nix generations settings",
						Setup: func(t *testing.T) (*domain.Config, error) {
							return newBaseNixConfig(t, true), nil
						},
					},
				},
				When: []BDDWhen{
					{
						Description: "the configuration is validated",
						Action: func(cfg *domain.Config) (*ValidationResult, error) {
							validator := NewConfigValidator()
							return validator.ValidateConfig(cfg), nil
						},
					},
				},
				Then: []BDDThen{
					{
						Description: "validation should succeed",
						Validate: func(result *ValidationResult) error {
							if !result.IsValid {
								return fmt.Errorf("expected valid configuration, got errors: %v", result.Errors)
							}
							return nil
						},
					},
					{
						Description: "no validation errors should be present",
						Validate: func(result *ValidationResult) error {
							if len(result.Errors) > 0 {
								return fmt.Errorf("expected no errors, got: %v", result.Errors)
							}
							return nil
						},
					},
				},
			},
			{
				Name:        "Invalid Nix generations below minimum",
				Description: "Should reject Nix generations below minimum threshold",
				Given: []BDDGiven{
					{
						Description: "a configuration with Nix generations below minimum",
						Setup: func(t *testing.T) (*domain.Config, error) {
							return withGenerations(t, newBaseNixConfig(t, true), 0), nil
						},
					},
				},
				When: []BDDWhen{
					{
						Description: "the configuration is validated",
						Action: func(cfg *domain.Config) (*ValidationResult, error) {
							validator := NewConfigValidator()
							return validator.ValidateConfig(cfg), nil
						},
					},
				},
				Then: []BDDThen{
					{
						Description: "validation should fail",
						Validate: func(result *ValidationResult) error {
							if result.IsValid {
								return fmt.Errorf("expected invalid configuration")
							}
							return nil
						},
					},
					{
						Description: "validation errors should be present",
						Validate: func(result *ValidationResult) error {
							if len(result.Errors) == 0 {
								return fmt.Errorf("expected validation errors")
							}
							return nil
						},
					},
					{
						Description: "error should mention generations constraint",
						Validate: func(result *ValidationResult) error {
							found := false
							for _, err := range result.Errors {
								if strings.Contains(err.Message, "generations") {
									found = true
									break
								}
							}
							if !found {
								return fmt.Errorf("expected error mentioning generations constraint")
							}
							return nil
						},
					},
				},
			},
			{
				Name:        "Invalid Nix generations above maximum",
				Description: "Should reject Nix generations above maximum threshold",
				Given: []BDDGiven{
					{
						Description: "a configuration with Nix generations above maximum",
						Setup: func(t *testing.T) (*domain.Config, error) {
							return withGenerations(t, newBaseNixConfig(t, true), 15), nil
						},
					},
				},
				When: []BDDWhen{
					{
						Description: "the configuration is validated",
						Action: func(cfg *domain.Config) (*ValidationResult, error) {
							validator := NewConfigValidator()
							return validator.ValidateConfig(cfg), nil
						},
					},
				},
				Then: []BDDThen{
					{
						Description: "validation should fail",
						Validate: func(result *ValidationResult) error {
							if result.IsValid {
								return fmt.Errorf("expected invalid configuration")
							}
							return nil
						},
					},
					{
						Description: "validation errors should be present",
						Validate: func(result *ValidationResult) error {
							if len(result.Errors) == 0 {
								return fmt.Errorf("expected validation errors")
							}
							return nil
						},
					},
					{
						Description: "error should mention generations constraint",
						Validate: func(result *ValidationResult) error {
							found := false
							for _, err := range result.Errors {
								if strings.Contains(err.Message, "generations") {
									found = true
									break
								}
							}
							if !found {
								return fmt.Errorf("expected error mentioning generations constraint")
							}
							return nil
						},
					},
				},
			},
			{
				Name:        "Critical Nix operation in unsafe mode",
				Description: "Should reject critical Nix operations when safe mode is disabled",
				Given: []BDDGiven{
					{
						Description: "a configuration with critical Nix operation in unsafe mode",
						Setup: func(t *testing.T) (*domain.Config, error) {
							return withRiskLevel(t, withGenerations(t, withOptimize(t, newBaseNixConfig(t, false), false), 1), domain.RiskCritical), nil
						},
					},
				},
				When: []BDDWhen{
					{
						Description: "the configuration is validated",
						Action: func(cfg *domain.Config) (*ValidationResult, error) {
							validator := NewConfigValidator()
							return validator.ValidateConfig(cfg), nil
						},
					},
				},
				Then: []BDDThen{
					{
						Description: "validation should fail with security error",
						Validate: func(result *ValidationResult) error {
							if result.IsValid {
								return fmt.Errorf("expected invalid configuration")
							}
							return nil
						},
					},
					{
						Description: "security validation error should be present",
						Validate: func(result *ValidationResult) error {
							found := false
							for _, err := range result.Errors {
								if err.Rule == "security" && strings.Contains(err.Message, "Critical risk operation") {
									found = true
									break
								}
							}
							if !found {
								return fmt.Errorf("expected security validation error for critical operation in unsafe mode")
							}
							return nil
						},
					},
				},
			},
		},
	}

	runner := NewBDDTestRunner(t, feature)
	runner.RunFeature()
}
