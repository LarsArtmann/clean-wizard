package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// TestValidationLevelTestCase defines validation level test cases
type TestValidationLevelTestCase struct {
	name         string
	config       *domain.Config
	level        ValidationLevel
	expectValid  bool
	expectErrors int
}

// CreateTestConfigurations creates test configurations for validation testing
// Note: Delegates to shared factory in test_data.go to eliminate duplication
func CreateTestConfigurations() map[string]*domain.Config {
	return CreateValidationTestConfigs()
}

// GetSanitizationTestCases returns all sanitization test cases
// Note: Delegates to shared test data in test_data.go
func GetSanitizationTestCases() []SanitizationTestCase {
	return GetStandardTestCasesCompatWrapper()
}

// GetStandardTestCasesCompatWrapper converts standard test cases to sanitization test cases
func GetStandardTestCasesCompatWrapper() []SanitizationTestCase {
	standardCases := GetStandardTestCases()
	result := make([]SanitizationTestCase, len(standardCases))
	for i, tc := range standardCases {
		result[i] = SanitizationTestCase(tc)
	}
	return result
}
