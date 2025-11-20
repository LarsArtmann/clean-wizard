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
func (cf *ConfigFixer) fixProfileStatuses(profiles map[string]*domain.CleaningProfile) {
	for name, profile := range profiles {
		// Convert boolean enabled to StatusType enum for profiles
		var profileEnabled bool
		if cf.v.IsSet(fmt.Sprintf("profiles.%s.enabled", name)) {
			if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled); err != nil {
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile enabled flag")
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
				log.Warn().Err(err).Str("profile", name).Msg("Failed to unmarshal profile status")
			}
			switch strings.ToUpper(strings.TrimSpace(profileStatusStr)) {
			case "DISABLED":
				profile.Status = domain.StatusDisabled
			case "ENABLED":
				profile.Status = domain.StatusEnabled
			case "NOT_SELECTED":
				profile.Status = domain.StatusNotSelected
			case "SELECTED":
				profile.Status = domain.StatusSelected
			case "INHERITED":
				profile.Status = domain.StatusInherited
			default:
				if profileStatusStr != "" {
					log.Warn().Str("profile", name).Str("status", profileStatusStr).Msg("Invalid profile status, defaulting to ENABLED")
				}
				profile.Status = domain.StatusEnabled
			}
		}
	}
}

// fixOperationStatusesAndRiskLevels converts operation status and risk level fields to proper enums
func (cf *ConfigFixer) fixOperationStatusesAndRiskLevels(profiles map[string]*domain.CleaningProfile) {
	for name, profile := range profiles {
		for i, op := range profile.Operations {
			// Fix risk levels
			cf.fixOperationRiskLevel(name, i, op)
			
			// Fix operation statuses
			cf.fixOperationStatus(name, i, op)
		}
	}
}

// fixOperationRiskLevel converts string risk level to RiskLevel enum
func (cf *ConfigFixer) fixOperationRiskLevel(profileName string, opIndex int, op *domain.OperationSettings) {
	var riskLevelStr string
	if err := cf.v.UnmarshalKey(fmt.Sprintf("profiles.%s.operations.%d.risk_level", profileName, opIndex), &riskLevelStr); err != nil {
		log.Warn().Err(err).Str("profile", profileName).Int("operation", opIndex).Msg("Failed to unmarshal risk level")
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
		log.Warn().Str("risk_level", riskLevelStr).Msg("Invalid risk level, defaulting to LOW")
		op.RiskLevel = domain.RiskLow
	}
}

// fixOperationStatus converts operation enabled/status fields to proper StatusType enum
func (cf *ConfigFixer) fixOperationStatus(profileName string, opIndex int, op *domain.OperationSettings) {
	// Try boolean enabled first
	operationEnabledKey := fmt.Sprintf("profiles.%s.operations.%d.enabled", profileName, opIndex)
	if cf.v.IsSet(operationEnabledKey) {
		var operationEnabled bool
		if err := cf.v.UnmarshalKey(operationEnabledKey, &operationEnabled); err != nil {
			log.Warn().Err(err).Str("profile", profileName).Int("operation", opIndex).Msg("Failed to unmarshal operation enabled flag")
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
		log.Warn().Err(err).Str("profile", profileName).Int("operation", opIndex).Msg("Failed to unmarshal operation status")
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
			log.Warn().Str("profile", profileName).Int("operation", opIndex).Str("status", opStatusStr).Msg("Invalid operation status, defaulting to ENABLED")
		}
		op.Status = domain.StatusEnabled
	}
}

// Legacy helper functions for backward compatibility

// setupDefaults configures viper with default values
func setupDefaults(v *viper.Viper) {
	v.SetDefault(ConfigKeyVersion, "1.0.0")
	v.SetDefault("safe_mode", DefaultSafeMode)
	v.SetDefault(ConfigKeyMaxDiskUsage, DefaultDiskUsagePercent)
	v.SetDefault(ConfigKeyProtected, []string{DefaultProtectedPathSystem, DefaultProtectedPathLibrary}) // Basic protection
}

// readConfigFile attempts to read configuration file with error handling
func readConfigFile(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found - this is handled by caller
			return nil
		}
		return err
	}
	return nil
}

// fixProfileStatuses converts profile enabled/status fields to proper StatusType enums (DEPRECATED - use ConfigFixer)
func fixProfileStatuses(v *viper.Viper, profiles map[string]*domain.CleaningProfile) {
	fixer := NewConfigFixer(v)
	fixer.fixProfileStatuses(profiles)
}