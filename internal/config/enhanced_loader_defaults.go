package config

import (
	"time"
)

// getDefaultLoadOptions returns default load options
func getDefaultLoadOptions() *ConfigLoadOptions {
	return &ConfigLoadOptions{
		ForceRefresh:       false,
		EnableCache:        true,
		EnableSanitization: true,
		ValidationLevel:    ValidationLevelComprehensive,
		Timeout:            30 * time.Second,
	}
}

// getDefaultSaveOptions returns default save options
func getDefaultSaveOptions() *ConfigSaveOptions {
	return &ConfigSaveOptions{
		EnableSanitization: true,
		CreateBackup:       true,  // Create backup by default for safety
		ValidationLevel:    ValidationLevelComprehensive,
		ForceSave:          false,
	}
}

// getDefaultRetryPolicy returns default retry policy
func getDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}
