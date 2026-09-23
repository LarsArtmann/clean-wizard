package cleaner

import (
	"reflect"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
)

// ToSimpleCleanerConstructor converts a constructor with additional methods to one that
// only exposes Clean and IsAvailable.
func ToSimpleCleanerConstructor(
	fullConstructor CleanerConstructorWithSettings,
) SimpleCleanerConstructor {
	return func(verbose, dryRun bool) SimpleCleaner {
		return fullConstructor(verbose, dryRun)
	}
}

// SimpleCleanerConstructorFromInstance creates a SimpleCleanerConstructor from an existing cleaner instance.
func SimpleCleanerConstructorFromInstance[T SimpleCleaner](cleaner T) SimpleCleanerConstructor {
	return func(verbose, dryRun bool) SimpleCleaner {
		return cleaner
	}
}

// cleanerConstructorWithSettingsAdapter wraps a constructor into CleanerConstructorWithSettings.
func cleanerConstructorWithSettingsAdapter[T CleanerWithSettings](
	constructor CleanerConstructor[T],
) CleanerConstructorWithSettings {
	return func(verbose, dryRun bool) CleanerWithSettings {
		return constructor(verbose, dryRun)
	}
}

// NewCleanerConstructorWithSettings creates a CleanerConstructorWithSettings from a constructor function
// that takes additional manager types parameter.
func NewCleanerConstructorWithSettings[T CleanerWithSettings, M any](
	constructor func(verbose, dryRun bool, managers []M) T,
	availableManagers func() []M,
) CleanerConstructorWithSettings {
	return cleanerConstructorWithSettingsAdapter(func(verbose, dryRun bool) T {
		return constructor(verbose, dryRun, availableManagers())
	})
}

// NewBooleanSettingsCleanerTestConstructor is a helper that creates a CleanerConstructorWithSettings
// from a cleaner constructor function.
func NewBooleanSettingsCleanerTestConstructor[T CleanerWithSettings](
	constructor CleanerConstructor[T],
) CleanerConstructorWithSettings {
	return cleanerConstructorWithSettingsAdapter(constructor)
}

// NewBooleanSettingsCleanerTestConfig creates a BooleanSettingsCleanerTestConfig with standardized values.
func NewBooleanSettingsCleanerTestConfig[T CleanerWithSettings](
	testName string,
	toolName string,
	settingsFieldName string,
	expectedItems uint,
	newCleanerFunc CleanerConstructor[T],
	createSettings func(bool) *operations.OperationSettings,
) BooleanSettingsCleanerTestConfig {
	return BooleanSettingsCleanerTestConfig{
		TestName:          testName,
		ToolName:          toolName,
		SettingsFieldName: settingsFieldName,
		ExpectedItems:     expectedItems,
		Constructor:       NewBooleanSettingsCleanerTestConstructor(newCleanerFunc),
		CreateSettings:    createSettings,
	}
}

// NewBooleanSettingsCleanerTestConfigFn creates a BooleanSettingsCleanerTestConfig
// from constructor and settings creation function.
func NewBooleanSettingsCleanerTestConfigFn[T CleanerWithSettings](
	testName, toolName, settingsFieldName string,
	expectedItems uint,
	constructor CleanerConstructor[T],
	createSettings func(bool) *operations.OperationSettings,
) BooleanSettingsCleanerTestConfig {
	return BooleanSettingsCleanerTestConfig{
		TestName:          testName,
		ToolName:          toolName,
		SettingsFieldName: settingsFieldName,
		ExpectedItems:     expectedItems,
		Constructor:       NewBooleanSettingsCleanerTestConstructor(constructor),
		CreateSettings:    createSettings,
	}
}

// NewTestCleaner creates a cleaner with default test settings (verbose=false, dryRun=false).
func NewTestCleaner[T any](constructor CleanerConstructor[T]) func() T {
	return func() T {
		return constructor(false, false)
	}
}

// VerifyNewCleanerConstructor tests a cleaner constructor function with different
// combinations of verbose and dryRun parameters.
// This eliminates duplicate test code across multiple cleaner test files.
//
// Usage:
//
//	func TestNewXxxCleaner(t *testing.T) {
//	    VerifyNewCleanerConstructor(t, NewXxxCleaner, "NewXxxCleaner")
//	}
//
// Type Parameters:
//   - T: The cleaner type that must have verbose and dryRun fields
//
// Parameters:
//   - t: The testing.T object
//   - constructor: Function that creates a cleaner with given verbose and dryRun flags
//   - cleanerName: Name of the cleaner for error messages
func VerifyNewCleanerConstructor[T any](
	t *testing.T,
	constructor func(bool, bool) T,
	cleanerName string,
) {
	t.Helper()

	tests := []struct {
		name    string
		verbose bool
		dryRun  bool
	}{
		{
			name:    "standard configuration",
			verbose: false,
			dryRun:  false,
		},
		{
			name:    "verbose mode",
			verbose: true,
			dryRun:  false,
		},
		{
			name:    "dry-run mode",
			verbose: false,
			dryRun:  true,
		},
		{
			name:    "verbose dry-run mode",
			verbose: true,
			dryRun:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := constructor(tt.verbose, tt.dryRun)

			cleanerValue := reflect.ValueOf(cleaner)
			if cleanerValue.Kind() == reflect.Pointer {
				if cleanerValue.IsNil() {
					t.Fatalf("%s(%v, %v) returned nil cleaner", cleanerName, tt.verbose, tt.dryRun)
				}

				cleanerValue = cleanerValue.Elem()
			} else {
				cleanerValue = cleanerValue.Elem()
			}

			verboseField := cleanerValue.FieldByName("verbose")
			if !verboseField.IsValid() {
				t.Fatalf("%s cleaner does not have 'verbose' field", cleanerName)
			}

			if verboseField.Bool() != tt.verbose {
				t.Errorf("verbose = %v, want %v", verboseField.Bool(), tt.verbose)
			}

			dryRunField := cleanerValue.FieldByName("dryRun")
			if !dryRunField.IsValid() {
				t.Fatalf("%s cleaner does not have 'dryRun' field", cleanerName)
			}

			if dryRunField.Bool() != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", dryRunField.Bool(), tt.dryRun)
			}
		})
	}
}
