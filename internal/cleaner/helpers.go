package cleaner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DryRunBytesPerItem is the estimated bytes freed per item in dry run mode.
const DryRunBytesPerItem = 300 * 1024 * 1024 // 300MB per item

// CleanItemFunc is a function that cleans a single item of type T.
type CleanItemFunc[T any] func(ctx context.Context, item T, homeDir string) result.Result[domain.CleanResult]

// AvailableCheckFunc is a function that checks if the cleaner is available.
type AvailableCheckFunc func(ctx context.Context) bool

// cleanWithIterator is a shared helper function that performs the common clean pattern.
// It iterates over items, calls the cleanFunc for each, and aggregates results.
func cleanWithIterator[T any](
	ctx context.Context,
	cleanerName string,
	availableCheck AvailableCheckFunc,
	items []T,
	cleanFunc CleanItemFunc[T],
	verbose bool,
	dryRun bool,
) result.Result[domain.CleanResult] {
	if !availableCheck(ctx) {
		return result.Err[domain.CleanResult](fmt.Errorf("%s not available", cleanerName))
	}

	if dryRun {
		totalBytes := int64(len(items)) * DryRunBytesPerItem
		cleanResult := conversions.NewCleanResult(domain.StrategyDryRun, len(items), totalBytes)
		return result.Ok(cleanResult)
	}

	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[domain.CleanResult](fmt.Errorf("failed to get home directory: %w", err))
	}

	for _, item := range items {
		result := cleanFunc(ctx, item, homeDir)
		if result.IsErr() {
			itemsFailed++
			if verbose {
				fmt.Printf("Warning: failed to clean %v: %v\n", item, result.Error())
			}
			continue
		}

		cleanResult := result.Value()
		itemsRemoved++
		bytesFreed += int64(cleanResult.FreedBytes)
	}

	duration := time.Since(startTime)
	cleanResult := domain.CleanResult{
		FreedBytes:   uint64(bytesFreed),
		ItemsRemoved: uint(itemsRemoved),
		ItemsFailed:  uint(itemsFailed),
		CleanTime:    duration,
		CleanedAt:    time.Now(),
		Strategy:     domain.StrategyConservative,
	}

	return result.Ok(cleanResult)
}

// ValidateToolTypes validates configured tool types against a set of available types.
// This eliminates duplicate validation code across different cleaner implementations.
func ValidateToolTypes(
	configuredTypes []string,
	availableTypes []string,
	typeName string,
) error {
	// Build map of valid types for O(1) lookup
	validTypes := make(map[string]bool, len(availableTypes))
	for _, t := range availableTypes {
		validTypes[t] = true
	}

	// Validate each configured type
	for _, t := range configuredTypes {
		if !validTypes[t] {
			return fmt.Errorf("invalid %s type: %s", typeName, t)
		}
	}

	return nil
}

// ValidateSettingsWithTypes validates settings that have a types slice field.
// This is a generic helper to eliminate duplication across similar validators.
// The getter function extracts the types slice from the settings struct.
func ValidateSettingsWithTypes[S any](
	settings S,
	getSlice func(S) []string,
	availableTypes []string,
	typeName string,
) error {
	slice := getSlice(settings)
	return ValidateToolTypes(slice, availableTypes, typeName)
}

// ValidateOptionalSettingsWithTypes validates optional settings with type validation.
// It handles the common pattern of checking if a field is nil, and if not,
// validating its types against available types.
func ValidateOptionalSettingsWithTypes[F any](
	settings *domain.OperationSettings,
	getField func(*domain.OperationSettings) *F,
	getSlice func(*F) []string,
	availableTypes []string,
	typeName string,
) error {
	if settings == nil {
		return nil
	}
	field := getField(settings)
	if field == nil {
		return nil
	}
	return ValidateToolTypes(getSlice(field), availableTypes, typeName)
}

// LangVersionManagerAvailableTypes defines all valid language version manager types.
var LangVersionManagerAvailableTypes = []string{
	domain.VersionManagerNvm.String(),
	domain.VersionManagerPyenv.String(),
	domain.VersionManagerGvm.String(),
	domain.VersionManagerRbenv.String(),
	domain.VersionManagerSdkman.String(),
	domain.VersionManagerJenv.String(),
}

// BuildCacheAvailableTypes defines all valid build cache tool types.
var BuildCacheAvailableTypes = []string{
	domain.BuildToolGo.String(),
	domain.BuildToolRust.String(),
	domain.BuildToolNode.String(),
	domain.BuildToolPython.String(),
	domain.BuildToolJava.String(),
	domain.BuildToolScala.String(),
}

// VersionManagerTypeToStringSlice converts domain.VersionManagerType slice to string slice.
func VersionManagerTypeToStringSlice(types []domain.VersionManagerType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}
	return result
}

// BuildToolTypeToStringSlice converts domain.BuildToolType slice to string slice.
func BuildToolTypeToStringSlice(types []domain.BuildToolType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}
	return result
}

// PackageManagerTypeToStringSlice converts domain.PackageManagerType slice to string slice.
func PackageManagerTypeToStringSlice(types []domain.PackageManagerType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}
	return result
}

// PackageManagerTypeToLowerSlice converts domain.PackageManagerType slice to lowercase string slice.
func PackageManagerTypeToLowerSlice(types []domain.PackageManagerType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = strings.ToLower(t.String())
	}
	return result
}

// CacheTypeToStringSlice converts domain.CacheType slice to string slice.
func CacheTypeToStringSlice(types []domain.CacheType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}
	return result
}

// CacheTypeToLowerSlice converts domain.CacheType slice to lowercase string slice.
func CacheTypeToLowerSlice(types []domain.CacheType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = strings.ToLower(t.String())
	}
	return result
}

// ValidateLangVersionManagerSettings validates language version manager settings.
func ValidateLangVersionManagerSettings(settings *domain.OperationSettings) error {
	return ValidateOptionalSettingsWithTypes(
		settings,
		func(s *domain.OperationSettings) *domain.LangVersionManagerSettings { return s.LangVersionManager },
		func(f *domain.LangVersionManagerSettings) []string { return VersionManagerTypeToStringSlice(f.ManagerTypes) },
		LangVersionManagerAvailableTypes,
		"manager",
	)
}

// ValidateBuildCacheSettings validates build cache settings.
func ValidateBuildCacheSettings(settings *domain.OperationSettings) error {
	return ValidateOptionalSettingsWithTypes(
		settings,
		func(s *domain.OperationSettings) *domain.BuildCacheSettings { return s.BuildCache },
		func(f *domain.BuildCacheSettings) []string { return BuildToolTypeToStringSlice(f.ToolTypes) },
		BuildCacheAvailableTypes,
		"tool",
	)
}

// ScanItemFunc is a function that scans for items of type T and returns scan results.
type ScanItemFunc[T any] func(ctx context.Context, item T, homeDir string) result.Result[[]domain.ScanItem]

// scanWithIterator is a shared helper function that performs the common scan pattern.
// It iterates over types, calls the scanFunc for each, and aggregates results.
func scanWithIterator[T any](
	ctx context.Context,
	types []T,
	scanFunc ScanItemFunc[T],
	verbose bool,
) result.Result[[]domain.ScanItem] {
	items := make([]domain.ScanItem, 0)

	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[[]domain.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
	}

	for _, item := range types {
		result := scanFunc(ctx, item, homeDir)
		if result.IsErr() {
			if verbose {
				fmt.Printf("Warning: failed to scan %v: %v\n", item, result.Error())
			}
			continue
		}

		items = append(items, result.Value()...)
	}

	return result.Ok(items)
}
