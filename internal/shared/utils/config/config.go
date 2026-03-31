package config

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// LoadConfigWithFallback loads configuration with proper error handling and user feedback
// This eliminates duplicate config loading patterns across commands.
func LoadConfigWithFallback(ctx context.Context, logger *log.Logger) (*domain.Config, error) {
	return loadConfig(ctx, logger, true)
}

// LoadConfigOrContinue loads configuration but allows caller to continue on error
// Returns (config, error) where error can be ignored for graceful degradation.
func LoadConfigOrContinue(ctx context.Context, logger *log.Logger) (*domain.Config, error) {
	return loadConfig(ctx, logger, false)
}

// loadConfig is a helper that loads configuration and optionally propagates errors.
// When propagateErrors is true, errors are returned; otherwise, default config is returned.
func loadConfig(
	ctx context.Context,
	logger *log.Logger,
	propagateErrors bool,
) (*domain.Config, error) {
	loadedCfg, err := config.LoadWithContext(ctx)
	if err != nil {
		logger.Warn("Could not load configuration, using defaults", "error", err)

		if propagateErrors {
			return nil, err
		}

		// Graceful degradation: return default config
		return config.GetDefaultConfig(), nil
	}

	logger.Info("Using configuration from ~/.clean-wizard.yaml")

	return loadedCfg, nil
}

// PrintConfigSuccess prints configuration success message.
func PrintConfigSuccess(cfg *domain.Config) {
	fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n",
		cfg.SafeMode, len(cfg.Profiles))
}
