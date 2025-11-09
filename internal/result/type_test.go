package result

import (
	"errors"
	"testing"
)

func TestResult_Ok(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			expected: true,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.IsOk() != tt.expected {
				t.Errorf("IsOk() = %v, want %v", tt.result.IsOk(), tt.expected)
			}
		})
	}
}

func TestResult_IsErr(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			expected: false,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.IsErr() != tt.expected {
				t.Errorf("IsErr() = %v, want %v", tt.result.IsErr(), tt.expected)
			}
		})
	}
}

func TestResult_Value(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected int
		wantPanic bool
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			expected: 42,
			wantPanic: false,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
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
		name     string
		result   Result[int]
		expected string
		wantPanic bool
	}{
		{
			name:     "ok result",
			result:   Ok(42),
			wantPanic: true,
		},
		{
			name:     "error result",
			result:   Err[int](errors.New("test error")),
			expected: "test error",
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
				// If both are Err, check error messages
				if mapped.Error().Error() != tt.expected.Error().Error() {
					t.Errorf("Map() Error() = %v, want %v", mapped.Error().Error(), tt.expected.Error().Error())
				}
			}
		})
	}
}