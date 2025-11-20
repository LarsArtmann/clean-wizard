package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"Valid simple semver", "1.0.0", true},
		{"Valid with zeros", "0.1.2", true},
		{"Valid large numbers", "10.20.30", true},
		{"Valid pre-release", "1.0.0-alpha", true},
		{"Valid pre-release with numbers", "1.0.0-alpha.1", true},
		{"Valid build metadata", "1.0.0+20130313144700", true},
		{"Valid pre-release and build", "1.0.0-alpha+001", true},
		{"Invalid empty string", "", false},
		{"Invalid no patch", "1.0", false},
		{"Invalid no minor", "1", false},
		{"Invalid negative numbers", "1.0.-1", false},
		{"Invalid non-numeric", "1.a.0", false},
		{"Invalid leading zero", "01.0.0", false},
		{"Invalid with letters", "v1.0.0", false},
		{"Invalid format", "1/0/0", false},
		{"Invalid whitespace", " 1.0.0 ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSemver(tt.version)
			if result != tt.expected {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

// createTestConfig creates a test configuration with the specified version
func createTestConfig(version string) *domain.Config {
	return &domain.Config{
		Version: version,
		Profiles: map[string]*domain.Profile{
			"test": {
				Name:        "test",
				Description: "Test profile",
				Operations: []domain.CleanupOperation{{
					Name:        "test-op",
					Description: "Test operation",
					RiskLevel:   domain.RiskLow,
					Status:      domain.StatusEnabled,
					Settings:    &domain.OperationSettings{NixGenerations: &domain.NixGenerationsSettings{Generations: 5}},
				}},
				Status: domain.StatusEnabled,
			},
		},
		Protected: []string{"/System"},
	}
}

func TestBasicStructureValidation_Semver(t *testing.T) {
	cv := NewConfigValidator()

	t.Run("Valid semver version", func(t *testing.T) {
		cfg := &domain.Config{
			Version: "1.2.3",
			Profiles: map[string]*domain.Profile{
				"test": {
					Name:        "test",
					Description: "Test profile",
					Operations: []domain.CleanupOperation{{
						Name:        "test-op",
						Description: "Test operation",
						RiskLevel:   domain.RiskLow,
						Status:      domain.StatusEnabled,
						Settings:    &domain.OperationSettings{NixGenerations: &domain.NixGenerationsSettings{Generations: 5}},
					}},
					Status: domain.StatusEnabled,
				},
			},
			Protected: []string{"/System"},
		}

		result := cv.ValidateConfig(cfg)
		if len(result.Errors) != 0 {
			t.Errorf("Expected no validation errors for valid semver, got: %v", result.Errors)
		}
	})

	t.Run("Invalid semver version", func(t *testing.T) {
		cfg := createTestConfig("invalid.version.format")

		result := cv.ValidateConfig(cfg)
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for invalid semver version")
		}

		// Check if error is related to version format
		foundVersionError := false
		for _, err := range result.Errors {
			if err.Field == "version" && err.Rule == "semver_format" {
				foundVersionError = true
				break
			}
		}
		if !foundVersionError {
			t.Errorf("Expected semver_format error for version field, got errors: %v", result.Errors)
		}
	})

	t.Run("Missing version", func(t *testing.T) {
		cfg := createTestConfig("")

		result := cv.ValidateConfig(cfg)
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for missing version")
		}

		// Check if error is related to missing version
		foundMissingError := false
		for _, err := range result.Errors {
			if err.Field == "version" && err.Rule == "required" {
				foundMissingError = true
				break
			}
		}
		if !foundMissingError {
			t.Errorf("Expected required error for version field, got errors: %v", result.Errors)
		}
	})
}
