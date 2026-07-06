package cleaner

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// ValidateSettingsTestCase represents a test case for ValidateSettings.
type ValidateSettingsTestCase struct {
	Name     string
	Settings *domain.OperationSettings
	WantErr  bool
}

// IsAvailableConstructor is a function type for creating cleaners in tests that need IsAvailable.
type IsAvailableConstructor func() CleanerCore

// IsAvailableTestCase represents a test case for IsAvailable tests.
type IsAvailableTestCase struct {
	Name        string
	Constructor IsAvailableConstructor
}

// CleanerCore is the minimum cleaner interface used in test helpers.
// CleanerWithSettings and SimpleCleaner both build on this.
type CleanerCore interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
}

// CleanerConstructor is the standard constructor signature for test cleaners:
// they take verbose and dryRun flags and return a value of type T.
type CleanerConstructor[T any] func(verbose, dryRun bool) T

// CleanerWithSettings represents a cleaner interface with settings validation
// This eliminates duplicate interface declarations in test helper functions.
type CleanerWithSettings interface {
	CleanerCore
	ValidateSettings(*domain.OperationSettings) error
}

// CleanerConstructorWithSettings is a function type for creating cleaners in tests that need ValidateSettings.
type CleanerConstructorWithSettings = CleanerConstructor[CleanerWithSettings]

// SimpleCleanerConstructor is a function type for creating cleaners in tests that only need Clean and IsAvailable.
type SimpleCleanerConstructor = CleanerConstructor[CleanerCore]

// SimpleCleaner represents a cleaner interface without settings validation.
type SimpleCleaner = CleanerCore

// GetHomeDirTestCase represents a test case for GetHomeDir tests.
type GetHomeDirTestCase struct {
	Name         string
	HomeValue    string
	ProfileValue string
	WantErr      bool
	WantHome     string
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
