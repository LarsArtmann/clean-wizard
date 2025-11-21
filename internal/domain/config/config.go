package config

import (
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// Config represents application configuration with type safety
// TODO: CRITICAL - Replace all primitive types with value types:
// - MaxDiskUsage int -> MaxDiskUsage uint8
// - CurrentProfile string -> ProfileName
// - Protected []string -> ProtectedPaths type
// - All profile fields with proper value types
type Config struct {
	Version        string              `json:"version" yaml:"version"`
	SafetyLevel    shared.SafetyLevelType     `json:"safety_level" yaml:"safety_level"`
	MaxDiskUsage   int                 `json:"max_disk_usage" yaml:"max_disk_usage"` // TODO: Replace with MaxDiskUsage
	Protected      []string            `json:"protected" yaml:"protected"`           // TODO: Replace with ProtectedPaths type
	Profiles       map[string]*Profile `json:"profiles" yaml:"profiles"`
	CurrentProfile string              `json:"current_profile,omitempty" yaml:"current_profile,omitempty"` // TODO: Replace with ProfileName
	LastClean      time.Time           `json:"last_clean" yaml:"last_clean"`
	Updated        time.Time           `json:"updated" yaml:"updated"`
}

// CreateDefaultConfig creates a working default configuration
// This eliminates the #1 user barrier: manual configuration requirements
func CreateDefaultConfig() (*Config, error) {
	// Create type-safe values
	maxDiskUsage, err := shared.NewMaxDiskUsage(50) // Conservative 50% disk usage limit
	if err != nil {
		return nil, fmt.Errorf("failed to create max disk usage: %w", err)
	}

	currentProfile, err := shared.NewProfileName("daily")
	if err != nil {
		return nil, fmt.Errorf("failed to create current profile name: %w", err)
	}

	quickProfile, err := shared.NewProfileName("quick")
	if err != nil {
		return nil, fmt.Errorf("failed to create quick profile name: %w", err)
	}

	comprehensiveProfile, err := shared.NewProfileName("comprehensive")
	if err != nil {
		return nil, fmt.Errorf("failed to create comprehensive profile name: %w", err)
	}

	config := &Config{
		Version:      "1.0.0",
		SafetyLevel:  shared.SafetyLevelEnabled, // Safe by default
		MaxDiskUsage: int(maxDiskUsage.Uint8()),
		Protected: []string{
			"/System",
			"/Applications", 
			"/Library",
			"/usr/local",
			"/Users/*/Documents",
			"/Users/*/Desktop",
		},
		CurrentProfile: currentProfile.String(),
		LastClean:      time.Now(),
		Updated:        time.Now(),
		Profiles: make(map[string]*Profile),
	}

	// Create default profiles with working operations
	config.Profiles["quick"] = &Profile{
		Name:        quickProfile.String(),
		Description: "Quick daily cleanup (safe operations only)",
		Status:      shared.StatusEnabled,
		Operations: []CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "temp-files", 
				Description: "Clean temporary files from user directories",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
		},
	}

	config.Profiles["comprehensive"] = &Profile{
		Name:        comprehensiveProfile.String(),
		Description: "Comprehensive system cleanup with development tools",
		Status:      shared.StatusEnabled,
		Operations: []CleanupOperation{
			{
				Name:        "nix-generations",
				Description: "Remove old Nix store generations", 
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "homebrew",
				Description: "Homebrew cleanup, autoremove and cache cleaning",
				RiskLevel:   shared.RiskLevelMediumType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "npm-cache",
				Description: "Node.js npm cache cleanup",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "pnpm-store",
				Description: "pnpm store cleanup",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "go-cache",
				Description: "Go build and module cache cleanup", 
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "cargo-cache",
				Description: "Rust Cargo cache cleanup",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "temp-files",
				Description: "System and user temporary file cleanup",
				RiskLevel:   shared.RiskLevelLowType,
				Status:      shared.StatusEnabled,
			},
			{
				Name:        "docker",
				Description: "Docker system cleanup (containers, images, volumes)",
				RiskLevel:   shared.RiskLevelMediumType,
				Status:      shared.StatusDisabled, // Disabled by default for safety
			},
		},
	}

	return config, nil
}

// ShouldCreateDefaultConfig determines if default config should be created
func ShouldCreateDefaultConfig(configPath string) bool {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return true
	}
	return false
}

// IsValid validates configuration
func (c *Config) IsValid() bool {
	if c.MaxDiskUsage < 0 || c.MaxDiskUsage > 100 {
		return false
	}

	if len(c.Protected) == 0 {
		return false
	}

	if len(c.Profiles) == 0 {
		return false
	}

	for name, profile := range c.Profiles {
		if !profile.IsValid(name) {
			return false
		}
	}

	return true
}

// Validate returns errors for invalid configuration
func (c *Config) Validate() error {
	if c.MaxDiskUsage < 0 || c.MaxDiskUsage > 100 {
		return fmt.Errorf("MaxDiskUsage must be between 0 and 100, got: %d", c.MaxDiskUsage)
	}

	if len(c.Protected) == 0 {
		return fmt.Errorf("protected paths cannot be empty")
	}

	for i, path := range c.Protected {
		if path == "" {
			return fmt.Errorf("protected path %d cannot be empty", i)
		}
	}

	if len(c.Profiles) == 0 {
		return fmt.Errorf("configuration must have at least one profile")
	}

	for name, profile := range c.Profiles {
		if err := profile.Validate(name); err != nil {
			return err
		}
	}

	return nil
}

// Profile represents cleanup profile
type Profile struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Operations  []CleanupOperation `json:"operations" yaml:"operations"`
	Status      shared.StatusType         `json:"status" yaml:"status"`
}

// IsValid validates profile
func (p *Profile) IsValid(name string) bool {
	if p.Name == "" {
		return false
	}
	if p.Description == "" {
		return false
	}
	if len(p.Operations) == 0 {
		return false
	}

	for _, op := range p.Operations {
		if !op.IsValid() {
			return false
		}
	}

	return p.Status.IsValid()
}

// Validate returns errors for invalid profile
func (p *Profile) Validate(name string) error {
	if p.Name == "" {
		return fmt.Errorf("Profile %s: name cannot be empty", name)
	}
	if p.Description == "" {
		return fmt.Errorf("Profile %s: description cannot be empty", name)
	}
	if len(p.Operations) == 0 {
		return fmt.Errorf("Profile %s: must have at least one operation", name)
	}

	for i, op := range p.Operations {
		if err := op.Validate(); err != nil {
			return fmt.Errorf("Profile %s: operation %d invalid: %w", name, i, err)
		}
	}

	if !p.Status.IsValid() {
		return fmt.Errorf("Profile %s: invalid status: %s", name, p.Status.String())
	}

	return nil
}

// CleanupOperation represents single cleanup operation with type-safe settings
type CleanupOperation struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	RiskLevel   shared.RiskLevelType      `json:"risk_level" yaml:"risk_level"`
	Status      shared.StatusType         `json:"status" yaml:"status"`
	Settings    *shared.OperationSettings `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// IsValid validates cleanup operation
func (op CleanupOperation) IsValid() bool {
	if op.Name == "" {
		return false
	}
	if op.Description == "" {
		return false
	}
	if !op.RiskLevel.IsValid() {
		return false
	}
	if !op.Status.IsValid() {
		return false
	}
	return true
}

// Validate returns errors for invalid cleanup operation
func (op CleanupOperation) Validate() error {
	if op.Name == "" {
		return fmt.Errorf("Operation name cannot be empty")
	}
	if op.Description == "" {
		return fmt.Errorf("Operation description cannot be empty")
	}

// 	// Validate settings if present
// 	if op.Settings != nil {
// 		opType := GetOperationType(op.Name)
// 		if err := op.Settings.ValidateSettings(opType); err != nil {
// 			return fmt.Errorf("Operation settings validation failed: %w", err)
// 		}
// 	}

	return nil
}
