package cleaner

import (
	stdtesting "testing"

	"github.com/LarsArtmann/clean-wizard/internal/testing"
)

// TestAvailableTypesGeneric tests that the available types function returns the expected list
func TestAvailableTypesGeneric[T comparable](t *stdtesting.T, name string, getAvailable func() []T, expected []T) {
	t.Run(name, func(t *stdtesting.T) {
		types := getAvailable()

		if len(types) != len(expected) {
			t.Errorf("%s() returned %d types, want %d", name, len(types), len(expected))
		}

		for i, typ := range types {
			if typ != expected[i] {
				t.Errorf("%s()[%d] = %v, want %v", name, i, typ, expected[i])
			}
		}
	})
}

// TestTypeStringCases creates test cases for String() method testing
func TestTypeStringCases[T ~string](cases []T) []testing.ValueTestCase[T, string] {
	result := make([]testing.ValueTestCase[T, string], len(cases))

	for i, c := range cases {
		result[i] = testing.ValueTestCase[T, string]{
			Value:    c,
			Expected: string(c),
		}
	}

	return result
}

// TestTypeStringGeneric tests the string representation of a type
func TestTypeStringGeneric[T ~string](t *stdtesting.T, name string, getTestCases func() []testing.ValueTestCase[T, string]) {
	tests := getTestCases()

	for _, tt := range tests {
		t.Run(tt.Expected, func(t *stdtesting.T) {
			got := string(tt.Value)
			if got != tt.Expected {
				t.Errorf("string(%v) = %v, want %v", tt.Value, got, tt.Expected)
			}
		})
	}
}

// TestTypeString tests the String() method of a type with the given cases
func TestTypeString[T ~string](t *stdtesting.T, name string, cases []T) {
	TestTypeStringGeneric(t, name, func() []testing.ValueTestCase[T, string] {
		return TestTypeStringCases(cases)
	})
}

// TestEnumString tests that all enum values produce their expected string representation
// This is a simpler alternative for types where the enum values are string constants
func TestEnumString[T ~string](t *stdtesting.T, name string, values []T) {
	expected := map[T]string{}
	for _, v := range values {
		expected[v] = string(v)
	}

	for _, value := range values {
		t.Run(expected[value], func(t *stdtesting.T) {
			got := string(value)
			want := expected[value]
			if got != want {
				t.Errorf("string(%[1]v) = %[2]q, want %[3]q", value, got, want)
			}
		})
	}
}