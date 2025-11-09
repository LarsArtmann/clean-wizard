package config

import (
	"strings"
	"testing"
	"time"
)

// contains helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    RiskLevel
		expected string
	}{
		{"low risk", RiskLow, "LOW"},
		{"medium risk", RiskMedium, "MEDIUM"},
		{"high risk", RiskHigh, "HIGH"},
		{"critical risk", RiskCritical, "CRITICAL"},
		{"unknown risk", RiskLevel(999), "UNKNOWN"},
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
		level    RiskLevel
		expected string
	}{
		{"low risk", RiskLow, "🟢"},
		{"medium risk", RiskMedium, "🟡"},
		{"high risk", RiskHigh, "🟠"},
		{"critical risk", RiskCritical, "🔴"},
		{"unknown risk", RiskLevel(999), "⚪"},
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
		level    RiskLevel
		expected bool
	}{
		{"low risk", RiskLow, true},
		{"medium risk", RiskMedium, true},
		{"high risk", RiskHigh, true},
		{"critical risk", RiskCritical, true},
		{"negative risk", RiskLevel(-1), false},
		{"too high risk", RiskLevel(100), false},
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
		name     string
		cleanType CleanType
		expected bool
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
					AddOperation(CleanTypeNixStore, RiskLow).
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
			name: "build config with invalid risk level",
			builderFunc: func() *SafeConfigBuilder {
				builder := NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, RiskLow).
					Done()
				// Force invalid risk level
				builder.maxRisk = RiskLevel(999)
				return builder
			},
			expectError: true,
			errorMsg:    "invalid risk level: 999",
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
					AddOperation(CleanTypeNixStore, RiskLow)
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
			name: "build profile with critical risk",
			builderFunc: func() *SafeProfileBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, RiskCritical)
			},
			expectError: true,
			errorMsg:    "cannot add critical risk operation to profile",
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