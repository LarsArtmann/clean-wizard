package domain

import (
	"testing"
	"time"
)

// makeTestCleanResult creates a CleanResult for testing with common defaults.
// This reduces duplication across test cases while maintaining readability.
func makeTestCleanResult(known uint64, itemsRemoved, itemsFailed uint, strategy CleanStrategyType) CleanResult {
	return CleanResult{
		SizeEstimate: SizeEstimate{Known: known, Status: SizeEstimateStatusKnown},
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  itemsFailed,
		CleanTime:    time.Second,
		CleanedAt:    time.Now(),
		Strategy:     strategy,
	}
}

// TestCleanResultValidation tests the enhanced CleanResult validation.
func TestCleanResultValidation(t *testing.T) {
	tests := []struct {
		name        string
		result      CleanResult
		shouldValid bool
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid CleanResult with aggressive strategy",
			result:      makeTestCleanResult(1024, 5, 0, StrategyAggressive),
			shouldValid: true,
			shouldError: false,
		},
		{
			name:        "Valid CleanResult with conservative strategy",
			result:      makeTestCleanResult(512, 2, 1, StrategyConservative),
			shouldValid: true,
			shouldError: false,
		},
		{
			name: "Valid CleanResult with dry-run strategy",
			result: CleanResult{
				FreedBytes:   0,
				ItemsRemoved: 0,
				ItemsFailed:  0,
				CleanTime:    0,
				CleanedAt:    time.Now(),
				Strategy:     CleanStrategyType(StrategyDryRunType),
			},
			shouldValid: true,
			shouldError: false,
		},
		{
			name:        "Invalid CleanResult - zero freed bytes with items removed",
			result:      makeTestCleanResult(0, 1, 0, StrategyAggressive),
			shouldValid: true, // IsValid() returns true (quick check), but Validate() returns error
			shouldError: true,
			errorMsg: "cannot have zero SizeEstimate when ItemsRemoved is > 0 " +
				"(set Status: Unknown if size cannot be determined)",
		},
		{
			name:        "Invalid CleanResult - failed items with no freed bytes",
			result:      makeTestCleanResult(0, 0, 5, StrategyConservative),
			shouldValid: false,
			shouldError: true,
			errorMsg:    "cannot have failed items when no items were processed",
		},
		{
			name: "Invalid CleanResult - zero cleaned at time",
			result: CleanResult{
				SizeEstimate: SizeEstimate{Known: 100, Status: SizeEstimateStatusKnown},
				ItemsRemoved: 1,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Time{}, // Zero time
				Strategy:     CleanStrategyType(StrategyDryRunType),
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "cleanedAt cannot be zero",
		},
		{
			name: "Invalid CleanResult - invalid strategy",
			result: CleanResult{
				SizeEstimate: SizeEstimate{Known: 100, Status: SizeEstimateStatusKnown},
				ItemsRemoved: 1,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Now(),
				Strategy:     CleanStrategyType(999), // Invalid strategy (out of range)
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "invalid strategy: unknown (must be 'aggressive', 'conservative', or 'dry-run')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test IsValid method
			isValid := tt.result.IsValid()
			if isValid != tt.shouldValid {
				t.Errorf("IsValid() = %v, expected %v", isValid, tt.shouldValid)
			}

			// Test Validate method
			err := tt.result.Validate()
			if tt.shouldError {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Validate() error = %v, expected %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() expected no error but got %v", err)
				}
			}
		})
	}
}
