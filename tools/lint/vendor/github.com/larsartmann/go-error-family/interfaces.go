package errorfamily

// Consumer interfaces — small, focused, composable.
//
// Each interface embeds error so it can be used as a type parameter with
// Go 1.26's errors.AsType[T](), which requires T to satisfy the error interface.
//
// Each error type implements the combination that makes sense for it.
// Consumers ask for exactly what they need via errors.AsType[T]().
//
// This is the Go way: accept interfaces, return structs,
// share protocols, not implementations.

// Coded provides machine-readable identity.
// Required for: metric labels, log fields, message templates, exit code mapping.
type Coded interface {
	error
	ErrorCode() string
}

// Classified provides behavioral classification.
// Required for: retry decisions, exit codes, circuit breakers, tone selection.
type Classified interface {
	error
	ErrorFamily() Family
}

// Contextual provides factual details about the error.
// Required for: structured logging, specific error messages, debug context,
// and triggering diagnostic rules.
type Contextual interface {
	error
	ErrorContext() map[string]string
}

// Retryable errors explicitly declare whether they should be retried.
// Overrides Family-based retry decisions when present.
type Retryable interface {
	error
	IsRetryable() bool
}

// ExitCoder errors carry a custom process exit code that overrides the
// family-based default from [Family.ExitCode]. When [ExitCode] encounters
// an error implementing this interface with a non-zero code, that code wins.
//
// This allows individual errors to deviate from their family's canonical
// exit code without changing the family classification itself.
type ExitCoder interface {
	error
	ExitCode() int
}

// HTTPStatuser errors carry a custom HTTP response status code that overrides
// the family-based default from [Family.HTTPStatus]. When [HTTPStatus]
// encounters an error implementing this interface with a non-zero status,
// that status wins.
//
// This allows individual errors to deviate from their family's canonical
// HTTP status (e.g. a Rejection error that should be 404 Not Found instead
// of the family default 400 Bad Request) without changing the classification.
type HTTPStatuser interface {
	error
	HTTPStatus() int
}
