// Package domain provides enum generation macros to reduce boilerplate.
//
// Usage example:
//
//	//go:generate go run ./cmd/enumgen -type=MyEnum -values=Value1,Value2,Value3
//	type MyEnum int
//
//	const (
//		MyEnumValue1 MyEnum = iota
//		MyEnumValue2
//		MyEnumValue3
//	)
//
//	// Implement using macros
//	func (m MyEnum) String() string { return EnumString(m, MyEnumStrings) }
//	func (m MyEnum) IsValid() bool  { return EnumIsValid(m, MyEnumValue3) }
//	func (m MyEnum) Values() []MyEnum { return EnumValues[MyEnum]() }
//
// This reduces 40+ lines per enum to ~10 lines.
package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnumString is a macro for generating String() method.
// Pass the enum value and a string slice mapping.
func EnumString[T ~int](val T, stringsMap []string) string {
	idx := int(val)
	if idx < 0 || idx >= len(stringsMap) {
		return "UNKNOWN"
	}

	return stringsMap[idx]
}

// EnumIsValid is a macro for generating IsValid() method.
// Pass the enum value and the maximum valid value.
func EnumIsValid[T ~int](val, maxVal T) bool {
	return val >= 0 && val <= maxVal
}

// EnumValues is a macro for generating Values() method.
// Returns a slice of all enum values from 0 to max.
func EnumValues[T ~int](maxVal T) []T {
	max := int(maxVal)

	values := make([]T, max+1)
	for i := 0; i <= max; i++ {
		values[i] = T(i)
	}

	return values
}

// EnumMarshalJSON is a macro for JSON marshaling.
func EnumMarshalJSON[T ~int](val T, stringsMap []string) ([]byte, error) {
	return json.Marshal(EnumString(val, stringsMap))
}

// EnumUnmarshalJSON is a macro for JSON unmarshaling.
func EnumUnmarshalJSON[T ~int](
	data []byte,
	target *T,
	stringsMap []string,
	name string,
) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("%s must be a string: %w", name, err)
	}

	upper := strings.ToUpper(s)
	for i, str := range stringsMap {
		if strings.ToUpper(str) == upper {
			*target = T(i)

			return nil
		}
	}

	return fmt.Errorf("invalid %s: %s. Valid: %v", name, s, stringsMap)
}

// EnumMarshalYAML is a macro for YAML marshaling.
func EnumMarshalYAML[T ~int](val T, stringsMap []string) (any, error) {
	return EnumString(val, stringsMap), nil
}

// EnumUnmarshalYAML is a macro for YAML unmarshaling.
func EnumUnmarshalYAML[T ~int](
	value *yaml.Node,
	target *T,
	stringsMap []string,
	name string,
) error {
	var s string

	err := value.Decode(&s)
	if err != nil {
		return fmt.Errorf("%s must be a string: %w", name, err)
	}

	upper := strings.ToUpper(s)
	for i, str := range stringsMap {
		if strings.ToUpper(str) == upper {
			*target = T(i)

			return nil
		}
	}

	return fmt.Errorf("invalid %s: %s. Valid: %v", name, s, stringsMap)
}

// EnumValueMaps holds string mappings for enums.
// Use these with the macro functions above.
var (
	// RiskLevelStrings maps RiskLevelType to strings.
	RiskLevelStrings = []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}

	// StrategyTypeStrings maps StrategyType to strings.
	StrategyTypeStrings = []string{"CONSERVATIVE", "BALANCED", "AGGRESSIVE"}

	// CleanTypeStrings maps CleanType to strings.
	CleanTypeStrings = []string{"QUICK", "STANDARD", "DEEP"}

	// ScanTypeStrings maps ScanType to strings.
	ScanTypeStrings = []string{"TEMP", "CACHE", "LOG", "TRASH", "DERIVED", "BINARY"}

	// OperationTypeStrings maps OperationType to strings.
	OperationTypeStrings = []string{
		"TEMP_FILES",
		"DOCKER",
		"NPM",
		"CARGO",
		"GO_PACKAGES",
		"HOMEBREW",
		"XCODE",
		"BASH",
		"BUILD_CACHE",
		"SYSTEM_CACHE",
		"GIT",
		"PROJECT_EXECUTABLES",
		"COMPILED_BINARIES",
		"NIX",
	}
)
