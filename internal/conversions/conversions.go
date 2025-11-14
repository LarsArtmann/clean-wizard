package conversions

import (
	"fmt"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// DOMAIN BUILDERS - Single source of truth for domain object construction

// NewCleanResult creates a valid CleanResult with basic strategy and metrics.
//
// Parameters:
//   - strategy: The cleaning strategy used (e.g., "NIX_GC", "DRY RUN")
//   - itemsRemoved: Number of items successfully removed
//   - freedBytes: Total bytes freed by the operation
//
// Returns:
//   - domain.CleanResult: A fully initialized CleanResult with current timestamp
//
// Example:
//
//	result := NewCleanResult("NIX_GC", 5, 1024*1024*100)
//	fmt.Printf("Freed %d bytes", result.FreedBytes)
func NewCleanResult(strategy string, itemsRemoved int, freedBytes int64) domain.CleanResult {
	return domain.CleanResult{
		FreedBytes:   freedBytes,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  0,
		CleanTime:    0, // Will be set by caller
		CleanedAt:    time.Now(),
		Strategy:     strategy,
	}
}

// NewCleanResultWithTiming creates a CleanResult with custom timing information.
//
// Use this function when you have precise timing measurements for the operation.
// This provides more accurate metrics than the basic NewCleanResult function.
//
// Parameters:
//   - strategy: The cleaning strategy used
//   - itemsRemoved: Number of items successfully removed
//   - freedBytes: Total bytes freed by the operation
//   - cleanTime: Time duration of the cleaning operation
//
// Returns:
//   - domain.CleanResult: A fully initialized CleanResult with timing data
//
// Example:
//
//	startTime := time.Now()
//	// ... perform cleaning ...
//	cleanTime := time.Since(startTime)
//	result := NewCleanResultWithTiming("NIX_GC", 5, 1024*1024*100, cleanTime)
func NewCleanResultWithTiming(strategy string, itemsRemoved int, freedBytes int64, cleanTime time.Duration) domain.CleanResult {
	return domain.CleanResult{
		FreedBytes:   freedBytes,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  0,
		CleanTime:    cleanTime,
		CleanedAt:    time.Now(),
		Strategy:     strategy,
	}
}

// NewCleanResultWithFailures creates a CleanResult with failure tracking for detailed reporting.
//
// Use this function when some operations failed and you need to track both successful
// and failed operations separately. This is useful for complex multi-step cleaning operations.
//
// Parameters:
//   - strategy: The cleaning strategy used
//   - itemsRemoved: Number of items successfully removed
//   - itemsFailed: Number of items that failed to remove
//   - freedBytes: Total bytes freed by successful operations
//   - cleanTime: Time duration of the cleaning operation
//
// Returns:
//   - domain.CleanResult: A fully initialized CleanResult with failure tracking
//
// Example:
//
//	result := NewCleanResultWithFailures("NIX_CLEANUP", 5, 2, 1024*1024*100, time.Second*30)
//	fmt.Printf("Success: %d, Failed: %d", result.ItemsRemoved, result.ItemsFailed)
func NewCleanResultWithFailures(strategy string, itemsRemoved, itemsFailed int, freedBytes int64, cleanTime time.Duration) domain.CleanResult {
	return domain.CleanResult{
		FreedBytes:   freedBytes,
		ItemsRemoved: itemsRemoved,
		ItemsFailed:  itemsFailed,
		CleanTime:    cleanTime,
		CleanedAt:    time.Now(),
		Strategy:     strategy,
	}
}

// NewScanResult creates a valid ScanResult with all required metrics and metadata.
//
// This is the central function for creating scan results throughout the application.
// All scanning operations should use this function to ensure consistent data.
//
// Parameters:
//   - totalBytes: Total bytes found during scanning
//   - totalItems: Total number of items discovered
//   - scannedPaths: List of paths that were scanned
//   - scanDuration: Time taken to perform the scan
//
// Returns:
//   - domain.ScanResult: A fully initialized ScanResult with current timestamp
//
// Example:
//
//	paths := []string{"/nix/store", "/tmp"}
//	result := NewScanResult(1024*1024*500, 1000, paths, time.Second*10)
//	fmt.Printf("Scanned %d items in %v", result.TotalItems, result.ScanTime)
func NewScanResult(totalBytes int64, totalItems int, scannedPaths []string, scanDuration time.Duration) domain.ScanResult {
	return domain.ScanResult{
		TotalBytes:   totalBytes,
		TotalItems:   totalItems,
		ScannedPaths: scannedPaths,
		ScanTime:     scanDuration,
		ScannedAt:    time.Now(),
	}
}

// GENERIC CONVERSION FUNCTIONS - Centralized primitive→domain transformations

// ToCleanResult converts primitive Result[int64] to domain Result[domain.CleanResult] with default strategy.
//
// This is the simplest conversion function that automatically uses "default" strategy.
// Use this when you don't need custom strategy information.
//
// Parameters:
//   - bytesResult: Result[int64] containing bytes freed from primitive operation
//
// Returns:
//   - result.Result[domain.CleanResult]: Converted result with default strategy
//
// Example:
//
//	bytesResult := adapter.GetStoreSize(ctx)
//	cleanResult := ToCleanResult(bytesResult)
//	if cleanResult.IsOk() {
//		fmt.Printf("Freed %d bytes", cleanResult.Value().FreedBytes)
//	}
func ToCleanResult(bytesResult result.Result[int64]) result.Result[domain.CleanResult] {
	return ToCleanResultWithStrategy(bytesResult, "default")
}

// ToCleanResultWithStrategy converts primitive Result[int64] to domain.Result[domain.CleanResult] with custom strategy.
//
// Use this function when you need to specify the cleaning strategy used.
// This provides more detailed tracking of which operation type was performed.
//
// Parameters:
//   - bytesResult: Result[int64] containing bytes freed from primitive operation
//   - strategy: String identifier for the cleaning strategy (e.g., "NIX_GC", "REMOVE_GENERATION")
//
// Returns:
//   - result.Result[domain.CleanResult]: Converted result with specified strategy
//
// Example:
//
//	bytesResult := adapter.CollectGarbage(ctx)
//	cleanResult := ToCleanResultWithStrategy(bytesResult, "NIX_GC")
func ToCleanResultWithStrategy(bytesResult result.Result[int64], strategy string) result.Result[domain.CleanResult] {
	if bytesResult.IsErr() {
		return result.Err[domain.CleanResult](bytesResult.Error())
	}

	bytes := bytesResult.Value()
	cleanResult := NewCleanResult(strategy, 1, bytes)
	return result.Ok(cleanResult)
}

// ToCleanResultFromItems converts items count and bytes to domain Result[domain.CleanResult].
//
// Use this function when you have both the number of items removed and the bytes freed.
// This provides more detailed metrics than just bytes conversion alone.
//
// Parameters:
//   - itemsRemoved: Number of items successfully removed
//   - bytesResult: Result[int64] containing bytes freed from operation
//   - strategy: String identifier for the cleaning strategy
//
// Returns:
//   - result.Result[domain.CleanResult]: Converted result with items and bytes data
//
// Example:
//
//	bytesResult := adapter.CollectGarbage(ctx)
//	cleanResult := ToCleanResultFromItems(5, bytesResult, "NIX_GC")
//	if cleanResult.IsOk() {
//		fmt.Printf("Removed %d items, freed %d bytes",
//			cleanResult.Value().ItemsRemoved, cleanResult.Value().FreedBytes)
//	}
func ToCleanResultFromItems(itemsRemoved int, bytesResult result.Result[int64], strategy string) result.Result[domain.CleanResult] {
	if bytesResult.IsErr() {
		return result.Err[domain.CleanResult](bytesResult.Error())
	}

	bytes := bytesResult.Value()
	cleanResult := NewCleanResult(strategy, itemsRemoved, bytes)
	return result.Ok(cleanResult)
}

// ToTimedCleanResult creates a timed CleanResult from bytes and duration
func ToTimedCleanResult(bytesResult result.Result[int64], strategy string, cleanTime time.Duration) result.Result[domain.CleanResult] {
	if bytesResult.IsErr() {
		return result.Err[domain.CleanResult](bytesResult.Error())
	}

	bytes := bytesResult.Value()
	cleanResult := NewCleanResultWithTiming(strategy, 1, bytes, cleanTime)
	return result.Ok(cleanResult)
}

// ToScanResult converts primitive scanning results to domain.ScanResult
func ToScanResult(totalBytes int64, totalItems int, scannedPaths []string, scanDuration time.Duration) domain.ScanResult {
	return NewScanResult(totalBytes, totalItems, scannedPaths, scanDuration)
}

// UTILITY FUNCTIONS - Helper transformations

// CombineCleanResults combines multiple CleanResults into one
func CombineCleanResults(results []domain.CleanResult) domain.CleanResult {
	if len(results) == 0 {
		return NewCleanResult("combined", 0, 0)
	}

	totalItems := 0
	totalFailed := 0
	totalBytes := int64(0)
	maxTime := time.Duration(0)
	strategies := make([]string, 0, len(results))

	for _, result := range results {
		totalItems += result.ItemsRemoved
		totalFailed += result.ItemsFailed
		totalBytes += result.FreedBytes
		if result.CleanTime > maxTime {
			maxTime = result.CleanTime
		}
		strategies = append(strategies, result.Strategy)
	}

	combinedStrategy := fmt.Sprintf("combined(%v)", strategies)
	return NewCleanResultWithFailures(combinedStrategy, totalItems, totalFailed, totalBytes, maxTime)
}

// ExtractBytesFromCleanResult extracts int64 from domain.CleanResult (for adapter compatibility)
func ExtractBytesFromCleanResult(cleanResult result.Result[domain.CleanResult]) result.Result[int64] {
	if cleanResult.IsErr() {
		return result.Err[int64](cleanResult.Error())
	}

	cleanValue := cleanResult.Value()
	return result.Ok(cleanValue.FreedBytes)
}

// ToCleanResultFromError converts error to Result[domain.CleanResult]
func ToCleanResultFromError(err error) result.Result[domain.CleanResult] {
	return result.Err[domain.CleanResult](err)
}

// ToScanResultFromError converts error to Result[domain.ScanResult]
func ToScanResultFromError(err error) result.Result[domain.ScanResult] {
	return result.Err[domain.ScanResult](err)
}

// VALIDATION HELPERS

// validateAndConvert is a generic helper for validating domain types
func validateAndConvert[T interface{ Validate() error }](item T, typeName string) result.Result[T] {
	if err := item.Validate(); err != nil {
		return result.Err[T](fmt.Errorf("invalid %s: %w", typeName, err))
	}
	return result.Ok(item)
}

// ValidateAndConvertCleanResult ensures CleanResult is valid before returning
func ValidateAndConvertCleanResult(cleanResult domain.CleanResult) result.Result[domain.CleanResult] {
	return validateAndConvert(cleanResult, "CleanResult")
}

// ValidateAndConvertScanResult ensures ScanResult is valid before returning
func ValidateAndConvertScanResult(scanResult domain.ScanResult) result.Result[domain.ScanResult] {
	return validateAndConvert(scanResult, "ScanResult")
}
