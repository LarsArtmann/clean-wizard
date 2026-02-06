package config

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/sirupsen/logrus"
)

// LoadConfigWithFallback loads configuration with proper error handling and user feedback
// This eliminates duplicate config loading patterns across commands.
func LoadConfigWithFallback(ctx context.Context, logger *logrus.Logger) (*domain.Config, error) {
	return loadConfig(ctx, logger, true)
}

// LoadConfigOrContinue loads configuration but allows caller to continue on error
// Returns (config, error) where error can be ignored for graceful degradation.
func LoadConfigOrContinue(ctx context.Context, logger *logrus.Logger) (*domain.Config, error) {
	return loadConfig(ctx, logger, false)
}

// loadConfig is a helper that loads configuration and optionally propagates errors.
// When propagateErrors is true, errors are returned; otherwise, errors are logged and nil config is returned.
func loadConfig(ctx context.Context, logger *logrus.Logger, propagateErrors bool) (*domain.Config, error) {
	loadedCfg, err := config.LoadWithContext(ctx)
	if err != nil {
		logger.Warnf("Could not load default configuration: %v", err)
		if propagateErrors {
			return nil, err
		}
		return nil, nil
	}
	logger.Info("Using configuration from ~/.clean-wizard.yaml")
	return loadedCfg, nil
}

// PrintConfigSuccess prints configuration success message.
func PrintConfigSuccess(cfg *domain.Config) {
	fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n",
		cfg.SafeMode, len(cfg.Profiles))
}
