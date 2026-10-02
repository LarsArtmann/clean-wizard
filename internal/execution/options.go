package execution

// RunOption configures a RunCleaners invocation.
type RunOption func(*runConfig)

type runConfig struct {
	maxConcurrency int
	verbose        bool
	retry          *RetryConfig
	progress       ProgressEmitter
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

// WithProgress injects a live progress renderer (ADR-0002). Passing nil or
// omitting the option keeps execution completely silent — the no-op emitter
// makes non-progress runs byte-identical to the pre-progress behavior.
func WithProgress(emitter ProgressEmitter) RunOption {
	return func(c *runConfig) {
		if emitter != nil {
			c.progress = emitter
		}
	}
}

// emitter returns the configured progress emitter, substituting the no-op
// when none was injected so call sites never nil-check.
func (c runConfig) emitter() ProgressEmitter {
	if c.progress == nil {
		return noopEmitter
	}

	return c.progress
}

func resolveRunOptions(opts []RunOption) runConfig {
	var c runConfig
	for _, opt := range opts {
		opt(&c)
	}

	return c
}
