package types

import (
	"context"
	"time"
)

// DEPRECATED: Use internal/domain package instead
// TODO: Migrate all consumers to domain.ScanResult
// TODO: Remove this entire file once migration complete
type LegacyScanResult struct {
	Name        string  `json:"name" yaml:"name"`
	SizeGB      float64 `json:"size_gb" yaml:"size_gb"`
	Description string  `json:"description" yaml:"description"`
	Cleanable   bool    `json:"cleanable" yaml:"cleanable"`
}

// DEPRECATED: Use internal/domain package instead  
type LegacyScanResults struct {
	TotalSizeGB float64        `json:"total_size_gb" yaml:"total_size_gb"`
	Results     []LegacyScanResult `json:"results" yaml:"results"`
	Timestamp   time.Time       `json:"timestamp" yaml:"timestamp"`
}

// DEPRECATED: Use domain.CleanRequest instead
type CleanupOperation struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	RiskLevel   RiskLevel      `json:"risk_level" yaml:"risk_level"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Settings    map[string]any `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// DEPRECATED: Use domain package instead
type Profile struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Operations  []CleanupOperation `json:"operations" yaml:"operations"`
	CreatedAt   time.Time          `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" yaml:"updated_at"`
}

// DEPRECATED: Use domain configuration pattern
type Config struct {
	Version      string             `json:"version" yaml:"version"`
	SafeMode     bool               `json:"safe_mode" yaml:"safe_mode"`
	DryRun       bool               `json:"dry_run" yaml:"dry_run"`
	Verbose      bool               `json:"verbose" yaml:"verbose"`
	Backup       bool               `json:"backup" yaml:"backup"`
	MaxDiskUsage int                `json:"max_disk_usage_percent" yaml:"max_disk_usage_percent"`
	Protected    []string           `json:"protected_paths" yaml:"protected_paths"`
	Profiles     map[string]Profile `json:"profiles" yaml:"profiles"`
}

// DEPRECATED: Use domain.CleanResult instead
// TODO: This is a SPLIT BRAIN pattern! Success field + ErrorMessage should be Result[T]
type CleanupResult struct {
	Operation   string        `json:"operation" yaml:"operation"`
	Success     bool          `json:"success" yaml:"success"` // SPLIT BRAIN!
	SizeFreedGB float64       `json:"size_freed_gb" yaml:"size_freed_gb"`
	Error       error         `json:"error,omitempty" yaml:"error,omitempty"` // SPLIT BRAIN!
	Duration    time.Duration `json:"duration" yaml:"duration"`
}

// DEPRECATED: Use domain package instead
type CleanupResults struct {
	TotalSizeFreedGB float64         `json:"total_size_freed_gb" yaml:"total_size_freed_gb"`
	Results          []CleanupResult `json:"results" yaml:"results"`
	StartTime        time.Time       `json:"start_time" yaml:"start_time"`
	EndTime          time.Time       `json:"end_time" yaml:"end_time"`
	Duration         time.Duration   `json:"duration" yaml:"duration"`
}

// TODO: REMOVE DEPRECATED PATTERNS!
// These interfaces should be replaced by domain interfaces

// DEPRECATED: Use domain.Scanner instead
type Scanner interface {
	Scan(ctx context.Context) (*LegacyScanResults, error)
	Name() string
}

// DEPRECATED: Use domain.Cleaner instead
type Cleaner interface {
	Clean(ctx context.Context, operation CleanupOperation) (*CleanupResult, error)
	Name() string
	RiskLevel() RiskLevel
}