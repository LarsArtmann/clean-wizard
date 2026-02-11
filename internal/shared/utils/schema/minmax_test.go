package schema

import (
	"math"
	"testing"
)

func TestExtractMinMax(t *testing.T) {
	minVal := 10
	maxVal := 95

	tests := []struct {
		name        string
		min         *int
		max         *int
		minFallback float64
		maxFallback float64
		wantMin     *float64
		wantMax     *float64
	}{
		{
			name:        "both values provided",
			min:         &minVal,
			max:         &maxVal,
			minFallback: 0,
			maxFallback: 100,
			wantMin:     float64Ptr(10),
			wantMax:     float64Ptr(95),
		},
		{
			name:        "nil min uses fallback",
			min:         nil,
			max:         &maxVal,
			minFallback: 5,
			maxFallback: 100,
			wantMin:     float64Ptr(5),
			wantMax:     float64Ptr(95),
		},
		{
			name:        "nil max uses fallback",
			min:         &minVal,
			max:         nil,
			minFallback: 0,
			maxFallback: 90,
			wantMin:     float64Ptr(10),
			wantMax:     float64Ptr(90),
		},
		{
			name:        "both nil uses fallbacks",
			min:         nil,
			max:         nil,
			minFallback: 0,
			maxFallback: 100,
			wantMin:     float64Ptr(0),
			wantMax:     float64Ptr(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMinMax(tt.min, tt.max, tt.minFallback, tt.maxFallback)

			if (tt.wantMin == nil) != (result.Min == nil) {
				t.Errorf("Min = %v, want %v", result.Min, tt.wantMin)
			} else if tt.wantMin != nil && *result.Min != *tt.wantMin {
				t.Errorf("Min = %v, want %v", *result.Min, *tt.wantMin)
			}

			if (tt.wantMax == nil) != (result.Max == nil) {
				t.Errorf("Max = %v, want %v", result.Max, tt.wantMax)
			} else if tt.wantMax != nil && *result.Max != *tt.wantMax {
				t.Errorf("Max = %v, want %v", *result.Max, *tt.wantMax)
			}
		})
	}
}

func TestMinMax_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		minMax    MinMax
		wantEmpty bool
	}{
		{
			name:      "both nil is empty",
			minMax:    MinMax{Min: nil, Max: nil},
			wantEmpty: true,
		},
		{
			name:      "min only is not empty",
			minMax:    MinMax{Min: float64Ptr(10), Max: nil},
			wantEmpty: false,
		},
		{
			name:      "max only is not empty",
			minMax:    MinMax{Min: nil, Max: float64Ptr(90)},
			wantEmpty: false,
		},
		{
			name:      "both present is not empty",
			minMax:    MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.minMax.IsEmpty(); got != tt.wantEmpty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.wantEmpty)
			}
		})
	}
}

func TestMinMax_Range(t *testing.T) {
	tests := []struct {
		name   string
		minMax MinMax
		want   float64
	}{
		{
			name:   "valid range",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			want:   80,
		},
		{
			name:   "nil min returns NaN",
			minMax: MinMax{Min: nil, Max: float64Ptr(90)},
			want:   math.NaN(),
		},
		{
			name:   "nil max returns NaN",
			minMax: MinMax{Min: float64Ptr(10), Max: nil},
			want:   math.NaN(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.minMax.Range()
			if math.IsNaN(tt.want) {
				if !math.IsNaN(got) {
					t.Errorf("Range() = %v, want NaN", got)
				}
			} else if got != tt.want {
				t.Errorf("Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinMax_Contains(t *testing.T) {
	tests := []struct {
		name   string
		minMax MinMax
		value  float64
		want   bool
	}{
		{
			name:   "value in range",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			value:  50,
			want:   true,
		},
		{
			name:   "value below min",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			value:  5,
			want:   false,
		},
		{
			name:   "value above max",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			value:  95,
			want:   false,
		},
		{
			name:   "value at min boundary",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			value:  10,
			want:   true,
		},
		{
			name:   "value at max boundary",
			minMax: MinMax{Min: float64Ptr(10), Max: float64Ptr(90)},
			value:  90,
			want:   true,
		},
		{
			name:   "only min constraint - below",
			minMax: MinMax{Min: float64Ptr(10), Max: nil},
			value:  5,
			want:   false,
		},
		{
			name:   "only min constraint - above",
			minMax: MinMax{Min: float64Ptr(10), Max: nil},
			value:  50,
			want:   true,
		},
		{
			name:   "only max constraint - below",
			minMax: MinMax{Min: nil, Max: float64Ptr(90)},
			value:  50,
			want:   true,
		},
		{
			name:   "only max constraint - above",
			minMax: MinMax{Min: nil, Max: float64Ptr(90)},
			value:  95,
			want:   false,
		},
		{
			name:   "no constraints - any value",
			minMax: MinMax{Min: nil, Max: nil},
			value:  50,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.minMax.Contains(tt.value); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	val := 42
	tests := []struct {
		name  string
		input *int
		want  *float64
	}{
		{
			name:  "converts value",
			input: &val,
			want:  float64Ptr(42),
		},
		{
			name:  "nil returns nil",
			input: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFloat64(tt.input)
			if (tt.want == nil) != (got == nil) {
				t.Errorf("ToFloat64() = %v, want %v", got, tt.want)
			} else if tt.want != nil && *got != *tt.want {
				t.Errorf("ToFloat64() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestGetConstraint(t *testing.T) {
	val := float64(50)
	tests := []struct {
		name     string
		input    *float64
		fallback float64
		want     float64
	}{
		{
			name:     "returns value",
			input:    &val,
			fallback: 0,
			want:     50,
		},
		{
			name:     "nil returns fallback",
			input:    nil,
			fallback: 75,
			want:     75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConstraint(tt.input, tt.fallback)
			if got != tt.want {
				t.Errorf("GetConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}
