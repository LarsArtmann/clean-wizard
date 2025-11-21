package middleware

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
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
func (vm *ValidationMiddleware) ValidateScanRequest(ctx context.Context, req domain.ScanRequest) result.Result[domain.ScanRequest] {
	return validateRequest(req, "scan")
}

// ValidateCleanRequest validates clean request before processing
func (vm *ValidationMiddleware) ValidateCleanRequest(ctx context.Context, req domain.CleanRequest) result.Result[domain.CleanRequest] {
	return validateRequest(req, "clean")
}

// ValidateCleanerSettings validates cleaner settings with type safety
func (vm *ValidationMiddleware) ValidateCleanerSettings(ctx context.Context, cleaner domain.Cleaner, settings *domain.OperationSettings) result.Result[*domain.OperationSettings] {
	if err := cleaner.ValidateSettings(settings); err != nil {
		return result.Err[*domain.OperationSettings](fmt.Errorf("invalid cleaner settings: %w", err))
	}
	return result.Ok(settings)
}
