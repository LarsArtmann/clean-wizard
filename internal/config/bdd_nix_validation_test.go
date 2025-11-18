package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestBDD_NixGenerationsValidation provides comprehensive BDD tests for Nix generations
func TestBDD_NixGenerationsValidation(t *testing.T) {
	feature := BDDFeature{
		Name:       "Nix Generations Configuration Validation",
		Description: "As a system administrator, I want to configure Nix generations cleanup with proper validation and constraints",
		Background: "The system should validate all Nix generations settings against business rules and safety constraints",
		Scenarios: []BDDScenario{
			{
				Name:        "Valid Nix generations within acceptable range",
				Description: "Should accept valid Nix generations configuration",
				Given: []BDDGiven{
					{
						Description: "a configuration with valid Nix generations settings",
						Setup: func() (*domain.Config, error) {
							return &domain.Config{
								Version:    "1.0.0",
								SafeMode:   true,
								MaxDiskUsage: 50,
								Protected:  []string{"/System", "/Applications"},
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
							}, nil
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
							return &domain.Config{
								Version:    "1.0.0",
								SafeMode:   true,
								MaxDiskUsage: 50,
								Protected:  []string{"/System", "/Applications"},
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
														Generations: 0, // Below minimum
														Optimize:    true,
													},
												},
											},
										},
										Enabled: true,
									},
								},
							}, nil
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
							return &domain.Config{
								Version:    "1.0.0",
								SafeMode:   true,
								MaxDiskUsage: 50,
								Protected:  []string{"/System", "/Applications"},
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
														Generations: 15, // Above maximum
														Optimize:    true,
													},
												},
											},
										},
										Enabled: true,
									},
								},
							}, nil
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
							return &domain.Config{
								Version:    "1.0.0",
								SafeMode:   false, // Unsafe mode
								MaxDiskUsage: 50,
								Protected:  []string{"/System", "/Applications"},
								Profiles: map[string]*domain.Profile{
									"critical-nix": {
										Name:        "Critical Nix Cleanup",
										Description: "Critical Nix generations cleanup",
										Operations: []domain.CleanupOperation{
											{
												Name:        "nix-generations",
												Description: "Critical Nix cleanup",
												RiskLevel:   domain.RiskCritical, // Critical risk
												Enabled:     true,
												Settings: &domain.OperationSettings{
													NixGenerations: &domain.NixGenerationsSettings{
														Generations: 1,
														Optimize:    false,
													},
												},
											},
										},
										Enabled: true,
									},
								},
							}, nil
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