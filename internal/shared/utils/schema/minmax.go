package schema

import (
	"math"
)

// MinMax represents minimum and maximum bounds for a numeric value.
type MinMax struct {
	Min *float64
	Max *float64
}

// NewMinMax creates a new MinMax with the provided min and max values.
func NewMinMax(min, max *float64) MinMax {
	return MinMax{Min: min, Max: max}
}

// ToFloat64 safely converts an int pointer to float64 pointer.
func ToFloat64(v *int) *float64 {
	if v == nil {
		return nil
	}
	result := float64(*v)
	return &result
}

// GetConstraint extracts a constraint value, returning fallback if nil.
func GetConstraint(constraint *float64, fallback float64) float64 {
	if constraint != nil {
		return *constraint
	}
	return fallback
}

// ExtractMinMax extracts min and max values from int pointers, converting to float64.
func ExtractMinMax(min, max *int, minFallback, maxFallback float64) MinMax {
	return MinMax{
		Min: extractFloat64(min, minFallback),
		Max: extractFloat64(max, maxFallback),
	}
}

// extractFloat64 converts int pointer to float64 with fallback.
func extractFloat64(v *int, fallback float64) *float64 {
	if v == nil {
		return &fallback
	}
	result := float64(*v)
	return &result
}

// DefaultMinMax returns default min/max values when constraints are not specified.
func DefaultMinMax(min, max float64) MinMax {
	return MinMax{
		Min: &min,
		Max: &max,
	}
}

// IsEmpty checks if both min and max are nil (no constraints).
func (m MinMax) IsEmpty() bool {
	return m.Min == nil && m.Max == nil
}

// Range returns the total range (max - min) if both values exist.
func (m MinMax) Range() float64 {
	if m.Min == nil || m.Max == nil {
		return math.NaN()
	}
	return *m.Max - *m.Min
}

// Contains checks if a value is within the min/max bounds.
func (m MinMax) Contains(value float64) bool {
	if m.Min != nil && value < *m.Min {
		return false
	}
	if m.Max != nil && value > *m.Max {
		return false
	}
	return true
}
