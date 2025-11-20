package domain

import (
	"fmt"
	"time"
)

// Config represents application configuration with type safety
// TODO: CRITICAL - Replace all primitive types with value types:
// - MaxDiskUsage int -> MaxDiskUsage uint8
// - CurrentProfile string -> ProfileName
// - Protected []string -> ProtectedPaths type
// - All profile fields with proper value types
type Config struct {
	Version        string              `json:"version" yaml:"version"`
	SafetyLevel    SafetyLevelType     `json:"safety_level" yaml:"safety_level"`
	MaxDiskUsage   int                 `json:"max_disk_usage" yaml:"max_disk_usage"` // TODO: Replace with MaxDiskUsage
	Protected      []string            `json:"protected" yaml:"protected"`           // TODO: Replace with ProtectedPaths type
	Profiles       map[string]*Profile `json:"profiles" yaml:"profiles"`
	CurrentProfile string              `json:"current_profile,omitempty" yaml:"current_profile,omitempty"` // TODO: Replace with ProfileName
	LastClean      time.Time           `json:"last_clean" yaml:"last_clean"`
	Updated        time.Time           `json:"updated" yaml:"updated"`
}

// TODO: Add type-safe constructor with validation
func NewConfig(version string, safetyLevel SafetyLevelType, maxDiskUsage MaxDiskUsage) (*Config, error) {
	if !maxDiskUsage.IsValid() {
		return nil, fmt.Errorf("invalid max disk usage: %d", maxDiskUsage.Uint8())
	}
	
	// TODO: Add complete validation for all fields
	return &Config{
		Version:      version,
		SafetyLevel:  safetyLevel,
		MaxDiskUsage: int(maxDiskUsage.Uint8()), // TODO: Remove type conversion
		Protected:    []string{"/System", "/Applications", "/Library"}, // TODO: Use ProtectedPaths
		Profiles:     make(map[string]*Profile),
		LastClean:    time.Now(),
		Updated:      time.Now(),
	}, nil
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
	Status      StatusType         `json:"status" yaml:"status"`
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
	RiskLevel   RiskLevel          `json:"risk_level" yaml:"risk_level"`
	Status      StatusType         `json:"status" yaml:"status"`
	Settings    *OperationSettings `json:"settings,omitempty" yaml:"settings,omitempty"`
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

	// Validate settings if present
	if op.Settings != nil {
		opType := GetOperationType(op.Name)
		if err := op.Settings.ValidateSettings(opType); err != nil {
			return fmt.Errorf("Operation settings validation failed: %w", err)
		}
	}

	return nil
}
