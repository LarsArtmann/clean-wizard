package errors

// This file acts as the main interface for the errors package
// All functionality is split across multiple files for better organization:
// - error_codes.go: ErrorCode constants and string methods
// - error_levels.go: ErrorLevel constants and string methods  
// - error_types.go: ErrorDetails and CleanWizardError structs
// - error_constructors.go: NewError and helper constructors
// - error_methods.go: With* and Is* methods, logging functionality

// Re-export all public types and functions from split files
// No additional imports needed - Go package system handles this automatically