package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// loadConfigWithRetry loads configuration with retry logic.
func (ecl *EnhancedConfigLoader) loadConfigWithRetry(ctx context.Context, options *ConfigLoadOptions) (*domain.Config, error) {
	var lastErr error

	for attempt := 0; attempt <= ecl.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ecl.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Use existing Load function with timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, options.Timeout)
		config, err := LoadWithContext(timeoutCtx)
		cancel()

		if err == nil {
			return config, nil
		}

		lastErr = err

		if !ecl.shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("failed to load config after %d attempts: %w", ecl.retryPolicy.MaxRetries+1, lastErr)
}

// saveConfigWithRetry saves configuration with retry logic.
func (ecl *EnhancedConfigLoader) saveConfigWithRetry(ctx context.Context, config *domain.Config, options *ConfigSaveOptions) error {
	var lastErr error

	for attempt := 0; attempt <= ecl.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := ecl.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := Save(config)
		if err == nil {
			return nil
		}

		lastErr = err

		if !ecl.shouldRetry(err) {
			break
		}
	}

	return fmt.Errorf("failed to save config after %d attempts: %w", ecl.retryPolicy.MaxRetries+1, lastErr)
}

// createBackup creates a backup of the current configuration.
func (ecl *EnhancedConfigLoader) createBackup(ctx context.Context, config *domain.Config) error {
	// Read current config file and copy to backup location
	originalConfigPath := filepath.Join(os.Getenv("HOME"), ".clean-wizard.yaml")
	backupPath := fmt.Sprintf("%s.backup.%d", originalConfigPath, time.Now().Unix())

	// Read current config file
	configData, err := os.ReadFile(originalConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read current config for backup: %w", err)
	}

	// Write backup copy
	err = os.WriteFile(backupPath, configData, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create backup at %s: %w", backupPath, err)
	}

	if ecl.enableMonitoring {
		fmt.Printf("💾 Configuration backup created\n")
	}
	return nil
}

// shouldRetry determines if an error should trigger a retry.
func (ecl *EnhancedConfigLoader) shouldRetry(err error) bool {
	// Don't retry on certain errors
	if err.Error() == "validation failed" {
		return false
	}
	return true
}

// calculateDelay calculates exponential backoff delay.
func (ecl *EnhancedConfigLoader) calculateDelay(attempt int) time.Duration {
	delay := float64(ecl.retryPolicy.InitialDelay) *
		float64(ecl.retryPolicy.BackoffFactor) *
		float64(attempt)

	if delay > float64(ecl.retryPolicy.MaxDelay) {
		delay = float64(ecl.retryPolicy.MaxDelay)
	}

	return time.Duration(delay)
}
