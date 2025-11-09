package domain

import (
	"time"
)

// Config represents application configuration with type safety
type Config struct {
	Version      string             `json:"version" yaml:"version"`
	SafeMode     bool               `json:"safe_mode" yaml:"safe_mode"`
	DryRun       bool               `json:"dry_run" yaml:"dry_run"`
	Verbose      bool               `json:"verbose" yaml:"verbose"`
	Backup       bool               `json:"backup" yaml:"backup"`
	MaxDiskUsage int                `json:"max_disk_usage_percent" yaml:"max_disk_usage_percent"`
	Protected    []string           `json:"protected_paths" yaml:"protected_paths"`
	Profiles     map[string]Profile `json:"profiles" yaml:"profiles"`
}

// Profile represents cleaning profile with operations
type Profile struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Operations  []CleanupOperation `json:"operations" yaml:"operations"`
	CreatedAt   time.Time          `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" yaml:"updated_at"`
}

// CleanupOperation represents single cleanup operation
type CleanupOperation struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	RiskLevel   RiskLevel      `json:"risk_level" yaml:"risk_level"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Settings    map[string]any `json:"settings,omitempty" yaml:"settings,omitempty"`
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

// IsValid validates cleanup operation
func (op *CleanupOperation) IsValid() bool {
	if op.Name == "" {
		return false
	}
	if !op.RiskLevel.IsValid() {
		return false
	}
	return true
}