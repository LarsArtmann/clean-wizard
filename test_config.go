package main

import (
	"fmt"
	"log"

	"github.com/LarsArtmann/clean-wizard/internal/pkg/config"
	"github.com/LarsArtmann/clean-wizard/internal/pkg/types"
)

func main() {
	fmt.Println("Testing config loading...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Config loaded successfully:\n")
	fmt.Printf("Protected paths: %v\n", cfg.Protected)
	fmt.Printf("Profiles count: %d\n", len(cfg.Profiles))

	// Try to marshal back to see the structure
	fmt.Printf("\nConfig structure:\n")
	fmt.Printf("Version: %s\n", cfg.Version)
	fmt.Printf("SafeMode: %t\n", cfg.SafeMode)
	fmt.Printf("DryRun: %t\n", cfg.DryRun)
	fmt.Printf("Verbose: %t\n", cfg.Verbose)
	fmt.Printf("Backup: %t\n", cfg.Backup)
	fmt.Printf("MaxDiskUsage: %d\n", cfg.MaxDiskUsage)
	fmt.Printf("Protected: %v\n", cfg.Protected)
	fmt.Printf("Profiles: %d\n", len(cfg.Profiles))

	// Test with a minimal config
	minimalCfg := &types.Config{
		Version:      "test",
		SafeMode:     false,
		DryRun:       false,
		Verbose:      false,
		Backup:       false,
		MaxDiskUsage: 80,
		Protected:    []string{"/test"},
		Profiles: map[string]types.Profile{
			"test": {
				Name:        "test",
				Description: "test profile",
				Operations: []types.CleanupOperation{
					{
						Name:        "test-op",
						Description: "test operation",
						RiskLevel:   types.RiskLow,
						Enabled:     true,
					},
				},
			},
		},
	}

	fmt.Printf("\nTesting minimal config validation...\n")
	err = config.Save(minimalCfg)
	if err != nil {
		log.Fatalf("Failed to save minimal config: %v", err)
	}

	// Now try to load it again
	cfg2, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to reload config: %v", err)
	}

	fmt.Printf("Reloaded config successfully:\n")
	fmt.Printf("Protected paths: %v\n", cfg2.Protected)
	fmt.Printf("Profiles count: %d\n", len(cfg2.Profiles))
}
