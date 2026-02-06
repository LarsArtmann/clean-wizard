package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// maxRiskLevelFromOperations calculates the maximum risk level from a slice of operations.
// This function is used across validation middleware and business logic components.
func maxRiskLevelFromOperations(operations []domain.CleanupOperation, currentMax domain.RiskLevel) domain.RiskLevel {
	maxRisk := currentMax
	for _, op := range operations {
		if op.RiskLevel == domain.RiskCritical {
			return domain.RiskCritical
		}
		if op.RiskLevel == domain.RiskHigh {
			maxRisk = domain.RiskHigh
		} else if op.RiskLevel == domain.RiskMedium && maxRisk == domain.RiskLow {
			maxRisk = domain.RiskMedium
		}
	}
	return maxRisk
}