package validation_test

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// ValidationLevel alias for backward compatibility
type ValidationLevel = shared.ValidationLevelType

// SanitizationTestCase defines sanitization test case structure
type SanitizationTestCase struct {
	Name          string
	Input         interface{}
	Expected      interface{}
	ShouldSucceed bool
}
