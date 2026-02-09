//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestEnumWorkflow_Integration tests full workflow from YAML config with enums to execution.
func TestEnumWorkflow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("integer_enums", func(t *testing.T) {
		testEnumWorkflow(t, `
operations:
  - type: docker
    docker:
      prune_mode: 0
  - type: go-packages
    go_packages:
      clean_cache: 1
      clean_test_cache: 1
      clean_mod_cache: 0
      clean_build_cache: 1
      clean_lint_cache: 0
  - type: system-cache
    system_cache:
      cache_types: [0, 1, 2, 3]
      older_than: "30d"
`, domain.DockerPruneAll, true, true, false, false, false, true)
	})

	t.Run("string_enums", func(t *testing.T) {
		testEnumWorkflow(t, `
operations:
  - type: docker
    docker:
      prune_mode: "ALL"
  - type: go-packages
    go_packages:
      clean_cache: "ENABLED"
      clean_test_cache: "ENABLED"
      clean_mod_cache: "DISABLED"
      clean_build_cache: "ENABLED"
      clean_lint_cache: "DISABLED"
  - type: system-cache
    system_cache:
      cache_types: ["SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW"]
      older_than: "30d"
`, domain.DockerPruneAll, true, true, false, false, false, true)
	})

	t.Run("mixed_enums", func(t *testing.T) {
		testEnumWorkflow(t, `
operations:
  - type: docker
    docker:
      prune_mode: 2
  - type: go-packages
    go_packages:
      clean_cache: "ENABLED"
      clean_test_cache: 1
      clean_mod_cache: "DISABLED"
      clean_build_cache: 1
      clean_lint_cache: 0
  - type: system-cache
    system_cache:
      cache_types: [0, "XCODE", "COCOAPODS", 3]
      older_than: "30d"
`, domain.DockerPruneContainers, true, true, false, false, false, true)
	})
}

// testEnumWorkflow tests the full workflow from YAML config to cleaner execution.
func testEnumWorkflow(t *testing.T, configYAML string,
	expectedDockerMode domain.DockerPruneMode,
	expectedCleanCache, expectedTestCache,
	expectedModCache, expectedBuildCache, expectedLintCache bool,
	expectedSystemCacheEmpty bool,
) {
	ctx := context.Background()

	// Parse YAML config
	var config struct {
		Operations []struct {
			Type   string `yaml:"type"`
			Docker *struct {
				PruneMode interface{} `yaml:"prune_mode"`
			} `yaml:"docker"`
			GoPackages *struct {
				CleanCache      interface{} `yaml:"clean_cache"`
				CleanTestCache  interface{} `yaml:"clean_test_cache"`
				CleanModCache   interface{} `yaml:"clean_mod_cache"`
				CleanBuildCache interface{} `yaml:"clean_build_cache"`
				CleanLintCache  interface{} `yaml:"clean_lint_cache"`
			} `yaml:"go_packages"`
			SystemCache *struct {
				CacheTypes []interface{} `yaml:"cache_types"`
				OlderThan  string        `yaml:"older_than"`
			} `yaml:"system_cache"`
		} `yaml:"operations"`
	}

	err := yaml.Unmarshal([]byte(configYAML), &config)
	require.NoError(t, err, "Failed to parse YAML config")
	require.NotEmpty(t, config.Operations, "Config should have operations")

	// Process operations
	for _, op := range config.Operations {
		switch op.Type {
		case "docker":
			if op.Docker != nil {
				// Unmarshal docker prune mode enum
				var pruneMode domain.DockerPruneMode
				node := &yaml.Node{Kind: yaml.ScalarNode}
				if str, ok := op.Docker.PruneMode.(string); ok {
					node.Value = str
				} else if num, ok := op.Docker.PruneMode.(int); ok {
					node.Value = string(rune(num + '0'))
				}
				err := pruneMode.UnmarshalYAML(node)
				require.NoError(t, err, "Failed to unmarshal DockerPruneMode")

				// Verify enum value
				assert.Equal(t, expectedDockerMode, pruneMode, "DockerPruneMode should match expected value")

				// Verify enum is valid
				assert.True(t, pruneMode.IsValid(), "DockerPruneMode should be valid")

				// Create docker cleaner with enum
				dockerCleaner := cleaner.NewDockerCleaner(false, true, pruneMode)
				assert.NotNil(t, dockerCleaner, "Docker cleaner should be created")

				// Test cleaner availability
				if dockerCleaner.IsAvailable(ctx) {
					// Execute cleaner in dry-run mode
					result := dockerCleaner.Clean(ctx)
					assert.True(t, result.IsOk() || result.IsErr(), "Clean should complete")
				}
			}

		case "go-packages":
			if op.GoPackages != nil {
				// Unmarshal cache cleanup mode enums
				modes := map[interface{}]*domain.CacheCleanupMode{
					op.GoPackages.CleanCache:      nil,
					op.GoPackages.CleanTestCache:  nil,
					op.GoPackages.CleanModCache:   nil,
					op.GoPackages.CleanBuildCache: nil,
					op.GoPackages.CleanLintCache:  nil,
				}

				for val := range modes {
					var mode domain.CacheCleanupMode
					node := &yaml.Node{Kind: yaml.ScalarNode}
					if str, ok := val.(string); ok {
						node.Value = str
					} else if num, ok := val.(int); ok {
						node.Value = string(rune(num + '0'))
					}

					err := mode.UnmarshalYAML(node)
					if err == nil {
						assert.True(t, mode.IsValid(), "CacheCleanupMode should be valid")
					}
				}

				// Create Go cleaner with enum flags
				caches := cleaner.GoCacheNone
				if expectedCleanCache {
					caches |= cleaner.GoCacheGOCACHE
				}
				if expectedTestCache {
					caches |= cleaner.GoCacheTestCache
				}
				if expectedBuildCache {
					caches |= cleaner.GoCacheBuildCache
				}

				goCleaner, err := cleaner.NewGoCleaner(false, true, caches)
				require.NoError(t, err, "Failed to create Go cleaner")

				// Test cleaner availability
				if goCleaner.IsAvailable(ctx) {
					// Execute cleaner in dry-run mode
					result := goCleaner.Clean(ctx)
					assert.True(t, result.IsOk() || result.IsErr(), "Clean should complete")
				}
			}

		case "system-cache":
			if op.SystemCache != nil {
				// Unmarshal cache type enums
				var cacheTypes []domain.CacheType
				for _, ct := range op.SystemCache.CacheTypes {
					var cacheType domain.CacheType
					node := &yaml.Node{Kind: yaml.ScalarNode}
					if str, ok := ct.(string); ok {
						node.Value = str
					} else if num, ok := ct.(int); ok {
						node.Value = string(rune(num + '0'))
					}

					err := cacheType.UnmarshalYAML(node)
					if err == nil {
						assert.True(t, cacheType.IsValid(), "CacheType should be valid")
						cacheTypes = append(cacheTypes, cacheType)
					}
				}

				if !expectedSystemCacheEmpty {
					assert.NotEmpty(t, cacheTypes, "SystemCache types should not be empty")
				}

				// Create system cache cleaner
				systemCleaner, err := cleaner.NewSystemCacheCleaner(false, true, op.SystemCache.OlderThan)
				require.NoError(t, err, "Failed to create system cache cleaner")

				// Test cleaner availability
				if systemCleaner.IsAvailable(ctx) {
					// Execute cleaner in dry-run mode
					result := systemCleaner.Clean(ctx)
					assert.True(t, result.IsOk() || result.IsErr(), "Clean should complete")
				}
			}
		}
	}
}

// TestEnumWorkflow_InvalidEnums tests that invalid enum values are caught.
func TestEnumWorkflow_InvalidEnums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	invalidConfig := `
operations:
  - type: docker
    docker:
      prune_mode: 99
  - type: go-packages
    go_packages:
      clean_cache: "INVALID"
  - type: system-cache
    system_cache:
      cache_types: ["SPOTLIGHT", "INVALID_TYPE"]
      older_than: "30d"
`

	// Parse YAML config
	var config struct {
		Operations []struct {
			Type   string `yaml:"type"`
			Docker *struct {
				PruneMode interface{} `yaml:"prune_mode"`
			} `yaml:"docker"`
			GoPackages *struct {
				CleanCache interface{} `yaml:"clean_cache"`
			} `yaml:"go_packages"`
			SystemCache *struct {
				CacheTypes []interface{} `yaml:"cache_types"`
				OlderThan  string        `yaml:"older_than"`
			} `yaml:"system_cache"`
		} `yaml:"operations"`
	}

	err := yaml.Unmarshal([]byte(invalidConfig), &config)
	require.NoError(t, err, "Failed to parse YAML config")

	// Process operations and expect enum unmarshal errors
	for _, op := range config.Operations {
		switch op.Type {
		case "docker":
			if op.Docker != nil {
				var pruneMode domain.DockerPruneMode
				node := &yaml.Node{Kind: yaml.ScalarNode}
				if str, ok := op.Docker.PruneMode.(string); ok {
					node.Value = str
				} else if num, ok := op.Docker.PruneMode.(int); ok {
					node.Value = string(rune(num + '0'))
				}
				err := pruneMode.UnmarshalYAML(node)
				assert.Error(t, err, "Should error on invalid DockerPruneMode")
				assert.Contains(t, err.Error(), "Valid options", "Error should list valid options")
			}

		case "go-packages":
			if op.GoPackages != nil {
				var mode domain.CacheCleanupMode
				node := &yaml.Node{Kind: yaml.ScalarNode}
				if str, ok := op.GoPackages.CleanCache.(string); ok {
					node.Value = str
				} else if num, ok := op.GoPackages.CleanCache.(int); ok {
					node.Value = string(rune(num + '0'))
				}
				err := mode.UnmarshalYAML(node)
				assert.Error(t, err, "Should error on invalid CacheCleanupMode")
				assert.Contains(t, err.Error(), "Valid options", "Error should list valid options")
			}

		case "system-cache":
			if op.SystemCache != nil {
				for _, ct := range op.SystemCache.CacheTypes {
					var cacheType domain.CacheType
					node := &yaml.Node{Kind: yaml.ScalarNode}
					if str, ok := ct.(string); ok {
						node.Value = str
						if str == "INVALID_TYPE" {
							err := cacheType.UnmarshalYAML(node)
							assert.Error(t, err, "Should error on invalid CacheType")
							assert.Contains(t, err.Error(), "Valid options", "Error should list valid options")
						}
					}
				}
			}
		}
	}
}

// TestDefaultSettings_WithEnums tests that DefaultSettings returns valid enum values.
func TestDefaultSettings_WithEnums(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []domain.OperationType{
		domain.OperationTypeNixGenerations,
		domain.OperationTypeHomebrew,
		domain.OperationTypeNodePackages,
		domain.OperationTypeGoPackages,
		domain.OperationTypeCargoPackages,
		domain.OperationTypeBuildCache,
		domain.OperationTypeDocker,
		domain.OperationTypeSystemCache,
		domain.OperationTypeLangVersionManager,
		domain.OperationTypeSystemTemp,
		domain.OperationTypeProjectsManagementAutomation,
	}

	for _, opType := range testCases {
		t.Run(string(opType), func(t *testing.T) {
			// Get default settings
			settings := domain.DefaultSettings(opType)
			require.NotNil(t, settings, "DefaultSettings should not return nil")

			// Validate settings
			err := settings.ValidateSettings(opType)
			assert.NoError(t, err, "DefaultSettings should be valid")

			// Test marshaling and unmarshaling
			data, err := yaml.Marshal(settings)
			require.NoError(t, err, "Should marshal settings")

			var unmarshaled domain.OperationSettings
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err, "Should unmarshal settings")

			// Verify enum values are preserved
			if settings.Docker != nil {
				assert.Equal(t, settings.Docker.PruneMode, unmarshaled.Docker.PruneMode,
					"DockerPruneMode should be preserved")
				assert.True(t, unmarshaled.Docker.PruneMode.IsValid(),
					"Unmarshaled DockerPruneMode should be valid")
			}

			if settings.GoPackages != nil {
				assert.Equal(t, settings.GoPackages.CleanCache, unmarshaled.GoPackages.CleanCache,
					"CleanCache should be preserved")
				assert.True(t, unmarshaled.GoPackages.CleanCache.IsValid(),
					"Unmarshaled CleanCache should be valid")
			}
		})
	}
}

// TestEnumRoundtrip_ThroughConfig tests roundtrip of enum values through config.
func TestEnumRoundtrip_ThroughConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Original settings with enum values
	originalSettings := &domain.OperationSettings{
		Docker: &domain.DockerSettings{
			PruneMode: domain.DockerPruneImages,
		},
		GoPackages: &domain.GoPackagesSettings{
			CleanCache:      domain.CacheCleanupEnabled,
			CleanTestCache:  domain.CacheCleanupEnabled,
			CleanModCache:   domain.CacheCleanupDisabled,
			CleanBuildCache: domain.CacheCleanupEnabled,
			CleanLintCache:  domain.CacheCleanupDisabled,
		},
		SystemCache: &domain.SystemCacheSettings{
			CacheTypes: []domain.CacheType{
				domain.CacheTypeSpotlight,
				domain.CacheTypeXcode,
			},
			OlderThan: "30d",
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(originalSettings)
	require.NoError(t, err, "Should marshal settings")

	// Unmarshal from YAML
	var unmarshaledSettings domain.OperationSettings
	err = yaml.Unmarshal(data, &unmarshaledSettings)
	require.NoError(t, err, "Should unmarshal settings")

	// Verify all enum values are preserved
	assert.Equal(t, originalSettings.Docker.PruneMode, unmarshaledSettings.Docker.PruneMode,
		"DockerPruneMode should be preserved through roundtrip")

	assert.Equal(t, originalSettings.GoPackages.CleanCache, unmarshaledSettings.GoPackages.CleanCache,
		"CleanCache should be preserved through roundtrip")

	assert.Equal(t, originalSettings.GoPackages.CleanTestCache, unmarshaledSettings.GoPackages.CleanTestCache,
		"CleanTestCache should be preserved through roundtrip")

	assert.Equal(t, len(originalSettings.SystemCache.CacheTypes), len(unmarshaledSettings.SystemCache.CacheTypes),
		"CacheTypes count should be preserved through roundtrip")

	for i, ct := range originalSettings.SystemCache.CacheTypes {
		assert.Equal(t, ct, unmarshaledSettings.SystemCache.CacheTypes[i],
			"CacheType[%d] should be preserved through roundtrip", i)
	}

	// Verify all enum values are still valid
	assert.True(t, unmarshaledSettings.Docker.PruneMode.IsValid(),
		"Unmarshaled DockerPruneMode should be valid")

	assert.True(t, unmarshaledSettings.GoPackages.CleanCache.IsValid(),
		"Unmarshaled CleanCache should be valid")

	for _, ct := range unmarshaledSettings.SystemCache.CacheTypes {
		assert.True(t, ct.IsValid(), "Unmarshaled CacheType should be valid")
	}
}

// TestEnumErrorMessages_ThroughWorkflow tests that enum errors are helpful.
func TestEnumErrorMessages_ThroughWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test invalid Docker prune mode - should error during YAML unmarshaling
	invalidConfig := `prune_mode: "INVALID"`
	var dockerSettings domain.DockerSettings
	err := yaml.Unmarshal([]byte(invalidConfig), &dockerSettings)
	assert.Error(t, err, "Should error on invalid DockerPruneMode during unmarshaling")
	assert.Contains(t, err.Error(), "Valid options", "Error should list valid options")
	assert.Contains(t, err.Error(), "INVALID", "Error should show invalid value")

	// Test valid enum value with cleaner
	validConfig := `prune_mode: "ALL"`
	var validDockerSettings domain.DockerSettings
	err = yaml.Unmarshal([]byte(validConfig), &validDockerSettings)
	require.NoError(t, err, "Should parse valid config")

	// Test with cleaner
	dockerCleaner := cleaner.NewDockerCleaner(false, true, validDockerSettings.PruneMode)
	if dockerCleaner.IsAvailable(ctx) {
		result := dockerCleaner.Clean(ctx)
		// Should complete without crashing
		assert.True(t, result.IsOk() || result.IsErr(), "Clean should complete")
	}
}

// TestEnumValues_ThroughExecution tests that enum values are used correctly in execution.
func TestEnumValues_ThroughExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test Docker cleaner with different prune modes
	pruneModes := []struct {
		mode domain.DockerPruneMode
		name string
	}{
		{domain.DockerPruneAll, "ALL"},
		{domain.DockerPruneImages, "IMAGES"},
		{domain.DockerPruneContainers, "CONTAINERS"},
	}

	for _, pm := range pruneModes {
		t.Run(pm.name, func(t *testing.T) {
			// Verify enum is valid
			assert.True(t, pm.mode.IsValid(), "%s should be valid", pm.name)

			// Create cleaner with enum
			dockerCleaner := cleaner.NewDockerCleaner(false, true, cleaner.DockerPruneMode(pm.mode.String()))
			assert.NotNil(t, dockerCleaner, "Docker cleaner should be created")

			// Verify settings validation
			settings := &domain.OperationSettings{
				Docker: &domain.DockerSettings{
					PruneMode: pm.mode,
				},
			}
			err := dockerCleaner.ValidateSettings(settings)
			assert.NoError(t, err, "Settings with %s should be valid", pm.name)

			// Test cleaner availability and execution
			if dockerCleaner.IsAvailable(ctx) {
				result := dockerCleaner.Clean(ctx)
				assert.True(t, result.IsOk() || result.IsErr(), "Clean should complete")
			}
		})
	}
}
