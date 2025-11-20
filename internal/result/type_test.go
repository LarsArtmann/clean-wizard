package result

import (
	"errors"
	"strings"
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

	tests := []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{
			name:     testData[0].name,
			result:   testData[0].result,
			expected: okExpected,
		},
		{
			name:     testData[1].name,
			result:   testData[1].result,
			expected: errExpected,
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

func TestResult_Ok(t *testing.T) {
	testResultMethod(t, "IsOk", func(r Result[int]) bool { return r.IsOk() }, true, false)
}

func TestResult_IsErr(t *testing.T) {
	testResultMethod(t, "IsErr", func(r Result[int]) bool { return r.IsErr() }, false, true)
}

func TestResult_Value(t *testing.T) {
	tests := []struct {
		name      string
		result    Result[int]
		expected  int
		wantPanic bool
	}{
		{
			name:      "ok result",
			result:    Ok(42),
			expected:  42,
			wantPanic: false,
		},
		{
			name:      "error result",
			result:    Err[int](errors.New("test error")),
			wantPanic: true,
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
		{
			name:      "ok result",
			result:    Ok(42),
			wantPanic: true,
		},
		{
			name:      "error result",
			result:    Err[int](errors.New("test error")),
			expected:  "test error",
			wantPanic: false,
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

			err := tt.result.Error()
			if !tt.wantPanic && !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Error() = %v, want to contain %v", err.Error(), tt.expected)
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
		{
			name:     "ok result",
			result:   Ok(42),
			default_: 0,
			expected: 42,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			default_: 99,
			expected: 99,
		},
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
		{
			name:     "ok result",
			result:   Ok(42),
			fn:       func(i int) string { return "42" },
			expected: Ok("42"),
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			fn:       func(i int) string { return "42" },
			expected: Err[string](errors.New("test error")),
		},
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
				// If both are Err, check error messages using substring matching
				if !strings.Contains(mapped.Error().Error(), tt.expected.Error().Error()) {
					t.Errorf("Map() Error() = %v, want to contain %v", mapped.Error().Error(), tt.expected.Error().Error())
				}
			}
		})
	}
}
