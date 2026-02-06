// Package testing provides generic test utilities for running test cases.
// This package is separate from internal/testhelper to avoid import cycles.
package testing

import (
	"fmt"
	"testing"
)

// TestCase represents a standard test case with input and expected output.
// This is the common structure used across test helpers for running parameterized tests.
type TestCase[T any, E any] struct {
	Name     string
	Input    T
	Expected E
}

// RunTestCases runs a suite of test cases with a given test function.
// This eliminates duplicate for-range-t.Run patterns across test files.
//
// Usage:
//
//	func TestMyFunction(t *testing.T) {
//	    tests := []testing.TestCase[InputType, string]{
//	        {"test name", inputValue, "expected"},
//	        // ... more cases
//	    }
//	    testing.RunTestCases(t, tests, func(tc TestCase[InputType, string]) string {
//	        return MyFunction(tc.Input)
//	    })
//	}
//
// Type Parameters:
//   - T: The input type for the test
//   - E: The expected output type
//
// Parameters:
//   - t: The testing.T object
//   - tests: Slice of TestCase to run
//   - testFn: Function that takes a test case and returns the actual result
func RunTestCases[T any, E any](t *testing.T, tests []TestCase[T, E], testFn func(TestCase[T, E]) E) {
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			result := testFn(tc)
			if any(result) != any(tc.Expected) {
				t.Errorf("%s(%v) = %v, want %v", tc.Name, tc.Input, result, tc.Expected)
			}
		})
	}
}

// ValueTestCase represents a test case with a value and expected result.
// This is a simpler pattern for tests that don't need a separate name field.
type ValueTestCase[T any, E any] struct {
	Value    T
	Expected E
}

// RunValueTestCases runs a suite of value test cases with a given test function.
// The test name is derived from the expected value using fmt.Sprintf.
// This is a lighter-weight alternative to RunTestCases for simpler tests.
//
// Usage:
//
//	func TestMyFunction(t *testing.T) {
//	    tests := []testing.ValueTestCase[InputType, string]{
//	        {inputValue, "expected"},
//	        // ... more cases
//	    }
//	    testing.RunValueTestCases(t, tests, func(tc ValueTestCase[InputType, string]) string {
//	        return MyFunction(tc.Value)
//	    })
//	}
//
// Type Parameters:
//   - T: The input type for the test
//   - E: The expected output type
//
// Parameters:
//   - t: The testing.T object
//   - tests: Slice of ValueTestCase to run
//   - testFn: Function that takes a test case and returns the actual result
func RunValueTestCases[T any, E any](t *testing.T, tests []ValueTestCase[T, E], testFn func(ValueTestCase[T, E]) E) {
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.Expected), func(t *testing.T) {
			result := testFn(tc)
			if any(result) != any(tc.Expected) {
				t.Errorf("%v = %v, want %v", tc.Value, result, tc.Expected)
			}
		})
	}
}