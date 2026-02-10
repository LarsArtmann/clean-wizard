package cleaner

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ToSimpleCleanerConstructor converts a constructor with additional methods to one that only exposes Clean and IsAvailable.
func ToSimpleCleanerConstructor(fullConstructor CleanerConstructorWithSettings) SimpleCleanerConstructor {
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
	constructor func(verbose, dryRun bool) T,
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
func NewBooleanSettingsCleanerTestConstructor[T CleanerWithSettings](constructor func(verbose, dryRun bool) T) CleanerConstructorWithSettings {
	return cleanerConstructorWithSettingsAdapter(constructor)
}

// NewBooleanSettingsCleanerTestConfig creates a BooleanSettingsCleanerTestConfig with standardized values.
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
func NewBooleanSettingsCleanerTestConfigFn[T CleanerWithSettings](
	testName, toolName, settingsFieldName string,
	expectedItems uint,
	constructor func(verbose, dryRun bool) T,
	createSettings func(bool) *domain.OperationSettings,
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
func NewTestCleaner[T any](constructor func(verbose, dryRun bool) T) func() T {
	return func() T {
		return constructor(false, false)
	}
}
