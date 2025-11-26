package api

import (
	domainconfig "github.com/LarsArtmann/clean-wizard/internal/domain/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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
