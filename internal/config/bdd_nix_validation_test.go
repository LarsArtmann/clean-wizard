package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Builder helpers for BDD test scenarios

// newBaseNixConfig creates a common config skeleton for Nix tests
func newBaseNixConfig(safeMode bool) *domain.Config {
	return &domain.Config{
		Version:      "1.0.0",
		SafeMode:     safeMode,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Applications"},
		Profiles: map[string]*domain.Profile{
			"nix-cleanup": {
				Name:        "Nix Cleanup",
				Description: "Clean Nix generations",
				Operations: []domain.CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean old Nix generations",
						RiskLevel:   domain.RiskLow,
						Enabled:     true,
						Settings: &domain.OperationSettings{
							NixGenerations: &domain.NixGenerationsSettings{
								Generations: 3,
								Optimize:    true,
							},
						},
					},
				},
				Enabled: true,
			},
		},
	}
}

// withGenerations sets/updates the nix-generations operation Generations value
func withGenerations(cfg *domain.Config, generations int) *domain.Config {
	// Find the nix-cleanup profile and its nix-generations operation
	if profile, exists := cfg.Profiles["nix-cleanup"]; exists {
		for i, op := range profile.Operations {
			if op.Name == "nix-generations" && op.Settings != nil && op.Settings.NixGenerations != nil {
				profile.Operations[i].Settings.NixGenerations.Generations = generations
				break
			}
		}
	}
	return cfg
}

// withRiskLevel adjusts the operation RiskLevel and Enabled flags
func withRiskLevel(cfg *domain.Config, level domain.RiskLevelType) *domain.Config {
	// Find the nix-cleanup profile and its nix-generations operation
	if profile, exists := cfg.Profiles["nix-cleanup"]; exists {
		for i, op := range profile.Operations {
			if op.Name == "nix-generations" {
				profile.Operations[i].RiskLevel = level
				// Auto-disable critical operations in unsafe mode
				if level == domain.RiskCritical && !cfg.SafeMode {
					profile.Operations[i].Enabled = false
				}
				break
			}
		}
	}
	return cfg
}

// withOptimize sets the Optimize flag for nix-generations
func withOptimize(cfg *domain.Config, optimize bool) *domain.Config {
	// Find the nix-cleanup profile and its nix-generations operation
	if profile, exists := cfg.Profiles["nix-cleanup"]; exists {
		for i, op := range profile.Operations {
			if op.Name == "nix-generations" && op.Settings != nil && op.Settings.NixGenerations != nil {
				profile.Operations[i].Settings.NixGenerations.Optimize = optimize
				break
			}
		}
	}
	return cfg
}

// withEnabled sets the Enabled flag for nix-generations operation
func withEnabled(cfg *domain.Config, enabled bool) *domain.Config {
	// Find the nix-cleanup profile and its nix-generations operation
	if profile, exists := cfg.Profiles["nix-cleanup"]; exists {
		for i, op := range profile.Operations {
			if op.Name == "nix-generations" {
				profile.Operations[i].Enabled = enabled
				break
			}
		}
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
						Setup: func() (*domain.Config, error) {
							return newBaseNixConfig(true), nil
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
						Setup: func() (*domain.Config, error) {
							return withGenerations(newBaseNixConfig(true), 0), nil
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
						Setup: func() (*domain.Config, error) {
							return withGenerations(newBaseNixConfig(true), 15), nil
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
						Setup: func() (*domain.Config, error) {
							return withRiskLevel(withGenerations(withOptimize(newBaseNixConfig(false), false), 1), domain.RiskCritical), nil
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
