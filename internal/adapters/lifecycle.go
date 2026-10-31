package adapters

// ShutdownerWithError mirrors do.ShutdownerWithError so the adapters package
// stays free of the DI library while proving the shutdown contract its types
// satisfy when (re-)registered as container services.
type ShutdownerWithError interface {
	Shutdown() error
}

var (
	_ ShutdownerWithError = (*CacheManager)(nil)
	_ ShutdownerWithError = (*HTTPClient)(nil)
)
