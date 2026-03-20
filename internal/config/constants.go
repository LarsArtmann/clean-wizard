package config

import (
	"time"
)

// Configuration constants for Clean Wizard
// Centralized to eliminate magic numbers and ensure consistency

// Disk usage constraints.
const (
	MinDiskUsagePercent = 10 // Minimum allowed disk usage percentage
	MaxDiskUsagePercent = 95 // Maximum allowed disk usage percentage
	DefaultMaxDiskUsage = 50 // Default disk usage percentage
	RoundingIncrement   = 10 // Round percentages to nearest increment
)

// Retry policy constants.
const (
	DefaultMaxRetries        = 3 // Default maximum retry attempts
	DefaultInitialRetryDelay = 100 * time.Millisecond
	DefaultMaxRetryDelay     = 5 * time.Second
	DefaultBackoffFactor     = 2.0 // Default exponential backoff factor
)

// Load timeout constants.
const (
	DefaultLoadTimeout = 30 * time.Second // Default timeout for configuration loading
)

// Cache constants.
const (
	DefaultCacheDuration = 30 * time.Minute // Default cache duration
)

// Nix generation constants.
const (
	DefaultNixMaxGenerations = 10 // Default maximum generations to keep
	MaxNixGenerations        = 10 // Maximum allowed generations to keep
	MockNixStoreSizeGB       = 300
	MockNixGenerationSizeMB  = 50
)

// File permission constants.
const (
	ConfigFilePermission = 0o644 // rw-r--r--
	ConfigDirPermission  = 0o755 // rwxr-xr-x
)

// Validation constants.
const (
	MinProtectedPaths      = 1   // Minimum number of protected paths required
	MaxProfiles            = 50  // Maximum number of profiles allowed
	MaxOperations          = 100 // Maximum number of operations per profile
	MaxProfileCountWarning = 20  // Warning threshold for profile count
	SchemaMinDiskUsage     = 10.0
	SchemaMaxDiskUsage     = 95.0
)
