package config

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/types"
)

// maxRiskLevelFromOperations calculates the maximum risk level from a slice of operations.
// This function is used across validation middleware and business logic components.
func maxRiskLevelFromOperations(
	operations []types.CleanupOperation, currentMax enums.RiskLevelType,
) enums.RiskLevelType {
	maxRisk := currentMax

	for _, op := range operations {
		if op.RiskLevel == enums.RiskLevelCriticalType {
			return enums.RiskLevelCriticalType
		}

		if op.RiskLevel == enums.RiskLevelHighType {
			maxRisk = enums.RiskLevelHighType
		} else if op.RiskLevel == enums.RiskLevelMediumType && maxRisk == enums.RiskLevelLowType {
			maxRisk = enums.RiskLevelMediumType
		}
	}

	return maxRisk
}
