package shared

// ValidationResult represents the result of validation operations
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors,omitempty"`
}

// NewValidationResult creates a new validation result
func NewValidationResult(isValid bool, errors ...string) *ValidationResult {
	return &ValidationResult{
		IsValid: isValid,
		Errors:  errors,
	}
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(error string) {
	vr.Errors = append(vr.Errors, error)
	vr.IsValid = false
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}
