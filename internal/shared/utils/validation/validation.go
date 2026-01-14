package validation

import (
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// Validator interface for types that can be validated.
type Validator interface {
	Validate() error
}

// ValidateAndWrap provides a generic validation wrapper that returns a Result[T]
// This eliminates duplicate validation patterns across the codebase.
func ValidateAndWrap[T Validator](item T, itemType string) result.Result[T] {
	if err := item.Validate(); err != nil {
		return result.Err[T](fmt.Errorf("invalid %s: %w", itemType, err))
	}
	return result.Ok(item)
}

// ValidateWithCustomError provides a generic validation wrapper with custom error message.
func ValidateWithCustomError[T Validator](item T, errorMsg string) result.Result[T] {
	if err := item.Validate(); err != nil {
		return result.Err[T](fmt.Errorf("%s: %w", errorMsg, err))
	}
	return result.Ok(item)
}

// ValidateAndConvert provides a generic validation and conversion pattern
// Useful when you need to validate and then convert to a different type.
func ValidateAndConvert[T Validator, U any](item T, converter func(T) U, itemType string) result.Result[U] {
	if err := item.Validate(); err != nil {
		return result.Err[U](fmt.Errorf("invalid %s: %w", itemType, err))
	}
	return result.Ok(converter(item))
}
