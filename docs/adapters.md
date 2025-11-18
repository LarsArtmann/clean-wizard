package adapters // import "github.com/LarsArtmann/clean-wizard/internal/adapters"

func ErrCacheMiss(key string) error
func ErrForbidden(operation string) error
func ErrHTTPError(statusCode int, message string) error
func ErrInvalidArgument(arg, message string) error
func ErrInvalidConfig(message string) error
func ErrNotFound(resource string) error
func ErrNotImplemented(feature string) error
func ErrRateLimit(limit float64) error
func ErrServiceUnavailable(service string) error
func ErrTimeout(operation string) error
func ErrUnauthorized(operation string) error
func GetEnvBool(key string, defaultValue bool) bool
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration
func GetEnvInt(key string, defaultValue int) int
func GetEnvWithDefault(key, defaultValue string) string
type CacheManager struct{ ... }
    func NewCacheManager(defaultExpiration, cleanupInterval time.Duration) *CacheManager
type CacheStats struct{ ... }
type EnvironmentConfig struct{ ... }
    func LoadEnvironmentConfig() (*EnvironmentConfig, error)
    func LoadEnvironmentConfigWithPrefix(prefix string) (*EnvironmentConfig, error)
type HTTPClient struct{ ... }
    func NewHTTPClient() *HTTPClient
type HTTPResponse struct{ ... }
type NixAdapter struct{ ... }
    func NewNixAdapter(timeout time.Duration, retries int) *NixAdapter
type RateLimitStats struct{ ... }
type RateLimiter struct{ ... }
    func NewRateLimiter(rps float64, burst int) *RateLimiter
