package config

// Configuration constants for Clean Wizard
// Centralized to eliminate magic numbers and ensure consistency

// Disk usage constraints
const (
	MinDiskUsagePercent     = 10 // Minimum allowed disk usage percentage
	MaxDiskUsagePercent     = 95 // Maximum allowed disk usage percentage
	DefaultDiskUsagePercent = 50 // Default disk usage percentage
	RoundingIncrement       = 10 // Round percentages to nearest increment
)

// Retry policy constants
const (
	DefaultMaxRetries    = 3    // Default maximum retry attempts
	DefaultInitialDelay  = 100  // Default initial delay in milliseconds
	DefaultMaxDelay      = 5000 // Default maximum delay in milliseconds
	DefaultBackoffFactor = 2.0  // Default exponential backoff factor
)

// Nix store constants
const (
	MockStoreSizeGB      = 300 // Mock Nix store size in GB
	MaxGenerations       = 10  // Maximum number of generations to keep
	MockGenerationSizeMB = 50  // Mock generation size in MB
)

// Validation constants
const (
	MinProtectedPaths = 1   // Minimum number of protected paths required
	MaxProfiles       = 50  // Maximum number of profiles allowed
	MaxOperations     = 100 // Maximum number of operations per profile
)

// Configuration key constants to eliminate magic strings
const (
	// Top-level configuration keys
	ConfigKeySafetyLevel = "safety_level"
	ConfigKeySafeMode    = "safe_mode"
	ConfigKeyVersion      = "version"
	ConfigKeyMaxDiskUsage = "max_disk_usage_percent"
	ConfigKeyProtected    = "protected"
	ConfigKeyLastClean    = "last_clean"
	ConfigKeyUpdated     = "updated"
	ConfigKeyProfiles    = "profiles"
	
	// Profile field key templates
	ProfileTemplateName     = "profiles.%s.name"
	ProfileTemplateDesc     = "profiles.%s.description"
	ProfileTemplateStatus  = "profiles.%s.status"
	ProfileTemplateEnabled = "profiles.%s.enabled"
	
	// Operation field key templates
	OperationTemplateName        = "profiles.%s.operations.%d.name"
	OperationTemplateDesc        = "profiles.%s.operations.%d.description"
	OperationTemplateRiskLevel   = "profiles.%s.operations.%d.risk_level"
	OperationTemplateStatus      = "profiles.%s.operations.%d.status"
	OperationTemplateEnabled     = "profiles.%s.operations.%d.enabled"
	OperationTemplateSettings    = "profiles.%s.operations.%d.settings"
)

// Default configuration values
const (
	DefaultSafetyLevel   = 1  // SafetyLevelEnabled
	DefaultSafeMode     = true
	DefaultMaxDiskUsage  = 50
	DefaultVersion      = "1.0.0"
	
	// Default protected paths
	DefaultProtectedPathSystem     = "/System"
	DefaultProtectedPathLibrary    = "/Library" 
	DefaultProtectedPathApplications = "/Applications"
)
