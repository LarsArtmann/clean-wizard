package config

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// SafeConfig represents a validated cleaning configuration
type SafeConfig struct {
	safeMode bool
	dryRun   bool
	backup   bool
	maxRisk  domain.RiskLevel
	profiles []SafeProfile
	created  time.Time
}

// SafeProfile represents a validated cleaning profile
type SafeProfile struct {
	name        string
	description string
	operations  []SafeOperation
	maxRisk     domain.RiskLevel
}

// SafeOperation represents a validated cleaning operation
type SafeOperation struct {
	name    CleanType
	risk    domain.RiskLevel
	enabled bool
	backup  bool
}

// String returns string representation (removed - use domain.RiskLevel methods)

// Icon returns emoji for risk level - moved to domain package

// CleanType represents type-safe cleaning types
type CleanType string

const (
	CleanTypeNixStore     CleanType = "nix_store"
	CleanTypeHomebrew     CleanType = "homebrew"
	CleanTypePackageCache CleanType = "package_cache"
	CleanTypeTempFiles    CleanType = "temp_files"
)

// IsValid checks if clean type is valid
func (ct CleanType) IsValid() bool {
	switch ct {
	case CleanTypeNixStore, CleanTypeHomebrew, CleanTypePackageCache, CleanTypeTempFiles:
		return true
	default:
		return false
	}
}

// NewSafeConfigBuilder creates a type-safe configuration builder
func NewSafeConfigBuilder() *SafeConfigBuilder {
	return &SafeConfigBuilder{
		profiles: []SafeProfile{},
		maxRisk:  domain.RiskLow,
	}
}

// SafeConfigBuilder builds type-safe configurations
type SafeConfigBuilder struct {
	safeMode bool
	dryRun   bool
	backup   bool
	maxRisk  domain.RiskLevel
	profiles []SafeProfile
	err      error
}

// SafeMode enables safe mode
func (scb *SafeConfigBuilder) SafeMode() *SafeConfigBuilder {
	scb.safeMode = true
	return scb
}

// DryRun enables dry-run mode
func (scb *SafeConfigBuilder) DryRun() *SafeConfigBuilder {
	scb.dryRun = true
	return scb
}

// Backup enables backup mode
func (scb *SafeConfigBuilder) Backup() *SafeConfigBuilder {
	scb.backup = true
	return scb
}

// AddProfile adds a safe profile
func (scb *SafeConfigBuilder) AddProfile(name, description string) *SafeProfileBuilder {
	if scb.err != nil {
		return &SafeProfileBuilder{err: scb.err}
	}

	return &SafeProfileBuilder{
		name:        name,
		description: description,
		config:      scb,
		operations:  []SafeOperation{},
		maxRisk:     domain.RiskLow,
	}
}

// Build creates safe configuration
func (scb *SafeConfigBuilder) Build() (SafeConfig, error) {
	if scb.err != nil {
		return SafeConfig{}, scb.err
	}

	if len(scb.profiles) == 0 {
		return SafeConfig{}, fmt.Errorf("config must have at least one profile")
	}

	if !scb.maxRisk.IsValid() {
		return SafeConfig{}, fmt.Errorf("invalid risk level: %s", scb.maxRisk)
	}

	return SafeConfig{
		safeMode: scb.safeMode,
		dryRun:   scb.dryRun,
		backup:   scb.backup,
		maxRisk:  scb.maxRisk,
		profiles: scb.profiles,
		created:  time.Now(),
	}, nil
}

// SafeProfileBuilder builds type-safe profiles
type SafeProfileBuilder struct {
	name        string
	description string
	config      *SafeConfigBuilder
	operations  []SafeOperation
	maxRisk     domain.RiskLevel
	err         error
}

// AddOperation adds a safe operation
func (spb *SafeProfileBuilder) AddOperation(opType CleanType, risk domain.RiskLevel) *SafeProfileBuilder {
	if spb.err != nil {
		return spb
	}

	if !opType.IsValid() {
		spb.err = fmt.Errorf("invalid clean type: %s", opType)
		return spb
	}

	if !risk.IsValid() {
		spb.err = fmt.Errorf("invalid risk level: %s", risk)
		return spb
	}

	if risk.IsHigherThan(domain.RiskHigh) && spb.err == nil {
		spb.err = fmt.Errorf("cannot add critical risk operation to profile")
		return spb
	}

	op := SafeOperation{
		name:    opType,
		risk:    risk,
		enabled: true,
		backup:  risk.IsHigherOrEqualThan(domain.RiskMedium),
	}

	spb.operations = append(spb.operations, op)
	if risk.IsHigherThan(spb.maxRisk) {
		spb.maxRisk = risk
	}

	return spb
}

// Done finishes profile building
func (spb *SafeProfileBuilder) Done() *SafeConfigBuilder {
	if spb.err != nil {
		spb.config.err = spb.err
		return spb.config
	}

	if len(spb.operations) == 0 {
		spb.config.err = fmt.Errorf("profile must have at least one operation")
		return spb.config
	}

	if spb.maxRisk.IsHigherThan(domain.RiskHigh) {
		spb.config.err = fmt.Errorf("profile risk level cannot exceed HIGH")
		return spb.config
	}

	profile := SafeProfile{
		name:        spb.name,
		description: spb.description,
		operations:  spb.operations,
		maxRisk:     spb.maxRisk,
	}

	spb.config.profiles = append(spb.config.profiles, profile)
	if spb.maxRisk > spb.config.maxRisk {
		spb.config.maxRisk = spb.maxRisk
	}

	return spb.config
}
