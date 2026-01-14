package middleware

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	"github.com/LarsArtmann/clean-wizard/internal/shared/utils/validation"
)

// ValidationMiddleware provides validation for all operations.
type ValidationMiddleware struct{}

// NewValidationMiddleware creates validation middleware.
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

// ValidateScanRequest validates scan request before processing.
func (vm *ValidationMiddleware) ValidateScanRequest(ctx context.Context, req domain.ScanRequest) result.Result[domain.ScanRequest] {
	return validation.ValidateAndWrap(req, "scan request")
}

// ValidateCleanRequest validates clean request before processing.
func (vm *ValidationMiddleware) ValidateCleanRequest(ctx context.Context, req domain.CleanRequest) result.Result[domain.CleanRequest] {
	return validation.ValidateAndWrap(req, "clean request")
}

// ValidateCleanerSettings validates cleaner settings with type safety.
func (vm *ValidationMiddleware) ValidateCleanerSettings(ctx context.Context, cleaner domain.Cleaner, settings *domain.OperationSettings) result.Result[*domain.OperationSettings] {
	if err := cleaner.ValidateSettings(settings); err != nil {
		return result.Err[*domain.OperationSettings](fmt.Errorf("invalid cleaner settings for %s: %w", cleaner.Type(), err))
	}
	return result.Ok(settings)
}
