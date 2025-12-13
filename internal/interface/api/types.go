package api

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
)

// Type definitions with proper type safety - DOMAIN DRIVEN DESIGN
// TODO: Generate from TypeSpec compiler in production

type PublicConfig struct {
	Version        string                    `json:"version"`
	SafetyLevel    shared.SafetyLevelType    `json:"safetyLevel"` // FIXED: bool → enum - invalid states UNREPRESENTABLE!
	MaxDiskUsage   int32                     `json:"maxDiskUsage"`
	ProtectedPaths []string                  `json:"protectedPaths"`
	Profiles       map[string]*PublicProfile `json:"profiles"`
}

type PublicProfile struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      shared.StatusType `json:"status"` // FIXED: bool → enum - invalid states UNREPRESENTABLE!
	Operations  []PublicOperation `json:"operations"`
}

type PublicOperation struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	RiskLevel   PublicRiskLevel   `json:"riskLevel"`
	Status      shared.StatusType `json:"status"` // FIXED: bool → enum - invalid states UNREPRESENTABLE!
	Settings    OperationSettings `json:"settings"`
}

type OperationSettings struct {
	ExecutionMode       shared.ExecutionModeType `json:"executionMode"` // FIXED: bool → enum - invalid states UNREPRESENTABLE!
	Verbose             bool                     `json:"verbose"`
	TimeoutSeconds      int32                    `json:"timeoutSeconds"`
	ConfirmBeforeDelete bool                     `json:"confirmBeforeDelete"`
}

type PublicCleanResult struct {
	Success      bool           `json:"success"`
	FreedBytes   uint64         `json:"freedBytes"`
	ItemsRemoved uint32         `json:"itemsRemoved"`
	ItemsFailed  uint32         `json:"itemsFailed"`
	CleanTime    string         `json:"cleanTime"` // ISO 8601 duration
	CleanedAt    string         `json:"cleanedAt"` // ISO 8601 timestamp
	Strategy     PublicStrategy `json:"strategy"`
	Errors       []string       `json:"errors,omitempty"`
}

type PublicRiskLevel string

const (
	PublicRiskLow      PublicRiskLevel = "LOW"
	PublicRiskMedium   PublicRiskLevel = "MEDIUM"
	PublicRiskHigh     PublicRiskLevel = "HIGH"
	PublicRiskCritical PublicRiskLevel = "CRITICAL"
)

type PublicStrategy string

const (
	PublicStrategyAggressive   PublicStrategy = "aggressive"
	PublicStrategyConservative PublicStrategy = "conservative"
	PublicStrategyDryRun       PublicStrategy = "dry-run"
)

// OperationType represents types of cleanup operations
type OperationType string

const (
	OperationTypeNixGenerations OperationType = "nix-generations"
	OperationTypeTempFiles      OperationType = "temp-files"
	OperationTypeLogFiles       OperationType = "log-files"
	OperationTypeCacheFiles     OperationType = "cache-files"
)

// CleanRequest represents a cleaning request with type-safe operations
type CleanRequest struct {
	Config     PublicConfig    `json:"config"`
	Strategy   PublicStrategy  `json:"strategy"`
	Operations []OperationType `json:"operations"`
	DryRun     *bool           `json:"dryRun,omitempty"`
}

// PublicValidationResult represents validation result
type PublicValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings,omitempty"`
}

// PublicScanResult represents scan result
type PublicScanResult struct {
	Success   bool             `json:"success"`
	ScanTime  string           `json:"scanTime"`
	TotalSize uint64           `json:"totalSize"`
	ItemCount uint32           `json:"itemCount"`
	Items     []PublicScanItem `json:"items"`
	Errors    []string         `json:"errors,omitempty"`
}

// PublicScanItem represents a single scanned item
type PublicScanItem struct {
	Path     string `json:"path"`
	Size     uint64 `json:"size"`
	Type     string `json:"type"`
	Created  string `json:"created,omitempty"`
	Modified string `json:"modified,omitempty"`
}
