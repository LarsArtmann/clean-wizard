package api

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// === MAPPING LAYER: API ↔ DOMAIN TYPES ===
// Converts between public API types and internal domain models

// MapConfigToDomain converts public API config to internal domain model
func MapConfigToDomain(publicConfig *PublicConfig) result.Result[*domain.Config] {
	if publicConfig == nil {
		return result.Err[*domain.Config](fmt.Errorf("public config cannot be nil"))
	}

	// Map profiles
	profiles := make(map[string]*domain.Profile, len(publicConfig.Profiles))
	for name, publicProfile := range publicConfig.Profiles {
		domainProfile, err := MapProfileToDomain(publicProfile)
		if err != nil {
			return result.Err[*domain.Config](fmt.Errorf("failed to map profile %s: %w", name, err))
		}
		profiles[name] = domainProfile
	}

	// Create domain config
	config := &domain.Config{
		Version:     publicConfig.Version,
		SafeMode:    publicConfig.SafeMode,
		MaxDiskUsage: int(publicConfig.MaxDiskUsage),
		Protected:   publicConfig.ProtectedPaths,
		Profiles:    profiles,
	}

	// Validate domain config
	if err := config.Validate(); err != nil {
		return result.Err[*domain.Config](fmt.Errorf("domain config validation failed: %w", err))
	}

	return result.Ok(config)
}

// MapConfigToPublic converts internal domain config to public API type
func MapConfigToPublic(domainConfig *domain.Config) result.Result[*PublicConfig] {
	if domainConfig == nil {
		return result.Err[*PublicConfig](fmt.Errorf("domain config cannot be nil"))
	}

	// Map profiles
	publicProfiles := make(map[string]*PublicProfile, len(domainConfig.Profiles))
	for name, domainProfile := range domainConfig.Profiles {
		publicProfile := MapProfileToPublic(domainProfile)
		publicProfiles[name] = publicProfile
	}

	publicConfig := &PublicConfig{
		Version:      domainConfig.Version,
		SafeMode:     domainConfig.SafeMode,
		MaxDiskUsage:  int32(domainConfig.MaxDiskUsage),
		ProtectedPaths: domainConfig.Protected,
		Profiles:      publicProfiles,
	}

	return result.Ok(publicConfig)
}

// MapProfileToDomain converts public API profile to domain model
func MapProfileToDomain(publicProfile *PublicProfile) (*domain.Profile, error) {
	if publicProfile == nil {
		return nil, fmt.Errorf("public profile cannot be nil")
	}

	// Map operations
	operations := make([]domain.CleanupOperation, len(publicProfile.Operations))
	for i, publicOp := range publicProfile.Operations {
		domainOp, err := MapOperationToDomain(&publicOp)
		if err != nil {
			return nil, fmt.Errorf("failed to map operation %d: %w", i, err)
		}
		operations[i] = *domainOp
	}

	return &domain.Profile{
		Name:        publicProfile.Name,
		Description: publicProfile.Description,
		Enabled:     publicProfile.Enabled,
		Operations:  operations,
	}, nil
}

// MapProfileToPublic converts domain profile to public API type
func MapProfileToPublic(domainProfile *domain.Profile) *PublicProfile {
	if domainProfile == nil {
		return nil
	}

	// Map operations
	publicOperations := make([]PublicOperation, len(domainProfile.Operations))
	for i, domainOp := range domainProfile.Operations {
		publicOperations[i] = *MapOperationToPublic(&domainOp)
	}

	return &PublicProfile{
		Name:        domainProfile.Name,
		Description: domainProfile.Description,
		Enabled:     domainProfile.Enabled,
		Operations:  publicOperations,
	}
}

// MapOperationToDomain converts public API operation to domain model
func MapOperationToDomain(publicOperation *PublicOperation) (*domain.CleanupOperation, error) {
	if publicOperation == nil {
		return nil, fmt.Errorf("public operation cannot be nil")
	}

	// Map risk level
	riskLevel, err := MapRiskLevelToDomain(publicOperation.RiskLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid risk level: %w", err)
	}

	return &domain.CleanupOperation{
		Name:        publicOperation.Name,
		Description: publicOperation.Description,
		RiskLevel:   riskLevel,
		Enabled:     publicOperation.Enabled,
		Settings:    domain.DefaultSettings(domain.OperationTypeNixGenerations), // Simplified for PoC
	}, nil
}

// MapOperationToPublic converts domain operation to public API type
func MapOperationToPublic(domainOperation *domain.CleanupOperation) *PublicOperation {
	if domainOperation == nil {
		return nil
	}

	return &PublicOperation{
		Name:        domainOperation.Name,
		Description: domainOperation.Description,
		RiskLevel:   MapRiskLevelToPublic(domainOperation.RiskLevel),
		Enabled:     domainOperation.Enabled,
		Settings: OperationSettings{
			DryRun:             domainOperation.Settings.DryRun,
			Verbose:            domainOperation.Settings.Verbose,
			TimeoutSeconds:     int32(domainOperation.Settings.Timeout.Seconds()),
			ConfirmBeforeDelete: domainOperation.Settings.ConfirmBeforeDelete,
		},
	}
}

// MapRiskLevelToDomain converts public risk level string to domain enum
func MapRiskLevelToDomain(publicRisk PublicRiskLevel) (domain.RiskLevelType, error) {
	switch publicRisk {
	case PublicRiskLow:
		return domain.RiskLowType, nil
	case PublicRiskMedium:
		return domain.RiskMediumType, nil
	case PublicRiskHigh:
		return domain.RiskHighType, nil
	case PublicRiskCritical:
		return domain.RiskCriticalType, nil
	default:
		return domain.RiskLowType, fmt.Errorf("unknown risk level: %s", publicRisk)
	}
}

// MapRiskLevelToPublic converts domain risk level enum to public string
func MapRiskLevelToPublic(domainRisk domain.RiskLevelType) PublicRiskLevel {
	switch domainRisk {
	case domain.RiskLowType:
		return PublicRiskLow
	case domain.RiskMediumType:
		return PublicRiskMedium
	case domain.RiskHighType:
		return PublicRiskHigh
	case domain.RiskCriticalType:
		return PublicRiskCritical
	default:
		return PublicRiskLow // Default fallback
	}
}

// MapCleanResultToPublic converts domain clean result to public API type
func MapCleanResultToPublic(domainResult domain.CleanResult) result.Result[*PublicCleanResult] {
	// Map strategy
	strategy := MapStrategyToPublic(domainResult.Strategy)

	publicResult := &PublicCleanResult{
		Success:     domainResult.IsValid(),
		FreedBytes:  domainResult.FreedBytes,
		ItemsRemoved: uint32(domainResult.ItemsRemoved),
		ItemsFailed:  uint32(domainResult.ItemsFailed),
		CleanTime:   domainResult.CleanTime.String(),
		CleanedAt:   domainResult.CleanedAt.Format(time.RFC3339),
		Strategy:    strategy,
	}

	// Add validation errors if any
	if err := domainResult.Validate(); err != nil {
		publicResult.Errors = []string{err.Error()}
		publicResult.Success = false
	}

	return result.Ok(publicResult)
}

// MapStrategyToPublic converts domain strategy enum to public string
func MapStrategyToPublic(domainStrategy domain.CleanStrategyType) PublicStrategy {
	switch domainStrategy {
	case domain.StrategyAggressiveType:
		return PublicStrategyAggressive
	case domain.StrategyConservativeType:
		return PublicStrategyConservative
	case domain.StrategyDryRunType:
		return PublicStrategyDryRun
	default:
		return PublicStrategyDryRun // Safe default
	}
}