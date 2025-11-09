package conversions

import (
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
)

// ToCleanResult converts primitive Result[int64] to domain Result[domain.CleanResult]
func ToCleanResult(bytesResult result.Result[int64]) result.Result[domain.CleanResult] {
	if bytesResult.IsErr() {
		return result.Err[domain.CleanResult](bytesResult.Error())
	}

	bytes := bytesResult.Value()
	return result.Ok(domain.CleanResult{
		FreedBytes: bytes,
		Strategy:   "default",
		CleanedAt:  time.Now(),
	})
}

// ToCleanResultWithStrategy converts primitive Result[int64] to domain.Result[domain.CleanResult] with strategy
func ToCleanResultWithStrategy(bytesResult result.Result[int64], strategy string) result.Result[domain.CleanResult] {
	if bytesResult.IsErr() {
		return result.Err[domain.CleanResult](bytesResult.Error())
	}

	bytes := bytesResult.Value()
	return result.Ok(domain.CleanResult{
		FreedBytes: bytes,
		Strategy:   strategy,
		CleanedAt:  time.Now(),
	})
}

// ToScanResult converts primitive scanning results to domain.ScanResult
func ToScanResult(totalBytes int64, totalItems int, scannedPaths []string, scanDuration time.Duration) domain.ScanResult {
	return domain.ScanResult{
		TotalBytes:   totalBytes,
		TotalItems:   totalItems,
		ScannedPaths: scannedPaths,
		ScanTime:     scanDuration,
		ScannedAt:    time.Now(),
	}
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