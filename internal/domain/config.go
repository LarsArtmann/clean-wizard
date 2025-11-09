package domain

import (
	"fmt"
	"time"
)

// Config represents application configuration with type safety
type Config struct {
	Version      string             `json:"version" yaml:"version"`
	SafeMode     bool               `json:"safe_mode" yaml:"safe_mode"`
	MaxDiskUsage int                `json:"max_disk_usage" yaml:"max_disk_usage"`
	Protected    []string          `json:"protected" yaml:"protected"`
	Profiles    map[string]*Profile `json:"profiles" yaml:"profiles"`
	LastClean    time.Time          `json:"last_clean" yaml:"last_clean"`
	Updated     time.Time          `json:"updated" yaml:"updated"`
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
		return fmt.Errorf("Protected paths cannot be empty")
	}
	
	for i, path := range c.Protected {
		if path == "" {
			return fmt.Errorf("Protected path %d cannot be empty", i)
		}
	}
	
	if len(c.Profiles) == 0 {
		return fmt.Errorf("Configuration must have at least one profile")
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
	Enabled     bool               `json:"enabled" yaml:"enabled"`
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
	
	return true
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
	
	return nil
}

// CleanupOperation represents single cleanup operation
type CleanupOperation struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	RiskLevel   RiskLevel      `json:"risk_level" yaml:"risk_level"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Settings    map[string]any `json:"settings,omitempty" yaml:"settings,omitempty"`
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
	if !op.RiskLevel.IsValid() {
		return fmt.Errorf("Invalid risk level: %s", op.RiskLevel)
	}
	return nil
}