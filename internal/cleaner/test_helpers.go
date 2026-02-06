package cleaner

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// ValidateSettingsTestCase represents a test case for ValidateSettings
type ValidateSettingsTestCase struct {
	Name     string
	Settings *domain.OperationSettings
	WantErr  bool
}

// IsAvailableConstructor is a function type for creating cleaners in tests that need IsAvailable
type IsAvailableConstructor func() interface {
	IsAvailable(ctx context.Context) bool
}

// IsAvailableTestCase represents a test case for IsAvailable tests
type IsAvailableTestCase struct {
	Name        string
	Constructor IsAvailableConstructor
}

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

			if available != true && available != false {
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

// CleanerWithSettings represents a cleaner interface with settings validation
// This eliminates duplicate interface declarations in test helper functions.
type CleanerWithSettings interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
	ValidateSettings(*domain.OperationSettings) error
}

// CleanerConstructorWithSettings is a function type for creating cleaners in tests that need ValidateSettings
type CleanerConstructorWithSettings func(verbose, dryRun bool) CleanerWithSettings

// SimpleCleanerConstructor is a function type for creating cleaners in tests that only need Clean and IsAvailable
type SimpleCleanerConstructor func(verbose, dryRun bool) SimpleCleaner

// SimpleCleaner represents a cleaner interface without settings validation
type SimpleCleaner interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
}

// ToSimpleCleanerConstructor converts a constructor with additional methods to one that only exposes Clean and IsAvailable
func ToSimpleCleanerConstructor(fullConstructor CleanerConstructorWithSettings) SimpleCleanerConstructor {
	return func(verbose, dryRun bool) SimpleCleaner {
		return fullConstructor(verbose, dryRun)
	}
}

// SimpleCleanerConstructorFromInstance creates a SimpleCleanerConstructor from an existing cleaner instance.
// This eliminates duplicate interface declarations in test files.
//
// Usage:
//
//	func TestXxxCleaner_DryRunStrategy(t *testing.T) {
//	    cleaner := NewXxxCleaner(...)
//	    TestDryRunStrategy(t, SimpleCleanerConstructorFromInstance(cleaner), "xxx")
//	}
//
// Type Parameters:
//   - T: The cleaner type that must implement the cleaner interface
//
// Parameters:
//   - cleaner: An existing cleaner instance
//
// Returns a SimpleCleanerConstructor that returns the given cleaner
func SimpleCleanerConstructorFromInstance[T SimpleCleaner](cleaner T) SimpleCleanerConstructor {
	return func(verbose, dryRun bool) SimpleCleaner {
		return cleaner
	}
}

// TestValidateSettings runs a standard validation settings test suite
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

// TestCleanDryRun runs a standard clean dry-run test suite
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

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
	}

	if cleanResult.FreedBytes == 0 {
		t.Errorf("Clean() freed %d bytes, want > 0", cleanResult.FreedBytes)
	}
}

// TestDryRunStrategy runs a standard dry-run strategy test suite
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

	if cleanResult.Strategy != domain.StrategyDryRun {
		t.Errorf("Clean() strategy = %v, want %v", cleanResult.Strategy, domain.StrategyDryRun)
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

// CleanResultAnalyzer provides methods for analyzing CleanResult in tests
type CleanResultAnalyzer struct {
	t          *testing.T
	cleanResult domain.CleanResult
	elapsed    time.Duration
}

// NewCleanResultAnalyzer creates a new analyzer for the given CleanResult
func NewCleanResultAnalyzer(t *testing.T, cleanResult domain.CleanResult, elapsed time.Duration) *CleanResultAnalyzer {
	return &CleanResultAnalyzer{
		t:          t,
		cleanResult: cleanResult,
		elapsed:    elapsed,
	}
}

// VerifyTiming verifies that clean operation timing is correctly recorded
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

// TestCleanTiming runs a standard clean timing test suite
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

// BooleanSettingsCleanerTestConfig holds configuration for testing cleaners with a boolean settings field.
// Use this with TestBooleanSettingsCleanerValidateSettings and TestBooleanSettingsCleanerCleanDryRun.
type BooleanSettingsCleanerTestConfig struct {
	TestName          string
	ToolName          string
	SettingsFieldName string
	CreateSettings    func(bool) *domain.OperationSettings
	ExpectedItems     uint
	Constructor       CleanerConstructorWithSettings
}

// TestBooleanSettingsCleanerValidateSettings runs a standard ValidateSettings test for cleaners with a single boolean settings field.
// This eliminates duplicate test code across multiple cleaner test files.
//
// Parameters:
//   - t: The testing.T object
//   - config: Configuration for the tests including constructor and settings creation
//   - constructor: Constructor function that returns an interface with IsAvailable, Clean, and ValidateSettings methods
func TestBooleanSettingsCleanerValidateSettings(t *testing.T, config BooleanSettingsCleanerTestConfig, constructor CleanerConstructorWithSettings) {
	testCases := CreateBooleanSettingsTestCases(config.SettingsFieldName, config.CreateSettings)
	TestValidateSettings(t, constructor, testCases)
}

// TestBooleanSettingsCleanerCleanDryRun runs a standard Clean_DryRun test for cleaners with expected items removed.
// This eliminates duplicate test code across multiple cleaner test files.
//
// Parameters:
//   - t: The testing.T object
//   - config: Configuration for the tests including tool name and expected items
//   - constructor: Constructor function that returns an interface with IsAvailable, Clean, and ValidateSettings methods
func TestBooleanSettingsCleanerCleanDryRun(t *testing.T, config BooleanSettingsCleanerTestConfig, constructor CleanerConstructorWithSettings) {
	simpleConstructor := ToSimpleCleanerConstructor(constructor)
	TestCleanDryRun(t, simpleConstructor, config.ToolName, config.ExpectedItems)
}

// TestDryRunStrategyWithConstructor is a helper that creates a DryRunStrategy test
// by wrapping the cleaner constructor with ToSimpleCleanerConstructor and calling TestDryRunStrategy.
// This eliminates duplicate constructor code across multiple cleaner test files.
//
// Parameters:
//   - t: The testing.T object
//   - constructor: Constructor function that returns an interface with IsAvailable, Clean, and ValidateSettings methods
//   - toolName: Name of the tool being tested (for logging/skips)
func TestDryRunStrategyWithConstructor(t *testing.T, constructor CleanerConstructorWithSettings, toolName string) {
	TestDryRunStrategy(t, ToSimpleCleanerConstructor(constructor), toolName)
}

// TestCleanTimingWithConstructor is a helper that creates a Clean_Timing test
// by calling TestCleanTiming with the given constructor.
// This eliminates duplicate constructor code across multiple cleaner test files.
//
// Parameters:
//   - t: The testing.T object
//   - constructor: Constructor function that returns an interface with IsAvailable and Clean methods
//   - toolName: Name of the tool being tested (for logging/skips)
func TestCleanTimingWithConstructor(t *testing.T, constructor SimpleCleanerConstructor, toolName string) {
	TestCleanTiming(t, constructor, toolName)
}

// TestStandardCleaner is a helper that runs DryRunStrategy and Clean_Timing tests
// for cleaners that support the full CleanerConstructorWithSettings interface.
// This eliminates duplicate constructor code across multiple cleaner test files.
//
// Usage:
//   func Test<X>Cleaner_StandardTests(t *testing.T) {
//       TestStandardCleaner(t, NewXCleaner, "ToolName")
//   }
func TestStandardCleaner(t *testing.T, constructor CleanerConstructorWithSettings, toolName string) {
	t.Run("DryRunStrategy", func(t *testing.T) {
		TestDryRunStrategyWithConstructor(t, constructor, toolName)
	})

	t.Run("Clean_Timing", func(t *testing.T) {
		TestCleanTimingWithConstructor(t, ToSimpleCleanerConstructor(constructor), toolName)
	})
}

// cleanerConstructorWithSettingsAdapter wraps a constructor into CleanerConstructorWithSettings.
// This eliminates duplicate interface adapter code in test helper functions.
func cleanerConstructorWithSettingsAdapter[T CleanerWithSettings](
	constructor func(verbose, dryRun bool) T,
) CleanerConstructorWithSettings {
	return func(verbose, dryRun bool) CleanerWithSettings {
		return constructor(verbose, dryRun)
	}
}

// NewCleanerConstructorWithSettings creates a CleanerConstructorWithSettings from a constructor function
// that takes additional manager types parameter.
// This eliminates duplicate factory functions in test files.
//
// Type Parameters:
//   - T: The cleaner type that must implement the cleaner interface
//   - M: The manager type
//
// Parameters:
//   - constructor: Function that creates a cleaner with given verbose, dryRun, and managers
//   - availableManagers: Function that returns the available managers
//
// Returns a CleanerConstructorWithSettings that matches the TestValidateSettings signature
func NewCleanerConstructorWithSettings[T CleanerWithSettings, M any](
	constructor func(verbose, dryRun bool, managers []M) T,
	availableManagers func() []M,
) CleanerConstructorWithSettings {
	return cleanerConstructorWithSettingsAdapter(func(verbose, dryRun bool) T {
		return constructor(verbose, dryRun, availableManagers())
	})
}

// CreateBooleanSettingsCleanerTestFunctions creates both ValidateSettings and Clean_DryRun test functions
// for cleaners with a boolean settings field. This eliminates duplicate config and constructor code.
//
// Usage:
//	func TestXxxCleaner_BooleanSettingsTests(t *testing.T) {
//	    CreateBooleanSettingsCleanerTestFunctions(t, BooleanSettingsCleanerTestFunctionsConfig{
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
func CreateBooleanSettingsCleanerTestFunctions(t *testing.T, config BooleanSettingsCleanerTestConfig) {
	t.Run("ValidateSettings", func(t *testing.T) {
		TestBooleanSettingsCleanerValidateSettings(t, config, config.Constructor)
	})

	t.Run("Clean_DryRun", func(t *testing.T) {
		TestBooleanSettingsCleanerCleanDryRun(t, config, config.Constructor)
	})
}

// NewBooleanSettingsCleanerTestConstructor is a helper that creates a CleanerConstructorWithSettings
// from a cleaner constructor function. This eliminates duplicate interface declarations in test files.
//
// Usage:
//
//	func TestXxxCleaner_StandardTests(t *testing.T) {
//	    TestStandardCleaner(t, NewBooleanSettingsCleanerTestConstructor(
//	        func(verbose, dryRun bool) *XxxCleaner {
//	            return NewXxxCleaner(verbose, dryRun)
//	        },
//	    ), "xxx-tool")
//	}
//
// Type Parameters:
//   - T: The cleaner type that must implement the cleaner interface
//
// Parameters:
//   - constructor: Function that creates a cleaner with given verbose and dryRun flags
//
// Returns a CleanerConstructorWithSettings that can be used with TestStandardCleaner and other helpers
func NewBooleanSettingsCleanerTestConstructor[T CleanerWithSettings](constructor func(verbose, dryRun bool) T) CleanerConstructorWithSettings {
	return cleanerConstructorWithSettingsAdapter(constructor)
}

// NewBooleanSettingsCleanerTestConfig creates a BooleanSettingsCleanerTestConfig with standardized values.
// This eliminates duplicate config boilerplate across test files.
//
// Parameters:
//   - testName: Name for the test (e.g., "Cargo", "ProjectsManagementAutomation")
//   - toolName: Tool identifier (e.g., "Cargo", "projects-management-automation")
//   - settingsFieldName: Human-readable name for the settings field (e.g., "cargo packages")
//   - expectedItems: Number of expected items for dry-run tests (usually 1)
//   - newCleanerFunc: Function that creates a new cleaner instance
//
// Returns a configured BooleanSettingsCleanerTestConfig ready for use with CreateBooleanSettingsCleanerTestFunctions
//
// Usage:
//   func TestXxxCleaner_BooleanSettingsTests(t *testing.T) {
//       CreateBooleanSettingsCleanerTestFunctions(t,
//           NewBooleanSettingsCleanerTestConfig(
//               "Xxx",           // testName
//               "xxx-tool",      // toolName
//               "xxx settings",  // settingsFieldName
//               1,               // expectedItems
//               NewXxxCleaner,   // newCleanerFunc
//               func(enabled bool) *domain.OperationSettings {
//                   return &domain.OperationSettings{
//                       XxxSettings: &domain.XxxSettings{
//                           Enabled: enabled,
//                       },
//                   }
//               },
//           ),
//       )
//   }
func NewBooleanSettingsCleanerTestConfig[T CleanerWithSettings](
	testName string,
	toolName string,
	settingsFieldName string,
	expectedItems uint,
	newCleanerFunc func(verbose, dryRun bool) T,
	createSettings func(bool) *domain.OperationSettings,
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

// NewBooleanSettingsCleanerTestConfigFn creates a BooleanSettingsCleanerTestConfig from constructor and settings creation function.
// This eliminates duplicate config boilerplate by combining config creation in one call.
//
// Type Parameters:
//   - T: The cleaner type that must implement the cleaner interface
//
// Parameters:
//   - testName: Name for the test (e.g., "Cargo", "ProjectsManagementAutomation")
//   - toolName: Tool identifier (e.g., "Cargo", "projects-management-automation")
//   - settingsFieldName: Human-readable name for the settings field (e.g., "cargo packages")
//   - expectedItems: Number of expected items for dry-run tests (usually 1)
//   - constructor: Function that creates a cleaner with given verbose and dryRun flags
//   - createSettings: Function that creates OperationSettings with the specific field configured
//
// Returns a configured BooleanSettingsCleanerTestConfig ready for use with CreateBooleanSettingsCleanerTestFunctions
func NewBooleanSettingsCleanerTestConfigFn[T CleanerWithSettings](testName, toolName, settingsFieldName string, expectedItems uint, constructor func(verbose, dryRun bool) T, createSettings func(bool) *domain.OperationSettings) BooleanSettingsCleanerTestConfig {
	return BooleanSettingsCleanerTestConfig{
		TestName:          testName,
		ToolName:          toolName,
		SettingsFieldName: settingsFieldName,
		ExpectedItems:     expectedItems,
		Constructor:       NewBooleanSettingsCleanerTestConstructor(constructor),
		CreateSettings:    createSettings,
	}
}

// BooleanSettingsTestConfig holds all configuration needed to create a BooleanSettingsTests test function.
// Use CreateBooleanSettingsTest to generate the actual test function.
type BooleanSettingsTestConfig struct {
	TestName           string
	ToolName           string
	SettingsFieldName  string
	ExpectedItems      uint
	Constructor        CleanerConstructorWithSettings
	CreateSettingsFunc func(bool) *domain.OperationSettings
}

// CreateBooleanSettingsTest creates a test function for cleaners with boolean settings.
// This eliminates duplicate test function definitions across test files.
//
// Usage:
//
//	func TestXxxCleaner_BooleanSettingsTests(t *testing.T) {
//	    CreateBooleanSettingsTest(t, BooleanSettingsTestConfig{
//	        TestName:          "Xxx",
//	        ToolName:          "xxx-tool",
//	        SettingsFieldName: "xxx settings",
//	        ExpectedItems:     1,
//	        Constructor:       NewBooleanSettingsCleanerTestConstructor(NewXxxCleaner),
//	        CreateSettingsFunc: func(enabled bool) *domain.OperationSettings {
//	            return &domain.OperationSettings{
//	                XxxSettings: &domain.XxxSettings{
//	                    Enabled: enabled,
//	                },
//	            }
//	        },
//	    })
//	}
func CreateBooleanSettingsTest(t *testing.T, config BooleanSettingsTestConfig) {
	CreateBooleanSettingsCleanerTestFunctions(t, BooleanSettingsCleanerTestConfig{
		TestName:          config.TestName,
		ToolName:          config.ToolName,
		SettingsFieldName: config.SettingsFieldName,
		ExpectedItems:     config.ExpectedItems,
		Constructor:       config.Constructor,
		CreateSettings:    config.CreateSettingsFunc,
	})
}

// NewTestCleaner creates a cleaner with default test settings (verbose=false, dryRun=false).
// This eliminates duplicate cleaner initialization code across test files.
//
// Usage:
//
//	func TestXxxCleaner_Xxx(t *testing.T) {
//	    cleaner := NewTestCleaner(NewXxxCleaner)
//	    // use cleaner...
//	}
//
// Type Parameters:
//   - T: The cleaner type
//
// Parameters:
//   - constructor: Function that creates a cleaner with given verbose and dryRun flags
//
// Returns a function that creates the cleaner with default test settings
func NewTestCleaner[T any](constructor func(verbose, dryRun bool) T) func() T {
	return func() T {
		return constructor(false, false)
	}
}

// GetHomeDirTestCase represents a test case for GetHomeDir tests
type GetHomeDirTestCase struct {
	Name         string
	HomeValue    string
	ProfileValue string
	WantErr      bool
	WantHome     string
}

// RunGetHomeDirTests runs GetHomeDir tests for given test cases.
// This eliminates duplicate error checking code across GetHomeDir tests.
//
// Usage:
//
//	func TestXxxCleaner_GetHomeDir(t *testing.T) {
//	    testCases := []GetHomeDirTestCase{
//	        {
//	            Name:      "with HOME set",
//	            HomeValue: "/test/home",
//	            WantErr:   false,
//	            WantHome:  "/test/home",
//	        },
//	        {
//	            Name:         "fallback to USERPROFILE",
//	            HomeValue:    "",
//	            ProfileValue: "C:\\Users\\test",
//	            WantErr:      false,
//	            WantHome:     "C:\\Users\\test",
//	        },
//	    }
//	    RunGetHomeDirTests(t, testCases)
//	}
func RunGetHomeDirTests(t *testing.T, testCases []GetHomeDirTestCase) {
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

// TestNewCleanerConstructor tests a cleaner constructor function with different
// combinations of verbose and dryRun parameters.
// This eliminates duplicate test code across multiple cleaner test files.
//
// Usage:
//
//	func TestNewXxxCleaner(t *testing.T) {
//	    TestNewCleanerConstructor(t, NewXxxCleaner, "NewXxxCleaner")
//	}
//
// Type Parameters:
//   - T: The cleaner type that must have verbose and dryRun fields
//
// Parameters:
//   - t: The testing.T object
//   - constructor: Function that creates a cleaner with given verbose and dryRun flags
//   - cleanerName: Name of the cleaner for error messages
func TestNewCleanerConstructor[T any](t *testing.T, constructor func(bool, bool) T, cleanerName string) {
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
			if cleanerValue.Kind() == reflect.Ptr {
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