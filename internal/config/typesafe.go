package config

import (
	"fmt"
)

// SafeProfile represents a validated cleaning profile
type SafeProfile struct {
	name        string
	description string
	operations  []SafeOperation
	maxRisk     RiskLevel
}

// SafeOperation represents a validated cleaning operation
type SafeOperation struct {
	name      CleanType
	risk      RiskLevel
	enabled   bool
	backup    bool
	dryRun    bool
}

// RiskLevel represents operation risk with type safety
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

// String returns string representation
func (rl RiskLevel) String() string {
	switch rl {
	case RiskLow:
		return "LOW"
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	case RiskCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Icon returns emoji for risk level
func (rl RiskLevel) Icon() string {
	switch rl {
	case RiskLow:
		return "🟢"
	case RiskMedium:
		return "🟡"
	case RiskHigh:
		return "🟠"
	case RiskCritical:
		return "🔴"
	default:
		return "⚪"
	}
}

// NewSafeProfile creates a type-safe profile
func NewSafeProfile(name, description string) *SafeProfileBuilder {
	return &SafeProfileBuilder{
		name:        name,
		description: description,
		operations:  []SafeOperation{},
		maxRisk:     RiskLow,
	}
}

// SafeProfileBuilder builds type-safe profiles
type SafeProfileBuilder struct {
	name        string
	description string
	operations  []SafeOperation
	maxRisk     RiskLevel
	err         error
}

// AddOperation adds a safe operation
func (spb *SafeProfileBuilder) AddOperation(opType CleanType, risk RiskLevel) *SafeProfileBuilder {
	if risk > RiskHigh && spb.err == nil {
		spb.err = fmt.Errorf("cannot add critical risk operation to profile")
		return spb
	}
	
	op := SafeOperation{
		name:    opType,
		risk:    risk,
		enabled: true,
		backup:  risk >= RiskMedium,
		dryRun:  risk >= RiskHigh,
	}
	
	spb.operations = append(spb.operations, op)
	if risk > spb.maxRisk {
		spb.maxRisk = risk
	}
	
	return spb
}

// Build creates the safe profile
func (spb *SafeProfileBuilder) Build() (SafeProfile, error) {
	if spb.err != nil {
		return SafeProfile{}, spb.err
	}
	
	if len(spb.operations) == 0 {
		return SafeProfile{}, fmt.Errorf("profile must have at least one operation")
	}
	
	if spb.maxRisk > RiskHigh {
		return SafeProfile{}, fmt.Errorf("profile risk level cannot exceed HIGH")
	}
	
	return SafeProfile{
		name:        spb.name,
		description: spb.description,
		operations:  spb.operations,
		maxRisk:     spb.maxRisk,
	}, nil
}

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
