package result

import (
	"errors"
	"testing"
)

// testPredicateTestCases provides test cases for predicate testing with shared test data.
func testPredicateTestCases(okExpected, errExpected bool) []struct {
	name     string
	result   Result[int]
	expected bool
} {
	return []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{"ok result", testOkResult(), okExpected},
		{"error result", testErrResult(), errExpected},
	}
}

func TestResult_Ok(t *testing.T) {
	testPredicateCases(t, "IsOk", func(r Result[int]) bool { return r.IsOk() }, testPredicateTestCases(true, false))
}

func TestResult_IsErr(t *testing.T) {
	testPredicateCases(t, "IsErr", func(r Result[int]) bool { return r.IsErr() }, testPredicateTestCases(false, true))
}

// testOkResult returns a successful result for testing.
func testOkResult() Result[int] {
	return Ok(42)
}

// testErrResult returns an error result for testing.
func testErrResult() Result[int] {
	return Err[int](errors.New("test error"))
}

// testPredicateCases tests a predicate function against multiple cases.
func testPredicateCases(t *testing.T, methodName string, predicate func(Result[int]) bool, cases []struct {
	name     string
	result   Result[int]
	expected bool
},
) {
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if predicate(tt.result) != tt.expected {
				t.Errorf("%s() = %v, want %v", methodName, predicate(tt.result), tt.expected)
			}
		})
	}
}

func TestResult_Value(t *testing.T) {
	tests := []struct {
		name      string
		result    Result[int]
		expected  int
		wantPanic bool
	}{
		{name: "ok result", result: testOkResult(), expected: 42, wantPanic: false},
		{name: "error result", result: testErrResult(), wantPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				} else {
					if tt.wantPanic {
						t.Error("expected panic but got none")
					}
				}
			}()

			value := tt.result.Value()
			if !tt.wantPanic && value != tt.expected {
				t.Errorf("Value() = %v, want %v", value, tt.expected)
			}
		})
	}
}

func TestResult_Error(t *testing.T) {
	tests := []struct {
		name      string
		result    Result[int]
		expected  string
		wantPanic bool
	}{
		{name: "ok result", result: testOkResult(), wantPanic: true},
		{name: "error result", result: testErrResult(), expected: "test error", wantPanic: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				} else {
					if tt.wantPanic {
						t.Error("expected panic but got none")
					}
				}
			}()

			err := tt.result.Error()
			if !tt.wantPanic && err.Error() != tt.expected {
				t.Errorf("Error() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestResult_UnwrapOr(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		default_ int
		expected int
	}{
		{name: "ok result", result: testOkResult(), default_: 0, expected: 42},
		{name: "error result", result: testErrResult(), default_: 99, expected: 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.result.UnwrapOr(tt.default_)
			if value != tt.expected {
				t.Errorf("UnwrapOr(%v) = %v, want %v", tt.default_, value, tt.expected)
			}
		})
	}
}

func TestResult_Map(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		fn       func(int) string
		expected Result[string]
	}{
		{name: "ok result", result: testOkResult(), fn: func(i int) string { return "42" }, expected: Ok("42")},
		{name: "error result", result: testErrResult(), fn: func(i int) string { return "42" }, expected: Err[string](errors.New("test error"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := Map(tt.result, tt.fn)

			// Check if both are Ok
			if mapped.IsOk() != tt.expected.IsOk() {
				t.Errorf("Map() IsOk() = %v, want %v", mapped.IsOk(), tt.expected.IsOk())
				return
			}

			// If both are Ok, check values
			if mapped.IsOk() {
				if mapped.Value() != tt.expected.Value() {
					t.Errorf("Map() Value() = %v, want %v", mapped.Value(), tt.expected.Value())
				}
			} else {
				// If both are Err, check error messages
				if mapped.Error().Error() != tt.expected.Error().Error() {
					t.Errorf("Map() Error() = %v, want %v", mapped.Error().Error(), tt.expected.Error().Error())
				}
			}
		})
	}
}

func TestResult_AndThen(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		fn       func(int) Result[string]
		expected Result[string]
	}{
		{name: "ok result to ok result", result: testOkResult(), fn: func(i int) Result[string] { return Ok("success") }, expected: Ok("success")},
		{name: "ok result to error result", result: testOkResult(), fn: func(i int) Result[string] { return Err[string](errors.New("chained error")) }, expected: Err[string](errors.New("chained error"))},
		{name: "error result", result: testErrResult(), fn: func(i int) Result[string] { return Ok("should not be called") }, expected: Err[string](errors.New("test error"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AndThen(tt.result, tt.fn)

			if result.IsOk() != tt.expected.IsOk() {
				t.Errorf("AndThen() IsOk() = %v, want %v", result.IsOk(), tt.expected.IsOk())
				return
			}

			if result.IsOk() {
				if result.Value() != tt.expected.Value() {
					t.Errorf("AndThen() Value() = %v, want %v", result.Value(), tt.expected.Value())
				}
			} else {
				if result.Error().Error() != tt.expected.Error().Error() {
					t.Errorf("AndThen() Error() = %v, want %v", result.Error().Error(), tt.expected.Error().Error())
				}
			}
		})
	}
}

func TestResult_FlatMap(t *testing.T) {
	// FlatMap is an alias for AndThen, so we just verify it works
	result := FlatMap(Ok(10), func(i int) Result[int] { return Ok(i * 2) })
	if !result.IsOk() || result.Value() != 20 {
		t.Errorf("FlatMap() = %v, want Ok(20)", result)
	}
}

func TestResult_OrElse(t *testing.T) {
	tests := []struct {
		name          string
		result        Result[int]
		fallback      Result[int]
		expectedValue int
		expectedOk    bool
	}{
		{name: "ok result returns original", result: Ok(42), fallback: Ok(99), expectedValue: 42, expectedOk: true},
		{name: "error result returns fallback", result: testErrResult(), fallback: Ok(99), expectedValue: 99, expectedOk: true},
		{name: "error result with error fallback", result: testErrResult(), fallback: Err[int](errors.New("fallback error")), expectedValue: 0, expectedOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.OrElse(tt.fallback)

			if result.IsOk() != tt.expectedOk {
				t.Errorf("OrElse() IsOk() = %v, want %v", result.IsOk(), tt.expectedOk)
			}

			if tt.expectedOk && result.Value() != tt.expectedValue {
				t.Errorf("OrElse() Value() = %v, want %v", result.Value(), tt.expectedValue)
			}
		})
	}
}

func TestResult_Validate(t *testing.T) {
	tests := []struct {
		name          string
		result        Result[int]
		predicate     func(int) bool
		errorMsg      string
		expectedOk    bool
		expectedError string
	}{
		{name: "valid passes through", result: Ok(42), predicate: func(i int) bool { return i > 0 }, errorMsg: "must be positive", expectedOk: true},
		{name: "invalid returns error", result: Ok(-5), predicate: func(i int) bool { return i > 0 }, errorMsg: "must be positive", expectedOk: false, expectedError: "must be positive"},
		{name: "error passes through", result: testErrResult(), predicate: func(i int) bool { return i > 0 }, errorMsg: "must be positive", expectedOk: false, expectedError: "test error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.Validate(tt.predicate, tt.errorMsg)

			if result.IsOk() != tt.expectedOk {
				t.Errorf("Validate() IsOk() = %v, want %v", result.IsOk(), tt.expectedOk)
			}

			if !tt.expectedOk && result.IsErr() {
				if result.Error().Error() != tt.expectedError {
					t.Errorf("Validate() Error() = %v, want %v", result.Error().Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestResult_ValidateWithError(t *testing.T) {
	customErr := errors.New("custom validation error")
	result := Ok(10).ValidateWithError(func(i int) bool { return i > 100 }, customErr)

	if result.IsOk() {
		t.Errorf("ValidateWithError() should have returned error for value 10 with predicate > 100")
	}

	if result.Error().Error() != customErr.Error() {
		t.Errorf("ValidateWithError() Error() = %v, want %v", result.Error().Error(), customErr.Error())
	}
}

func TestResult_Tap(t *testing.T) {
	var tappedValue *int
	okResult := Ok(42).Tap(func(v int) { tappedValue = &v })

	if !okResult.IsOk() || okResult.Value() != 42 {
		t.Errorf("Tap() should not change the result")
	}

	if tappedValue == nil || *tappedValue != 42 {
		t.Errorf("Tap() should have called the function with the value")
	}

	// Tap should not call the function on error result
	errResult := testErrResult().Tap(func(v int) {
		t.Errorf("Tap should not be called on error result")
	})

	// Tap should return the original error result
	if !errResult.IsErr() {
		t.Errorf("Tap() should return the original error result when called on error")
	}
}
