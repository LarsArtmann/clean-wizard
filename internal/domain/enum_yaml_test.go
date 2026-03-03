package domain

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestEnumYAMLMarshaling tests that all enum types can be properly marshaled to YAML.
func TestEnumYAMLMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", CacheCleanupDisabled, "0\n"},
		{"CacheCleanupMode Enabled", CacheCleanupEnabled, "1\n"},

		// DockerPruneMode
		{"DockerPruneMode All", DockerPruneAll, "0\n"},
		{"DockerPruneMode Images", DockerPruneImages, "1\n"},
		{"DockerPruneMode Containers", DockerPruneContainers, "2\n"},
		{"DockerPruneMode Volumes", DockerPruneVolumes, "3\n"},
		{"DockerPruneMode Builds", DockerPruneBuilds, "4\n"},

		// BuildToolType
		{"BuildToolType Go", BuildToolGo, "0\n"},
		{"BuildToolType Rust", BuildToolRust, "1\n"},
		{"BuildToolType Node", BuildToolNode, "2\n"},
		{"BuildToolType Python", BuildToolPython, "3\n"},
		{"BuildToolType Java", BuildToolJava, "4\n"},
		{"BuildToolType Scala", BuildToolScala, "5\n"},

		// CacheType
		{"CacheType Spotlight", CacheTypeSpotlight, "0\n"},
		{"CacheType Xcode", CacheTypeXcode, "1\n"},
		{"CacheType Cocoapods", CacheTypeCocoapods, "2\n"},
		{"CacheType Homebrew", CacheTypeHomebrew, "3\n"},
		{"CacheType Pip", CacheTypePip, "4\n"},
		{"CacheType Npm", CacheTypeNpm, "5\n"},
		{"CacheType Yarn", CacheTypeYarn, "6\n"},
		{"CacheType Ccache", CacheTypeCcache, "7\n"},

		// PackageManagerType
		{"PackageManagerType Npm", PackageManagerNpm, "0\n"},
		{"PackageManagerType Pnpm", PackageManagerPnpm, "1\n"},
		{"PackageManagerType Yarn", PackageManagerYarn, "2\n"},
		{"PackageManagerType Bun", PackageManagerBun, "3\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Marshal() = %q, want %q", string(data), tt.expected)
			}
		})
	}
}

// enumUnmarshalTestCase represents a single test case for enum YAML unmarshaling.
type enumUnmarshalTestCase struct {
	name     string
	input    string
	target   any
	expected any
}

// enumValueTestCase represents a single test case for enum method testing.
type enumValueTestCase[T comparable] struct {
	name     string
	value    any
	expected T
}

// enumTypeInfo holds metadata about an enum type for test case generation.
type enumTypeInfo struct {
	name     string
	typeName string
	values   []string
}

// enumTypeDefinitions contains all enum type definitions for test case generation.
var enumTypeDefinitions = []enumTypeInfo{
	{name: "CacheCleanupMode", typeName: "CacheCleanupMode", values: []string{"DISABLED", "ENABLED", "0", "1"}},
	{name: "DockerPruneMode", typeName: "DockerPruneMode", values: []string{"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS", "0", "1", "2", "3", "4"}},
	{name: "BuildToolType", typeName: "BuildToolType", values: []string{"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA", "0", "1", "2", "3", "4", "5"}},
	{name: "CacheType", typeName: "CacheType", values: []string{"SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW", "PIP", "NPM", "YARN", "CCACHE", "0", "1", "2", "3", "4", "5", "6", "7"}},
	{name: "PackageManagerType", typeName: "PackageManagerType", values: []string{"NPM", "PNPM", "YARN", "BUN", "0", "1", "2", "3"}},
}

// generateEnumUnmarshalTestCases generates test cases for enum YAML unmarshaling.
func generateEnumUnmarshalTestCases(useInt bool) []enumUnmarshalTestCase {
	var cases []enumUnmarshalTestCase

	for _, enumType := range enumTypeDefinitions {
		var stringVals []string
		count := len(enumType.values) / 2
		if useInt {
			for i := 0; i < count; i++ {
				stringVals = append(stringVals, fmt.Sprintf("%d", i))
			}
		} else {
			stringVals = enumType.values[:count]
		}

		var typePtr any
		switch enumType.typeName {
		case "CacheCleanupMode":
			typePtr = new(CacheCleanupMode)
		case "DockerPruneMode":
			typePtr = new(DockerPruneMode)
		case "BuildToolType":
			typePtr = new(BuildToolType)
		case "CacheType":
			typePtr = new(CacheType)
		case "PackageManagerType":
			typePtr = new(PackageManagerType)
		}

		for i, val := range stringVals {
			var expected any
			switch enumType.typeName {
			case "CacheCleanupMode":
				expected = CacheCleanupMode(i)
			case "DockerPruneMode":
				expected = DockerPruneMode(i)
			case "BuildToolType":
				expected = BuildToolType(i)
			case "CacheType":
				expected = CacheType(i)
			case "PackageManagerType":
				expected = PackageManagerType(i)
			}

			suffix := "string"
			if useInt {
				suffix = "int"
			}
			cases = append(cases, enumUnmarshalTestCase{
				name:     fmt.Sprintf("%s %s %s", enumType.name, val, suffix),
				input:    val,
				target:   typePtr,
				expected: expected,
			})
		}
	}

	return cases
}

// enumStringTestCases contains test cases for string-based YAML unmarshaling.
var enumStringTestCases = generateEnumUnmarshalTestCases(false)

// enumIntTestCases contains test cases for integer-based YAML unmarshaling.
var enumIntTestCases = generateEnumUnmarshalTestCases(true)

// runEnumMethodTests executes common test logic for enum method testing.
func runEnumMethodTests[T comparable](
	t *testing.T,
	tests []enumValueTestCase[T],
	extract func(any) T,
	methodName string,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := extract(tt.value)
			if actual != tt.expected {
				t.Errorf("%s = %v, want %v", methodName, actual, tt.expected)
			}
		})
	}
}

// enumStringer is an interface for enum types that can return their string representation.
type enumStringer interface {
	String() string
}

// enumValidator is an interface for enum types that can validate themselves.
type enumValidator interface {
	IsValid() bool
}

// extractEnumString extracts the string representation from any enum value.
// Returns empty string if the type is not supported.
func extractEnumString(v any) string {
	switch val := v.(type) {
	case CacheCleanupMode:
		return val.String()
	case DockerPruneMode:
		return val.String()
	case BuildToolType:
		return val.String()
	case CacheType:
		return val.String()
	case PackageManagerType:
		return val.String()
	}

	return ""
}

// extractEnumValidity extracts the validity status from any enum value.
// Returns false if the type is not supported.
func extractEnumValidity(v any) bool {
	switch val := v.(type) {
	case CacheCleanupMode:
		return val.IsValid()
	case DockerPruneMode:
		return val.IsValid()
	case BuildToolType:
		return val.IsValid()
	case CacheType:
		return val.IsValid()
	case PackageManagerType:
		return val.IsValid()
	}

	return false
}

// runEnumYAMLUnmarshalingTests executes the common test logic for enum unmarshaling.
func runEnumYAMLUnmarshalingTests(t *testing.T, tests []enumUnmarshalTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yaml.Unmarshal([]byte(tt.input), tt.target)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Get the actual value by dereferencing the pointer
			var actual any

			switch v := tt.target.(type) {
			case *CacheCleanupMode:
				actual = *v
			case *DockerPruneMode:
				actual = *v
			case *BuildToolType:
				actual = *v
			case *CacheType:
				actual = *v
			case *PackageManagerType:
				actual = *v
			}

			if actual != tt.expected {
				t.Errorf("Unmarshal() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

// TestEnumYAMLUnmarshalingFromString tests that all enum types can be unmarshaled from YAML strings.
func TestEnumYAMLUnmarshalingFromString(t *testing.T) {
	runEnumYAMLUnmarshalingTests(t, enumStringTestCases)
}

// TestEnumYAMLUnmarshalingFromInt tests that all enum types can be unmarshaled from YAML integers.
func TestEnumYAMLUnmarshalingFromInt(t *testing.T) {
	runEnumYAMLUnmarshalingTests(t, enumIntTestCases)
}

// TestEnumStringMethod tests that all enum types implement String() correctly.
func TestEnumStringMethod(t *testing.T) {
	tests := []enumValueTestCase[string]{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", CacheCleanupDisabled, "DISABLED"},
		{"CacheCleanupMode Enabled", CacheCleanupEnabled, "ENABLED"},
		{"CacheCleanupMode Invalid", CacheCleanupMode(99), "UNKNOWN"},

		// DockerPruneMode
		{"DockerPruneMode All", DockerPruneAll, "ALL"},
		{"DockerPruneMode Images", DockerPruneImages, "IMAGES"},
		{"DockerPruneMode Containers", DockerPruneContainers, "CONTAINERS"},
		{"DockerPruneMode Volumes", DockerPruneVolumes, "VOLUMES"},
		{"DockerPruneMode Builds", DockerPruneBuilds, "BUILDS"},
		{"DockerPruneMode Invalid", DockerPruneMode(99), "UNKNOWN"},

		// BuildToolType
		{"BuildToolType Go", BuildToolGo, "GO"},
		{"BuildToolType Rust", BuildToolRust, "RUST"},
		{"BuildToolType Node", BuildToolNode, "NODE"},
		{"BuildToolType Python", BuildToolPython, "PYTHON"},
		{"BuildToolType Java", BuildToolJava, "JAVA"},
		{"BuildToolType Scala", BuildToolScala, "SCALA"},
		{"BuildToolType Invalid", BuildToolType(99), "UNKNOWN"},

		// CacheType
		{"CacheType Spotlight", CacheTypeSpotlight, "SPOTLIGHT"},
		{"CacheType Xcode", CacheTypeXcode, "XCODE"},
		{"CacheType Cocoapods", CacheTypeCocoapods, "COCOAPODS"},
		{"CacheType Homebrew", CacheTypeHomebrew, "HOMEBREW"},
		{"CacheType Pip", CacheTypePip, "PIP"},
		{"CacheType Npm", CacheTypeNpm, "NPM"},
		{"CacheType Yarn", CacheTypeYarn, "YARN"},
		{"CacheType Ccache", CacheTypeCcache, "CCACHE"},
		{"CacheType Invalid", CacheType(99), "UNKNOWN"},

		// PackageManagerType
		{"PackageManagerType Npm", PackageManagerNpm, "NPM"},
		{"PackageManagerType Pnpm", PackageManagerPnpm, "PNPM"},
		{"PackageManagerType Yarn", PackageManagerYarn, "YARN"},
		{"PackageManagerType Bun", PackageManagerBun, "BUN"},
		{"PackageManagerType Invalid", PackageManagerType(99), "UNKNOWN"},
	}

	runEnumMethodTests(t, tests, extractEnumString, "String()")
}

// TestEnumIsValidMethod tests that all enum types implement IsValid() correctly.
func TestEnumIsValidMethod(t *testing.T) {
	tests := []enumValueTestCase[bool]{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", CacheCleanupDisabled, true},
		{"CacheCleanupMode Enabled", CacheCleanupEnabled, true},
		{"CacheCleanupMode Invalid", CacheCleanupMode(99), false},

		// DockerPruneMode
		{"DockerPruneMode All", DockerPruneAll, true},
		{"DockerPruneMode Images", DockerPruneImages, true},
		{"DockerPruneMode Containers", DockerPruneContainers, true},
		{"DockerPruneMode Volumes", DockerPruneVolumes, true},
		{"DockerPruneMode Builds", DockerPruneBuilds, true},
		{"DockerPruneMode Invalid", DockerPruneMode(99), false},

		// BuildToolType
		{"BuildToolType Go", BuildToolGo, true},
		{"BuildToolType Rust", BuildToolRust, true},
		{"BuildToolType Node", BuildToolNode, true},
		{"BuildToolType Python", BuildToolPython, true},
		{"BuildToolType Java", BuildToolJava, true},
		{"BuildToolType Scala", BuildToolScala, true},
		{"BuildToolType Invalid", BuildToolType(99), false},

		// CacheType
		{"CacheType Spotlight", CacheTypeSpotlight, true},
		{"CacheType Xcode", CacheTypeXcode, true},
		{"CacheType Cocoapods", CacheTypeCocoapods, true},
		{"CacheType Homebrew", CacheTypeHomebrew, true},
		{"CacheType Pip", CacheTypePip, true},
		{"CacheType Npm", CacheTypeNpm, true},
		{"CacheType Yarn", CacheTypeYarn, true},
		{"CacheType Ccache", CacheTypeCcache, true},
		{"CacheType Invalid", CacheType(99), false},

		// PackageManagerType
		{"PackageManagerType Npm", PackageManagerNpm, true},
		{"PackageManagerType Pnpm", PackageManagerPnpm, true},
		{"PackageManagerType Yarn", PackageManagerYarn, true},
		{"PackageManagerType Bun", PackageManagerBun, true},
		{"PackageManagerType Invalid", PackageManagerType(99), false},
	}

	runEnumMethodTests(t, tests, extractEnumValidity, "IsValid()")
}

// testBuildCacheSettings creates a BuildCacheSettings with Java and Scala tools.
// Used for testing YAML marshaling/unmarshaling.
func testBuildCacheSettings() *BuildCacheSettings {
	return &BuildCacheSettings{
		ToolTypes: []BuildToolType{BuildToolJava, BuildToolScala},
		OlderThan: "30d",
	}
}

// testSystemCacheSettings creates a SystemCacheSettings with Spotlight and Xcode caches.
// Used for testing YAML marshaling/unmarshaling.
func testSystemCacheSettings() *SystemCacheSettings {
	return &SystemCacheSettings{
		CacheTypes: []CacheType{CacheTypeSpotlight, CacheTypeXcode},
		OlderThan:  "30d",
	}
}

// TestOperationSettingsWithEnums tests that OperationSettings can be marshaled/unmarshaled with enums.
func TestOperationSettingsWithEnums(t *testing.T) {
	settings := &OperationSettings{
		NodePackages: &NodePackagesSettings{
			PackageManagers: []PackageManagerType{
				PackageManagerNpm,
				PackageManagerPnpm,
			},
		},
		BuildCache:  testBuildCacheSettings(),
		Docker:      &DockerSettings{PruneMode: DockerPruneAll},
		SystemCache: testSystemCacheSettings(),
	}

	// Marshal to YAML
	data, err := yaml.Marshal(settings)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal from YAML
	var unmarshaled OperationSettings
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify NodePackages
	if len(unmarshaled.NodePackages.PackageManagers) != 2 {
		t.Errorf(
			"NodePackages.PackageManagers length = %d, want 2",
			len(unmarshaled.NodePackages.PackageManagers),
		)
	}

	if unmarshaled.NodePackages.PackageManagers[0] != PackageManagerNpm {
		t.Errorf("NodePackages.PackageManagers[0] = %v, want %v",
			unmarshaled.NodePackages.PackageManagers[0], PackageManagerNpm)
	}

	// Verify BuildCache
	if len(unmarshaled.BuildCache.ToolTypes) != 2 {
		t.Errorf("BuildCache.ToolTypes length = %d, want 2", len(unmarshaled.BuildCache.ToolTypes))
	}

	if unmarshaled.BuildCache.ToolTypes[0] != BuildToolJava {
		t.Errorf(
			"BuildCache.ToolTypes[0] = %v, want %v",
			unmarshaled.BuildCache.ToolTypes[0],
			BuildToolJava,
		)
	}

	// Verify Docker
	if unmarshaled.Docker.PruneMode != DockerPruneAll {
		t.Errorf("Docker.PruneMode = %v, want %v", unmarshaled.Docker.PruneMode, DockerPruneAll)
	}

	// Verify SystemCache
	if len(unmarshaled.SystemCache.CacheTypes) != 2 {
		t.Errorf(
			"SystemCache.CacheTypes length = %d, want 2",
			len(unmarshaled.SystemCache.CacheTypes),
		)
	}

	if unmarshaled.SystemCache.CacheTypes[0] != CacheTypeSpotlight {
		t.Errorf(
			"SystemCache.CacheTypes[0] = %v, want %v",
			unmarshaled.SystemCache.CacheTypes[0],
			CacheTypeSpotlight,
		)
	}
}

// TestEnumErrorMessages tests that invalid enum values produce helpful error messages.
func TestEnumErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantStrings []string
	}{
		{
			name:  "invalid DockerPruneMode integer",
			input: "99",
			wantStrings: []string{
				"invalid docker prune mode value: 99",
				"Strings:",
				"Integers:",
				"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS",
				"0", "1", "2", "3", "4",
				"docs/YAML_ENUM_FORMATS.md",
			},
		},
		{
			name:  "invalid BuildToolType integer",
			input: "99",
			wantStrings: []string{
				"invalid build tool type value: 99",
				"Strings:",
				"Integers:",
				"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA",
				"0", "1", "2", "3", "4", "5",
				"docs/YAML_ENUM_FORMATS.md",
			},
		},
		{
			name:  "invalid ProfileStatus binary enum string",
			input: "DISAABLED",
			wantStrings: []string{
				"invalid profile status value: DISAABLED",
				"Strings:",
				"Integers:",
				"DISABLED", "ENABLED",
				"0", "1",
				"docs/YAML_ENUM_FORMATS.md",
			},
		},
		{
			name:  "invalid CacheCleanupMode binary enum integer",
			input: "99",
			wantStrings: []string{
				"invalid cache cleanup mode value: 99",
				"Strings:",
				"Integers:",
				"DISABLED", "ENABLED",
				"0", "1",
				"docs/YAML_ENUM_FORMATS.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				dockerMode    DockerPruneMode
				buildTool     BuildToolType
				profileStatus ProfileStatus
				cacheMode     CacheCleanupMode
				err           error
			)

			// Test appropriate enum type based on input

			switch {
			case strings.Contains(tt.name, "DockerPruneMode"):
				err = yaml.Unmarshal([]byte(tt.input), &dockerMode)
			case strings.Contains(tt.name, "BuildToolType"):
				err = yaml.Unmarshal([]byte(tt.input), &buildTool)
			case strings.Contains(tt.name, "ProfileStatus"):
				err = yaml.Unmarshal([]byte(tt.input), &profileStatus)
			case strings.Contains(tt.name, "CacheCleanupMode"):
				err = yaml.Unmarshal([]byte(tt.input), &cacheMode)
			}

			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			errMsg := err.Error()
			t.Logf("Error message:\n%s\n", errMsg)

			// Verify all expected strings are in error message
			for _, want := range tt.wantStrings {
				if !contains(errMsg, want) {
					t.Errorf(
						"Error message does not contain expected string %q\nGot: %s",
						want,
						errMsg,
					)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
