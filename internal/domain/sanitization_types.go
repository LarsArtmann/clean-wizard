package domain

// SanitizationWarning represents a warning during sanitization.
type SanitizationWarning struct {
	Field     string `json:"field"`
	Message   string `json:"message,omitempty"`
	Original  any    `json:"original,omitempty"`
	Sanitized any    `json:"sanitized,omitempty"`
	OldValue  any    `json:"old_value,omitempty"`
	NewValue  any    `json:"new_value,omitempty"`
	Reason    string `json:"reason,omitempty"`
}
