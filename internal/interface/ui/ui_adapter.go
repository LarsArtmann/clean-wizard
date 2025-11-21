package adapters

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
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

// StatusIcon returns appropriate emoji icon for status
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) StatusIcon(status domain.StatusType) string {
	switch status {
	case domain.StatusDisabled:
		return "🔴"
	case domain.StatusEnabled:
		return "🟢"
	case domain.StatusInherited:
		return "🔵"
	default:
		return "⚪"
	}
}

// EnforcementLevelIcon returns appropriate emoji icon for enforcement level
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) EnforcementLevelIcon(level domain.EnforcementLevelType) string {
	switch level {
	case domain.EnforcementLevelNone:
		return "⚪"
	case domain.EnforcementLevelWarning:
		return "🟡"
	case domain.EnforcementLevelError:
		return "🔴"
	case domain.EnforcementLevelStrict:
		return "🚫"
	default:
		return "❓"
	}
}

// SelectedStatusIcon returns appropriate emoji icon for selected status
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) SelectedStatusIcon(status domain.SelectedStatusType) string {
	switch status {
	case domain.SelectedStatusNotSelected:
		return "⭕"
	case domain.SelectedStatusSelected:
		return "✅"
	case domain.SelectedStatusDefault:
		return "🌟"
	default:
		return "❓"
	}
}

// RecursionLevelIcon returns appropriate emoji icon for recursion level
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) RecursionLevelIcon(level domain.RecursionLevelType) string {
	switch level {
	case domain.RecursionLevelNone:
		return "➡️"
	case domain.RecursionLevelDirect:
		return "⬇️"
	case domain.RecursionLevelFull:
		return "🔄"
	case domain.RecursionLevelInfinite:
		return "♾️"
	default:
		return "❓"
	}
}

// OptimizationLevelIcon returns appropriate emoji icon for optimization level
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) OptimizationLevelIcon(level domain.OptimizationLevelType) string {
	switch level {
	case domain.OptimizationLevelNone:
		return "⚪"
	case domain.OptimizationLevelConservative:
		return "🟡"
	case domain.OptimizationLevelAggressive:
		return "🔴"
	default:
		return "❓"
	}
}

// FileSelectionStrategyIcon returns appropriate emoji icon for file selection strategy
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) FileSelectionStrategyIcon(strategy domain.FileSelectionStrategyType) string {
	switch strategy {
	case domain.FileSelectionStrategyAll:
		return "📁"
	case domain.FileSelectionStrategyUnusedOnly:
		return "🗑️"
	case domain.FileSelectionStrategyManual:
		return "✏️"
	default:
		return "❓"
	}
}

// SafetyLevelIcon returns appropriate emoji icon for safety level
// UI CONCERN: Properly separated from domain layer
func (ui *UIAdapter) SafetyLevelIcon(level domain.SafetyLevelType) string {
	switch level {
	case domain.SafetyLevelDisabled:
		return "🔴"
	case domain.SafetyLevelEnabled:
		return "🟢"
	case domain.SafetyLevelStrict:
		return "🟡"
	case domain.SafetyLevelParanoid:
		return "🚫"
	default:
		return "❓"
	}
}
