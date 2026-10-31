package errorfamily

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"time"
)

// Error is the reference implementation of a classified, structured error.
//
// This is a convenience type, not the contract. The public protocol is the
// Coded/Classified/Contextual/Retryable interfaces — your own domain error
// types implement only those and need not embed or use this struct.
//
// Projects with simple error needs can use this directly.
// Projects with domain-specific needs (e.g., FindingError with Position/File)
// can build their own struct and implement the Coded/Classified/Contextual interfaces.
//
// Implements: error, Coded, Classified, Contextual, Retryable, ExitCoder, fmt.Formatter.
type Error struct {
	code       string            // machine-readable identity (e.g. "db.timeout")
	message    string            // human-readable technical message
	family     Family            // behavioral classification
	context    map[string]string // factual details about the error
	cause      error             // underlying error in the chain
	timestamp  time.Time         // when the error occurred
	exitCode   int               // optional override; 0 = use family default
	httpStatus int               // optional override; 0 = use family default
}

// Error implements the error interface.
// Returns a compact format: [family:code] message.
// The cause string is extracted via [safeCauseString] to guard against
// panics from misbehaving wrapped error types.
func (e *Error) Error() string {
	if e.cause != nil {
		causeMsg := safeCauseString(e.cause)
		if causeMsg != "" {
			return fmt.Sprintf("[%s:%s] %s: %s", e.family, e.code, e.message, causeMsg)
		}
	}

	return fmt.Sprintf("[%s:%s] %s", e.family, e.code, e.message)
}

// Unwrap returns the underlying cause for error chain traversal.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

// Is supports errors.Is by matching error code and family.
// Two errors match if they have the same code AND family.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.code == t.code && e.family == t.family
}

// ErrorCode returns the machine-readable error code.
//
// This is the canonical accessor: it satisfies the [Coded] interface, which is
// the public contract for code extraction (used by [Code], [HandleError],
// metrics, and log fields). Prefer this over [Error.Code] when you need the
// interface behavior. See [Error.Code] for the difference.
func (e *Error) ErrorCode() string { return e.code }

// ErrorFamily returns the behavioral classification.
func (e *Error) ErrorFamily() Family { return e.family }

// ErrorContext returns a copy of the error's factual context.
func (e *Error) ErrorContext() map[string]string {
	if e.context == nil {
		return map[string]string{}
	}

	return maps.Clone(e.context)
}

// IsRetryable reports whether this error is worth retrying.
func (e *Error) IsRetryable() bool { return e.family.IsRetryable() }

// Timestamp returns when this error was created.
func (e *Error) Timestamp() time.Time { return e.timestamp }

// Family returns the error's behavioral classification.
// Convenience accessor for direct use without interface assertion.
func (e *Error) Family() Family { return e.family }

// Code returns the machine-readable error code.
//
// This is a direct accessor — a convenience sibling of [Error.Family] and
// [Error.Message] for use when you already hold a concrete *Error and want the
// field without an interface assertion. It returns the SAME value as
// [Error.ErrorCode]; the two exist because ErrorCode is the [Coded] interface
// contract (used via errors.AsType / [Code](err)), while Code is the ergonomic
// accessor on the concrete type. Neither is deprecated.
func (e *Error) Code() string { return e.code }

// Message returns the human-readable technical message.
func (e *Error) Message() string { return e.message }

// Cause returns the underlying error in the chain.
func (e *Error) Cause() error {
	return e.cause
}

// Format implements fmt.Formatter.
//
//	%v    → [family:code] message
//	%+v   → verbose multi-line with context and cause chain
//	%s    → message only (no classification prefix)
func (e *Error) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		_, _ = fmt.Fprint( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
			f,
			e.message,
		)
	case 'v':
		if f.Flag('+') {
			e.formatVerbose(f)

			return
		}

		_, _ = fmt.Fprint(
			f,
			e.Error(),
		) //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
	default:
		_, _ = fmt.Fprint(
			f,
			e.Error(),
		) //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
	}
}

func (e *Error) formatVerbose(f fmt.State) {
	_, _ = fmt.Fprintf( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
		f,
		"[%s] %s: %s",
		e.family,
		e.code,
		e.message,
	)

	if len(e.context) > 0 {
		_, _ = fmt.Fprint(
			f,
			"\n  context:",
		) //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
		for k, v := range e.context {
			_, _ = fmt.Fprintf( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
				f,
				"\n    %s: %s",
				k,
				v,
			)
		}
	}

	if !e.timestamp.IsZero() {
		_, _ = fmt.Fprintf( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
			f,
			"\n  at: %s",
			e.timestamp.Format(time.RFC3339),
		)
	}

	if e.exitCode != 0 {
		_, _ = fmt.Fprintf( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
			f,
			"\n  exit_code: %d",
			e.exitCode,
		)
	}

	if e.cause != nil {
		causeMsg := safeCauseString(e.cause)
		if causeMsg != "" {
			_, _ = fmt.Fprintf( //nolint:legacyerrors // fmt.State write errors cannot be propagated (fmt.Formatter has no error return)
				f,
				"\n  caused by: %s",
				causeMsg,
			)
		}
	}
}

// WithContext adds a key-value pair to the error's context.
// Returns a new Error, leaving the original unchanged — safe for shared/sentinel errors.
func (e *Error) WithContext(key, value string) *Error {
	clone := e.clone()
	clone.context[key] = value

	return clone
}

// WithContextMap merges a map of key-value pairs into the error's context.
// Returns a new Error, leaving the original unchanged. Nil or empty input
// returns a clone with no added context.
func (e *Error) WithContextMap(ctx map[string]string) *Error {
	clone := e.clone()
	maps.Insert(clone.context, maps.All(ctx))

	return clone
}

// WithContextf adds a formatted key-value pair to the error's context.
// The value is produced by fmt.Sprintf(format, args...).
// Returns a new Error, leaving the original unchanged.
func (e *Error) WithContextf(key, format string, args ...any) *Error {
	clone := e.clone()
	clone.context[key] = fmt.Sprintf(format, args...)

	return clone
}

// WithContextAny adds a key-value pair to the error's context where the value
// can be of any type. The value is converted to string via a type switch for
// common types (string, int, int64, uint, uint64, float64, bool) and falls back
// to fmt.Sprint for everything else.
// Returns a new Error, leaving the original unchanged — safe for shared/sentinel errors.
func (e *Error) WithContextAny(key string, value any) *Error {
	return e.WithContext(key, contextValueToString(value))
}

// WithCause sets the underlying cause and returns a new error for chaining.
func (e *Error) WithCause(cause error) *Error {
	clone := e.clone()
	clone.cause = cause

	return clone
}

// WithTimestamp sets the error timestamp and returns a new error for chaining.
// Useful for testing and deterministic construction.
func (e *Error) WithTimestamp(ts time.Time) *Error {
	clone := e.clone()
	clone.timestamp = ts

	return clone
}

// WithExitCode sets a custom process exit code that overrides the family-based
// default from [Family.ExitCode]. A value of 0 means "not set" — callers should
// fall back to the family's canonical exit code.
//
// Note: os.Exit wraps the code to the range 0-255 on POSIX (e.g., -1 becomes 255).
// Use values in the 1-125 range for maximum portability (126+ have special meaning
// in some shells).
//
// Returns a new Error, leaving the original unchanged — safe for shared/sentinel errors.
func (e *Error) WithExitCode(code int) *Error {
	clone := e.clone()
	clone.exitCode = code

	return clone
}

// WithHTTPStatus sets a custom HTTP response status code that overrides the
// family-based default from [Family.HTTPStatus]. A value of 0 means "not set" —
// callers should fall back to the family's canonical HTTP status.
//
// Use this for errors whose family default is misleading at the HTTP layer:
//
//	notFound := NewRejection("battle.not_found", "battle not found").
//	    WithHTTPStatus(http.StatusNotFound) // 404 instead of family default 400
//
// Returns a new Error, leaving the original unchanged — safe for shared/sentinel errors.
func (e *Error) WithHTTPStatus(status int) *Error {
	clone := e.clone()
	clone.httpStatus = status

	return clone
}

// clone returns a shallow copy of the error with a deep-copied context map.
func (e *Error) clone() *Error {
	cloned := &Error{
		code:       e.code,
		message:    e.message,
		family:     e.family,
		cause:      e.cause,
		timestamp:  e.timestamp,
		exitCode:   e.exitCode,
		httpStatus: e.httpStatus,
		context:    make(map[string]string, len(e.context)),
	}
	maps.Copy(cloned.context, e.context)

	return cloned
}

// Summary returns a one-line human-readable summary suitable for logs and CLI output.
// Format: "code: message" without family prefix.
// The cause string is extracted via [safeCauseString] to guard against panics.
func (e *Error) Summary() string {
	if e.cause != nil {
		causeMsg := safeCauseString(e.cause)
		if causeMsg != "" {
			return fmt.Sprintf("%s: %s — %s", e.code, e.message, causeMsg)
		}
	}

	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// ExitCode returns the custom process exit code if one was set via [WithExitCode].
// Returns 0 when no custom code is set, signaling callers to fall back to the
// family-based default from [Family.ExitCode].
//
// This satisfies the [ExitCoder] interface so that [ExitCode] (the package-level
// function) can discover per-error overrides.
func (e *Error) ExitCode() int { return e.exitCode }

// HTTPStatus returns the custom HTTP response status code if one was set via
// [WithHTTPStatus]. Returns 0 when no custom status is set, signaling callers
// to fall back to the family-based default from [Family.HTTPStatus].
//
// This satisfies the [HTTPStatuser] interface so that [HTTPStatus] (the
// package-level function) can discover per-error overrides.
func (e *Error) HTTPStatus() int { return e.httpStatus }

// safeCauseString calls cause.Error() with panic recovery.
// Certain third-party error types panic when their Error() method encounters
// nil internal values. This guard prevents the panic from propagating through
// fmt.Sprintf callers, returning an empty string instead.
func safeCauseString(cause error) string {
	defer func() {
		_ = recover() //nolint:legacyerrors // the panic value is not an error; deliberate silent recovery so a misbehaving cause cannot crash the process
	}()

	return cause.Error()
}

// contextValueToString converts any context value to its string representation.
// Handles complex types directly and delegates numeric/boolean scalars to scalarToString.
func contextValueToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case nil:
		return ""
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	case time.Duration:
		return val.String()
	case error:
		return safeCauseString(val)
	default:
		return scalarToString(v)
	}
}

// scalarToString converts numeric and boolean primitives to strings.
// Falls back to fmt.Sprint for any type not explicitly handled.
func scalarToString(v any) string {
	switch val := v.(type) {
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprint(v)
	}
}

// HasContext reports whether the error has a specific context key.
func (e *Error) HasContext(key string) bool {
	if e.context == nil {
		return false
	}

	_, ok := e.context[key]

	return ok
}

// ContextValue returns a specific context value, or empty string if not present.
func (e *Error) ContextValue(key string) string {
	if e.context == nil {
		return ""
	}

	return e.context[key]
}

// jsonError is the JSON view of an Error for API responses.
// exitCode is intentionally excluded: it is a CLI/POSIX concept. API consumers
// use HTTP status codes (via Family.HTTPStatus) or the family/code fields for
// behavioral decisions, not process exit codes.
type jsonError struct {
	Family    string            `json:"family"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Context   map[string]string `json:"context,omitempty"`
	Retryable bool              `json:"retryable"`
	Timestamp string            `json:"timestamp,omitempty"`
}

// JSON returns a canonical JSON encoding of the error for API boundaries.
// The shape is stable: {family, code, message, context, retryable, timestamp}.
// Use this for HTTP/REST error responses where a structured body is preferable
// to the [transient:code] message format of Error().
func (e *Error) JSON() ([]byte, error) {
	view := jsonError{
		Family:    e.family.String(),
		Code:      e.code,
		Message:   e.message,
		Context:   e.context,
		Retryable: e.IsRetryable(),
	}
	if !e.timestamp.IsZero() {
		view.Timestamp = e.timestamp.Format(time.RFC3339)
	}

	data, err := json.Marshal(view)
	if err != nil {
		return nil, fmt.Errorf("marshal error view: %w", err)
	}

	return data, nil
}
