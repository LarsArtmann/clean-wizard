package domain

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"
)

// benchmarkTestConfig is a shared Config instance used across benchmark tests.
var benchmarkTestConfig = &Config{
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

// BenchmarkMarshalYAML_DockerPruneMode benchmarks marshaling DockerPruneMode enum to YAML.
func BenchmarkMarshalYAML_DockerPruneMode(b *testing.B) {
	testCases := []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}
	runMarshalBenchmark(b, testCases)
}

// BenchmarkUnmarshalYAML_DockerPruneMode_String benchmarks unmarshaling DockerPruneMode from string YAML.
func BenchmarkUnmarshalYAML_DockerPruneMode_String(b *testing.B) {
	testCases := []string{"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS"}
	runUnmarshalStringBenchmark(b, testCases, func(tc string) error {
		var result DockerPruneMode

		yamlData := fmt.Sprintf(`"%s"`, tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
}

// runUnmarshalIntBenchmark runs a benchmark for unmarshaling an enum from integer YAML values.
func runUnmarshalIntBenchmark(b *testing.B, testCases []int, unmarshal func(int) error) {
	runUnmarshalBenchmark(b, testCases, strconv.Itoa, unmarshal)
}

// runUnmarshalStringBenchmark runs a benchmark for unmarshaling an enum from string YAML values.
func runUnmarshalStringBenchmark(b *testing.B, testCases []string, unmarshal func(string) error) {
	runUnmarshalBenchmark(b, testCases, func(s string) string { return s }, unmarshal)
}

// runUnmarshalBenchmark runs a benchmark for unmarshaling enum values with a custom label function.
func runUnmarshalBenchmark[T any](
	b *testing.B,
	testCases []T,
	labelFunc func(T) string,
	unmarshal func(T) error,
) {
	for _, tc := range testCases {
		b.Run(labelFunc(tc), func(b *testing.B) {
			for range b.N {
				err := unmarshal(tc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// yamlMarshaler is a constraint for types that can marshal to YAML and have a String method.
type yamlMarshaler interface {
	fmt.Stringer
	MarshalYAML() (any, error)
}

// yamlRoundTripper is a constraint for types that support full YAML round-trip (marshal + unmarshal).
type yamlRoundTripper interface {
	comparable
	fmt.Stringer
	MarshalYAML() (any, error)
}

// runRoundTripBenchmark runs a benchmark for full marshal→unmarshal round-trip.
func runRoundTripBenchmark[T yamlRoundTripper](b *testing.B, testCases []T) {
	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for range b.N {
				marshaled, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}

				yamlBytes, err := yaml.Marshal(marshaled)
				if err != nil {
					b.Fatal(err)
				}

				var result T
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

// runMarshalBenchmark runs a benchmark for marshaling enum values to YAML.
func runMarshalBenchmark[T yamlMarshaler](b *testing.B, testCases []T) {
	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for range b.N {
				_, err := tc.MarshalYAML()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkUnmarshalYAML_DockerPruneMode_Int benchmarks unmarshaling DockerPruneMode from integer YAML.
func BenchmarkUnmarshalYAML_DockerPruneMode_Int(b *testing.B) {
	testCases := []int{0, 1, 2, 3, 4}
	runUnmarshalIntBenchmark(b, testCases, func(tc int) error {
		var result DockerPruneMode

		yamlData := strconv.Itoa(tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
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
	runMarshalBenchmark(b, testCases)
}

// BenchmarkUnmarshalYAML_BuildToolType_String benchmarks unmarshaling BuildToolType from string YAML.
func BenchmarkUnmarshalYAML_BuildToolType_String(b *testing.B) {
	testCases := []string{"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA"}
	runUnmarshalStringBenchmark(b, testCases, func(tc string) error {
		var result BuildToolType

		yamlData := fmt.Sprintf(`"%s"`, tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
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
	runMarshalBenchmark(b, testCases)
}

// BenchmarkUnmarshalYAML_CacheType benchmarks unmarshaling CacheType enum.
func BenchmarkUnmarshalYAML_CacheType_String(b *testing.B) {
	testCases := []string{
		"SPOTLIGHT",
		"XCODE",
		"COCOAPODS",
		"HOMEBREW",
		"PIP",
		"NPM",
		"YARN",
		"CCACHE",
	}
	runUnmarshalStringBenchmark(b, testCases, func(tc string) error {
		var result CacheType

		yamlData := fmt.Sprintf(`"%s"`, tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
}

// BenchmarkMarshalYAML_PackageManagerType benchmarks marshaling PackageManagerType enum to YAML.
func BenchmarkMarshalYAML_PackageManagerType(b *testing.B) {
	testCases := []PackageManagerType{
		PackageManagerNpm,
		PackageManagerPnpm,
		PackageManagerYarn,
		PackageManagerBun,
	}
	runMarshalBenchmark(b, testCases)
}

// BenchmarkUnmarshalYAML_PackageManagerType benchmarks unmarshaling PackageManagerType enum.
func BenchmarkUnmarshalYAML_PackageManagerType_String(b *testing.B) {
	testCases := []string{"NPM", "PNPM", "YARN", "BUN"}
	runUnmarshalStringBenchmark(b, testCases, func(tc string) error {
		var result PackageManagerType

		yamlData := fmt.Sprintf(`"%s"`, tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
}

// BenchmarkMarshalYAML_CacheCleanupMode benchmarks marshaling CacheCleanupMode enum to YAML.
func BenchmarkMarshalYAML_CacheCleanupMode(b *testing.B) {
	testCases := []CacheCleanupMode{CacheCleanupDisabled, CacheCleanupEnabled}
	runMarshalBenchmark(b, testCases)
}

// BenchmarkUnmarshalYAML_BinaryEnums_String benchmarks unmarshaling binary enums from string YAML.
func BenchmarkUnmarshalYAML_CacheCleanupMode_String(b *testing.B) {
	testCases := []string{"DISABLED", "ENABLED"}
	runUnmarshalStringBenchmark(b, testCases, func(tc string) error {
		var result CacheCleanupMode

		yamlData := fmt.Sprintf(`"%s"`, tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
}

// BenchmarkUnmarshalYAML_BinaryEnums_Int benchmarks unmarshaling binary enums from integer YAML.
func BenchmarkUnmarshalYAML_CacheCleanupMode_Int(b *testing.B) {
	testCases := []int{0, 1}
	runUnmarshalIntBenchmark(b, testCases, func(tc int) error {
		var result CacheCleanupMode

		yamlData := strconv.Itoa(tc)

		return yaml.Unmarshal([]byte(yamlData), &result)
	})
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
	runRoundTripBenchmark(b, testCases)
}

// BenchmarkRoundTrip_CacheCleanupMode benchmarks full marshal→unmarshal round-trip for CacheCleanupMode.
func BenchmarkRoundTrip_CacheCleanupMode(b *testing.B) {
	testCases := []CacheCleanupMode{CacheCleanupDisabled, CacheCleanupEnabled}
	runRoundTripBenchmark(b, testCases)
}

// BenchmarkFullConfigMarshal benchmarks marshaling a complete configuration with enums.
func BenchmarkFullConfigMarshal(b *testing.B) {
	for b.Loop() {
		_, err := yaml.Marshal(benchmarkTestConfig)
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

	for b.Loop() {
		var config Config

		err := yaml.Unmarshal([]byte(yamlConfig), &config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullConfigRoundTrip benchmarks marshal→unmarshal round-trip for complete configuration.
func BenchmarkFullConfigRoundTrip(b *testing.B) {
	for b.Loop() {
		yamlBytes, err := yaml.Marshal(benchmarkTestConfig)
		if err != nil {
			b.Fatal(err)
		}

		var result Config
		if err := yaml.Unmarshal(yamlBytes, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// stringer is a constraint for types that have a String() method.
type stringer interface {
	String() string
}

// runStringMethodBenchmark runs a benchmark for the String() method.
func runStringMethodBenchmark[T stringer](b *testing.B, testCases []T) {
	for _, tc := range testCases {
		b.Run(tc.String(), func(b *testing.B) {
			for range b.N {
				_ = tc.String()
			}
		})
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
	runStringMethodBenchmark(b, testCases)
}

// validatable is a constraint for types that have an IsValid() method.
type validatable interface {
	~int
	IsValid() bool
}

// runIsValidBenchmark runs a benchmark for the IsValid() method.
func runIsValidBenchmark[T validatable](b *testing.B, testCases []T) {
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("%d", tc), func(b *testing.B) {
			for range b.N {
				_ = tc.IsValid()
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
	runIsValidBenchmark(b, testCases)
}

// benchmarkYAMLDecodeRaw runs a benchmark for raw YAML decoding into type T.
func benchmarkYAMLDecodeRaw[T any](b *testing.B, data []byte) {
	var result T
	for b.Loop() {
		err := yaml.Unmarshal(data, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkYAMLDecodeRaw_Int benchmarks raw YAML int decoding for comparison.
func BenchmarkYAMLDecodeRaw_Int(b *testing.B) {
	benchmarkYAMLDecodeRaw[int](b, []byte("0"))
}

// BenchmarkYAMLDecodeRaw_String benchmarks raw YAML string decoding for comparison.
func BenchmarkYAMLDecodeRaw_String(b *testing.B) {
	benchmarkYAMLDecodeRaw[string](b, []byte("ALL"))
}

// benchmarkNodeDecode runs a benchmark for yaml.Node.Decode into type T.
func benchmarkNodeDecode[T any](b *testing.B, yamlData string) {
	node := &yaml.Node{}

	err := yaml.Unmarshal([]byte(yamlData), node)
	if err != nil {
		b.Fatal(err)
	}

	var result T
	for b.Loop() {
		err := node.Decode(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNodeDecode_Int benchmarks yaml.Node.Decode performance for ints.
func BenchmarkNodeDecode_Int(b *testing.B) {
	benchmarkNodeDecode[int](b, "0")
}

// BenchmarkNodeDecode_String benchmarks yaml.Node.Decode for strings.
func BenchmarkNodeDecode_String(b *testing.B) {
	benchmarkNodeDecode[string](b, `"ALL"`)
}

// BenchmarkStringComparison benchmarks string comparison operations used in enum unmarshaling.
func BenchmarkStringComparison_CaseInsensitive(b *testing.B) {
	testCases := []string{"ALL", "all", "All", "aLl", "aLL"}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			target := "ALL"
			for range b.N {
				_ = tc == target
			}
		})
	}
}

// BenchmarkBufferWrite benchmarks writing YAML to buffer.
func BenchmarkBufferWrite_Config(b *testing.B) {
	for b.Loop() {
		var buf bytes.Buffer

		err := yaml.NewEncoder(&buf).Encode(benchmarkTestConfig)
		if err != nil {
			b.Fatal(err)
		}
	}
}
