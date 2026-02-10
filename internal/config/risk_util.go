package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// maxRiskLevelFromOperations calculates the maximum risk level from a slice of operations.
// This function is used across validation middleware and business logic components.
func maxRiskLevelFromOperations(operations []domain.CleanupOperation, currentMax domain.RiskLevelType) domain.RiskLevelType {
	maxRisk := currentMax
	for _, op := range operations {
		if op.RiskLevel == domain.RiskLevelType(domain.RiskLevelCriticalType) {
			return domain.RiskLevelType(domain.RiskLevelCriticalType)
		}
		if op.RiskLevel == domain.RiskLevelType(domain.RiskLevelHighType) {
			maxRisk = domain.RiskLevelType(domain.RiskLevelHighType)
		} else if op.RiskLevel == domain.RiskLevelType(domain.RiskLevelMediumType) && maxRisk == domain.RiskLevelType(domain.RiskLevelLowType) {
			maxRisk = domain.RiskLevelType(domain.RiskLevelMediumType)
		}
	}
	return maxRisk
}
