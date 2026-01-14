package config

import (
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Test constants for invalid risk levels to improve readability and maintainability.
const (
	testInvalidRiskUnknown  = domain.RiskLevelType(99)  // Unknown risk value outside valid range
	testInvalidRiskNegative = domain.RiskLevelType(-1)  // Negative risk value
	testInvalidRiskTooHigh  = domain.RiskLevelType(100) // Risk value above maximum
)

// contains helper function.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    domain.RiskLevel
		expected string
	}{
		{"low risk", domain.RiskLow, "LOW"},
		{"medium risk", domain.RiskMedium, "MEDIUM"},
		{"high risk", domain.RiskHigh, "HIGH"},
		{"critical risk", domain.RiskCritical, "CRITICAL"},
		{"unknown risk", testInvalidRiskUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRiskLevel_Icon(t *testing.T) {
	tests := []struct {
		name     string
		level    domain.RiskLevel
		expected string
	}{
		{"low risk", domain.RiskLow, "🟢"},
		{"medium risk", domain.RiskMedium, "🟡"},
		{"high risk", domain.RiskHigh, "🟠"},
		{"critical risk", domain.RiskCritical, "🔴"},
		{"unknown risk", testInvalidRiskUnknown, "⚪"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.Icon()
			if result != tt.expected {
				t.Errorf("Icon() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRiskLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    domain.RiskLevel
		expected bool
	}{
		{"low risk", domain.RiskLow, true},
		{"medium risk", domain.RiskMedium, true},
		{"high risk", domain.RiskHigh, true},
		{"critical risk", domain.RiskCritical, true},
		{"unknown risk", testInvalidRiskUnknown, false},
		{"negative risk", testInvalidRiskNegative, false},
		{"too high risk", testInvalidRiskTooHigh, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		cleanType CleanType
		expected  bool
	}{
		{"nix store", CleanTypeNixStore, true},
		{"homebrew", CleanTypeHomebrew, true},
		{"package cache", CleanTypePackageCache, true},
		{"temp files", CleanTypeTempFiles, true},
		{"invalid type", CleanType("invalid"), false},
		{"empty type", CleanType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cleanType.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSafeConfigBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		builderFunc func() *SafeConfigBuilder
		expectError bool
		errorMsg    string
	}{
		{
			name: "build valid config",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, domain.RiskLow).
					Done()
			},
			expectError: false,
		},
		{
			name: "build config with no profiles",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder()
			},
			expectError: true,
			errorMsg:    "config must have at least one profile",
		},
		{
			name: "build config with valid risk level only",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, domain.RiskLow).
					Done()
			},
			expectError: false,
		},
		{
			name: "build config with critical risk operation should fail",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, domain.RiskCritical).
					Done()
			},
			expectError: true,
			errorMsg:    "cannot add critical risk operation to profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := tt.builderFunc().Build()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !config.created.Before(time.Now().Add(time.Second)) {
				t.Error("config creation time seems incorrect")
			}
		})
	}
}

func TestSafeProfileBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		builderFunc func() *SafeProfileBuilder
		expectError bool
		errorMsg    string
	}{
		{
			name: "build valid profile",
			builderFunc: func() *SafeProfileBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, domain.RiskLow)
			},
			expectError: false,
		},
		{
			name: "build profile with no operations",
			builderFunc: func() *SafeProfileBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile")
			},
			expectError: true,
			errorMsg:    "profile must have at least one operation",
		},
		{
			name: "build profile with high risk (valid)",
			builderFunc: func() *SafeProfileBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, domain.RiskHigh)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to build the full config to test profile building
			config, err := tt.builderFunc().Done().Build()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorMsg != "" {
					errMsg := err.Error()
					if !contains(errMsg, tt.errorMsg) {
						t.Errorf("expected error message containing %q, got %q", tt.errorMsg, errMsg)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check if the profile was created correctly
			if len(config.profiles) == 0 {
				t.Error("expected at least one profile")
				return
			}

			profile := config.profiles[0]
			if profile.name != "test" {
				t.Errorf("expected profile name 'test', got %q", profile.name)
			}

			if profile.description != "test profile" {
				t.Errorf("expected profile description 'test profile', got %q", profile.description)
			}

			if len(profile.operations) == 0 {
				t.Error("expected at least one operation")
			}
		})
	}
}
