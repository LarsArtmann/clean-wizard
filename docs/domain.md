package domain // import "github.com/LarsArtmann/clean-wizard/internal/domain"

var RiskLow = RiskLevelType(RiskLevelLowType) ...
var StrategyAggressive = CleanStrategyType(StrategyAggressiveType) ...
func ParseCustomDuration(durationStr string) (time.Duration, error)
func ValidateCustomDuration(durationStr string) error
type ChangeOperation = ChangeOperationType
type ChangeOperationType int
const ChangeOperationAddedType ChangeOperationType = iota ...
type CleanRequest struct{ ... }
type CleanResult struct{ ... }
type CleanStrategy = CleanStrategyType
type CleanStrategyType int
const StrategyAggressiveType CleanStrategyType = iota ...
type Cleaner interface{ ... }
type CleanupOperation struct{ ... }
type Config struct{ ... }
type GenerationCleaner interface{ ... }
type HomebrewSettings struct{ ... }
type NixGeneration struct{ ... }
type NixGenerationsSettings struct{ ... }
type OperationSettings struct{ ... }
func DefaultSettings(opType OperationType) \*OperationSettings
type OperationType string
const OperationTypeNixGenerations OperationType = "nix-generations" ...
func GetOperationType(name string) OperationType
type PackageCleaner interface{ ... }
type Profile struct{ ... }
type RiskLevel = RiskLevelType
type RiskLevelType int
const RiskLevelLowType RiskLevelType = iota ...
type ScanItem struct{ ... }
type ScanRequest struct{ ... }
type ScanResult struct{ ... }
type ScanType string
const ScanTypeNixStore ScanType = "nix_store" ...
type Scanner interface{ ... }
type SystemTempSettings struct{ ... }
type TempFilesSettings struct{ ... }
type TypeSafeEnum[T any] interface{ ... }
type ValidationError struct{ ... }
type ValidationLevel = ValidationLevelType
type ValidationLevelType int
const ValidationLevelNoneType ValidationLevelType = iota ...
