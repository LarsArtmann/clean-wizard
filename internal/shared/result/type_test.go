package result

import (
	"errors"
	"testing"
)

// testResultMethod tests a boolean method on Result type with expected values
func testResultMethod(t *testing.T, methodName string, methodFunc func(Result[int]) bool, okExpected, errExpected bool) {
	testData := []struct {
		name   string
		result Result[int]
	}{
		{
			name:   "ok result",
			result: Ok(42),
		},
		{
			name:   "error result",
			result: Err[int](errors.New("test error")),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			actual := methodFunc(tt.result)
			expected := okExpected
			if tt.result.IsErr() {
				expected = errExpected
			}

			if actual != expected {
				t.Errorf("%s() = %v, want %v", methodName, actual, expected)
			}
		})
	}
}

// runMethodTest tests a method that returns a comparable type
func runMethodTest[T comparable](t *testing.T, methodName string, methodFunc func(Result[int]) T, okResult, errResult T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected T
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			expected: okResult,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			expected: errResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if methodFunc(tt.result) != tt.expected {
				t.Errorf("%s() = %v, want %v", methodName, methodFunc(tt.result), tt.expected)
			}
		})
	}
}

// runMethodTestWithErrorHandling tests a method with panic handling and custom validation
func runMethodTestWithErrorHandling[T comparable](t *testing.T, methodName string, methodFunc func(Result[int]) T, okResult, errResult T, okWantPanic, errWantPanic bool, validateFunc func(T, T) bool) {
	tests := []struct {
		name      string
		result    Result[int]
		expected  T
		wantPanic bool
	}{
		{
			name:      "ok result",
			result:    Ok(42),
			expected:  okResult,
			wantPanic: okWantPanic,
		},
		{
			name:      "error result",
			result:    Err[int](errors.New("test error")),
			expected:  errResult,
			wantPanic: errWantPanic,
		},
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

			if !tt.wantPanic {
				result := methodFunc(tt.result)
				if !validateFunc(result, tt.expected) {
					t.Errorf("%s() = %v, want %v", methodName, result, tt.expected)
				}
			} else {
				_ = methodFunc(tt.result)
			}
		})
	}
}

func TestResult_IsOk(t *testing.T) {
	testResultMethod(t, "IsOk", func(r Result[int]) bool { return r.IsOk() }, true, false)
}

func TestResult_IsErr(t *testing.T) {
	testResultMethod(t, "IsErr", func(r Result[int]) bool { return r.IsErr() }, false, true)
}

func TestResult_Value(t *testing.T) {
	runMethodTestWithErrorHandling(t, "Value", func(r Result[int]) int { return r.Value() }, 42, 0, false, true, func(actual, expected int) bool { return actual == expected })
}

func TestResult_Error(t *testing.T) {
	runMethodTestWithErrorHandling(t, "Error", func(r Result[int]) error { return r.Error() }, error(nil), errors.New("test error"), true, false, func(actual, expected error) bool {
		if actual == nil && expected == nil {
			return true
		}
		if actual == nil || expected == nil {
			return false
		}
		return actual.Error() == expected.Error()
	})
}

func TestResult_UnwrapOr(t *testing.T) {
	runMethodTest(t, "UnwrapOr", func(r Result[int]) int { return r.UnwrapOr(99) }, 42, 99)
}

func TestResult_Map(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		fn       func(int) string
		expected Result[string]
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			fn:       func(i int) string { return "ok" },
			expected: Ok("ok"),
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			fn:       func(i int) string { return "ok" },
			expected: Err[string](errors.New("test error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.result, tt.fn)
			if (result.IsOk() != tt.expected.IsOk()) || (result.IsErr() != tt.expected.IsErr()) {
				t.Errorf("Map() = %v, want %v", result, tt.expected)
			}
		})
	}
}
