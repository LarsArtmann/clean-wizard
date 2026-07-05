package commands

import (
	"context"

	"github.com/LarsArtmann/clean-wizard/internal/cleaner"
)

// CleanerConfig holds configuration for each cleaner type.
type CleanerConfig struct {
	Type        CleanerType
	Name        string
	Description string
	Icon        string
	Available   CleanerAvailability
}

// GetCleanerConfigs returns all cleaner configurations with availability status.
// Uses the provided registry for dynamic discovery and availability checking.
func GetCleanerConfigs(ctx context.Context, registry *cleaner.Registry) []CleanerConfig {
	allNames := registry.Names()

	configs := make([]CleanerConfig, 0, len(allNames))
	for _, name := range allNames {
		c, ok := registry.Get(name)
		if !ok {
			continue
		}

		cleanerType, ok := registryNameToCleanerType[name]
		if !ok {
			continue // Skip unknown cleaners
		}

		configs = append(configs, CleanerConfig{
			Type:        cleanerType,
			Name:        getCleanerName(cleanerType),
			Description: getCleanerDescription(cleanerType),
			Icon:        getCleanerIcon(cleanerType),
			Available:   toCleanerAvailability(c.IsAvailable(ctx)),
		})
	}

	return configs
}

// toCleanerAvailability converts a boolean availability to the enum type.
func toCleanerAvailability(available bool) CleanerAvailability {
	if available {
		return CleanerAvailabilityAvailable
	}

	return CleanerAvailabilityUnavailable
}
