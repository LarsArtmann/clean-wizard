package conversions

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

func TestNewCleanResult(t *testing.T) {
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	items := 5
	bytes := int64(1024)

	cleanResult := NewCleanResult(strategy, items, bytes)

	if cleanResult.Strategy != strategy {
		t.Errorf("Expected strategy %s, got %s", strategy.String(), cleanResult.Strategy.String())
	}
	if cleanResult.ItemsRemoved != uint(items) {
		t.Errorf("Expected items %d, got %d", items, cleanResult.ItemsRemoved)
	}
	if cleanResult.FreedBytes != uint64(bytes) {
		t.Errorf("Expected bytes %d, got %d", bytes, cleanResult.FreedBytes)
	}
	if cleanResult.ItemsFailed != 0 {
		t.Errorf("Expected items failed 0, got %d", cleanResult.ItemsFailed)
	}
	if cleanResult.CleanTime != 0 {
		t.Errorf("Expected clean time 0, got %d", cleanResult.CleanTime)
	}
	if cleanResult.CleanedAt.IsZero() {
		t.Error("Expected cleaned at to be set, got zero time")
	}
}

func TestNewCleanResultWithTiming(t *testing.T) {
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	items := 3
	bytes := int64(2048)
	cleanTime := time.Duration(5) * time.Second

	cleanResult := NewCleanResultWithTiming(strategy, items, bytes, cleanTime)

	if cleanResult.CleanTime != cleanTime {
		t.Errorf("Expected clean time %v, got %v", cleanTime, cleanResult.CleanTime)
	}
	if cleanResult.Strategy != strategy {
		t.Errorf("Expected strategy %s, got %s", strategy.String(), cleanResult.Strategy.String())
	}
	if cleanResult.ItemsRemoved != uint(items) {
		t.Errorf("Expected items %d, got %d", items, cleanResult.ItemsRemoved)
	}
	if cleanResult.FreedBytes != uint64(bytes) {
		t.Errorf("Expected bytes %d, got %d", bytes, cleanResult.FreedBytes)
	}
}

func TestNewCleanResultWithFailures(t *testing.T) {
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	itemsRemoved := 8
	itemsFailed := 2
	bytes := int64(4096)
	cleanTime := time.Duration(10) * time.Second

	cleanResult := NewCleanResultWithFailures(strategy, itemsRemoved, itemsFailed, bytes, cleanTime)

	if cleanResult.Strategy != strategy {
		t.Errorf("Expected strategy %s, got %s", strategy.String(), cleanResult.Strategy.String())
	}
	if cleanResult.ItemsRemoved != uint(itemsRemoved) {
		t.Errorf("Expected items removed %d, got %d", itemsRemoved, cleanResult.ItemsRemoved)
	}
	if cleanResult.ItemsFailed != uint(itemsFailed) {
		t.Errorf("Expected items failed %d, got %d", itemsFailed, cleanResult.ItemsFailed)
	}
	if cleanResult.FreedBytes != uint64(bytes) {
		t.Errorf("Expected bytes %d, got %d", bytes, cleanResult.FreedBytes)
	}
	if cleanResult.CleanTime != cleanTime {
		t.Errorf("Expected clean time %v, got %v", cleanTime, cleanResult.CleanTime)
	}
}

func TestNewScanResult(t *testing.T) {
	totalBytes := int64(8192)
	totalItems := 12
	scannedPaths := []string{"/path1", "/path2"}
	scanDuration := time.Duration(3) * time.Second

	scanResult := NewScanResult(totalBytes, totalItems, scannedPaths, scanDuration)

	if scanResult.TotalBytes != totalBytes {
		t.Errorf("Expected total bytes %d, got %d", totalBytes, scanResult.TotalBytes)
	}
	if scanResult.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, scanResult.TotalItems)
	}
	if len(scanResult.ScannedPaths) != len(scannedPaths) {
		t.Errorf("Expected %d scanned paths, got %d", len(scannedPaths), len(scanResult.ScannedPaths))
	}
	if scanResult.ScanTime != scanDuration {
		t.Errorf("Expected scan time %v, got %v", scanDuration, scanResult.ScanTime)
	}
	if scanResult.ScannedAt.IsZero() {
		t.Error("Expected scanned at to be set, got zero time")
	}
}

func TestToCleanResult(t *testing.T) {
	bytes := int64(1024)
	bytesResult := result.Ok(bytes)

	cleanResult := ToCleanResult(bytesResult)

	if cleanResult.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", cleanResult.Error())
	}

	value := cleanResult.Value()
	if value.FreedBytes != uint64(bytes) {
		t.Errorf("Expected freed bytes %d, got %d", bytes, value.FreedBytes)
	}
	if value.Strategy != domain.CleanStrategyType(domain.StrategyConservativeType) {
		t.Errorf("Expected strategy 'conservative', got %s", value.Strategy)
	}
}

func TestToCleanResultWithError(t *testing.T) {
	expectedErr := errors.New("test error")
	bytesResult := result.Err[int64](expectedErr)

	cleanResult := ToCleanResult(bytesResult)

	if cleanResult.IsOk() {
		t.Error("Expected error result, got Ok")
	}

	if cleanResult.Error().Error() != expectedErr.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedErr.Error(), cleanResult.Error().Error())
	}
}

func TestToCleanResultWithStrategy(t *testing.T) {
	bytes := int64(2048)
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	bytesResult := result.Ok(bytes)

	cleanResult := ToCleanResultWithStrategy(bytesResult, strategy)

	if cleanResult.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", cleanResult.Error())
	}

	value := cleanResult.Value()
	if value.FreedBytes != uint64(bytes) {
		t.Errorf("Expected freed bytes %d, got %d", bytes, value.FreedBytes)
	}
	if value.Strategy.String() != strategy.String() {
		t.Errorf("Expected strategy '%s', got %s", strategy.String(), value.Strategy.String())
	}
}

func TestToCleanResultFromItems(t *testing.T) {
	itemsRemoved := 5
	bytes := int64(4096)
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	bytesResult := result.Ok(bytes)

	cleanResult := ToCleanResultFromItems(itemsRemoved, bytesResult, strategy)

	if cleanResult.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", cleanResult.Error())
	}

	value := cleanResult.Value()
	if value.ItemsRemoved != uint(itemsRemoved) {
		t.Errorf("Expected items removed %d, got %d", itemsRemoved, value.ItemsRemoved)
	}
	if value.FreedBytes != uint64(bytes) {
		t.Errorf("Expected freed bytes %d, got %d", bytes, value.FreedBytes)
	}
	if value.Strategy.String() != strategy.String() {
		t.Errorf("Expected strategy '%s', got %s", strategy.String(), value.Strategy.String())
	}
}

func TestToTimedCleanResult(t *testing.T) {
	bytes := int64(8192)
	strategy := domain.CleanStrategyType(domain.StrategyDryRunType)
	cleanTime := time.Duration(7) * time.Second
	bytesResult := result.Ok(bytes)

	cleanResult := ToTimedCleanResult(bytesResult, strategy, cleanTime)

	if cleanResult.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", cleanResult.Error())
	}

	value := cleanResult.Value()
	if value.FreedBytes != uint64(bytes) {
		t.Errorf("Expected freed bytes %d, got %d", bytes, value.FreedBytes)
	}
	if value.Strategy.String() != strategy.String() {
		t.Errorf("Expected strategy '%s', got %s", strategy.String(), value.Strategy.String())
	}
	if value.CleanTime != cleanTime {
		t.Errorf("Expected clean time %v, got %v", cleanTime, value.CleanTime)
	}
}

func TestToScanResult(t *testing.T) {
	totalBytes := int64(16384)
	totalItems := 20
	scannedPaths := []string{"/path1", "/path2", "/path3"}
	scanDuration := time.Duration(15) * time.Second

	scanResult := ToScanResult(totalBytes, totalItems, scannedPaths, scanDuration)

	if scanResult.TotalBytes != totalBytes {
		t.Errorf("Expected total bytes %d, got %d", totalBytes, scanResult.TotalBytes)
	}
	if scanResult.TotalItems != totalItems {
		t.Errorf("Expected total items %d, got %d", totalItems, scanResult.TotalItems)
	}
	if len(scanResult.ScannedPaths) != len(scannedPaths) {
		t.Errorf("Expected %d scanned paths, got %d", len(scannedPaths), len(scanResult.ScannedPaths))
	}
	if scanResult.ScanTime != scanDuration {
		t.Errorf("Expected scan time %v, got %v", scanDuration, scanResult.ScanTime)
	}
}

func TestCombineCleanResults(t *testing.T) {
	result1 := NewCleanResult(domain.CleanStrategyType(domain.StrategyAggressiveType), 3, int64(1024))
	result2 := NewCleanResult(domain.CleanStrategyType(domain.StrategyConservativeType), 5, int64(2048))

	results := []domain.CleanResult{result1, result2}

	combined := CombineCleanResults(results)

	if combined.ItemsRemoved != uint(8) { // 3 + 5
		t.Errorf("Expected items removed 8, got %d", combined.ItemsRemoved)
	}
	if combined.FreedBytes != uint64(3072) { // 1024 + 2048
		t.Errorf("Expected freed bytes 3072, got %d", combined.FreedBytes)
	}
	if combined.ItemsFailed != uint(0) {
		t.Errorf("Expected items failed 0, got %d", combined.ItemsFailed)
	}
	// When combining different strategies, should default to conservative
	if combined.Strategy != domain.CleanStrategyType(domain.StrategyConservativeType) {
		t.Errorf("Expected combined strategy to be 'conservative' for mixed strategies, got %s", combined.Strategy)
	}
}

func TestCombineCleanResultsWithFailures(t *testing.T) {
	result1 := NewCleanResultWithFailures(domain.CleanStrategyType(domain.StrategyAggressiveType), 3, 1, int64(1024), time.Second)
	result2 := NewCleanResultWithFailures(domain.CleanStrategyType(domain.StrategyConservativeType), 5, 2, int64(2048), 2*time.Second)

	results := []domain.CleanResult{result1, result2}

	combined := CombineCleanResults(results)

	if combined.ItemsRemoved != uint(8) { // 3 + 5
		t.Errorf("Expected items removed 8, got %d", combined.ItemsRemoved)
	}
	if combined.ItemsFailed != uint(3) { // 1 + 2
		t.Errorf("Expected items failed 3, got %d", combined.ItemsFailed)
	}
	if combined.FreedBytes != uint64(3072) { // 1024 + 2048
		t.Errorf("Expected freed bytes 3072, got %d", combined.FreedBytes)
	}
	if combined.CleanTime != 2*time.Second { // max(1s, 2s)
		t.Errorf("Expected clean time 2s, got %v", combined.CleanTime)
	}
}

func TestCombineCleanResultsEmpty(t *testing.T) {
	results := []domain.CleanResult{}

	combined := CombineCleanResults(results)

	if combined.ItemsRemoved != uint(0) {
		t.Errorf("Expected items removed 0, got %d", combined.ItemsRemoved)
	}
	if combined.FreedBytes != uint64(0) {
		t.Errorf("Expected freed bytes 0, got %d", combined.FreedBytes)
	}
	if combined.Strategy != domain.CleanStrategyType(domain.StrategyConservativeType) {
		t.Errorf("Expected strategy 'conservative' (default for empty results), got %s", combined.Strategy)
	}
}

func TestExtractBytesFromCleanResult(t *testing.T) {
	bytes := int64(4096)
	cleanResult := result.Ok(NewCleanResult(domain.CleanStrategyType(domain.StrategyConservativeType), 1, bytes))

	extracted := ExtractBytesFromCleanResult(cleanResult)

	if extracted.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", extracted.Error())
	}

	if extracted.Value() != bytes {
		t.Errorf("Expected bytes %d, got %d", bytes, extracted.Value())
	}
}

func TestExtractBytesFromCleanResultWithError(t *testing.T) {
	expectedErr := errors.New("test error")
	cleanResult := result.Err[domain.CleanResult](expectedErr)

	extracted := ExtractBytesFromCleanResult(cleanResult)

	if extracted.IsOk() {
		t.Error("Expected error result, got Ok")
	}

	if extracted.Error().Error() != expectedErr.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedErr.Error(), extracted.Error().Error())
	}
}

func TestToCleanResultFromError(t *testing.T) {
	expectedErr := errors.New("test error")

	cleanResult := ToCleanResultFromError(expectedErr)

	if cleanResult.IsOk() {
		t.Error("Expected error result, got Ok")
	}

	if cleanResult.Error().Error() != expectedErr.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedErr.Error(), cleanResult.Error().Error())
	}
}

func TestValidateAndConvertCleanResult(t *testing.T) {
	validResult := NewCleanResult(domain.CleanStrategyType(domain.StrategyConservativeType), 1, 1024)

	result := ValidateAndConvertCleanResult(validResult)

	if result.IsErr() {
		t.Errorf("Expected Ok result, got error: %v", result.Error())
	}
}

func TestValidateAndConvertCleanResultInvalid(t *testing.T) {
	invalidResult := domain.CleanResult{
		SizeEstimate: domain.SizeEstimate{Known: 0, Status: domain.SizeEstimateStatusKnown}, // Invalid zero bytes with ItemsRemoved > 0
		ItemsRemoved: 1,
		CleanedAt:    time.Now(),
		Strategy:     domain.CleanStrategyType(domain.StrategyDryRunType),
		CleanTime:    time.Second, // Non-zero clean time triggers validation
	}

	result := ValidateAndConvertCleanResult(invalidResult)

	if result.IsOk() {
		t.Error("Expected error result, got Ok")
	}

	if !strings.Contains(result.Error().Error(), "invalid CleanResult") {
		t.Errorf("Expected 'invalid CleanResult' in error, got %s", result.Error().Error())
	}
}
