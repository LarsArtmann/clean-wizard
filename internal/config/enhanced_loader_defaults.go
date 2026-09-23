package config

import "github.com/LarsArtmann/clean-wizard/internal/domain/enums"

// getDefaultLoadOptions returns default load options.
func getDefaultLoadOptions() *ConfigLoadOptions {
	return &ConfigLoadOptions{
		ForceRefresh:       RefreshOptionDisabled,
		EnableCache:        CacheOptionEnabled,
		EnableSanitization: SanitizeOptionEnabled,
		ValidationLevel:    enums.ValidationLevelComprehensiveType,
		Timeout:            DefaultLoadTimeout,
	}
}

// getDefaultSaveOptions returns default save options.
func getDefaultSaveOptions() *ConfigSaveOptions {
	return &ConfigSaveOptions{
		EnableSanitization: SanitizeOptionEnabled,
		BackupEnabled:      BackupOptionEnabled,
		ValidationLevel:    enums.ValidationLevelComprehensiveType,
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
