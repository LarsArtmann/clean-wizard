package api

import (
	"fmt"
	"slices"
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
		Version:      publicConfig.Version,
		SafetyLevel:  boolToSafetyLevel(publicConfig.SafeMode),
		MaxDiskUsage: int(publicConfig.MaxDiskUsage),
		Protected:    publicConfig.ProtectedPaths,
		Profiles:     profiles,
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
		Version:        domainConfig.Version,
		SafeMode:       safetyLevelToBool(domainConfig.SafetyLevel),
		MaxDiskUsage:   int32(domainConfig.MaxDiskUsage),
		ProtectedPaths: domainConfig.Protected,
		Profiles:       publicProfiles,
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
		Status:      boolToStatus(publicProfile.Enabled),
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
		Enabled:     statusToBool(domainProfile.Status),
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

	// Derive operation type from operation name
	opType := domain.GetOperationType(publicOperation.Name)

	// Check if this is a known/standard operation type
	knownTypes := []domain.OperationType{
		domain.OperationTypeNixGenerations,
		domain.OperationTypeTempFiles,
		domain.OperationTypeHomebrew,
		domain.OperationTypeSystemTemp,
	}
	isKnown := slices.Contains(knownTypes, opType)

	if !isKnown {
		return nil, fmt.Errorf("unknown/unsupported operation type: %s", publicOperation.Name)
	}

	// Get default settings for the specific operation type
	settings := domain.DefaultSettings(opType)
	if settings == nil {
		return nil, fmt.Errorf("failed to get default settings for operation type: %s", opType)
	}

	return &domain.CleanupOperation{
		Name:        publicOperation.Name,
		Description: publicOperation.Description,
		RiskLevel:   riskLevel,
		Status:      boolToStatus(publicOperation.Enabled),
		Settings:    settings,
	}, nil
}

// MapOperationToPublic converts domain operation to public API type
func MapOperationToPublic(domainOperation *domain.CleanupOperation) *PublicOperation {
	if domainOperation == nil {
		return nil
	}

	// Map complex operation settings to simplified public API settings
	publicSettings := MapOperationSettingsToPublic(domainOperation.Settings)

	return &PublicOperation{
		Name:        domainOperation.Name,
		Description: domainOperation.Description,
		RiskLevel:   MapRiskLevelToPublic(domainOperation.RiskLevel),
		Enabled:     statusToBool(domainOperation.Status),
		Settings:    publicSettings,
	}
}

// MapOperationSettingsToPublic converts complex domain settings to simplified public API settings
func MapOperationSettingsToPublic(settings *domain.OperationSettings) OperationSettings {
	// Default values for simplified public API
	publicSettings := OperationSettings{
		DryRun:              true, // Safe default
		Verbose:             false,
		TimeoutSeconds:      300, // 5 minutes
		ConfirmBeforeDelete: false,
	}

	// Extract relevant values from domain-specific settings
	if settings.NixGenerations != nil {
		publicSettings.DryRun = false // Nix operations default to false
		if settings.NixGenerations.Optimization != domain.OptimizationLevelNone {
			publicSettings.Verbose = true
		}
	}

	return publicSettings
}

// MapRiskLevelToDomain converts public risk level string to domain enum
func MapRiskLevelToDomain(publicRisk PublicRiskLevel) (domain.RiskLevelType, error) {
	switch publicRisk {
	case PublicRiskLow:
		return domain.RiskLow, nil
	case PublicRiskMedium:
		return domain.RiskMedium, nil
	case PublicRiskHigh:
		return domain.RiskHigh, nil
	case PublicRiskCritical:
		return domain.RiskCritical, nil
	default:
		return domain.RiskLow, fmt.Errorf("unknown risk level: %s", publicRisk)
	}
}

// MapRiskLevelToPublic converts domain risk level enum to public string
func MapRiskLevelToPublic(domainRisk domain.RiskLevelType) PublicRiskLevel {
	switch domainRisk {
	case domain.RiskLow:
		return PublicRiskLow
	case domain.RiskMedium:
		return PublicRiskMedium
	case domain.RiskHigh:
		return PublicRiskHigh
	case domain.RiskCritical:
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
		Success:      domainResult.IsValid(),
		FreedBytes:   domainResult.FreedBytes,
		ItemsRemoved: uint32(domainResult.ItemsRemoved),
		ItemsFailed:  uint32(domainResult.ItemsFailed),
		CleanTime:    domainResult.CleanTime.String(),
		CleanedAt:    domainResult.CleanedAt.Format(time.RFC3339),
		Strategy:     strategy,
	}

	// Add validation errors if any
	if err := domainResult.Validate(); err != nil {
		publicResult.Errors = []string{err.Error()}
		publicResult.Success = false
	}

	return result.Ok(publicResult)
}

// MapCleanRequestToDomain converts public API clean request to domain model
func MapCleanRequestToDomain(publicRequest *CleanRequest) result.Result[*domain.CleanRequest] {
	if publicRequest == nil {
		return result.Err[*domain.CleanRequest](fmt.Errorf("public clean request cannot be nil"))
	}

	// Map config to domain
	domainConfigResult := MapConfigToDomain(&publicRequest.Config)
	if domainConfigResult.IsErr() {
		err, _ := domainConfigResult.SafeError()
		return result.Err[*domain.CleanRequest](fmt.Errorf("failed to map config: %w", err))
	}

	// Map strategy
	domainStrategy, err := MapStrategyToDomain(publicRequest.Strategy)
	if err != nil {
		return result.Err[*domain.CleanRequest](fmt.Errorf("invalid strategy: %w", err))
	}

	// Map operations to domain scan items
	items := make([]domain.ScanItem, 0, len(publicRequest.Operations))
	for _, opType := range publicRequest.Operations {
		// Convert operation type to domain scan item
		item := domain.ScanItem{
			Path:     string(opType),      // Use operation type as path identifier
			ScanType: domain.ScanTypeTemp, // Use temp scan type for operations
			Size:     0,                   // Will be calculated during scanning
			Created:  time.Time{},         // Will be set during scanning
		}
		items = append(items, item)
	}

	domainRequest := &domain.CleanRequest{
		Items:    items,
		Strategy: domainStrategy,
	}

	// Validate domain request
	if err := domainRequest.Validate(); err != nil {
		return result.Err[*domain.CleanRequest](fmt.Errorf("domain clean request validation failed: %w", err))
	}

	return result.Ok(domainRequest)
}

// MapStrategyToDomain converts public strategy to domain enum
func MapStrategyToDomain(publicStrategy PublicStrategy) (domain.CleanStrategyType, error) {
	switch publicStrategy {
	case PublicStrategyAggressive:
		return domain.StrategyAggressive, nil
	case PublicStrategyConservative:
		return domain.StrategyConservative, nil
	case PublicStrategyDryRun:
		return domain.StrategyDryRun, nil
	default:
		return domain.StrategyDryRun, fmt.Errorf("unknown strategy: %s", publicStrategy)
	}
}

// MapStrategyToPublic converts domain strategy enum to public string
func MapStrategyToPublic(domainStrategy domain.CleanStrategyType) PublicStrategy {
	switch domainStrategy {
	case domain.StrategyAggressive:
		return PublicStrategyAggressive
	case domain.StrategyConservative:
		return PublicStrategyConservative
	case domain.StrategyDryRun:
		return PublicStrategyDryRun
	default:
		return PublicStrategyDryRun // Safe default
	}
}

// MapScanResultToPublic converts domain scan result to public API format
func MapScanResultToPublic(domainResult *domain.ScanResult) result.Result[*PublicScanResult] {
	if domainResult == nil {
		return result.Err[*PublicScanResult](fmt.Errorf("domain scan result cannot be nil"))
	}

	publicResult := &PublicScanResult{
		Success:   true, // Assuming success if result exists
		ScanTime:  domainResult.ScannedAt.Format(time.RFC3339),
		TotalSize: uint64(domainResult.TotalBytes),
		ItemCount: uint32(domainResult.TotalItems),
		Items:     make([]PublicScanItem, 0), // No individual items in ScanResult
	}

	return result.Ok(publicResult)
}

// MapValidationResultToPublic converts domain validation result to public API format
func MapValidationResultToPublic(domainResult any) result.Result[*PublicValidationResult] {
	// For now, return a simple valid result
	// This would be implemented with actual domain validation result type
	publicResult := &PublicValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	return result.Ok(publicResult)
}

// Helper functions for boolean-to-enum conversion

// boolToSafetyLevel converts boolean SafeMode to SafetyLevelType
func boolToSafetyLevel(safeMode bool) domain.SafetyLevelType {
	if safeMode {
		return domain.SafetyLevelEnabled
	}
	return domain.SafetyLevelDisabled
}

// safetyLevelToBool converts SafetyLevelType to boolean (backward compatibility)
func safetyLevelToBool(level domain.SafetyLevelType) bool {
	return level == domain.SafetyLevelEnabled || level == domain.SafetyLevelStrict || level == domain.SafetyLevelParanoid
}

// boolToStatus converts boolean Enabled to StatusType
func boolToStatus(enabled bool) domain.StatusType {
	if enabled {
		return domain.StatusEnabled
	}
	return domain.StatusDisabled
}

// statusToBool converts StatusType to boolean (backward compatibility)
func statusToBool(status domain.StatusType) bool {
	return status == domain.StatusEnabled
}
