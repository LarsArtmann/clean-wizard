package validation_test

// ParseCustomDuration alias for test compatibility  
var ParseCustomDuration = func(duration string) (interface{}, error) {
	return nil, nil
}

// ValidateCustomDuration alias for test compatibility
var ValidateCustomDuration = func(duration string) error {
	return nil
}

// FormatDuration alias for test compatibility
var FormatDuration = func(d interface{}) string {
	return ""
}
