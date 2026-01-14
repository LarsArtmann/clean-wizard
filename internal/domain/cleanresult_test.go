package domain

import (
	"testing"
	"time"
)

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
			name: "Valid CleanResult with aggressive strategy",
			result: CleanResult{
				FreedBytes:   1024,
				ItemsRemoved: 5,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Now(),
				Strategy:     StrategyAggressive,
			},
			shouldValid: true,
			shouldError: false,
		},
		{
			name: "Valid CleanResult with conservative strategy",
			result: CleanResult{
				FreedBytes:   512,
				ItemsRemoved: 2,
				ItemsFailed:  1,
				CleanTime:    time.Second * 2,
				CleanedAt:    time.Now(),
				Strategy:     StrategyConservative,
			},
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
				Strategy:     StrategyDryRun,
			},
			shouldValid: true,
			shouldError: false,
		},
		{
			name: "Invalid CleanResult - zero freed bytes with items removed",
			result: CleanResult{
				FreedBytes:   0,
				ItemsRemoved: 1,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Now(),
				Strategy:     StrategyAggressive,
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "cannot have zero FreedBytes when ItemsRemoved is > 0",
		},
		{
			name: "Invalid CleanResult - failed items with no freed bytes",
			result: CleanResult{
				FreedBytes:   0,
				ItemsRemoved: 0,
				ItemsFailed:  5,
				CleanTime:    time.Second,
				CleanedAt:    time.Now(),
				Strategy:     StrategyConservative,
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "cannot have failed items when no items were processed",
		},
		{
			name: "Invalid CleanResult - zero cleaned at time",
			result: CleanResult{
				FreedBytes:   100,
				ItemsRemoved: 1,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Time{}, // Zero time
				Strategy:     StrategyDryRun,
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "CleanedAt cannot be zero",
		},
		{
			name: "Invalid CleanResult - invalid strategy",
			result: CleanResult{
				FreedBytes:   100,
				ItemsRemoved: 1,
				ItemsFailed:  0,
				CleanTime:    time.Second,
				CleanedAt:    time.Now(),
				Strategy:     CleanStrategyType(999), // Invalid strategy (out of range)
			},
			shouldValid: false,
			shouldError: true,
			errorMsg:    "Invalid strategy: unknown (must be 'aggressive', 'conservative', or 'dry-run')",
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
