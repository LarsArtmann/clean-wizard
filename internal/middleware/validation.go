package middleware

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// ValidationMiddleware provides validation for all operations
type ValidationMiddleware struct{}

// NewValidationMiddleware creates validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

// ValidateScanRequest validates scan request before processing
func (vm *ValidationMiddleware) ValidateScanRequest(ctx context.Context, req domain.ScanRequest) result.Result[domain.ScanRequest] {
	if err := req.Validate(); err != nil {
		return result.Err[domain.ScanRequest](fmt.Errorf("invalid scan request: %w", err))
	}
	return result.Ok(req)
}

// ValidateCleanRequest validates clean request before processing
func (vm *ValidationMiddleware) ValidateCleanRequest(ctx context.Context, req domain.CleanRequest) result.Result[domain.CleanRequest] {
	if err := req.Validate(); err != nil {
		return result.Err[domain.CleanRequest](fmt.Errorf("invalid clean request: %w", err))
	}
	return result.Ok(req)
}

// ValidateCleanerSettings validates cleaner settings with type safety
func (vm *ValidationMiddleware) ValidateCleanerSettings(ctx context.Context, cleaner domain.Cleaner, settings *domain.OperationSettings) result.Result[*domain.OperationSettings] {
	if err := cleaner.ValidateSettings(settings); err != nil {
		return result.Err[*domain.OperationSettings](fmt.Errorf("invalid cleaner settings: %w", err))
	}
	return result.Ok(settings)
}
