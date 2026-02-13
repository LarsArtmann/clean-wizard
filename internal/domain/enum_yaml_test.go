package domain

import (
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

// TestEnumYAMLUnmarshalingFromString tests that all enum types can be unmarshaled from YAML strings.
func TestEnumYAMLUnmarshalingFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   any
		expected any
	}{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled string", "DISABLED", new(CacheCleanupMode), CacheCleanupDisabled},
		{"CacheCleanupMode Enabled string", "ENABLED", new(CacheCleanupMode), CacheCleanupEnabled},

		// DockerPruneMode
		{"DockerPruneMode All string", "ALL", new(DockerPruneMode), DockerPruneAll},
		{"DockerPruneMode Images string", "IMAGES", new(DockerPruneMode), DockerPruneImages},
		{"DockerPruneMode Containers string", "CONTAINERS", new(DockerPruneMode), DockerPruneContainers},
		{"DockerPruneMode Volumes string", "VOLUMES", new(DockerPruneMode), DockerPruneVolumes},
		{"DockerPruneMode Builds string", "BUILDS", new(DockerPruneMode), DockerPruneBuilds},

		// BuildToolType
		{"BuildToolType Go string", "GO", new(BuildToolType), BuildToolGo},
		{"BuildToolType Rust string", "RUST", new(BuildToolType), BuildToolRust},
		{"BuildToolType Node string", "NODE", new(BuildToolType), BuildToolNode},
		{"BuildToolType Python string", "PYTHON", new(BuildToolType), BuildToolPython},
		{"BuildToolType Java string", "JAVA", new(BuildToolType), BuildToolJava},
		{"BuildToolType Scala string", "SCALA", new(BuildToolType), BuildToolScala},

		// CacheType
		{"CacheType Spotlight string", "SPOTLIGHT", new(CacheType), CacheTypeSpotlight},
		{"CacheType Xcode string", "XCODE", new(CacheType), CacheTypeXcode},
		{"CacheType Cocoapods string", "COCOAPODS", new(CacheType), CacheTypeCocoapods},
		{"CacheType Homebrew string", "HOMEBREW", new(CacheType), CacheTypeHomebrew},
		{"CacheType Pip string", "PIP", new(CacheType), CacheTypePip},
		{"CacheType Npm string", "NPM", new(CacheType), CacheTypeNpm},
		{"CacheType Yarn string", "YARN", new(CacheType), CacheTypeYarn},
		{"CacheType Ccache string", "CCACHE", new(CacheType), CacheTypeCcache},

		// PackageManagerType
		{"PackageManagerType Npm string", "NPM", new(PackageManagerType), PackageManagerNpm},
		{"PackageManagerType Pnpm string", "PNPM", new(PackageManagerType), PackageManagerPnpm},
		{"PackageManagerType Yarn string", "YARN", new(PackageManagerType), PackageManagerYarn},
		{"PackageManagerType Bun string", "BUN", new(PackageManagerType), PackageManagerBun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tt.input), tt.target); err != nil {
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

// TestEnumYAMLUnmarshalingFromInt tests that all enum types can be unmarshaled from YAML integers.
func TestEnumYAMLUnmarshalingFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   any
		expected any
	}{
		// CacheCleanupMode
		{"CacheCleanupMode Disabled int", "0", new(CacheCleanupMode), CacheCleanupDisabled},
		{"CacheCleanupMode Enabled int", "1", new(CacheCleanupMode), CacheCleanupEnabled},

		// DockerPruneMode
		{"DockerPruneMode All int", "0", new(DockerPruneMode), DockerPruneAll},
		{"DockerPruneMode Images int", "1", new(DockerPruneMode), DockerPruneImages},
		{"DockerPruneMode Containers int", "2", new(DockerPruneMode), DockerPruneContainers},
		{"DockerPruneMode Volumes int", "3", new(DockerPruneMode), DockerPruneVolumes},
		{"DockerPruneMode Builds int", "4", new(DockerPruneMode), DockerPruneBuilds},

		// BuildToolType
		{"BuildToolType Go int", "0", new(BuildToolType), BuildToolGo},
		{"BuildToolType Rust int", "1", new(BuildToolType), BuildToolRust},
		{"BuildToolType Node int", "2", new(BuildToolType), BuildToolNode},
		{"BuildToolType Python int", "3", new(BuildToolType), BuildToolPython},
		{"BuildToolType Java int", "4", new(BuildToolType), BuildToolJava},
		{"BuildToolType Scala int", "5", new(BuildToolType), BuildToolScala},

		// CacheType
		{"CacheType Spotlight int", "0", new(CacheType), CacheTypeSpotlight},
		{"CacheType Xcode int", "1", new(CacheType), CacheTypeXcode},
		{"CacheType Cocoapods int", "2", new(CacheType), CacheTypeCocoapods},
		{"CacheType Homebrew int", "3", new(CacheType), CacheTypeHomebrew},
		{"CacheType Pip int", "4", new(CacheType), CacheTypePip},
		{"CacheType Npm int", "5", new(CacheType), CacheTypeNpm},
		{"CacheType Yarn int", "6", new(CacheType), CacheTypeYarn},
		{"CacheType Ccache int", "7", new(CacheType), CacheTypeCcache},

		// PackageManagerType
		{"PackageManagerType Npm int", "0", new(PackageManagerType), PackageManagerNpm},
		{"PackageManagerType Pnpm int", "1", new(PackageManagerType), PackageManagerPnpm},
		{"PackageManagerType Yarn int", "2", new(PackageManagerType), PackageManagerYarn},
		{"PackageManagerType Bun int", "3", new(PackageManagerType), PackageManagerBun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tt.input), tt.target); err != nil {
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

// TestEnumStringMethod tests that all enum types implement String() correctly.
func TestEnumStringMethod(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual string
			switch v := tt.value.(type) {
			case CacheCleanupMode:
				actual = v.String()
			case DockerPruneMode:
				actual = v.String()
			case BuildToolType:
				actual = v.String()
			case CacheType:
				actual = v.String()
			case PackageManagerType:
				actual = v.String()
			}

			if actual != tt.expected {
				t.Errorf("String() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

// TestEnumIsValidMethod tests that all enum types implement IsValid() correctly.
func TestEnumIsValidMethod(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual bool
			switch v := tt.value.(type) {
			case CacheCleanupMode:
				actual = v.IsValid()
			case DockerPruneMode:
				actual = v.IsValid()
			case BuildToolType:
				actual = v.IsValid()
			case CacheType:
				actual = v.IsValid()
			case PackageManagerType:
				actual = v.IsValid()
			}

			if actual != tt.expected {
				t.Errorf("IsValid() = %v, want %v", actual, tt.expected)
			}
		})
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
		BuildCache: &BuildCacheSettings{
			ToolTypes: []BuildToolType{
				BuildToolJava,
				BuildToolScala,
			},
			OlderThan: "30d",
		},
		Docker: &DockerSettings{
			PruneMode: DockerPruneAll,
		},
		SystemCache: &SystemCacheSettings{
			CacheTypes: []CacheType{
				CacheTypeSpotlight,
				CacheTypeXcode,
			},
			OlderThan: "30d",
		},
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
		t.Errorf("NodePackages.PackageManagers length = %d, want 2", len(unmarshaled.NodePackages.PackageManagers))
	}
	if unmarshaled.NodePackages.PackageManagers[0] != PackageManagerNpm {
		t.Errorf("NodePackages.PackageManagers[0] = %v, want %v", unmarshaled.NodePackages.PackageManagers[0], PackageManagerNpm)
	}

	// Verify BuildCache
	if len(unmarshaled.BuildCache.ToolTypes) != 2 {
		t.Errorf("BuildCache.ToolTypes length = %d, want 2", len(unmarshaled.BuildCache.ToolTypes))
	}
	if unmarshaled.BuildCache.ToolTypes[0] != BuildToolJava {
		t.Errorf("BuildCache.ToolTypes[0] = %v, want %v", unmarshaled.BuildCache.ToolTypes[0], BuildToolJava)
	}

	// Verify Docker
	if unmarshaled.Docker.PruneMode != DockerPruneAll {
		t.Errorf("Docker.PruneMode = %v, want %v", unmarshaled.Docker.PruneMode, DockerPruneAll)
	}

	// Verify SystemCache
	if len(unmarshaled.SystemCache.CacheTypes) != 2 {
		t.Errorf("SystemCache.CacheTypes length = %d, want 2", len(unmarshaled.SystemCache.CacheTypes))
	}
	if unmarshaled.SystemCache.CacheTypes[0] != CacheTypeSpotlight {
		t.Errorf("SystemCache.CacheTypes[0] = %v, want %v", unmarshaled.SystemCache.CacheTypes[0], CacheTypeSpotlight)
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
			var dockerMode DockerPruneMode
			var buildTool BuildToolType
			var profileStatus ProfileStatus
			var cacheMode CacheCleanupMode
			var err error

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
					t.Errorf("Error message does not contain expected string %q\nGot: %s", want, errMsg)
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
