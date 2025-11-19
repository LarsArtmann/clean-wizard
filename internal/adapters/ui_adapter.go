package adapters

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// UIAdapter provides UI-specific transformations of domain objects
// This keeps domain layer pure of UI concerns like emojis and display formatting
type UIAdapter struct{}

// NewUIAdapter creates a new UI adapter
func NewUIAdapter() *UIAdapter {
	return &UIAdapter{}
}

// RiskLevelIcon returns the appropriate emoji icon for a risk level
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) RiskLevelIcon(risk domain.RiskLevelType) string {
	switch risk {
	case domain.RiskLevelLowType:
		return "🟢"
	case domain.RiskLevelMediumType:
		return "🟡"
	case domain.RiskLevelHighType:
		return "🟠"
	case domain.RiskLevelCriticalType:
		return "🔴"
	default:
		return "⚪"
	}
}

// CleanStrategyIcon returns the appropriate emoji icon for a clean strategy
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) CleanStrategyIcon(strategy domain.CleanStrategyType) string {
	switch strategy {
	case domain.StrategyAggressiveType:
		return "🔥"
	case domain.StrategyConservativeType:
		return "🛡️"
	case domain.StrategyDryRunType:
		return "🔍"
	default:
		return "❓"
	}
}

// RiskLevelColor returns CSS color for risk level
func (ui *UIAdapter) RiskLevelColor(risk domain.RiskLevelType) string {
	switch risk {
	case domain.RiskLevelLowType:
		return "#22c55e" // green
	case domain.RiskLevelMediumType:
		return "#eab308" // yellow
	case domain.RiskLevelHighType:
		return "#f97316" // orange
	case domain.RiskLevelCriticalType:
		return "#ef4444" // red
	default:
		return "#6b7280" // gray
	}
}

// ScanTypeIcon returns appropriate emoji icon for a scan type
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) ScanTypeIcon(scanType domain.ScanTypeType) string {
	switch scanType {
	case domain.ScanTypeNixStoreType:
		return "📦"
	case domain.ScanTypeHomebrewType:
		return "🍺"
	case domain.ScanTypeSystemType:
		return "💻"
	case domain.ScanTypeTempType:
		return "🗑️"
	default:
		return "❓"
	}
}

// ScanTypeDescription returns human-readable description for scan type
func (ui *UIAdapter) ScanTypeDescription(scanType domain.ScanTypeType) string {
	switch scanType {
	case domain.ScanTypeNixStoreType:
		return "Nix store garbage collection and cleanup"
	case domain.ScanTypeHomebrewType:
		return "Homebrew package cleanup and maintenance"
	case domain.ScanTypeSystemType:
		return "System-level temporary files cleanup"
	case domain.ScanTypeTempType:
		return "Temporary files and cache cleanup"
	default:
		return "Unknown scan type"
	}
}

// CleanStrategyDescription returns human-readable description for strategy
func (ui *UIAdapter) CleanStrategyDescription(strategy domain.CleanStrategyType) string {
	switch strategy {
	case domain.StrategyAggressiveType:
		return "Aggressive cleanup with maximum disk space recovery"
	case domain.StrategyConservativeType:
		return "Conservative cleanup with safety-first approach"
	case domain.StrategyDryRunType:
		return "Preview mode - shows what would be cleaned without making changes"
	default:
		return "Unknown cleaning strategy"
	}
}