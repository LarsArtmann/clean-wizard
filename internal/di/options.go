package di

// RunSettings holds the runtime flags parsed from the CLI.
// These are registered as an eager value in the DI container so that
// all cleaner providers can resolve them lazily.
type RunSettings struct {
	Verbose        bool
	DryRun         bool
	MaxConcurrency int

	// Profile is the name of the configuration profile selected on the CLI.
	// When set, the cleaners in the registry are constructed with the merged
	// OperationSettings of that profile; empty means no profile was selected
	// and cleaners use their factory defaults.
	Profile string
}
