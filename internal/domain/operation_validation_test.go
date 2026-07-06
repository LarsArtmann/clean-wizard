package domain

import (
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
)

func TestValidationError_ClassifiedAsRejection(t *testing.T) {
	t.Parallel()

	err := &ValidationError{
		Field:   "temp_files.older_than",
		Message: "older_than is required",
	}

	errorfamilytest.AssertFamily(t, err, errorfamily.Rejection)
	errorfamilytest.AssertCode(t, err, "validation.rejected")
	errorfamilytest.AssertRetryable(t, err, false)
}
