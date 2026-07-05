package execution

// RunOption configures a RunCleaners invocation.
type RunOption func(*runConfig)

type runConfig struct {
	maxConcurrency int
	verbose        bool
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

func resolveRunOptions(opts []RunOption) runConfig {
	var c runConfig
	for _, opt := range opts {
		opt(&c)
	}
	return c
}
