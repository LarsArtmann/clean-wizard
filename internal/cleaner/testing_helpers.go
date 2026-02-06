package cleaner

import "testing"

// TestAvailableTypesGeneric tests that the available types function returns the expected list
func TestAvailableTypesGeneric[T comparable](t *testing.T, name string, getAvailable func() []T, expected []T) {
	t.Run(name, func(t *testing.T) {
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
func TestTypeStringCases[T ~string](cases []T) []struct {
	Value T
	Want  string
} {
	result := make([]struct {
		Value T
		Want  string
	}, len(cases))

	for i, c := range cases {
		result[i] = struct {
			Value T
			Want  string
		}{
			Value: c,
			Want:  string(c),
		}
	}

	return result
}

// TestTypeStringGeneric tests the string representation of a type
func TestTypeStringGeneric[T ~string](t *testing.T, name string, getTestCases func() []struct {
	Value T
	Want  string
}) {
	tests := getTestCases()

	for _, tt := range tests {
		t.Run(tt.Want, func(t *testing.T) {
			got := string(tt.Value)
			if got != tt.Want {
				t.Errorf("string(%v) = %v, want %v", tt.Value, got, tt.Want)
			}
		})
	}
}

// TestTypeString tests the String() method of a type with the given cases
func TestTypeString[T ~string](t *testing.T, name string, cases []T) {
	TestTypeStringGeneric(t, name, func() []struct {
		Value T
		Want  string
	} {
		return TestTypeStringCases(cases)
	})
}