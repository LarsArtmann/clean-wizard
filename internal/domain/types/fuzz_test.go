package types

import (
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
)

// FuzzNixGenerationCreation tests Nix generation creation with fuzzed inputs.
func FuzzNixGenerationCreation(f *testing.F) {
	f.Add("2024-01-01")

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic on any string input
		// Randomly assign current status
		currentStatus := enums.GenerationStatusHistorical
		if len(data)%2 == 0 {
			currentStatus = enums.GenerationStatusCurrent
		}

		gen := NixGeneration{
			ID:      NixGenerationID(len(data)), // Convert string length to ID
			Path:    "/nix/store/" + data,
			Date:    time.Time{},   // Zero value for fuzzing
			Current: currentStatus, // Random current status
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
