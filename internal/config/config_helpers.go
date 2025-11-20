package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// setupDefaults configures viper with default values
func setupDefaults(v *viper.Viper) {
	v.SetDefault(ConfigKeyVersion, "1.0.0")
	v.SetDefault("safe_mode", DefaultSafeMode)
	v.SetDefault(ConfigKeyMaxDiskUsage, DefaultDiskUsagePercent)
	v.SetDefault(ConfigKeyProtected, []string{DefaultProtectedPathSystem, DefaultProtectedPathLibrary}) // Basic protection
}

// readConfigFile attempts to read the configuration file with error handling
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

// fixProfileStatuses converts profile enabled/status fields to proper StatusType enums
func fixProfileStatuses(v *viper.Viper, profiles map[string]*domain.CleaningProfile) {
	for name, profile := range profiles {
		// Convert boolean enabled to StatusType enum for profiles
		var profileEnabled bool
		if v.IsSet(fmt.Sprintf("profiles.%s.enabled", name)) {
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.enabled", name), &profileEnabled); err != nil {
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
			if err := v.UnmarshalKey(fmt.Sprintf("profiles.%s.status", name), &profileStatusStr); err != nil {
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
			default:
				profile.Status = domain.StatusNotSelected // Safe default
			}
		}
	}
}
