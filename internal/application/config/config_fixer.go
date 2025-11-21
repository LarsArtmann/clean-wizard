package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/spf13/viper"
)

// ConfigFixer consolidates all configuration fixing logic to eliminate duplications
type ConfigFixer struct {
	v *viper.Viper
}

// NewConfigFixer creates a new config fixer instance
func NewConfigFixer(v *viper.Viper) *ConfigFixer {
	return &ConfigFixer{v: v}
}

// FixAll fixes all configuration issues using consolidated logic
func (cf *ConfigFixer) FixAll(config *domain.Config) {
	cf.fixProfileStatuses(config.Profiles)
	cf.fixOperationStatusesAndRiskLevels(config.Profiles)
}

// fixProfileStatuses converts profile enabled/status fields to proper StatusType enums
func (cf *ConfigFixer) fixProfileStatuses(profiles map[string]*domain.Profile) {
	for name, profile := range profiles {
		// Convert boolean enabled to StatusType enum for profiles
		var profileEnabled bool
		if cf.v.IsSet(fmt.Sprintf("profiles.%s.enabled", name)) {
			if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled); err != nil {
				slog.Warn("Failed to unmarshal profile enabled flag", "err", err, "profile", name)
			}
			if profileEnabled {
				profile.Status = domain.StatusEnabled
			} else {
				profile.Status = domain.StatusDisabled
			}
		} else {
			// Fallback to string status parsing for backward compatibility
			var profileStatusStr string
			if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr); err != nil {
				slog.Warn("Failed to unmarshal profile status", "err", err, "profile", name)
			}
			switch strings.ToUpper(strings.TrimSpace(profileStatusStr)) {
			case "DISABLED":
				profile.Status = domain.StatusDisabled
			case "ENABLED":
				profile.Status = domain.StatusEnabled
			case "INHERITED":
				profile.Status = domain.StatusInherited
			default:
				if profileStatusStr != "" {
					slog.Warn("Invalid profile status, defaulting to ENABLED", "profile", name, "status", profileStatusStr)
				}
				profile.Status = domain.StatusEnabled
			}
		}
	}
}

// fixOperationStatusesAndRiskLevels converts operation status and risk level fields to proper enums
func (cf *ConfigFixer) fixOperationStatusesAndRiskLevels(profiles map[string]*domain.Profile) {
	for name, profile := range profiles {
		for i, op := range profile.Operations {
			// Fix risk levels
			cf.fixOperationRiskLevel(name, i, &op)

			// Fix operation statuses
			cf.fixOperationStatus(name, i, &op)
		}
	}
}

// fixOperationRiskLevel converts string risk level to RiskLevel enum
func (cf *ConfigFixer) fixOperationRiskLevel(profileName string, opIndex int, op *domain.CleanupOperation) {
	var riskLevelStr string
	if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", profileName, opIndex), &riskLevelStr); err != nil {
		slog.Warn("Failed to unmarshal risk level", "err", err, "profile", profileName, "operation", opIndex)
	}

	switch strings.ToUpper(strings.TrimSpace(riskLevelStr)) {
	case "LOW":
		op.RiskLevel = domain.RiskLow
	case "MEDIUM":
		op.RiskLevel = domain.RiskMedium
	case "HIGH":
		op.RiskLevel = domain.RiskHigh
	case "CRITICAL":
		op.RiskLevel = domain.RiskCritical
	default:
		slog.Warn("Invalid risk level, defaulting to LOW", "risk_level", riskLevelStr)
		op.RiskLevel = domain.RiskLow
	}
}

// fixOperationStatus converts operation enabled/status fields to proper StatusType enum
func (cf *ConfigFixer) fixOperationStatus(profileName string, opIndex int, op *domain.CleanupOperation) {
	// Try boolean enabled first
	operationEnabledKey := fmt.Sprintf("profiles.%s.operations.%d.enabled", profileName, opIndex)
	if cf.v.IsSet(operationEnabledKey) {
		var operationEnabled bool
		if err := cf.v.UnmarshalKey(operationEnabledKey, &operationEnabled); err != nil {
			slog.Warn("Failed to unmarshal operation enabled flag", "err", err, "profile", profileName, "operation", opIndex)
		}
		if operationEnabled {
			op.Status = domain.StatusEnabled
		} else {
			op.Status = domain.StatusDisabled
		}
		return
	}

	// Fallback to string status parsing
	var opStatusStr string
	if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.status", profileName, opIndex), &opStatusStr); err != nil {
		slog.Warn("Failed to unmarshal operation status", "err", err, "profile", profileName, "operation", opIndex)
	}

	switch strings.ToUpper(strings.TrimSpace(opStatusStr)) {
	case "DISABLED":
		op.Status = domain.StatusDisabled
	case "ENABLED":
		op.Status = domain.StatusEnabled
	case "INHERITED":
		op.Status = domain.StatusInherited
	default:
		if opStatusStr != "" {
			slog.Warn("Invalid operation status, defaulting to ENABLED", "profile", profileName, "operation", opIndex, "status", opStatusStr)
		}
		op.Status = domain.StatusEnabled
	}
}
