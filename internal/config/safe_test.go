package config

import (
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

// Test constants for invalid risk levels to improve readability and maintainability.
const (
	testInvalidRiskUnknown  = enums.RiskLevelType(99)  // Unknown risk value outside valid range
	testInvalidRiskNegative = enums.RiskLevelType(-1)  // Negative risk value
	testInvalidRiskTooHigh  = enums.RiskLevelType(100) // Risk value above maximum
)

// contains helper function.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// riskLevelTestCases provides reusable test cases for RiskLevel method testing.
var riskLevelTestCases = []struct {
	name  string
	level enums.RiskLevelType
}{
	{"low risk", enums.RiskLevelLowType},
	{"medium risk", enums.RiskLevelMediumType},
	{"high risk", enums.RiskLevelHighType},
	{"critical risk", enums.RiskLevelCriticalType},
	{"unknown risk", testInvalidRiskUnknown},
}

// riskLevelValues defines the ordered risk level keys used for test value maps.
var riskLevelValues = []enums.RiskLevelType{
	enums.RiskLevelLowType,
	enums.RiskLevelMediumType,
	enums.RiskLevelHighType,
	enums.RiskLevelCriticalType,
	testInvalidRiskUnknown,
}

// newRiskLevelValueMap creates a map of risk levels to values using the provided value function.
// The value function receives the index (0-4) and should return the corresponding string value.
func newRiskLevelValueMap(values ...string) map[enums.RiskLevelType]string {
	m := make(map[enums.RiskLevelType]string, len(riskLevelValues))
	for i, level := range riskLevelValues {
		if i < len(values) {
			m[level] = values[i]
		}
	}

	return m
}

// riskLevelTextValues provides expected text values for String() method testing.
var riskLevelTextValues = newRiskLevelValueMap("LOW", "MEDIUM", "HIGH", "CRITICAL", "UNKNOWN")

// riskLevelEmojiValues provides expected emoji values for Icon() method testing.
var riskLevelEmojiValues = newRiskLevelValueMap("🟢", "🟡", "🟠", "🔴", "⚪")

// testRiskLevelMethod is a helper function that tests RiskLevel methods with a value map.
func testRiskLevelMethod(
	t *testing.T, methodName string,
	method func(enums.RiskLevelType) string, expected map[enums.RiskLevelType]string,
) {
	t.Helper()

	for _, tc := range riskLevelTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := method(tc.level)

			expect := expected[tc.level]
			if result != expect {
				t.Errorf("%s() = %v, want %v", methodName, result, expect)
			}
		})
	}
}

// validatable is an interface for types that can validate themselves.
type validatable interface {
	IsValid() bool
}

// testIsValid is a generic helper function that tests IsValid() methods.
func testIsValid[T validatable](t *testing.T, tests []struct {
	name     string
	value    T
	expected bool
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.value.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRiskLevel_String(t *testing.T) {
	t.Parallel()
	testRiskLevelMethod(
		t,
		"String",
		func(level enums.RiskLevelType) string { return level.String() },
		riskLevelTextValues,
	)
}

func TestRiskLevel_Icon(t *testing.T) {
	t.Parallel()
	testRiskLevelMethod(
		t, "Icon", func(level enums.RiskLevelType) string { return level.Icon() },
		riskLevelEmojiValues,
	)
}

func TestRiskLevel_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    enums.RiskLevelType
		expected bool
	}{
		{"low risk", enums.RiskLevelLowType, true},
		{"medium risk", enums.RiskLevelMediumType, true},
		{"high risk", enums.RiskLevelHighType, true},
		{"critical risk", enums.RiskLevelCriticalType, true},
		{"unknown risk", testInvalidRiskUnknown, false},
		{"negative risk", testInvalidRiskNegative, false},
		{"too high risk", testInvalidRiskTooHigh, false},
	}

	testIsValid(t, tests)
}

func TestCleanType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    CleanType
		expected bool
	}{
		{"nix store", CleanTypeNixStore, true},
		{"homebrew", CleanTypeHomebrew, true},
		{"package cache", CleanTypePackageCache, true},
		{"temp files", CleanTypeTempFiles, true},
		{"invalid type", CleanType("invalid"), false},
		{"empty type", CleanType(""), false},
	}

	testIsValid(t, tests)
}

func TestSafeConfigBuilder_Build(t *testing.T) {
	t.Parallel()

	validBuilderFunc := func() *SafeConfigBuilder {
		return NewSafeConfigBuilder().
			AddProfile("test", "test profile").
			AddOperation(CleanTypeNixStore, enums.RiskLevelLowType).
			Done()
	}

	tests := []struct {
		name        string
		builderFunc func() *SafeConfigBuilder
		expectError bool
		errorMsg    string
	}{
		{
			name:        "build valid config",
			builderFunc: validBuilderFunc,
			expectError: false,
		},
		{
			name:        "build config with no profiles",
			builderFunc: NewSafeConfigBuilder,
			expectError: true,
			errorMsg:    "config must have at least one profile",
		},
		{
			name:        "build config with valid risk level only",
			builderFunc: validBuilderFunc,
			expectError: false,
		},
		{
			name: "build config with critical risk operation should fail",
			builderFunc: func() *SafeConfigBuilder {
				return NewSafeConfigBuilder().
					AddProfile("test", "test profile").
					AddOperation(CleanTypeNixStore, enums.RiskLevelCriticalType).
					Done()
			},
			expectError: true,
			errorMsg:    "cannot add critical risk operation to profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

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
					AddOperation(CleanTypeNixStore, enums.RiskLevelLowType)
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
					AddOperation(CleanTypeNixStore, enums.RiskLevelHighType)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
						t.Errorf(
							"expected error message containing %q, got %q",
							tt.errorMsg,
							errMsg,
						)
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
