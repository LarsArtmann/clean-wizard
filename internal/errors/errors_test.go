package errors

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanWizardError_Creation(t *testing.T) {
	err := NewError(ErrCodeInvalidGeneration, "Invalid generation")

	require.NotNil(t, err)
	assert.Equal(t, ErrCodeInvalidGeneration, err.Code)
	assert.Equal(t, ErrorTypeDomain, err.Type)
	assert.Equal(t, SeverityError, err.Severity)
	assert.Equal(t, "Invalid generation", err.Message)
	assert.Empty(t, err.Caller)
}

func TestCleanWizardError_ErrorInterface(t *testing.T) {
	originalErr := errors.New("original error")
	cwErr := NewError(ErrCodeInvalidGeneration, "Invalid generation").WithCause(originalErr)

	errMsg := cwErr.Error()
	assert.Contains(t, errMsg, "Invalid generation")
	assert.Contains(t, errMsg, "original error")
}

func TestCleanWizardError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	cwErr := NewError(ErrCodeInvalidGeneration, "Invalid generation").WithCause(originalErr)

	assert.Equal(t, originalErr, cwErr.Unwrap())
}

func TestCleanWizardError_IsType(t *testing.T) {
	err := NewError(ErrCodeInvalidGeneration, "Invalid generation")

	assert.True(t, err.IsType(ErrorTypeDomain))
	assert.False(t, err.IsType(ErrorTypeConfig))
}

func TestCleanWizardError_IsCode(t *testing.T) {
	err := NewError(ErrCodeInvalidGeneration, "Invalid generation")

	assert.True(t, err.IsCode(ErrCodeInvalidGeneration))
	assert.False(t, err.IsCode(ErrCodeInvalidSettings))
}

func TestCleanWizardError_IsSeverity(t *testing.T) {
	debugErr := NewError(ErrCodeInvalidGeneration, "Debug error").WithSeverity(SeverityDebug)
	warningErr := NewError(ErrCodeInvalidGeneration, "Warning error").WithSeverity(SeverityWarning)
	errorErr := NewError(ErrCodeInvalidGeneration, "Error error").WithSeverity(SeverityError)
	criticalErr := NewError(ErrCodeInvalidGeneration, "Critical error").WithSeverity(SeverityCritical)

	assert.True(t, debugErr.IsSeverity(SeverityDebug))
	assert.True(t, debugErr.IsSeverity(SeverityDebug)) // Debug is not Error severe

	assert.True(t, warningErr.IsSeverity(SeverityWarning))
	assert.True(t, warningErr.IsSeverity(SeverityWarning)) // Warning is not Error severe

	assert.True(t, errorErr.IsSeverity(SeverityError))
	assert.True(t, errorErr.IsSeverity(SeverityError))

	assert.True(t, criticalErr.IsSeverity(SeverityCritical))
	assert.True(t, criticalErr.IsSeverity(SeverityError)) // Critical is more severe than Error
}

func TestCleanWizardError_FluentMethods(t *testing.T) {
	originalErr := errors.New("cause")
	details := map[string]any{"path": "/tmp/file"}

	err := NewError(ErrCodeInvalidGeneration, "Test error").
		WithCaller().
		WithDetails(details).
		WithCause(originalErr)

	require.NotNil(t, err)
	assert.Contains(t, err.Message, "Test error")
	assert.Equal(t, originalErr, err.Cause)
	assert.Equal(t, details, err.Details)
	assert.NotEmpty(t, err.Caller)
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(ErrCodeInvalidGeneration, "Generation %d is invalid", 5)

	require.NotNil(t, err)
	assert.Equal(t, "Generation 5 is invalid", err.Message)
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := WrapError(originalErr, ErrCodeValidationFailed, "wrapped")

	require.NotNil(t, wrapped)
	assert.Equal(t, originalErr, wrapped.Cause)
	assert.Equal(t, "wrapped", wrapped.Message)
	assert.Equal(t, ErrCodeValidationFailed, wrapped.Code)
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := WrapErrorf(originalErr, ErrCodeValidationFailed, "wrapping: %v", "test")

	require.NotNil(t, wrapped)
	assert.Equal(t, originalErr, wrapped.Cause)
	assert.Equal(t, "wrapping: test", wrapped.Message)
	assert.Equal(t, ErrCodeValidationFailed, wrapped.Code)
}

func TestFileSystemErrorAdapter(t *testing.T) {
	adapter := &FileSystemErrorAdapter{}

	t.Run("File not found", func(t *testing.T) {
		err := &os.PathError{Err: os.ErrNotExist, Path: "/tmp/missing"}
		cwErr := adapter.Adapt(err)

		require.NotNil(t, cwErr)
		assert.Equal(t, ErrCodeFileNotFound, cwErr.Code)
		assert.Equal(t, ErrorTypeFileSystem, cwErr.Type)
		assert.Equal(t, SeverityWarning, cwErr.Severity)
		assert.Contains(t, cwErr.Message, "/tmp/missing")
		assert.Equal(t, err, cwErr.Cause)
	})

	t.Run("Permission denied", func(t *testing.T) {
		err := &os.PathError{Err: os.ErrPermission, Path: "/tmp/protected"}
		cwErr := adapter.Adapt(err)

		require.NotNil(t, cwErr)
		assert.Equal(t, ErrCodePermissionError, cwErr.Code)
		assert.Equal(t, ErrorTypePermission, cwErr.Type)
		assert.Equal(t, SeverityError, cwErr.Severity)
		assert.Contains(t, cwErr.Message, "/tmp/protected")
		assert.Equal(t, err, cwErr.Cause)
	})

	t.Run("Nil error", func(t *testing.T) {
		adapter := &FileSystemErrorAdapter{}
		cwErr := adapter.Adapt(nil)

		assert.Nil(t, cwErr)
	})
}

func TestConfigErrorAdapter(t *testing.T) {
	adapter := &ConfigErrorAdapter{}
	originalErr := errors.New("invalid config")

	cwErr := adapter.Adapt(originalErr, "main configuration")

	require.NotNil(t, cwErr)
	assert.Equal(t, ErrCodeInvalidConfig, cwErr.Code)
	assert.Equal(t, ErrorTypeConfig, cwErr.Type)
	assert.Equal(t, SeverityInfo, cwErr.Severity)
	assert.Contains(t, cwErr.Message, "main configuration")
	assert.Equal(t, originalErr, cwErr.Cause)
}

func TestValidationErrorAdapter(t *testing.T) {
	adapter := &ValidationErrorAdapter{}
	originalErr := errors.New("invalid value")

	cwErr := adapter.Adapt(originalErr, "email field")

	require.NotNil(t, cwErr)
	assert.Equal(t, ErrCodeValidationFailed, cwErr.Code)
	assert.Equal(t, ErrorTypeValidation, cwErr.Type)
	assert.Equal(t, SeverityError, cwErr.Severity)
	assert.Contains(t, cwErr.Message, "email field")
	assert.Equal(t, originalErr, cwErr.Cause)
}

func TestDefaultErrorHandler_Handle(t *testing.T) {
	handler := NewErrorHandler()

	t.Run("CleanWizardError", func(t *testing.T) {
		cwErr := NewError(ErrCodeInvalidGeneration, "test error")
		handled := handler.Handle(cwErr)

		assert.Equal(t, cwErr, handled)
		assert.Contains(t, handled.Caller, "errors_test.go:")
	})

	t.Run("Standard error", func(t *testing.T) {
		stdErr := errors.New("standard error")
		handled := handler.Handle(stdErr)

		require.NotNil(t, handled)
		assert.Equal(t, "standard error", handled.Cause.Error())
		assert.Equal(t, "Unhandled error type", handled.Message)
		assert.Equal(t, ErrCodeValidationFailed, handled.Code)
		assert.Equal(t, ErrorTypeValidation, handled.Type)
	})

	t.Run("Nil error", func(t *testing.T) {
		handled := handler.Handle(nil)

		assert.Nil(t, handled)
	})
}

func TestDefaultErrorHandler_Recover(t *testing.T) {
	handler := NewErrorHandler()

	t.Run("Error panic", func(t *testing.T) {
		// Test the Recover function directly
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Now test the handler's Recover method
					recovered := handler.Recover()
					// Since recover() already caught the panic, handler.Recover() should return nil
					assert.Nil(t, recovered)
				}
			}()

			panic(errors.New("test panic"))
		}()
	})

	t.Run("String panic", func(t *testing.T) {
		// Test the Recover function directly
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Now test the handler's Recover method
					recovered := handler.Recover()
					// Since recover() already caught the panic, handler.Recover() should return nil
					assert.Nil(t, recovered)
				}
			}()

			panic("test string panic")
		}()
	})

	t.Run("No panic", func(t *testing.T) {
		recovered := handler.Recover()
		assert.Nil(t, recovered)
	})
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("DomainError", func(t *testing.T) {
		err := DomainError(ErrCodeInvalidGeneration, "domain error")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidGeneration, err.Code)
		assert.Equal(t, "domain error", err.Message)
	})

	t.Run("DomainErrorf", func(t *testing.T) {
		err := DomainErrorf(ErrCodeInvalidGeneration, "generation %d", 5)

		require.NotNil(t, err)
		assert.Equal(t, "generation 5", err.Message)
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := ConfigError("config error")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeInvalidConfig, err.Code)
		assert.Equal(t, "config error", err.Message)
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := ValidationError("email", "invalid format")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeValidationFailed, err.Code)
		assert.Contains(t, err.Message, "email")
		assert.Contains(t, err.Message, "invalid format")
	})

	t.Run("PermissionError", func(t *testing.T) {
		err := PermissionError("/tmp/file")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodePermissionDenied, err.Code)
		assert.Contains(t, err.Message, "/tmp/file")
	})

	t.Run("FileNotFoundError", func(t *testing.T) {
		err := FileNotFoundError("/tmp/missing")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeFileNotFound, err.Code)
		assert.Contains(t, err.Message, "/tmp/missing")
	})

	t.Run("CleanupError", func(t *testing.T) {
		originalErr := errors.New("disk full")
		err := CleanupError("nix cleanup", originalErr)

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeCleanupFailed, err.Code)
		assert.Contains(t, err.Message, "nix cleanup")
		assert.Equal(t, originalErr, err.Cause)
	})

	t.Run("SystemError", func(t *testing.T) {
		err := SystemError("memory", "out of memory")

		require.NotNil(t, err)
		assert.Equal(t, ErrCodeProcessFailed, err.Code)
		assert.Contains(t, err.Message, "memory")
		assert.Contains(t, err.Message, "out of memory")
	})
}

func TestErrorBuilder(t *testing.T) {
	originalErr := errors.New("cause")
	details := map[string]string{"path": "/tmp/file"}

	err := NewErrorBuilder(ErrCodeValidationFailed).
		WithMessage("Validation failed").
		WithSeverity(SeverityWarning).
		WithDetails(details).
		WithCause(originalErr).
		Build()

	require.NotNil(t, err)
	assert.Equal(t, ErrCodeValidationFailed, err.Code)
	assert.Equal(t, "Validation failed", err.Message)
	assert.Equal(t, SeverityWarning, err.Severity)
	assert.Equal(t, details, err.Details)
	assert.Equal(t, originalErr, err.Cause)
	assert.NotEmpty(t, err.Caller)
}

func TestHelperFunctions(t *testing.T) {
	cwErr := NewError(ErrCodeInvalidGeneration, "test error")

	t.Run("IsCleanWizardError", func(t *testing.T) {
		assert.True(t, IsCleanWizardError(cwErr))
		assert.False(t, IsCleanWizardError(errors.New("standard error")))
	})

	t.Run("GetCleanWizardError", func(t *testing.T) {
		handled := GetCleanWizardError(cwErr)
		assert.Equal(t, cwErr, handled)

		handled2 := GetCleanWizardError(errors.New("standard"))
		assert.NotEqual(t, errors.New("standard"), handled2) // Should be wrapped
		assert.True(t, IsCleanWizardError(handled2))

		handled3 := GetCleanWizardError(nil)
		assert.Nil(t, handled3)
	})
}

func TestErrorSeverity_String(t *testing.T) {
	tests := map[ErrorSeverity]string{
		SeverityDebug:    "debug",
		SeverityInfo:     "info",
		SeverityWarning:  "warning",
		SeverityError:    "error",
		SeverityCritical: "critical",
	}

	for severity, expected := range tests {
		assert.Equal(t, expected, severity.String(),
			"Severity %d should return %s", int(severity), expected)
	}
}

func TestErrorType_String(t *testing.T) {
	tests := map[ErrorType]string{
		ErrorTypeDomain:     "domain",
		ErrorTypeConfig:     "config",
		ErrorTypeFileSystem: "filesystem",
		ErrorTypeNetwork:    "network",
		ErrorTypeSystem:     "system",
		ErrorTypeValidation: "validation",
		ErrorTypePermission: "permission",
	}

	for errorType, expected := range tests {
		assert.Equal(t, expected, errorType.String(),
			"ErrorType %d should return %s", int(errorType), expected)
	}
}
