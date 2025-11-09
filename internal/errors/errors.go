package errors

import "fmt"

// ConfigError represents configuration errors
type ConfigError struct {
	Operation string
	Err       error
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

func (e ConfigError) Unwrap() error {
	return e.Err
}

// ConfigLoadError creates config load error
func ConfigLoadError(err error) error {
	return ConfigError{Operation: "ConfigLoadError", Err: err}
}

// ConfigSaveError creates config save error  
func ConfigSaveError(err error) error {
	return ConfigError{Operation: "ConfigSaveError", Err: err}
}

// ConfigValidateError creates config validation error
func ConfigValidateError(message string) error {
	return ConfigError{Operation: "ConfigValidateError", Err: fmt.Errorf(message)}
}
