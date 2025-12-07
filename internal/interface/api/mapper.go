package api

import (
	"fmt"

	domainconfig "github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// Mapper provides type-safe mapping between domain and API types
type Mapper struct{}

// NewMapper creates a new type-safe mapper
func NewMapper() *Mapper {
	return &Mapper{}
}

// DomainConfigToPublic maps domain config to public API config
func (m *Mapper) DomainConfigToPublic(domain *domainconfig.Config) *PublicConfig {
	if domain == nil {
		return nil
	}

	public := &PublicConfig{
		Version:        domain.Version,
		SafetyLevel:    domain.SafetyLevel,
		MaxDiskUsage:   int32(domain.MaxDiskUsage),
		ProtectedPaths: append([]string(nil), domain.Protected...),
		Profiles:       make(map[string]*PublicProfile),
	}

	// Map profiles
	for name, profile := range domain.Profiles {
		public.Profiles[name] = m.DomainProfileToPublic(profile)
	}

	return public
}

// DomainProfileToPublic maps domain profile to public API profile
func (m *Mapper) DomainProfileToPublic(domain *domainconfig.Profile) *PublicProfile {
	if domain == nil {
		return nil
	}

	public := &PublicProfile{
		Name:        domain.Name,
		Description: domain.Description,
		Status:      domain.Status,
		Operations:  make([]PublicOperation, len(domain.Operations)),
	}

	// Map operations
	for i, op := range domain.Operations {
		public.Operations[i] = m.DomainOperationToPublic(op)
	}

	return public
}

// DomainOperationToPublic maps domain operation to public API operation
func (m *Mapper) DomainOperationToPublic(domain domainconfig.CleanupOperation) PublicOperation {
	publicRisk := PublicRiskLow
	switch domain.RiskLevel {
	case shared.RiskLevelLowType:
		publicRisk = PublicRiskLow
	case shared.RiskLevelMediumType:
		publicRisk = PublicRiskMedium
	case shared.RiskLevelHighType:
		publicRisk = PublicRiskHigh
	case shared.RiskLevelCriticalType:
		publicRisk = PublicRiskCritical
	}

	return PublicOperation{
		Name:        domain.Name,
		Description: domain.Description,
		RiskLevel:   publicRisk,
		Status:      domain.Status,
		Settings: OperationSettings{
			ExecutionMode:       shared.ExecutionModeDryRun, // Default safe mode
			Verbose:             false,
			TimeoutSeconds:      300, // 5 minutes default
			ConfirmBeforeDelete: true,
		},
	}
}

// MapConfigToDomain maps public API config to domain config
func MapConfigToDomain(public *PublicConfig) result.Result[*domainconfig.Config] {
	if public == nil {
		return result.Err[*domainconfig.Config](fmt.Errorf("public config is nil"))
	}

	domain := &domainconfig.Config{
		Version:      public.Version,
		SafetyLevel:  public.SafetyLevel,
		MaxDiskUsage: int(public.MaxDiskUsage),
		Protected:    append([]string(nil), public.ProtectedPaths...),
		Profiles:     make(map[string]*domainconfig.Profile),
	}

	// Map profiles
	for name, profile := range public.Profiles {
		domainProfile := mapPublicProfileToDomain(profile)
		if domainProfile == nil {
			return result.Err[*domainconfig.Config](fmt.Errorf("failed to map profile: %s", name))
		}
		domain.Profiles[name] = domainProfile
	}

	return result.Ok(domain)
}

// mapPublicProfileToDomain maps public API profile to domain profile
func mapPublicProfileToDomain(public *PublicProfile) *domainconfig.Profile {
	if public == nil {
		return nil
	}

	domain := &domainconfig.Profile{
		Name:        public.Name,
		Description: public.Description,
		Status:      public.Status,
		Operations:  make([]domainconfig.CleanupOperation, len(public.Operations)),
	}

	// Map operations
	for i, op := range public.Operations {
		domain.Operations[i] = mapPublicOperationToDomain(op)
	}

	return domain
}

// mapPublicOperationToDomain maps public API operation to domain operation
func mapPublicOperationToDomain(public PublicOperation) domainconfig.CleanupOperation {
	domainRisk := shared.RiskLevelLowType
	switch public.RiskLevel {
	case PublicRiskLow:
		domainRisk = shared.RiskLevelLowType
	case PublicRiskMedium:
		domainRisk = shared.RiskLevelMediumType
	case PublicRiskHigh:
		domainRisk = shared.RiskLevelHighType
	case PublicRiskCritical:
		domainRisk = shared.RiskLevelCriticalType
	}

	return domainconfig.CleanupOperation{
		Name:        public.Name,
		Description: public.Description,
		RiskLevel:   domainRisk,
		Status:      public.Status,
		Settings: &shared.OperationSettings{
			ExecutionMode:       public.Settings.ExecutionMode,
			Verbose:             public.Settings.Verbose,
			TimeoutSeconds:      int(public.Settings.TimeoutSeconds),
			ConfirmBeforeDelete: public.Settings.ConfirmBeforeDelete,
		},
	}
}
