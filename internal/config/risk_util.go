package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// maxRiskLevelFromOperations calculates the maximum risk level from a slice of operations.
// This function is used across validation middleware and business logic components.
func maxRiskLevelFromOperations(
	operations []domain.CleanupOperation, currentMax domain.RiskLevelType,
) domain.RiskLevelType {
	maxRisk := currentMax

	for _, op := range operations {
		if op.RiskLevel == domain.RiskLevelCriticalType {
			return domain.RiskLevelCriticalType
		}

		if op.RiskLevel == domain.RiskLevelHighType {
			maxRisk = domain.RiskLevelHighType
		} else if op.RiskLevel == domain.RiskLevelMediumType && maxRisk == domain.RiskLevelLowType {
			maxRisk = domain.RiskLevelMediumType
		}
	}

	return maxRisk
}
