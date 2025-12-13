package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/shared/utils/middleware"
)

// NewConfigSanitizer creates a new config sanitizer
func NewConfigSanitizer() *middleware.ValidationMiddleware {
	return middleware.NewValidationMiddleware()
}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *middleware.ValidationMiddleware {
	return middleware.NewValidationMiddleware()
}
