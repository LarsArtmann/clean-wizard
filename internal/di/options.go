package di

// RunSettings holds the runtime flags parsed from the CLI.
// These are registered as an eager value in the DI container so that
// all cleaner providers can resolve them lazily.
type RunSettings struct {
	Verbose        bool
	DryRun         bool
	MaxConcurrency int
}
