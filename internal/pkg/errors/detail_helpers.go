package errors

import "fmt"

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
