package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// enumMacroTestCase holds test data for enum macros (JSON/YAML unmarshal).
// This consolidates duplicate test case structures.
type enumMacroTestCase struct {
	name     string
	input    string
	expected int
	wantErr  bool
}

// assertEnumUnmarshalResult is a helper that runs assertions on enum unmarshal results.
func assertEnumUnmarshalResult(t *testing.T, tc enumMacroTestCase, err error, val int) {
	if tc.wantErr {
		assert.Error(t, err)
	} else {
		require.NoError(t, err)
		assert.Equal(t, tc.expected, val)
	}
}

// jsonEnumTestCases provides test cases specific to JSON unmarshal (with quoted strings).
var jsonEnumTestCases = []enumMacroTestCase{
	{"valid lowercase", `"medium"`, 1, false},
	{"valid uppercase", `"HIGH"`, 2, false},
	{"invalid value", `"INVALID"`, 0, true},
}

// yamlEnumTestCases provides test cases specific to YAML unmarshal.
var yamlEnumTestCases = []enumMacroTestCase{
	{"valid lowercase", "medium", 1, false},
	{"valid integer", "2", 2, false},
	{"invalid value", "INVALID", 0, true},
}

func TestEnumString(t *testing.T) {
	t.Parallel()

	strings := []string{"A", "B", "C"}
	tests := []struct {
		name     string
		val      int
		expected string
	}{
		{"valid 0", 0, "A"},
		{"valid 1", 1, "B"},
		{"valid 2", 2, "C"},
		{"negative", -1, "UNKNOWN"},
		{"too high", 3, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := EnumString(tt.val, strings)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnumIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		val      int
		maxVal   int
		expected bool
	}{
		{"valid min", 0, 2, true},
		{"valid middle", 1, 2, true},
		{"valid max", 2, 2, true},
		{"negative", -1, 2, false},
		{"too high", 3, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := EnumIsValid(tt.val, tt.maxVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnumValues(t *testing.T) {
	t.Parallel()

	result := EnumValues[int](2)
	assert.Equal(t, []int{0, 1, 2}, result)
}

func TestEnumMarshalJSON(t *testing.T) {
	t.Parallel()

	strings := []string{"LOW", "MEDIUM", "HIGH"}
	data, err := EnumMarshalJSON(1, strings)
	require.NoError(t, err)
	assert.Equal(t, `"MEDIUM"`, string(data))
}

func TestEnumUnmarshalJSON(t *testing.T) {
	t.Parallel()

	strings := []string{"LOW", "MEDIUM", "HIGH"}

	for _, tc := range jsonEnumTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var val int

			err := EnumUnmarshalJSON([]byte(tc.input), &val, strings, "test")
			assertEnumUnmarshalResult(t, tc, err, val)
		})
	}
}

func TestEnumMarshalYAML(t *testing.T) {
	t.Parallel()

	strings := []string{"LOW", "MEDIUM", "HIGH"}
	result, err := EnumMarshalYAML(1, strings)
	require.NoError(t, err)
	assert.Equal(t, "MEDIUM", result)
}

func TestEnumUnmarshalYAML(t *testing.T) {
	t.Parallel()

	strings := []string{"LOW", "MEDIUM", "HIGH"}

	for _, tc := range yamlEnumTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := &yaml.Node{Kind: yaml.ScalarNode, Value: tc.input}

			var val int

			err := EnumUnmarshalYAML(node, &val, strings, "test")
			assertEnumUnmarshalResult(t, tc, err, val)
		})
	}
}

// Example of using macros for a custom enum.
type testEnum int

const (
	testEnumA testEnum = iota
	testEnumB
	testEnumC
)

var testEnumStrings = []string{"A", "B", "C"}

func (e testEnum) String() string     { return EnumString(e, testEnumStrings) }
func (e testEnum) IsValid() bool      { return EnumIsValid(e, testEnumC) }
func (e testEnum) Values() []testEnum { return EnumValues[testEnum](testEnumC) }
func (e testEnum) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(e, testEnumStrings)
}

func (e *testEnum) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(data, e, testEnumStrings, "testEnum")
}

func TestCustomEnumWithMacros(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "A", testEnumA.String())
		assert.Equal(t, "B", testEnumB.String())
		assert.Equal(t, "UNKNOWN", testEnum(99).String())
	})

	t.Run("IsValid", func(t *testing.T) {
		t.Parallel()
		assert.True(t, testEnumA.IsValid())
		assert.True(t, testEnumC.IsValid())
		assert.False(t, testEnum(99).IsValid())
	})

	t.Run("Values", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []testEnum{0, 1, 2}, testEnumA.Values())
	})

	t.Run("JSON roundtrip", func(t *testing.T) {
		t.Parallel()

		original := testEnumB
		data, err := json.Marshal(original)
		require.NoError(t, err)
		assert.Equal(t, `"B"`, string(data))

		var decoded testEnum

		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})
}
