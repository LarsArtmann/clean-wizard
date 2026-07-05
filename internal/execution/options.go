package execution

// RunOption configures a RunCleaners invocation.
type RunOption func(*runConfig)

type runConfig struct {
	maxConcurrency int
	verbose        bool
	retry          *RetryConfig
}

// WithMaxConcurrency sets the maximum number of cleaners that may run
// concurrently. A value of 0 (the default) means unlimited.
func WithMaxConcurrency(n int) RunOption {
	return func(c *runConfig) { c.maxConcurrency = n }
}

// WithVerbose enables or disables per-step debug output during workflow execution.
func WithVerbose(verbose bool) RunOption {
	return func(c *runConfig) { c.verbose = verbose }
}

// WithRetry enables per-step retry with the given configuration.
// Passing nil disables retries (the default).
func WithRetry(cfg *RetryConfig) RunOption {
	return func(c *runConfig) { c.retry = cfg }
}

func resolveRunOptions(opts []RunOption) runConfig {
	var c runConfig
	for _, opt := range opts {
		opt(&c)
	}
	return c
}
