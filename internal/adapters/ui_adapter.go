package adapters

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// UIAdapter provides UI-related utility functions
type UIAdapter struct{}

// NewUIAdapter creates a new UI adapter
func NewUIAdapter() *UIAdapter {
	return &UIAdapter{}
}

// RiskLevelIcon returns the icon for a given risk level
func (u *UIAdapter) RiskLevelIcon(level shared.RiskLevel) string {
	switch level {
	case shared.RiskLow:
		return "🟢"
	case shared.RiskMedium:
		return "🟡"
	case shared.RiskHigh:
		return "🟠"
	case shared.RiskCritical:
		return "🔴"
	default:
		return "⚪"
	}
}
