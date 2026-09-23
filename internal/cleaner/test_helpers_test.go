package cleaner

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
