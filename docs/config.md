package config // import "github.com/LarsArtmann/clean-wizard/internal/config"

const MinDiskUsagePercent = 10 ...
const DefaultMaxRetries = 3 ...
const MockStoreSizeGB = 300 ...
const MinProtectedPaths = 1 ...
func GetCurrentTime() time.Time
func Load() (*domain.Config, error)
func LoadWithContext(ctx context.Context) (*domain.Config, error)
func Save(config *domain.Config) error
func WithDetailedLogging(enable bool) func(*ValidationMiddlewareOptions)
func WithEnvironment(env string) func(*ValidationMiddlewareOptions)
func WithMonitoring(enabled bool) func(*EnhancedConfigLoader)
func WithRequireSafeModeConfirmation(require bool) func(*ValidationMiddlewareOptions)
func WithRetryPolicy(policy *RetryPolicy) func(*EnhancedConfigLoader)
type BDDFeature struct{ ... }
type BDDGiven struct{ ... }
type BDDScenario struct{ ... }
type BDDTestRunner struct{ ... }
    func NewBDDTestRunner(t *testing.T, feature BDDFeature) *BDDTestRunner
type BDDThen struct{ ... }
type BDDWhen struct{ ... }
type ChangeOperation string
    const OperationAdded ChangeOperation = "added" ...
type CleanType string
    const CleanTypeNixStore CleanType = "nix_store" ...
type ConfigCache struct{ ... }
    func NewConfigCache(ttl time.Duration) *ConfigCache
type ConfigChange struct{ ... }
type ConfigChangeResult struct{ ... }
type ConfigLoadOptions struct{ ... }
type ConfigSanitizer struct{ ... }
    func NewConfigSanitizer() *ConfigSanitizer
    func NewConfigSanitizerWithRules(rules *SanitizationRules) *ConfigSanitizer
type ConfigSaveOptions struct{ ... }
type ConfigValidationRules struct{ ... }
type ConfigValidator struct{ ... }
    func NewConfigValidator() *ConfigValidator
    func NewConfigValidatorWithRules(rules *ConfigValidationRules) *ConfigValidator
type ContextValues struct{ ... }
type DefaultValidationLogger struct{ ... }
    func NewDefaultValidationLogger(enableDetailed bool) *DefaultValidationLogger
type EnhancedConfigLoader struct{ ... }
    func NewEnhancedConfigLoader(options ...func(*EnhancedConfigLoader)) *EnhancedConfigLoader
type NumericValidationRule struct{ ... }
type ProfileOperationResult struct{ ... }
type RetryPolicy struct{ ... }
type SafeConfig struct{ ... }
type SafeConfigBuilder struct{ ... }
    func NewSafeConfigBuilder() *SafeConfigBuilder
type SafeOperation struct{ ... }
type SafeProfile struct{ ... }
type SafeProfileBuilder struct{ ... }
type SanitizationResult struct{ ... }
type SanitizationRules struct{ ... }
type SanitizationWarning struct{ ... }
type StringValidationRule struct{ ... }
type TypeSafeValidationRules struct{ ... }
    func NewTypeSafeValidationRules() *TypeSafeValidationRules
type TypedContext struct{ ... }
    func FromMap(m map[string]any) (*TypedContext, error)
    func NewTypedContext(field, operation string) *TypedContext
type ValidationError struct{ ... }
type ValidationLevel int
    const ValidationLevelNone ValidationLevel = 0 ...
type ValidationLogger interface{ ... }
type ValidationMetadata struct{ ... }
type ValidationMiddleware struct{ ... }
    func NewValidationMiddleware() *ValidationMiddleware
    func NewValidationMiddlewareWithLogger(logger ValidationLogger) *ValidationMiddleware
    func NewValidationMiddlewareWithOptions(options ...func(*ValidationMiddlewareOptions)) *ValidationMiddleware
type ValidationMiddlewareOptions struct{ ... }
type ValidationResult struct{ ... }
type ValidationRule[T comparable] struct{ ... }
type ValidationSanitizedData struct{ ... }
type ValidationSeverity string
    const SeverityError ValidationSeverity = "error" ...
type ValidationWarning struct{ ... }
