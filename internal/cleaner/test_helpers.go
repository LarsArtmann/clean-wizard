package cleaner

import (
	"reflect"
	"testing"
)

// CreateBooleanSettingsCleanerTestFunctions creates both ValidateSettings and Clean_DryRun test functions
// for cleaners with a boolean settings field. This eliminates duplicate config and constructor code.
//
// Usage:
//
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
