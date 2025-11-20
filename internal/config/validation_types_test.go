package config

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestValidationLevelTestCase defines validation level test cases
type TestValidationLevelTestCase struct {
	name         string
	config       *domain.Config
	level        ValidationLevel
	expectValid  bool
	expectErrors int
}

// CreateTestConfigurations creates test configurations for validation testing
func CreateTestConfigurations() map[string]*domain.Config {
	return map[string]*domain.Config{
		"valid": {
			Version:      "1.0.0",
			SafetyLevel:  domain.SafetyLevelEnabled,
			MaxDiskUsage: 50,
			Protected:    []string{"/System", "/Library", "/Applications"},
			Profiles: map[string]*domain.Profile{
				"daily": {
					Name:        "daily",
					Description: "Daily cleanup",
					Operations: []domain.CleanupOperation{
						{
							Name:        "nix-generations",
							Description: "Clean Nix generations",
							RiskLevel:   domain.RiskLow,
							Status:      domain.StatusEnabled,
						},
					},
					Status: domain.StatusEnabled,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
		"invalid_high_disk": {
			Version:      "1.0.0",
			SafetyLevel:  domain.SafetyLevelEnabled,
			MaxDiskUsage: 150, // Invalid: too high
			Protected:    []string{"/System"},
			Profiles: map[string]*domain.Profile{
				"daily": {
					Name:        "daily",
					Description: "Daily cleanup",
					Operations: []domain.CleanupOperation{
						{
							Name:        "nix-generations",
							Description: "Clean Nix generations",
							RiskLevel:   domain.RiskLow,
							Status:      domain.StatusEnabled,
						},
					},
					Status: domain.StatusEnabled,
				},
			},
			LastClean: time.Now(),
			Updated:   time.Now(),
		},
	}
}

// GetSanitizationTestCases returns all sanitization test cases
// Note: Delegates to shared test data in test_data.go
func GetSanitizationTestCases() []SanitizationTestCase {
	return GetStandardTestCasesCompatWrapper()
}

// GetStandardTestCasesCompatWrapper converts standard test cases to sanitization test cases
func GetStandardTestCasesCompatWrapper() []SanitizationTestCase {
	standardCases := GetStandardTestCases()
	result := make([]SanitizationTestCase, len(standardCases))
	for i, tc := range standardCases {
		result[i] = SanitizationTestCase(tc)
	}
	return result
}
