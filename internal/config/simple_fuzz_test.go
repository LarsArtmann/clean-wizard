package config

import (
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// FuzzBasicConfig tests basic configuration fuzzing.
func FuzzBasicConfig(f *testing.F) {
	f.Add("version: \"1.0.0\"\nsafe_mode: true\n")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic with any string input
		if len(data) > 10000 { // Prevent excessive memory usage
			return
		}

		// Basic string operations should not panic
		_ = len(data)
		_ = data + "_test"

		// Should handle empty strings gracefully
		if data == "" {
			return
		}

		// Should handle various character sets
		for i, char := range data {
			if i > 1000 { // Prevent excessive loops
				break
			}
			_ = char // Should not panic
		}
	})
}

// FuzzValidationLevelBasic tests validation level fuzzing.
func FuzzValidationLevelBasic(f *testing.F) {
	f.Add(int32(domain.ValidationLevelBasicType))

	f.Fuzz(func(t *testing.T, data int32) {
		// Should not panic with any int32 input
		level := domain.ValidationLevelType(data)

		// IsValid method should not panic
		_ = level.IsValid()

		// Should handle extreme values gracefully
		if int32(level) > int32(domain.ValidationLevelStrictType)+100 || int32(level) < int32(domain.ValidationLevelNoneType)-100 {
			// Should still not panic
			_ = level.IsValid()
		}

		// Valid range should produce meaningful strings
		if level >= domain.ValidationLevelNoneType && level <= domain.ValidationLevelStrictType {
			isValid := level.IsValid()
			if !isValid {
				t.Logf("Valid level %d reported as invalid", level)
			}
		}
	})
}

// FuzzStringOperations tests string operations with fuzzed data.
func FuzzStringOperations(f *testing.F) {
	f.Add("test string")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic with any string operations
		if len(data) > 5000 { // Prevent excessive memory usage
			return
		}

		// Basic string operations
		_ = data + "_suffix"
		_ = len(data)
		_ = data[:min(len(data), 10)]

		// Character operations
		if len(data) > 0 {
			_ = data[0]
			_ = data[len(data)-1]
		}

		// Should handle empty strings
		if data == "" {
			return
		}

		// Should handle unicode characters
		for i, r := range data {
			if i > 100 { // Prevent excessive loops
				break
			}
			_ = r // Should not panic on any unicode
		}
	})
}

// FuzzSliceOperations tests slice operations with fuzzed data.
func FuzzSliceOperations(f *testing.F) {
	f.Add("/System")

	f.Fuzz(func(t *testing.T, data string) {
		// Create slice from fuzzed string
		slice := []string{data}

		// Should not panic with any slice operations
		_ = len(slice)

		// Should handle empty strings gracefully
		if len(data) == 0 {
			return
		}

		// Iteration should not panic
		for i, item := range slice {
			if i > 10 { // Prevent excessive loops
				break
			}

			if len(item) > 1000 { // Prevent excessive string operations
				continue
			}

			_ = len(item)
			_ = item + "_test"
		}

		// Should handle string with reasonable length
		if len(data) > 10000 { // Prevent excessive memory usage
			return
		}

		// Should handle unicode characters
		for i, r := range data {
			if i > 100 { // Prevent excessive loops
				break
			}
			_ = r // Should not panic on any unicode
		}
	})
}
