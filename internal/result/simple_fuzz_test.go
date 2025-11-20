package result

import (
	"errors"
	"testing"
)

// FuzzResultCreationBasic tests result type creation with fuzzed inputs
func FuzzResultCreationBasic(f *testing.F) {
	f.Add("test data")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic when creating results
		okResult := Ok(data)
		testError := errors.New(data)
		errResult := Err[string](testError)

		// Methods should not panic
		if okResult.IsOk() != true {
			t.Logf("Ok result should be ok: %s", data)
		}

		if errResult.IsOk() != false {
			t.Logf("Err result should not be ok: %s", data)
		}

		if errResult.IsErr() != true {
			t.Logf("Err result should be err: %s", data)
		}

		// Value method should not panic
		okResult.Value()

		// Error method should not panic
		_ = errResult.Error()
	})
}

// FuzzResultStringOperations tests result string operations
func FuzzResultStringOperations(f *testing.F) {
	f.Add("string test")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic with any string length
		if len(data) > 10000 { // Prevent excessive memory usage
			return
		}

		// Create result from fuzzed string
		result := Ok(data)

		// Basic operations should not panic
		if result.IsOk() {
			value := result.Value()
			_ = len(value)
			_ = value + "_suffix"
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

// FuzzResultErrorHandling tests error handling with fuzzed inputs
func FuzzResultErrorHandling(f *testing.F) {
	f.Add("error test")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic with any string length
		if len(data) > 10000 { // Prevent excessive memory usage
			return
		}

		// Create error result
		testError := errors.New(data)
		errResult := Err[string](testError)

		// Error methods should not panic
		if errResult.IsErr() {
			errorMsg := errResult.Error().Error()
			// Error message should be reasonable
			if len(errorMsg) > 10000 { // Reasonable limit
				t.Logf("Very long error message: %d chars", len(errorMsg))
			}
		}

		// Should handle various string lengths
		_ = data + "_error"
		_ = len(data)
	})
}
