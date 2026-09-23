package enums_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	"gopkg.in/yaml.v3"
)

// TestEnumYAMLMarshaling tests that all enum types can be properly marshaled to YAML.
func TestEnumYAMLMarshaling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", enums.CacheCleanupDisabled, "DISABLED\n"},
		{"CacheCleanupMode Enabled", enums.CacheCleanupEnabled, "ENABLED\n"},

		// DockerPruneMode
		{"DockerPruneMode All", enums.DockerPruneAll, "ALL\n"},
		{"DockerPruneMode Images", enums.DockerPruneImages, "IMAGES\n"},
		{"DockerPruneMode Containers", enums.DockerPruneContainers, "CONTAINERS\n"},
		{"DockerPruneMode Volumes", enums.DockerPruneVolumes, "VOLUMES\n"},
		{"DockerPruneMode Builds", enums.DockerPruneBuilds, "BUILDS\n"},

		// BuildToolType
		{"BuildToolType Go", enums.BuildToolGo, "GO\n"},
		{"BuildToolType Rust", enums.BuildToolRust, "RUST\n"},
		{"BuildToolType Node", enums.BuildToolNode, "NODE\n"},
		{"BuildToolType Python", enums.BuildToolPython, "PYTHON\n"},
		{"BuildToolType Java", enums.BuildToolJava, "JAVA\n"},
		{"BuildToolType Scala", enums.BuildToolScala, "SCALA\n"},

		// CacheType
		{"CacheType Spotlight", enums.CacheTypeSpotlight, "SPOTLIGHT\n"},
		{"CacheType Xcode", enums.CacheTypeXcode, "XCODE\n"},
		{"CacheType Cocoapods", enums.CacheTypeCocoapods, "COCOAPODS\n"},
		{"CacheType Homebrew", enums.CacheTypeHomebrew, "HOMEBREW\n"},
		{"CacheType Pip", enums.CacheTypePip, "PIP\n"},
		{"CacheType Npm", enums.CacheTypeNpm, "NPM\n"},
		{"CacheType Yarn", enums.CacheTypeYarn, "YARN\n"},
		{"CacheType Ccache", enums.CacheTypeCcache, "CCACHE\n"},

		// PackageManagerType
		{"PackageManagerType Npm", enums.PackageManagerNpm, "NPM\n"},
		{"PackageManagerType Pnpm", enums.PackageManagerPnpm, "PNPM\n"},
		{"PackageManagerType Yarn", enums.PackageManagerYarn, "YARN\n"},
		{"PackageManagerType Bun", enums.PackageManagerBun, "BUN\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	{
		name:     "CacheCleanupMode",
		typeName: "CacheCleanupMode",
		values:   []string{"DISABLED", "ENABLED", "0", "1"},
	},
	{
		name:     "DockerPruneMode",
		typeName: "DockerPruneMode",
		values: []string{
			"ALL",
			"IMAGES",
			"CONTAINERS",
			"VOLUMES",
			"BUILDS",
			"0",
			"1",
			"2",
			"3",
			"4",
		},
	},
	{
		name:     "BuildToolType",
		typeName: "BuildToolType",
		values: []string{
			"GO",
			"RUST",
			"NODE",
			"PYTHON",
			"JAVA",
			"SCALA",
			"0",
			"1",
			"2",
			"3",
			"4",
			"5",
		},
	},
	{
		name:     "CacheType",
		typeName: "CacheType",
		values: []string{
			"SPOTLIGHT",
			"XCODE",
			"COCOAPODS",
			"HOMEBREW",
			"PIP",
			"NPM",
			"YARN",
			"CCACHE",
			"XDG_CACHE",
			"THUMBNAILS",
			"PUPPETEER",
			"TERRAFORM",
			"GRADLE_WRAPPER",
			"KONAN",
			"RUSTUP",
			"GOPLS",
			"GOIMPORTS",
			"JETBRAINS",
			"BUN_CACHE",
			"PLAYWRIGHT",
			"MOZILLA",
			"NIX_CACHE",
			"ZIG",
			"UV",
			"TINYGO",
			"MESA_SHADER",
			"COMGR",
			"0",
			"1",
			"2",
			"3",
			"4",
			"5",
			"6",
			"7",
			"8",
			"9",
			"10",
			"11",
			"12",
			"13",
			"14",
			"15",
			"16",
			"17",
			"18",
			"19",
			"20",
			"21",
			"22",
			"23",
			"24",
			"25",
			"26",
		},
	},
	{
		name:     "PackageManagerType",
		typeName: "PackageManagerType",
		values:   []string{"NPM", "PNPM", "YARN", "BUN", "0", "1", "2", "3"},
	},
}

// generateEnumUnmarshalTestCases generates test cases for enum YAML unmarshaling.
func generateEnumUnmarshalTestCases(useInt bool) []enumUnmarshalTestCase {
	var cases []enumUnmarshalTestCase

	for _, enumType := range enumTypeDefinitions {
		var stringVals []string

		count := len(enumType.values) / 2
		if useInt {
			for i := range count {
				stringVals = append(stringVals, strconv.Itoa(i))
			}
		} else {
			stringVals = enumType.values[:count]
		}

		var typePtr any

		switch enumType.typeName {
		case "CacheCleanupMode":
			typePtr = new(enums.CacheCleanupMode)
		case "DockerPruneMode":
			typePtr = new(enums.DockerPruneMode)
		case "BuildToolType":
			typePtr = new(enums.BuildToolType)
		case "CacheType":
			typePtr = new(enums.CacheType)
		case "PackageManagerType":
			typePtr = new(enums.PackageManagerType)
		}

		for i, val := range stringVals {
			var expected any

			switch enumType.typeName {
			case "CacheCleanupMode":
				expected = enums.CacheCleanupMode(i)
			case "DockerPruneMode":
				expected = enums.DockerPruneMode(i)
			case "BuildToolType":
				expected = enums.BuildToolType(i)
			case "CacheType":
				expected = enums.CacheType(i)
			case "PackageManagerType":
				expected = enums.PackageManagerType(i)
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
			t.Parallel()

			actual := extract(tt.value)
			if actual != tt.expected {
				t.Errorf("%s = %v, want %v", methodName, actual, tt.expected)
			}
		})
	}
}

// extractEnumString extracts the string representation from any enum value.
// Returns empty string if the type is not supported.
func extractEnumString(v any) string {
	switch val := v.(type) {
	case enums.CacheCleanupMode:
		return val.String()
	case enums.DockerPruneMode:
		return val.String()
	case enums.BuildToolType:
		return val.String()
	case enums.CacheType:
		return val.String()
	case enums.PackageManagerType:
		return val.String()
	}

	return ""
}

// extractEnumValidity extracts the validity status from any enum value.
// Returns false if the type is not supported.
func extractEnumValidity(v any) bool {
	switch val := v.(type) {
	case enums.CacheCleanupMode:
		return val.IsValid()
	case enums.DockerPruneMode:
		return val.IsValid()
	case enums.BuildToolType:
		return val.IsValid()
	case enums.CacheType:
		return val.IsValid()
	case enums.PackageManagerType:
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
			case *enums.CacheCleanupMode:
				actual = *v
			case *enums.DockerPruneMode:
				actual = *v
			case *enums.BuildToolType:
				actual = *v
			case *enums.CacheType:
				actual = *v
			case *enums.PackageManagerType:
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
	t.Parallel()
	runEnumYAMLUnmarshalingTests(t, enumStringTestCases)
}

// TestEnumYAMLUnmarshalingFromInt tests that all enum types can be unmarshaled from YAML integers.
func TestEnumYAMLUnmarshalingFromInt(t *testing.T) {
	t.Parallel()
	runEnumYAMLUnmarshalingTests(t, enumIntTestCases)
}

// TestEnumStringMethod tests that all enum types implement String() correctly.
func TestEnumStringMethod(t *testing.T) {
	t.Parallel()

	tests := []enumValueTestCase[string]{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", enums.CacheCleanupDisabled, "DISABLED"},
		{"CacheCleanupMode Enabled", enums.CacheCleanupEnabled, "ENABLED"},
		{"CacheCleanupMode Invalid", enums.CacheCleanupMode(99), "UNKNOWN"},

		// DockerPruneMode
		{"DockerPruneMode All", enums.DockerPruneAll, "ALL"},
		{"DockerPruneMode Images", enums.DockerPruneImages, "IMAGES"},
		{"DockerPruneMode Containers", enums.DockerPruneContainers, "CONTAINERS"},
		{"DockerPruneMode Volumes", enums.DockerPruneVolumes, "VOLUMES"},
		{"DockerPruneMode Builds", enums.DockerPruneBuilds, "BUILDS"},
		{"DockerPruneMode Invalid", enums.DockerPruneMode(99), "UNKNOWN"},

		// BuildToolType
		{"BuildToolType Go", enums.BuildToolGo, "GO"},
		{"BuildToolType Rust", enums.BuildToolRust, "RUST"},
		{"BuildToolType Node", enums.BuildToolNode, "NODE"},
		{"BuildToolType Python", enums.BuildToolPython, "PYTHON"},
		{"BuildToolType Java", enums.BuildToolJava, "JAVA"},
		{"BuildToolType Scala", enums.BuildToolScala, "SCALA"},
		{"BuildToolType Invalid", enums.BuildToolType(99), "UNKNOWN"},

		// CacheType
		{"CacheType Spotlight", enums.CacheTypeSpotlight, "SPOTLIGHT"},
		{"CacheType Xcode", enums.CacheTypeXcode, "XCODE"},
		{"CacheType Cocoapods", enums.CacheTypeCocoapods, "COCOAPODS"},
		{"CacheType Homebrew", enums.CacheTypeHomebrew, "HOMEBREW"},
		{"CacheType Pip", enums.CacheTypePip, "PIP"},
		{"CacheType Npm", enums.CacheTypeNpm, "NPM"},
		{"CacheType Yarn", enums.CacheTypeYarn, "YARN"},
		{"CacheType Ccache", enums.CacheTypeCcache, "CCACHE"},
		{"CacheType Invalid", enums.CacheType(99), "UNKNOWN"},

		// PackageManagerType
		{"PackageManagerType Npm", enums.PackageManagerNpm, "NPM"},
		{"PackageManagerType Pnpm", enums.PackageManagerPnpm, "PNPM"},
		{"PackageManagerType Yarn", enums.PackageManagerYarn, "YARN"},
		{"PackageManagerType Bun", enums.PackageManagerBun, "BUN"},
		{"PackageManagerType Invalid", enums.PackageManagerType(99), "UNKNOWN"},
	}

	runEnumMethodTests(t, tests, extractEnumString, "String()")
}

// TestEnumIsValidMethod tests that all enum types implement IsValid() correctly.
func TestEnumIsValidMethod(t *testing.T) {
	t.Parallel()

	tests := []enumValueTestCase[bool]{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled", enums.CacheCleanupDisabled, true},
		{"CacheCleanupMode Enabled", enums.CacheCleanupEnabled, true},
		{"CacheCleanupMode Invalid", enums.CacheCleanupMode(99), false},

		// DockerPruneMode
		{"DockerPruneMode All", enums.DockerPruneAll, true},
		{"DockerPruneMode Images", enums.DockerPruneImages, true},
		{"DockerPruneMode Containers", enums.DockerPruneContainers, true},
		{"DockerPruneMode Volumes", enums.DockerPruneVolumes, true},
		{"DockerPruneMode Builds", enums.DockerPruneBuilds, true},
		{"DockerPruneMode Invalid", enums.DockerPruneMode(99), false},

		// BuildToolType
		{"BuildToolType Go", enums.BuildToolGo, true},
		{"BuildToolType Rust", enums.BuildToolRust, true},
		{"BuildToolType Node", enums.BuildToolNode, true},
		{"BuildToolType Python", enums.BuildToolPython, true},
		{"BuildToolType Java", enums.BuildToolJava, true},
		{"BuildToolType Scala", enums.BuildToolScala, true},
		{"BuildToolType Invalid", enums.BuildToolType(99), false},

		// CacheType
		{"CacheType Spotlight", enums.CacheTypeSpotlight, true},
		{"CacheType Xcode", enums.CacheTypeXcode, true},
		{"CacheType Cocoapods", enums.CacheTypeCocoapods, true},
		{"CacheType Homebrew", enums.CacheTypeHomebrew, true},
		{"CacheType Pip", enums.CacheTypePip, true},
		{"CacheType Npm", enums.CacheTypeNpm, true},
		{"CacheType Yarn", enums.CacheTypeYarn, true},
		{"CacheType Ccache", enums.CacheTypeCcache, true},
		{"CacheType Invalid", enums.CacheType(99), false},

		// PackageManagerType
		{"PackageManagerType Npm", enums.PackageManagerNpm, true},
		{"PackageManagerType Pnpm", enums.PackageManagerPnpm, true},
		{"PackageManagerType Yarn", enums.PackageManagerYarn, true},
		{"PackageManagerType Bun", enums.PackageManagerBun, true},
		{"PackageManagerType Invalid", enums.PackageManagerType(99), false},
	}

	runEnumMethodTests(t, tests, extractEnumValidity, "IsValid()")
}

// testBuildCacheSettings creates a BuildCacheSettings with Java and Scala tools.
// Used for testing YAML marshaling/unmarshaling.
func testBuildCacheSettings() *operations.BuildCacheSettings {
	return &operations.BuildCacheSettings{
		ToolTypes: []enums.BuildToolType{enums.BuildToolJava, enums.BuildToolScala},
		OlderThan: "30d",
	}
}

// testSystemCacheSettings creates a SystemCacheSettings with Spotlight and Xcode caches.
// Used for testing YAML marshaling/unmarshaling.
func testSystemCacheSettings() *operations.SystemCacheSettings {
	return &operations.SystemCacheSettings{
		CacheTypes: []enums.CacheType{enums.CacheTypeSpotlight, enums.CacheTypeXcode},
		OlderThan:  "30d",
	}
}

// TestOperationSettingsWithEnums tests that OperationSettings can be marshaled/unmarshaled with enums.
func TestOperationSettingsWithEnums(t *testing.T) {
	t.Parallel()

	settings := &operations.OperationSettings{
		NodePackages: &operations.NodePackagesSettings{
			PackageManagers: []enums.PackageManagerType{
				enums.PackageManagerNpm,
				enums.PackageManagerPnpm,
			},
		},
		BuildCache:  testBuildCacheSettings(),
		Docker:      &operations.DockerSettings{PruneMode: enums.DockerPruneAll},
		SystemCache: testSystemCacheSettings(),
	}

	// Marshal to YAML
	data, err := yaml.Marshal(settings) //nolint:musttag
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal from YAML
	var unmarshaled operations.OperationSettings
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil { //nolint:musttag
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify NodePackages
	if len(unmarshaled.NodePackages.PackageManagers) != 2 {
		t.Errorf(
			"NodePackages.PackageManagers length = %d, want 2",
			len(unmarshaled.NodePackages.PackageManagers),
		)
	}

	if unmarshaled.NodePackages.PackageManagers[0] != enums.PackageManagerNpm {
		t.Errorf("NodePackages.PackageManagers[0] = %v, want %v",
			unmarshaled.NodePackages.PackageManagers[0], enums.PackageManagerNpm)
	}

	// Verify BuildCache
	if len(unmarshaled.BuildCache.ToolTypes) != 2 {
		t.Errorf("BuildCache.ToolTypes length = %d, want 2", len(unmarshaled.BuildCache.ToolTypes))
	}

	if unmarshaled.BuildCache.ToolTypes[0] != enums.BuildToolJava {
		t.Errorf(
			"BuildCache.ToolTypes[0] = %v, want %v",
			unmarshaled.BuildCache.ToolTypes[0],
			enums.BuildToolJava,
		)
	}

	// Verify Docker
	if unmarshaled.Docker.PruneMode != enums.DockerPruneAll {
		t.Errorf("Docker.PruneMode = %v, want %v", unmarshaled.Docker.PruneMode, enums.DockerPruneAll)
	}

	// Verify SystemCache
	if len(unmarshaled.SystemCache.CacheTypes) != 2 {
		t.Errorf(
			"SystemCache.CacheTypes length = %d, want 2",
			len(unmarshaled.SystemCache.CacheTypes),
		)
	}

	if unmarshaled.SystemCache.CacheTypes[0] != enums.CacheTypeSpotlight {
		t.Errorf(
			"SystemCache.CacheTypes[0] = %v, want %v",
			unmarshaled.SystemCache.CacheTypes[0],
			enums.CacheTypeSpotlight,
		)
	}
}

// TestEnumErrorMessages tests that invalid enum values produce helpful error messages.
func TestEnumErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantStrings []string
	}{
		{
			name:  "invalid DockerPruneMode integer",
			input: "99",
			wantStrings: []string{
				"invalid docker prune mode",
				"99",
				"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS",
			},
		},
		{
			name:  "invalid BuildToolType integer",
			input: "99",
			wantStrings: []string{
				"invalid build tool type",
				"99",
				"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA",
			},
		},
		{
			name:  "invalid ProfileStatus binary enum string",
			input: "DISAABLED",
			wantStrings: []string{
				"invalid profile status",
				"DISAABLED",
				"DISABLED", "ENABLED",
			},
		},
		{
			name:  "invalid CacheCleanupMode binary enum integer",
			input: "99",
			wantStrings: []string{
				"invalid cache cleanup mode",
				"99",
				"DISABLED", "ENABLED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				dockerMode    enums.DockerPruneMode
				buildTool     enums.BuildToolType
				profileStatus enums.ProfileStatus
				cacheMode     enums.CacheCleanupMode
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
