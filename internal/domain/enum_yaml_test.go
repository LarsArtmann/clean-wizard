package domain

import (
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

		// VersionManagerType
		{"VersionManagerType Nvm", VersionManagerNvm, "0\n"},
		{"VersionManagerType Pyenv", VersionManagerPyenv, "1\n"},
		{"VersionManagerType Gvm", VersionManagerGvm, "2\n"},
		{"VersionManagerType Rbenv", VersionManagerRbenv, "3\n"},
		{"VersionManagerType Sdkman", VersionManagerSdkman, "4\n"},
		{"VersionManagerType Jenv", VersionManagerJenv, "5\n"},

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

		// VersionManagerType
		{"VersionManagerType Nvm string", "NVM", new(VersionManagerType), VersionManagerNvm},
		{"VersionManagerType Pyenv string", "PYENV", new(VersionManagerType), VersionManagerPyenv},
		{"VersionManagerType Gvm string", "GVM", new(VersionManagerType), VersionManagerGvm},
		{"VersionManagerType Rbenv string", "RBENV", new(VersionManagerType), VersionManagerRbenv},
		{"VersionManagerType Sdkman string", "SDKMAN", new(VersionManagerType), VersionManagerSdkman},
		{"VersionManagerType Jenv string", "JENV", new(VersionManagerType), VersionManagerJenv},

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
			case *VersionManagerType:
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

		// VersionManagerType
		{"VersionManagerType Nvm int", "0", new(VersionManagerType), VersionManagerNvm},
		{"VersionManagerType Pyenv int", "1", new(VersionManagerType), VersionManagerPyenv},
		{"VersionManagerType Gvm int", "2", new(VersionManagerType), VersionManagerGvm},
		{"VersionManagerType Rbenv int", "3", new(VersionManagerType), VersionManagerRbenv},
		{"VersionManagerType Sdkman int", "4", new(VersionManagerType), VersionManagerSdkman},
		{"VersionManagerType Jenv int", "5", new(VersionManagerType), VersionManagerJenv},

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
			case *VersionManagerType:
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

		// VersionManagerType
		{"VersionManagerType Nvm", VersionManagerNvm, "NVM"},
		{"VersionManagerType Pyenv", VersionManagerPyenv, "PYENV"},
		{"VersionManagerType Gvm", VersionManagerGvm, "GVM"},
		{"VersionManagerType Rbenv", VersionManagerRbenv, "RBENV"},
		{"VersionManagerType Sdkman", VersionManagerSdkman, "SDKMAN"},
		{"VersionManagerType Jenv", VersionManagerJenv, "JENV"},
		{"VersionManagerType Invalid", VersionManagerType(99), "UNKNOWN"},

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
			case VersionManagerType:
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

		// VersionManagerType
		{"VersionManagerType Nvm", VersionManagerNvm, true},
		{"VersionManagerType Pyenv", VersionManagerPyenv, true},
		{"VersionManagerType Gvm", VersionManagerGvm, true},
		{"VersionManagerType Rbenv", VersionManagerRbenv, true},
		{"VersionManagerType Sdkman", VersionManagerSdkman, true},
		{"VersionManagerType Jenv", VersionManagerJenv, true},
		{"VersionManagerType Invalid", VersionManagerType(99), false},

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
			case VersionManagerType:
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
		LangVersionManager: &LangVersionManagerSettings{
			ManagerTypes: []VersionManagerType{
				VersionManagerNvm,
				VersionManagerPyenv,
			},
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

	// Verify LangVersionManager
	if len(unmarshaled.LangVersionManager.ManagerTypes) != 2 {
		t.Errorf("LangVersionManager.ManagerTypes length = %d, want 2", len(unmarshaled.LangVersionManager.ManagerTypes))
	}
	if unmarshaled.LangVersionManager.ManagerTypes[0] != VersionManagerNvm {
		t.Errorf("LangVersionManager.ManagerTypes[0] = %v, want %v", unmarshaled.LangVersionManager.ManagerTypes[0], VersionManagerNvm)
	}
}
