package middleware

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// Validator defines interface for types that can validate themselves
type Validator interface {
	Validate() error
}

// ValidationMiddleware provides validation for all operations
type ValidationMiddleware struct{}

// NewValidationMiddleware creates validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

// validateRequest performs common validation pattern for Validator types
// TODO: Consider adding request type enum for better type safety
func validateRequest[T Validator](req T, requestType string) result.Result[T] {
	if err := req.Validate(); err != nil {
		return result.Err[T](fmt.Errorf("invalid %s request: %w", requestType, err))
	}
	return result.Ok(req)
}

// ValidateScanRequest validates scan request before processing
func (vm *ValidationMiddleware) ValidateScanRequest(ctx context.Context, req shared.ScanRequest) result.Result[shared.ScanRequest] {
	return validateRequest(req, "scan")
}

// ValidateCleanRequest validates clean request before processing
func (vm *ValidationMiddleware) ValidateCleanRequest(ctx context.Context, req shared.CleanRequest) result.Result[shared.CleanRequest] {
	return validateRequest(req, "clean")
}

// ValidateCleanerSettings validates cleaner settings with type safety
func (vm *ValidationMiddleware) ValidateCleanerSettings(ctx context.Context, cleaner shared.Cleaner, settings *shared.OperationSettings) result.Result[*shared.OperationSettings] {
	if err := cleaner.ValidateSettings(settings); err != nil {
		return result.Err[*shared.OperationSettings](fmt.Errorf("invalid cleaner settings: %w", err))
	}
	return result.Ok(settings)
}
