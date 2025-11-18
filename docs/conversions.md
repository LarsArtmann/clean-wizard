package conversions // import "github.com/LarsArtmann/clean-wizard/internal/conversions"

func CombineCleanResults(results []domain.CleanResult) domain.CleanResult
func ExtractBytesFromCleanResult(cleanResult result.Result[domain.CleanResult]) result.Result[int64]
func NewCleanResult(strategy domain.CleanStrategy, itemsRemoved int, freedBytes int64) domain.CleanResult
func NewCleanResultWithFailures(strategy domain.CleanStrategy, itemsRemoved, itemsFailed int, freedBytes int64, ...) domain.CleanResult
func NewCleanResultWithTiming(strategy domain.CleanStrategy, itemsRemoved int, freedBytes int64, ...) domain.CleanResult
func NewScanResult(totalBytes int64, totalItems int, scannedPaths []string, ...) domain.ScanResult
func ToCleanResult(bytesResult result.Result[int64]) result.Result[domain.CleanResult]
func ToCleanResultFromError(err error) result.Result[domain.CleanResult]
func ToCleanResultFromItems(itemsRemoved int, bytesResult result.Result[int64], ...) result.Result[domain.CleanResult]
func ToCleanResultWithStrategy(bytesResult result.Result[int64], strategy domain.CleanStrategy) result.Result[domain.CleanResult]
func ToScanResult(totalBytes int64, totalItems int, scannedPaths []string, ...) domain.ScanResult
func ToScanResultFromError(err error) result.Result[domain.ScanResult]
func ToTimedCleanResult(bytesResult result.Result[int64], strategy domain.CleanStrategy, ...) result.Result[domain.CleanResult]
func ValidateAndConvertCleanResult(cleanResult domain.CleanResult) result.Result[domain.CleanResult]
func ValidateAndConvertScanResult(scanResult domain.ScanResult) result.Result[domain.ScanResult]
