package cleaner

import (
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
