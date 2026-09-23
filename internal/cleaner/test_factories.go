package cleaner

import (
	"reflect"
	"testing"

	"github.com/onsi/gomega"

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

// availableItemsTestHelper is a helper function for testing Available* functions.
// This is called by type-specific test wrappers.
func AvailableItemsTestHelper[T comparable](
	t *testing.T,
	expectedItems []T,
	availableFn func() []T,
	testName string,
) {
	t.Helper()

	items := availableFn()

	if len(items) != len(expectedItems) {
		t.Errorf("%s() returned %d items, want %d", testName, len(items), len(expectedItems))
	}

	for i, item := range items {
		if !reflect.DeepEqual(item, expectedItems[i]) {
			t.Errorf("%s()[%d] = %v, want %v", testName, i, item, expectedItems[i])
		}
	}
}

// assertValidationError is a helper for testing that ValidateSettings
// returns expected errors. Consolidates duplicate validation test patterns
// across ginkgo test files.
func AssertValidationError(
	cleaner CleanerWithSettings, settings *operations.OperationSettings, expectedErrSubstring string,
) {
	err := cleaner.ValidateSettings(settings)
	gomega.Expect(err).To(gomega.HaveOccurred())
	gomega.Expect(err.Error()).To(gomega.ContainSubstring(expectedErrSubstring))
}

// DurationParseTestCase holds test data for duration parsing tests.
type DurationParseTestCase struct {
	Duration  string
	WantValid bool
}

// CommonDurationTestCases provides shared test cases for duration parsing tests.
// These are used by both BuildCacheCleaner and SystemCacheCleaner.
var CommonDurationTestCases = []DurationParseTestCase{
	{Duration: "1h", WantValid: true},
	{Duration: "24h", WantValid: true},
	{Duration: "7d", WantValid: true},
	{Duration: "30d", WantValid: true},
	{Duration: "1w", WantValid: false}, // Not supported
	{Duration: "invalid", WantValid: false},
}

// BooleanSettingsCleanerTestCase represents a test case for cleaners with boolean settings.
type BooleanSettingsCleanerTestCase struct {
	Name   string
	Config BooleanSettingsTestConfig
}

func CreateBooleanSettingsTest(t *testing.T, config BooleanSettingsTestConfig) {
	t.Helper()
	CreateBooleanSettingsCleanerTestFunctions(t, BooleanSettingsCleanerTestConfig{
		TestName:          config.TestName,
		ToolName:          config.ToolName,
		SettingsFieldName: config.SettingsFieldName,
		ExpectedItems:     config.ExpectedItems,
		Constructor:       config.Constructor,
		CreateSettings:    config.CreateSettingsFunc,
	})
}

func RunGetHomeDirTests(t *testing.T, testCases []GetHomeDirTestCase) {
	t.Helper()

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Setenv("HOME", tt.HomeValue)
			t.Setenv("USERPROFILE", tt.ProfileValue)

			home, err := GetHomeDir()

			if tt.WantErr {
				if err == nil {
					t.Errorf("GetHomeDir() error = %v, want error for missing home", err)
				}

				if home != "" {
					t.Errorf("GetHomeDir() = %v, want empty string", home)
				}
			} else {
				if err != nil {
					t.Errorf("GetHomeDir() error = %v", err)
				}

				if home != tt.WantHome {
					t.Errorf("GetHomeDir() = %v, want %v", home, tt.WantHome)
				}
			}
		})
	}
}

// CreateBooleanSettingsCleanerTestFunctions creates both ValidateSettings and Clean_DryRun test functions
// for cleaners with a boolean settings field. This eliminates duplicate config and constructor code.
//
// Usage:
//
//	func TestXxxCleaner_BooleanSettingsTests(t *testing.T) {
//	    CreateBooleanSettingsCleanerTestFunctions(t, BooleanSettingsCleanerTestConfig{
//	        TestName:          "Xxx",
//	        ToolName:          "xxx-tool",
//	        SettingsFieldName: "xxx settings",
//	        CreateSettings: func(enabled bool) *domain.OperationSettings {
//	            return &domain.OperationSettings{
//	                XxxSettings: &domain.XxxSettings{
//	                    Enabled: enabled,
//	                },
//	            }
//	        },
//	        ExpectedItems: 1,
//	        Constructor:   NewBooleanSettingsCleanerTestConstructor(NewXxxCleaner),
//	    })
//	}
func CreateBooleanSettingsCleanerTestFunctions(
	t *testing.T,
	config BooleanSettingsCleanerTestConfig,
) {
	t.Helper()
	t.Run("ValidateSettings", func(t *testing.T) {
		TestBooleanSettingsCleanerValidateSettings(t, config, config.Constructor)
	})

	t.Run("Clean_DryRun", func(t *testing.T) {
		TestBooleanSettingsCleanerCleanDryRun(t, config, config.Constructor)
	})
}
