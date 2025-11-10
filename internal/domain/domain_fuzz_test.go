package domain

import (
	"testing"
	"time"
)

// FuzzRiskLevelCreation tests risk level creation with fuzzed inputs
func FuzzRiskLevelCreation(f *testing.F) {
	f.Add(string(RiskLow))
	f.Add(string(RiskMedium))
	f.Add(string(RiskHigh))
	f.Add(string(RiskCritical))
	
	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string value
		level := RiskLevel(data)
		
		// IsValid method should not panic
		_ = level.IsValid()
		
		// Icon method should not panic
		_ = level.Icon()
		
		// Should handle any string value gracefully
		if level.IsValid() {
			// Valid level, string should be meaningful
			str := string(level)
			if str == "" {
				t.Logf("Valid level %s produced empty string", level)
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
			Date:    time.Time{}, // Zero value for fuzzing
			Current: len(data) % 2 == 0, // Random current status
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