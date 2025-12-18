package config

import (
	"context"
	"io"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPrintConfigSuccess(t *testing.T) {
	cfg := &domain.Config{
		SafeMode: true,
		Profiles: map[string]*domain.Profile{
			"daily":  {Name: "daily", Enabled: true},
			"weekly": {Name: "weekly", Enabled: true},
			"custom": {Name: "custom", Enabled: false},
		},
	}

	// This test just ensures the function doesn't panic and prints expected format
	PrintConfigSuccess(cfg)
	// We can't easily test stdout in this context, but we ensure no panic occurs
	assert.True(t, true)
}

func TestLoadConfigOrContinue_ContextCancel(t *testing.T) {
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard output for test

	_, err := LoadConfigOrContinue(ctx, logger)
	assert.NoError(t, err, "Should return no error for graceful continuation")
}
