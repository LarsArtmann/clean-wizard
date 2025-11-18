package domain

import (
	"testing"
	"time"
)

// FuzzRiskLevelCreation tests risk level creation with fuzzed inputs
func FuzzRiskLevelCreation(f *testing.F) {
	f.Add(RiskLow.String())
	f.Add(RiskMedium.String())
	f.Add(RiskHigh.String())
	f.Add(RiskCritical.String())

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string value
		// Convert string hash to int for RiskLevelType simulation
		level := RiskLevelType(len(data) % 10) // Simple hash to int conversion

		// IsValid method should not panic
		_ = level.IsValid()

		// Icon method should not panic
		_ = level.Icon()

		// Should handle any string value gracefully
		if level.IsValid() {
			// Valid level, string should be meaningful
			str := level.String()
			if str == "" {
				t.Logf("Valid level %s produced empty string", level.String())
			}
		}
	})
}

// FuzzNixGenerationCreation tests Nix generation creation with fuzzed inputs
func FuzzNixGenerationCreation(f *testing.F) {
	f.Add("2024-01-01")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string input
		gen := NixGeneration{
			ID:      len(data), // Convert string length to ID
			Path:    "/nix/store/" + data,
			Date:    time.Time{},      // Zero value for fuzzing
			Current: len(data)%2 == 0, // Random current status
		}

		// Should not panic on creation
		_ = gen.IsValid()

		// Should handle various string lengths gracefully
		if len(data) > 100000 { // Prevent excessive memory usage in fuzzing
			return
		}

		// ID validation should be reasonable
		if gen.ID < 0 {
			t.Logf("Invalid generation ID: %d", gen.ID)
		}
	})
}
