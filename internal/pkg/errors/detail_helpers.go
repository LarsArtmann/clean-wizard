package errors

import "fmt"

// ErrorDetailsBuilder provides a fluent API for constructing ErrorDetails.
type ErrorDetailsBuilder struct {
	details ErrorDetails
}

// NewErrorDetails creates a new builder with an initialized metadata map.
func NewErrorDetails() *ErrorDetailsBuilder {
	return &ErrorDetailsBuilder{
		details: ErrorDetails{
			Metadata: make(map[string]string),
		},
	}
}

// WithField sets the field name that caused the error.
func (b *ErrorDetailsBuilder) WithField(field string) *ErrorDetailsBuilder {
	b.details.Field = field
	return b
}

// WithValue sets the actual value that caused the error.
func (b *ErrorDetailsBuilder) WithValue(value string) *ErrorDetailsBuilder {
	b.details.Value = value
	return b
}

// WithExpected sets the expected value for the error.
func (b *ErrorDetailsBuilder) WithExpected(expected string) *ErrorDetailsBuilder {
	b.details.Expected = expected
	return b
}

// WithActual sets the actual value for the error.
func (b *ErrorDetailsBuilder) WithActual(actual string) *ErrorDetailsBuilder {
	b.details.Actual = actual
	return b
}

// WithOperation sets the operation context for the error.
func (b *ErrorDetailsBuilder) WithOperation(operation string) *ErrorDetailsBuilder {
	b.details.Operation = operation
	return b
}

// WithFilePath sets the file path associated with the error.
func (b *ErrorDetailsBuilder) WithFilePath(path string) *ErrorDetailsBuilder {
	b.details.FilePath = path
	return b
}

// WithLineNumber sets the line number associated with the error.
func (b *ErrorDetailsBuilder) WithLineNumber(line int) *ErrorDetailsBuilder {
	b.details.LineNumber = line
	return b
}

// WithRetryCount sets the retry count for the error.
func (b *ErrorDetailsBuilder) WithRetryCount(count int) *ErrorDetailsBuilder {
	b.details.RetryCount = count
	return b
}

// WithDuration sets the duration string for the error.
func (b *ErrorDetailsBuilder) WithDuration(duration string) *ErrorDetailsBuilder {
	b.details.Duration = duration
	return b
}

// WithMetadata adds a key-value pair to the metadata map.
func (b *ErrorDetailsBuilder) WithMetadata(key, value string) *ErrorDetailsBuilder {
	if b.details.Metadata == nil {
		b.details.Metadata = make(map[string]string)
	}
	b.details.Metadata[key] = value
	return b
}

// Build returns the constructed ErrorDetails.
func (b *ErrorDetailsBuilder) Build() *ErrorDetails {
	return &b.details
}

// setStringField sets a string field if value is a string, otherwise formats the value
// Returns true if value was set successfully as string.
func setStringField(target *string, value any) bool {
	if v, ok := value.(string); ok {
		*target = v
		return true
	}
	*target = fmt.Sprintf("%v", value)
	return false
}

// setStringFieldStrict sets a string field only if value is a string
// Returns true if value was set, false otherwise.
func setStringFieldStrict(target *string, value any) bool {
	if v, ok := value.(string); ok {
		*target = v
		return true
	}
	return false
}

// setIntField sets an int field only if value is an int
// Returns true if value was set, false otherwise.
func setIntField(target *int, value any) bool {
	if v, ok := value.(int); ok {
		*target = v
		return true
	}
	return false
}

// addToMetadata adds a key-value pair to metadata map with string formatting
// Returns the (potentially new) metadata map.
func addToMetadata(metadata map[string]string, key string, value any) map[string]string {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[key] = fmt.Sprintf("%v", value)
	return metadata
}

// ensureDetails initializes ErrorDetails if nil and ensures Metadata is initialized.
func ensureDetails(e **ErrorDetails) {
	if *e == nil {
		*e = &ErrorDetails{
			Metadata: make(map[string]string),
		}
	} else if (*e).Metadata == nil {
		(*e).Metadata = make(map[string]string)
	}
}
