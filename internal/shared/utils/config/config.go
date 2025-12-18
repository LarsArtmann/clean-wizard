package config

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/sirupsen/logrus"
)

// LoadConfigWithFallback loads configuration with proper error handling and user feedback
// This eliminates duplicate config loading patterns across commands
func LoadConfigWithFallback(ctx context.Context, logger *logrus.Logger) (*domain.Config, error) {
	loadedCfg, err := config.LoadWithContext(ctx)
	if err != nil {
		logger.Warnf("Could not load default configuration: %v", err)
		return nil, err // Return error for caller to handle
	}
	logger.Info("Using configuration from ~/.clean-wizard.yaml")
	return loadedCfg, nil
}

// LoadConfigOrContinue loads configuration but allows caller to continue on error
// Returns (config, error) where error can be ignored for graceful degradation
func LoadConfigOrContinue(ctx context.Context, logger *logrus.Logger) (*domain.Config, error) {
	loadedCfg, err := config.LoadWithContext(ctx)
	if err != nil {
		logger.Warnf("Could not load default configuration: %v", err)
		return nil, nil // Return nil config but no error for graceful continuation
	}
	logger.Info("Using configuration from ~/.clean-wizard.yaml")
	return loadedCfg, nil
}

// PrintConfigSuccess prints configuration success message
func PrintConfigSuccess(cfg *domain.Config) {
	fmt.Printf("✅ Configuration applied: safe_mode=%v, profiles=%d\n",
		cfg.SafeMode, len(cfg.Profiles))
}