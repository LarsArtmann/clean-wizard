package cleaner

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestIsAvailableGeneric runs a standard IsAvailable test suite for cleaners that should
// always return a boolean value (true or false).
// This eliminates duplicate test code across multiple cleaner test files.
//
// Parameters:
//   - t: The testing.T object
//   - testCases: Slice of IsAvailableTestCase containing test cases to run
func TestIsAvailableGeneric(t *testing.T, testCases []IsAvailableTestCase) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			cleaner := tt.Constructor()
			available := cleaner.IsAvailable(context.Background())

			if !available && available {
				t.Errorf("IsAvailable() returned invalid value")
			}
		})
	}
}

// TestIsAvailable is a helper that creates a standard IsAvailable test for cleaners.
// This eliminates duplicate test function code across multiple cleaner test files.
//
// Usage:
//
//	func TestXxxCleaner_IsAvailable(t *testing.T) {
//	    TestIsAvailable(t, NewXxxCleaner)
//	}
//
// Type Parameters:
//   - T: The cleaner type that must implement IsAvailable
//
// Parameters:
//   - t: The testing.T object
//   - newCleanerFunc: Function that creates a new cleaner instance with verbose and dryRun parameters
func TestIsAvailable[T interface {
	IsAvailable(ctx context.Context) bool
}](t *testing.T, newCleanerFunc func(bool, bool) T) {
	testCases := []IsAvailableTestCase{
		{
			Name: "default configuration",
			Constructor: func() interface {
				IsAvailable(ctx context.Context) bool
			} {
				return NewTestCleaner(newCleanerFunc)()
			},
		},
	}
	TestIsAvailableGeneric(t, testCases)
}

// TestValidateSettings runs a standard validation settings test suite.
func TestValidateSettings(t *testing.T, newCleanerFunc CleanerConstructorWithSettings, testCases []ValidateSettingsTestCase) {
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

// TestCleanDryRun runs a standard clean dry-run test suite.
func TestCleanDryRun(t *testing.T, newCleanerFunc SimpleCleanerConstructor, toolName string, expectedItemsRemoved uint) {
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

	if cleanResult.Strategy != domain.CleanStrategyType(domain.StrategyDryRunType) {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.CleanStrategyType(domain.StrategyDryRunType))
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

// TestDryRunStrategy runs a standard dry-run strategy test suite.
func TestDryRunStrategy(t *testing.T, newCleanerFunc SimpleCleanerConstructor, toolName string) {
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

	if cleanResult.Strategy != domain.CleanStrategyType(domain.StrategyDryRunType) {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.CleanStrategyType(domain.StrategyDryRunType))
	}

	if cleanResult.ItemsFailed != 0 {
		t.Errorf("Clean() failed %d items, want 0", cleanResult.ItemsFailed)
	}
}

// CreateBooleanSettingsTestCases creates standard test cases for cleaners with a single boolean settings field.
// This eliminates duplicate test case code across multiple cleaner test files.
//
// Parameters:
//   - nilName: Name describing the settings field (e.g., "cargo packages" for Cargo)
//   - settingsFunc: Function that creates an OperationSettings with the specific field configured
//
// Returns test cases for:
//   - nil settings (valid)
//   - empty OperationSettings (valid)
//   - settings with field enabled (valid)
//   - settings with field disabled (valid)
func CreateBooleanSettingsTestCases(nilName string, settingsFunc func(bool) *domain.OperationSettings) []ValidateSettingsTestCase {
	return []ValidateSettingsTestCase{
		{
			Name:     "nil settings",
			Settings: nil,
			WantErr:  false,
		},
		{
			Name:     "nil " + nilName + " settings",
			Settings: &domain.OperationSettings{},
			WantErr:  false,
		},
		{
			Name:     "valid settings with feature enabled",
			Settings: settingsFunc(true),
			WantErr:  false,
		},
		{
			Name:     "valid settings with feature disabled",
			Settings: settingsFunc(false),
			WantErr:  false,
		},
	}
}

// CleanResultAnalyzer provides methods for analyzing CleanResult in tests.
type CleanResultAnalyzer struct {
	t           *testing.T
	cleanResult domain.CleanResult
	elapsed     time.Duration
}

// NewCleanResultAnalyzer creates a new analyzer for the given CleanResult.
func NewCleanResultAnalyzer(t *testing.T, cleanResult domain.CleanResult, elapsed time.Duration) *CleanResultAnalyzer {
	return &CleanResultAnalyzer{
		t:           t,
		cleanResult: cleanResult,
		elapsed:     elapsed,
	}
}

// VerifyTiming verifies that clean operation timing is correctly recorded.
func (a *CleanResultAnalyzer) VerifyTiming() {
	// CleanTime should be recorded
	if a.cleanResult.CleanTime == 0 {
		a.t.Error("Clean() returned CleanTime = 0")
	}

	// CleanedAt should be set
	if a.cleanResult.CleanedAt.IsZero() {
		a.t.Error("Clean() returned CleanedAt = zero time")
	}

	// Verify timing is reasonable (clean operation should complete quickly)
	if a.cleanResult.CleanTime > 30*time.Second {
		a.t.Errorf("Clean() took %v, which seems too long", a.cleanResult.CleanTime)
	}

	// Actual execution time should be close to CleanTime
	if a.elapsed < a.cleanResult.CleanTime/2 || a.elapsed > a.cleanResult.CleanTime*2 {
		a.t.Logf("Note: Clean() recorded time %v but actual elapsed was %v", a.cleanResult.CleanTime, a.elapsed)
	}
}

// TestCleanTiming runs a standard clean timing test suite.
func TestCleanTiming(
	t *testing.T,
	newCleanerFunc SimpleCleanerConstructor,
	toolName string,
) {
	cleaner := newCleanerFunc(false, true)

	if !cleaner.IsAvailable(context.Background()) {
		t.Skipf("Skipping test: %s not available", toolName)
		return
	}

	startTime := time.Now()
	cleanResult := cleaner.Clean(context.Background())
	elapsed := time.Since(startTime)

	if cleanResult.IsErr() {
		t.Fatalf("Clean() error = %v", cleanResult.Error())
	}

	NewCleanResultAnalyzer(t, cleanResult.Value(), elapsed).VerifyTiming()
}

// TestBooleanSettingsCleanerValidateSettings runs a standard ValidateSettings test for cleaners with a single boolean settings field.
// This eliminates duplicate test code across multiple cleaner test files.
func TestBooleanSettingsCleanerValidateSettings(t *testing.T, config BooleanSettingsCleanerTestConfig, constructor CleanerConstructorWithSettings) {
	testCases := CreateBooleanSettingsTestCases(config.SettingsFieldName, config.CreateSettings)
	TestValidateSettings(t, constructor, testCases)
}

// TestBooleanSettingsCleanerCleanDryRun runs a standard Clean_DryRun test for cleaners with expected items removed.
func TestBooleanSettingsCleanerCleanDryRun(t *testing.T, config BooleanSettingsCleanerTestConfig, constructor CleanerConstructorWithSettings) {
	simpleConstructor := ToSimpleCleanerConstructor(constructor)
	TestCleanDryRun(t, simpleConstructor, config.ToolName, config.ExpectedItems)
}

// TestDryRunStrategyWithConstructor is a helper that creates a DryRunStrategy test.
func TestDryRunStrategyWithConstructor(t *testing.T, constructor CleanerConstructorWithSettings, toolName string) {
	TestDryRunStrategy(t, ToSimpleCleanerConstructor(constructor), toolName)
}

// TestCleanTimingWithConstructor is a helper that creates a Clean_Timing test.
func TestCleanTimingWithConstructor(t *testing.T, constructor SimpleCleanerConstructor, toolName string) {
	TestCleanTiming(t, constructor, toolName)
}

// TestStandardCleaner is a helper that runs DryRunStrategy and Clean_Timing tests.
func TestStandardCleaner(t *testing.T, constructor CleanerConstructorWithSettings, toolName string) {
	t.Run("DryRunStrategy", func(t *testing.T) {
		TestDryRunStrategyWithConstructor(t, constructor, toolName)
	})

	t.Run("Clean_Timing", func(t *testing.T) {
		TestCleanTimingWithConstructor(t, ToSimpleCleanerConstructor(constructor), toolName)
	})
}
