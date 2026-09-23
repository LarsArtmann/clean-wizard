package cleaner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/conversions"
	"github.com/LarsArtmann/clean-wizard/internal/domain/enums"
	"github.com/LarsArtmann/clean-wizard/internal/domain/operations"
	domaintypes "github.com/LarsArtmann/clean-wizard/internal/domain/types"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

const (
	// DryRunBytesPerItem is the estimated bytes freed per item in dry run mode when no size estimator is provided.
	DryRunBytesPerItem = 300 * 1024 * 1024 // 300MB per item
	// TrashPathTimeout is the timeout for trash operations.
	TrashPathTimeout = 30 * time.Second
)

// SizeEstimatorFunc is a function that estimates the size of an item for dry-run mode.
type SizeEstimatorFunc[T any] func(item T) int64

// CleanItemFunc is a function that cleans a single item of type T.
type CleanItemFunc[T any] func(ctx context.Context, item T, homeDir string) result.Result[domaintypes.CleanResult]

// AvailableCheckFunc is a function that checks if the cleaner is available.
type AvailableCheckFunc func(ctx context.Context) bool

// CleanWithIterator is a shared helper function that performs the common clean pattern.
// It iterates over items, calls the cleanFunc for each, and aggregates results.
// If sizeEstimator is provided, it uses that for dry-run estimates; otherwise uses DryRunBytesPerItem.
func CleanWithIterator[T any](
	ctx context.Context,
	cleanerName string,
	availableCheck AvailableCheckFunc,
	items []T,
	cleanFunc CleanItemFunc[T],
	verbose bool,
	dryRun bool,
	sizeEstimator SizeEstimatorFunc[T],
) result.Result[domaintypes.CleanResult] {
	if !availableCheck(ctx) {
		return result.Err[domaintypes.CleanResult](NewNotAvailableError(cleanerName, ""))
	}

	if dryRun {
		var totalBytes int64

		if sizeEstimator != nil {
			for _, item := range items {
				totalBytes += sizeEstimator(item)
			}
		} else {
			totalBytes = int64(len(items)) * DryRunBytesPerItem
		}

		cleanResult := conversions.NewCleanResult(
			enums.StrategyDryRunType,
			len(items),
			totalBytes,
		)

		return result.Ok(cleanResult)
	}

	startTime := time.Now()
	itemsRemoved := 0
	itemsFailed := 0
	bytesFreed := int64(0)

	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[domaintypes.CleanResult](
			fmt.Errorf("failed to get home directory for %s: %w", cleanerName, err),
		)
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

	return result.Ok(conversions.NewCleanResultWithFailures(
		enums.StrategyConservativeType,
		itemsRemoved,
		itemsFailed,
		bytesFreed,
		duration,
	))
}

// ValidateToolTypes validates configured tool types against a set of available domaintypes.
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

// ValidateOptionalSettings validates the typed subfield of OperationSettings
// when both settings and the subfield are non-nil, otherwise returns nil.
// This eliminates the repeated `if settings == nil || settings.X == nil { return nil }`
// boilerplate across every cleaner's ValidateSettings method.
func ValidateOptionalSettings[T any](
	settings *operations.OperationSettings,
	getField func(*operations.OperationSettings) *T,
	validate func(*T) error,
) error {
	if settings == nil {
		return nil
	}

	field := getField(settings)
	if field == nil {
		return nil
	}

	return validate(field)
}

// ValidateOptionalSettingsWithTypes validates optional settings with type validation.
// It handles the common pattern of checking if a field is nil, and if not,
// validating its types against available domaintypes.
func ValidateOptionalSettingsWithTypes[F any](
	settings *operations.OperationSettings,
	getField func(*operations.OperationSettings) *F,
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

// BuildCacheAvailableTypes defines all valid build cache tool domaintypes.
var BuildCacheAvailableTypes = []string{ //nolint:gochecknoglobals
	enums.BuildToolGo.String(),
	enums.BuildToolRust.String(),
	enums.BuildToolNode.String(),
	enums.BuildToolPython.String(),
	enums.BuildToolJava.String(),
	enums.BuildToolScala.String(),
}

type stringer interface {
	String() string
}

// toStringSlice converts a slice of types implementing Stringer to string slice.
func toStringSlice[T stringer](types []T) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = t.String()
	}

	return result
}

// toLowerStringSlice converts a slice of types implementing Stringer to lowercase string slice.
func toLowerStringSlice[T stringer](types []T) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = strings.ToLower(t.String())
	}

	return result
}

// BuildToolTypeToStringSlice converts enums.BuildToolType slice to string slice.
func BuildToolTypeToStringSlice(types []enums.BuildToolType) []string {
	return toStringSlice(types)
}

// PackageManagerTypeToStringSlice converts enums.PackageManagerType slice to string slice.
func PackageManagerTypeToStringSlice(types []enums.PackageManagerType) []string {
	return toStringSlice(types)
}

// PackageManagerTypeToLowerSlice converts enums.PackageManagerType slice to lowercase string slice.
func PackageManagerTypeToLowerSlice(types []enums.PackageManagerType) []string {
	return toLowerStringSlice(types)
}

// CacheTypeToStringSlice converts enums.CacheType slice to string slice.
func CacheTypeToStringSlice(types []enums.CacheType) []string {
	return toStringSlice(types)
}

// CacheTypeToLowerSlice converts enums.CacheType slice to lowercase string slice.
func CacheTypeToLowerSlice(types []enums.CacheType) []string {
	return toLowerStringSlice(types)
}

// ValidateBuildCacheSettings validates build cache settings.
func ValidateBuildCacheSettings(settings *operations.OperationSettings) error {
	return ValidateOptionalSettingsWithTypes(
		settings,
		func(s *operations.OperationSettings) *operations.BuildCacheSettings { return s.BuildCache },
		func(f *operations.BuildCacheSettings) []string { return BuildToolTypeToStringSlice(f.ToolTypes) },
		BuildCacheAvailableTypes,
		"tool",
	)
}

// ScanItemFunc is a function that scans for items of type T and returns scan results.
type ScanItemFunc[T any] func(ctx context.Context, item T, homeDir string) result.Result[[]domaintypes.ScanItem]

// ScanWithIterator is a shared helper function that performs the common scan pattern.
// It iterates over types, calls the scanFunc for each, and aggregates results.
func ScanWithIterator[T any](
	ctx context.Context,
	types []T,
	scanFunc ScanItemFunc[T],
	verbose bool,
) result.Result[[]domaintypes.ScanItem] {
	items := make([]domaintypes.ScanItem, 0)

	homeDir, err := GetHomeDir()
	if err != nil {
		return result.Err[[]domaintypes.ScanItem](fmt.Errorf("failed to get home directory: %w", err))
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

// CalculateTotalSizeFromScan calculates the total size from scan results.
// Returns 0 if the scan resulted in an error.
func CalculateTotalSizeFromScan(scanResult result.Result[[]domaintypes.ScanItem]) int64 {
	if scanResult.IsErr() {
		return 0
	}

	var total int64
	for _, item := range scanResult.Value() {
		total += item.Size
	}

	return total
}

// TrashPath moves a file or directory to the system trash using the `trash` command.
// It applies a 30-second timeout to the operation.
// This is a shared helper to eliminate duplicate trash implementations across cleaners.
func TrashPath(ctx context.Context, path string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, TrashPathTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "trash", path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("trash failed for %s: %w (output: %s)", path, err, string(output))
	}

	return nil
}
