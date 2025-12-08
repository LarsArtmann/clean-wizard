package config

import (
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/adapters"
	"github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// SafeConfigBuilder provides type-safe configuration building
type SafeConfigBuilder struct {
	profileName string
	profileDesc string
	operations  []config.CleanupOperation
}

// NewSafeConfigBuilder creates a new safe config builder
func NewSafeConfigBuilder() *SafeConfigBuilder {
	return &SafeConfigBuilder{
		operations: make([]config.CleanupOperation, 0),
	}
}

// AddProfile adds a profile to the builder
func (scb *SafeConfigBuilder) AddProfile(name, desc string) *SafeConfigBuilder {
	scb.profileName = name
	scb.profileDesc = desc
	return scb
}

// AddOperation adds an operation to the builder
func (scb *SafeConfigBuilder) AddOperation(opType shared.CleanType, riskLevel shared.RiskLevelType) *SafeConfigBuilder {
	operation := config.CleanupOperation{
		Name:        strings.ToLower(string(opType)),
		RiskLevel:   riskLevel,
		Description: opType.String() + " cleanup",
	}
	scb.operations = append(scb.operations, operation)
	return scb
}

// Done completes the builder and returns the config
func (scb *SafeConfigBuilder) Done() *SafeConfigBuilder {
	return scb
}

// Test constants for invalid risk levels to improve readability and maintainability
const (
	testInvalidRiskUnknown  = shared.RiskLevelType(99)  // Unknown risk value outside valid range
	testInvalidRiskNegative = shared.RiskLevelType(-1)  // Negative risk value
	testInvalidRiskTooHigh  = shared.RiskLevelType(100) // Risk value above maximum
)

// contains helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    shared.RiskLevel
		expected string
	}{
		{"low risk", shared.RiskLow, "LOW"},
		{"medium risk", shared.RiskMedium, "MEDIUM"},
		{"high risk", shared.RiskHigh, "HIGH"},
		{"critical risk", shared.RiskCritical, "CRITICAL"},
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
	uiAdapter := adapters.NewUIAdapter()

	tests := []struct {
		name     string
		level    shared.RiskLevel
		expected string
	}{
		{"low risk", shared.RiskLow, "🟢"},
		{"medium risk", shared.RiskMedium, "🟡"},
		{"high risk", shared.RiskHigh, "🟠"},
		{"critical risk", shared.RiskCritical, "🔴"},
		{"unknown risk", testInvalidRiskUnknown, "⚪"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uiAdapter.RiskLevelIcon(tt.level)
			if result != tt.expected {
				t.Errorf("RiskLevelIcon() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRiskLevel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		level    shared.RiskLevel
		expected bool
	}{
		{"low risk", shared.RiskLow, true},
		{"medium risk", shared.RiskMedium, true},
		{"high risk", shared.RiskHigh, true},
		{"critical risk", shared.RiskCritical, true},
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
		cleanType shared.CleanType
		expected  bool
	}{
		{"nix store", shared.CleanType("nix-store"), true},
		{"homebrew", shared.CleanType("homebrew"), true},
		{"package cache", shared.CleanType("package-cache"), true},
		{"temp files", shared.CleanType("temp-files"), true},
		{"invalid type", shared.CleanType("invalid"), false},
		{"empty type", shared.CleanType(""), false},
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
					AddOperation(shared.CleanType("nix-store"), shared.RiskLevelLowType).
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
					AddOperation(shared.CleanType("nix-store"), shared.RiskLevelLowType).
					Done()
			},
			expectError: false,
		},
		{
			name: "build config with critical risk operation should fail",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, shared.RiskCritical).
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
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !config.Updated.Before(time.Now().Add(time.Second)) {
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
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(shared.CleanType("nix-store"), shared.RiskLevelLowType)
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
					AddOperation(CleanTypeNixStore, shared.RiskHigh)
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
			if len(Configprofiles) == 0 {
				t.Error("expected at least one profile")
				return
			}

			profile := Configprofiles[0]
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

// Build creates a configuration from builder
func (scb *SafeConfigBuilder) Build() *config.Config {
	profile := &config.Profile{
		Name:        scb.profileName,
		Description: scb.profileDesc,
		Status:      shared.StatusEnabled,
		Operations:  scb.operations,
	}

	return &config.Config{
		Version:      "1.0.0",
		SafetyLevel:  shared.SafetyLevelMediumType,
		MaxDiskUsage: 50,
		Profiles: map[string]*config.Profile{
			scb.profileName: profile,
		},
		CurrentProfile: scb.profileName,
		LastClean:      time.Now(),
		Updated:        time.Now(),
	}
}
