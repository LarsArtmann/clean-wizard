package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// getDefaultLoadOptions returns default load options.
func getDefaultLoadOptions() *ConfigLoadOptions {
	return &ConfigLoadOptions{
		ForceRefresh:       RefreshOptionDisabled,
		EnableCache:        CacheOptionEnabled,
		EnableSanitization: SanitizeOptionEnabled,
		ValidationLevel:    domain.ValidationLevelComprehensiveType,
		Timeout:            DefaultLoadTimeout,
	}
}

// getDefaultSaveOptions returns default save options.
func getDefaultSaveOptions() *ConfigSaveOptions {
	return &ConfigSaveOptions{
		EnableSanitization: SanitizeOptionEnabled,
		BackupEnabled:      BackupOptionEnabled,
		ValidationLevel:    domain.ValidationLevelComprehensiveType,
		CreateBackup:       BackupOptionDisabled,
		ForceSave:          SaveOptionDisabled,
	}
}

// getDefaultRetryPolicy returns default retry policy.
func getDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    DefaultMaxRetries,
		InitialDelay:  DefaultInitialRetryDelay,
		MaxDelay:      DefaultMaxRetryDelay,
		BackoffFactor: DefaultBackoffFactor,
	}
}
