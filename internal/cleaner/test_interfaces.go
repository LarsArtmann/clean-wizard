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
type IsAvailableConstructor func() interface {
	IsAvailable(ctx context.Context) bool
}

// IsAvailableTestCase represents a test case for IsAvailable tests.
type IsAvailableTestCase struct {
	Name        string
	Constructor IsAvailableConstructor
}

// CleanerWithSettings represents a cleaner interface with settings validation
// This eliminates duplicate interface declarations in test helper functions.
type CleanerWithSettings interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
	ValidateSettings(*domain.OperationSettings) error
}

// CleanerConstructorWithSettings is a function type for creating cleaners in tests that need ValidateSettings.
type CleanerConstructorWithSettings func(verbose, dryRun bool) CleanerWithSettings

// SimpleCleanerConstructor is a function type for creating cleaners in tests that only need Clean and IsAvailable.
type SimpleCleanerConstructor func(verbose, dryRun bool) SimpleCleaner

// SimpleCleaner represents a cleaner interface without settings validation.
type SimpleCleaner interface {
	IsAvailable(ctx context.Context) bool
	Clean(ctx context.Context) result.Result[domain.CleanResult]
}

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
