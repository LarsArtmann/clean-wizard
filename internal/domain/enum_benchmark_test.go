package domain

import (
	"bytes"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

// BenchmarkMarshalYAML_DockerPruneMode benchmarks marshaling DockerPruneMode enum to YAML.
func BenchmarkMarshalYAML_DockerPruneMode(b *testing.B) {
	testCases := []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_DockerPruneMode_String benchmarks unmarshaling DockerPruneMode from string YAML.
func BenchmarkUnmarshalYAML_DockerPruneMode_String(b *testing.B) {
	testCases := []string{"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result DockerPruneMode
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_DockerPruneMode_Int benchmarks unmarshaling DockerPruneMode from integer YAML.
func BenchmarkUnmarshalYAML_DockerPruneMode_Int(b *testing.B) {
	testCases := []int{0, 1, 2, 3, 4}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("%d", tc), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result DockerPruneMode
				// Use actual YAML unmarshaling with int
				yamlData := fmt.Sprintf("%d", tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalYAML_BuildToolType benchmarks marshaling BuildToolType enum to YAML.
func BenchmarkMarshalYAML_BuildToolType(b *testing.B) {
	testCases := []BuildToolType{
		BuildToolGo,
		BuildToolRust,
		BuildToolNode,
		BuildToolPython,
		BuildToolJava,
		BuildToolScala,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_BuildToolType_String benchmarks unmarshaling BuildToolType from string YAML.
func BenchmarkUnmarshalYAML_BuildToolType_String(b *testing.B) {
	testCases := []string{"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result BuildToolType
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalYAML_CacheType benchmarks marshaling CacheType enum to YAML.
func BenchmarkMarshalYAML_CacheType(b *testing.B) {
	testCases := []CacheType{
		CacheTypeSpotlight,
		CacheTypeXcode,
		CacheTypeCocoapods,
		CacheTypeHomebrew,
		CacheTypePip,
		CacheTypeNpm,
		CacheTypeYarn,
		CacheTypeCcache,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_CacheType benchmarks unmarshaling CacheType enum.
func BenchmarkUnmarshalYAML_CacheType_String(b *testing.B) {
	testCases := []string{"SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW", "PIP", "NPM", "YARN", "CCACHE"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result CacheType
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalYAML_VersionManagerType benchmarks marshaling VersionManagerType enum to YAML.
func BenchmarkMarshalYAML_VersionManagerType(b *testing.B) {
	testCases := []VersionManagerType{
		VersionManagerNvm,
		VersionManagerPyenv,
		VersionManagerGvm,
		VersionManagerRbenv,
		VersionManagerSdkman,
		VersionManagerJenv,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_VersionManagerType benchmarks unmarshaling VersionManagerType enum.
func BenchmarkUnmarshalYAML_VersionManagerType_String(b *testing.B) {
	testCases := []string{"NVM", "PYENV", "GVM", "RBENV", "SDKMAN", "JENV"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result VersionManagerType
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalYAML_PackageManagerType benchmarks marshaling PackageManagerType enum to YAML.
func BenchmarkMarshalYAML_PackageManagerType(b *testing.B) {
	testCases := []PackageManagerType{
		PackageManagerNpm,
		PackageManagerPnpm,
		PackageManagerYarn,
		PackageManagerBun,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_PackageManagerType benchmarks unmarshaling PackageManagerType enum.
func BenchmarkUnmarshalYAML_PackageManagerType_String(b *testing.B) {
	testCases := []string{"NPM", "PNPM", "YARN", "BUN"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result PackageManagerType
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMarshalYAML_BinaryEnums benchmarks marshaling binary enums to YAML.
func BenchmarkMarshalYAML_CacheCleanupMode(b *testing.B) {
	testCases := []CacheCleanupMode{CacheCleanupDisabled, CacheCleanupEnabled}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_BinaryEnums_String benchmarks unmarshaling binary enums from string YAML.
func BenchmarkUnmarshalYAML_CacheCleanupMode_String(b *testing.B) {
	testCases := []string{"DISABLED", "ENABLED"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result CacheCleanupMode
				// Use actual YAML unmarshaling with string
				yamlData := fmt.Sprintf(`"%s"`, tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_BinaryEnums_Int benchmarks unmarshaling binary enums from integer YAML.
func BenchmarkUnmarshalYAML_CacheCleanupMode_Int(b *testing.B) {
	testCases := []int{0, 1}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("%d", tc), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var result CacheCleanupMode
				// Use actual YAML unmarshaling with int
				yamlData := fmt.Sprintf("%d", tc)
				if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRoundTrip_DockerPruneMode benchmarks full marshal→unmarshal round-trip for DockerPruneMode.
func BenchmarkRoundTrip_DockerPruneMode(b *testing.B) {
	testCases := []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				marshaled, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}

				yamlBytes, err := yaml.Marshal(marshaled)
				if err != nil {
					b.Fatal(err)
				}

				var result DockerPruneMode
				if err := yaml.Unmarshal(yamlBytes, &result); err != nil {
					b.Fatal(err)
				}

				if result != tc {
					b.Fatalf("round-trip failed: got %v, want %v", result, tc)
				}
			}
		})
	}
}

// BenchmarkRoundTrip_CacheCleanupMode benchmarks full marshal→unmarshal round-trip for CacheCleanupMode.
func BenchmarkRoundTrip_CacheCleanupMode(b *testing.B) {
	testCases := []CacheCleanupMode{CacheCleanupDisabled, CacheCleanupEnabled}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				marshaled, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}

				yamlBytes, err := yaml.Marshal(marshaled)
				if err != nil {
					b.Fatal(err)
				}

				var result CacheCleanupMode
				if err := yaml.Unmarshal(yamlBytes, &result); err != nil {
					b.Fatal(err)
				}

				if result != tc {
					b.Fatalf("round-trip failed: got %v, want %v", result, tc)
				}
			}
		})
	}
}

// BenchmarkFullConfigMarshal benchmarks marshaling a complete configuration with enums.
func BenchmarkFullConfigMarshal(b *testing.B) {
	config := &Config{
		Version:      "1.0.0",
		SafeMode:     SafeModeEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library"},
		Profiles: map[string]*Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Enabled:     ProfileStatusEnabled,
				Operations: []CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   RiskLevelType(RiskLevelLowType),
						Enabled:     ProfileStatusEnabled,
						Settings: &OperationSettings{
							NixGenerations: &NixGenerationsSettings{
								Generations: 1,
								Optimize:    OptimizationModeDisabled,
								DryRun:      ExecutionModeNormal,
							},
						},
					},
					{
						Name:        "docker",
						Description: "Clean Docker resources",
						RiskLevel:   RiskLevelType(RiskLevelMediumType),
						Enabled:     ProfileStatusEnabled,
						Settings: &OperationSettings{
							Docker: &DockerSettings{
								PruneMode: DockerPruneAll,
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := yaml.Marshal(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullConfigUnmarshal benchmarks unmarshaling a complete configuration with enums.
func BenchmarkFullConfigUnmarshal(b *testing.B) {
	yamlConfig := `version: "1.0.0"
safe_mode: 1
max_disk_usage: 50
protected:
  - "/System"
  - "/Library"
profiles:
  daily:
    name: "daily"
    description: "Daily cleanup"
    enabled: 1
    operations:
      - name: "nix-generations"
        description: "Clean Nix generations"
        risk_level: 0
        enabled: 1
        settings:
          nix_generations:
            generations: 1
            optimize: 0
            dry_run: 1
      - name: "docker"
        description: "Clean Docker resources"
        risk_level: 1
        enabled: 1
        settings:
          docker:
            prune_mode: 0`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var config Config
		if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullConfigRoundTrip benchmarks marshal→unmarshal round-trip for complete configuration.
func BenchmarkFullConfigRoundTrip(b *testing.B) {
	config := &Config{
		Version:      "1.0.0",
		SafeMode:     SafeModeEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library"},
		Profiles: map[string]*Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Enabled:     ProfileStatusEnabled,
				Operations: []CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   RiskLevelType(RiskLevelLowType),
						Enabled:     ProfileStatusEnabled,
						Settings: &OperationSettings{
							NixGenerations: &NixGenerationsSettings{
								Generations: 1,
								Optimize:    OptimizationModeDisabled,
								DryRun:      ExecutionModeNormal,
							},
						},
					},
					{
						Name:        "docker",
						Description: "Clean Docker resources",
						RiskLevel:   RiskLevelType(RiskLevelMediumType),
						Enabled:     ProfileStatusEnabled,
						Settings: &OperationSettings{
							Docker: &DockerSettings{
								PruneMode: DockerPruneAll,
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		yamlBytes, err := yaml.Marshal(config)
		if err != nil {
			b.Fatal(err)
		}

		var result Config
		if err := yaml.Unmarshal(yamlBytes, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEnumString benchmarks the String() method for all enums.
func BenchmarkEnumString_DockerPruneMode(b *testing.B) {
	testCases := []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}

	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.String()
			}
		})
	}
}

// BenchmarkEnumIsValid benchmarks the IsValid() method for all enums.
func BenchmarkEnumIsValid_DockerPruneMode(b *testing.B) {
	testCases := []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
		99, // invalid value
	}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("%d", tc), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.IsValid()
			}
		})
	}
}

// BenchmarkYAMLDecodeRaw benchmarks raw YAML decoding for comparison.
func BenchmarkYAMLDecodeRaw_Int(b *testing.B) {
	data := []byte("0")
	var result int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := yaml.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkYAMLDecodeRaw_String benchmarks raw YAML string decoding for comparison.
func BenchmarkYAMLDecodeRaw_String(b *testing.B) {
	data := []byte("ALL")
	var result string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := yaml.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNodeDecode benchmarks yaml.Node.Decode performance.
func BenchmarkNodeDecode_Int(b *testing.B) {
	yamlData := "0"
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(yamlData), node); err != nil {
		b.Fatal(err)
	}
	var result int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := node.Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNodeDecode_String benchmarks yaml.Node.Decode for strings.
func BenchmarkNodeDecode_String(b *testing.B) {
	yamlData := `"ALL"`
	node := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(yamlData), node); err != nil {
		b.Fatal(err)
	}
	var result string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := node.Decode(&result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStringComparison benchmarks string comparison operations used in enum unmarshaling.
func BenchmarkStringComparison_CaseInsensitive(b *testing.B) {
	testCases := []string{"ALL", "all", "All", "aLl", "aLL"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			target := "ALL"
			for i := 0; i < b.N; i++ {
				_ = tc == target
			}
		})
	}
}

// BenchmarkBufferWrite benchmarks writing YAML to buffer.
func BenchmarkBufferWrite_Config(b *testing.B) {
	config := &Config{
		Version:      "1.0.0",
		SafeMode:     SafeModeEnabled,
		MaxDiskUsage: 50,
		Protected:    []string{"/System", "/Library"},
		Profiles: map[string]*Profile{
			"daily": {
				Name:        "daily",
				Description: "Daily cleanup",
				Enabled:     ProfileStatusEnabled,
				Operations: []CleanupOperation{
					{
						Name:        "nix-generations",
						Description: "Clean Nix generations",
						RiskLevel:   RiskLevelType(RiskLevelLowType),
						Enabled:     ProfileStatusEnabled,
						Settings: &OperationSettings{
							NixGenerations: &NixGenerationsSettings{
								Generations: 1,
								Optimize:    OptimizationModeDisabled,
								DryRun:      ExecutionModeNormal,
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := yaml.NewEncoder(&buf).Encode(config); err != nil {
			b.Fatal(err)
		}
	}
}
