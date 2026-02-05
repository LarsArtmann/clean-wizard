package cleaner

import (
	"context"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ValidateSettingsTestCase represents a test case for ValidateSettings
type ValidateSettingsTestCase struct {
	Name     string
	Settings *domain.OperationSettings
	WantErr  bool
}

// CleanerConstructor is a function type for creating cleaners in tests
type CleanerConstructor func(verbose, dryRun bool) interface{
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) *domain.Result
	ValidateSettings(*domain.OperationSettings) error
}

// TestValidateSettings runs a standard validation settings test suite
func TestValidateSettings(t *testing.T, newCleanerFunc CleanerConstructor, testCases []ValidateSettingsTestCase) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			cleaner := newCleanerFunc(false, false)

			err := cleaner.ValidateSettings(tt.Settings)
			if (err != nil) != tt.WantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.WantErr)
			}
		})
	}
}

// TestCleanDryRun runs a standard clean dry-run test suite
func TestCleanDryRun(t *testing.T, newCleanerFunc CleanerConstructor, toolName string, expectedItemsRemoved int) {
	cleaner := newCleanerFunc(false, true)

	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: %s not available", toolName)
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	if cleanResult.ItemsRemoved != expectedItemsRemoved {
		t.Errorf("Clean() removed %d items, want %d", cleanResult.ItemsRemoved, expectedItemsRemoved)
	}

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

// TestDryRunStrategy runs a standard dry-run strategy test suite
func TestDryRunStrategy(t *testing.T, newCleanerFunc CleanerConstructor, toolName string) {
	cleaner := newCleanerFunc(false, true)

	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: %s not available", toolName)
		return
	}

	result := cleaner.Clean(context.Background())
	if result.IsErr() {
		t.Fatalf("Clean() error = %v", result.Error())
	}

	cleanResult := result.Value()

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.ItemsFailed != 0 {
		t.Errorf("Clean() failed %d items, want 0", cleanResult.ItemsFailed)
	}
}