package errorfamily

import "strings"

// Common string constants to satisfy goconst linter.
const (
	strRejection      = "rejection"
	strConflict       = "conflict"
	strTransient      = "transient"
	strCorruption     = "corruption"
	strInfrastructure = "infrastructure"
	strOrchestration  = "orchestration"
	strUnknown        = "unknown"
	strUser           = "user"
	strOps            = "ops"
	strAll            = "all"
	msgCheckInput     = "Check your input and try again."
	msgRefreshData    = "Refresh your data and try the operation again."
)

// Family classifies an error's behavioral profile for automated handling.
//
// One concept serving three audiences:
//   - Retry loops: "Should I try again?" (Transient = yes)
//   - Exit codes: "Which exit code for the shell?" (maps to BSD sysexits.h)
//   - Presentation: "Whose fault is it?" (determines tone and framing in user messages)
type Family int

const (
	// Rejection indicates bad input, unauthorized access, or resource not found.
	// No state changed. Not retryable. User's fault. Tone: helpful, instructional.
	Rejection Family = iota

	// Conflict indicates version mismatch, duplicate creation, or state machine violation.
	// No state changed for requester. Not retryable. User needs to resolve. Tone: explanatory.
	Conflict

	// Transient indicates a temporary infrastructure failure.
	// State unknown. Retryable with backoff. System's fault. Tone: reassuring.
	Transient

	// Corruption indicates the source of truth is damaged (unparseable payload, schema break).
	// Not self-healable. Not retryable. Serious. Tone: urgent, escalate to ops.
	Corruption

	// Infrastructure indicates the system cannot serve (closed, nil deps, startup failure).
	// Not retryable. System's fault. Tone: apologetic.
	Infrastructure

	// Orchestration indicates an internal coordination failure — the program's own
	// logic failed to complete an operation (rendering output, building a CLI command,
	// wiring dependencies, writing config). Not an I/O failure, not user input, not
	// a data-integrity break: a bug or misconfiguration in the program itself.
	// Not retryable. System's fault. No user-facing fix (report the bug).
	Orchestration
)

// familyInfo holds all per-family data in one place.
// Adding a new family requires exactly one entry here.
type familyInfo struct {
	Name     string
	Severity int // total order for multi-error classification: higher = worse (Transient=1 … Corruption=6)
	Exit     int
	HTTP     int // HTTP status code mapping
	Tone     Tone
	Audience Audience
	Message  string
	Why      string
	Fix      string
}

var familyData = [...]familyInfo{ //nolint:gochecknoglobals // Immutable lookup table for Family metadata.
	Rejection: {
		Name:     strRejection,
		Severity: 2, // user error — fixable by caller
		Exit:     1,
		HTTP:     400, // Bad Request
		Tone:     ToneInstructional,
		Audience: AudienceUser,
		Message:  "The request was invalid. Check your input and try again.",
		Fix:      msgCheckInput,
	},
	Conflict: {
		Name:     strConflict,
		Severity: 3, // user must resolve state before retrying
		Exit:     1,
		HTTP:     409, // Conflict
		Tone:     ToneExplanatory,
		Audience: AudienceUser,
		Message:  "A conflict was detected. Refresh and try again.",
		Fix:      msgRefreshData,
	},
	Transient: {
		Name:     strTransient,
		Severity: 1, // least bad — temporary, will likely pass on retry
		Exit:     75,
		HTTP:     503, // Service Unavailable
		Tone:     ToneReassuring,
		Audience: AudienceAll,
		Message:  "A temporary error occurred. Please try again in a few moments.",
		Why:      "This is a temporary issue. No data was lost.",
		Fix:      "Wait a moment and try again.",
	},
	Corruption: {
		Name:     strCorruption,
		Severity: 6, // worst — source of truth is damaged, data integrity at risk
		Exit:     65,
		HTTP:     500, // Internal Server Error (data integrity break is server-side)
		Tone:     ToneUrgent,
		Audience: AudienceOps,
		Message:  "Data appears to be corrupted. This requires manual intervention.",
		Why:      "Some data appears to be damaged. This requires attention.",
		Fix:      "This may require manual intervention. Check the logs for details.",
	},
	Infrastructure: {
		Name:     strInfrastructure,
		Severity: 4, // system cannot serve, but data is intact
		Exit:     69,
		HTTP:     503, // Service Unavailable
		Tone:     ToneApologetic,
		Audience: AudienceOps,
		Message:  "The service is currently unavailable. Please try again later.",
		Why:      "This is a system issue, not something you caused.",
		Fix:      "The service may be temporarily unavailable. Try again later.",
	},
	Orchestration: {
		Name:     strOrchestration,
		Severity: 5,   // internal logic bug — worse than Infrastructure, less bad than Corruption
		Exit:     70,  // EX_SOFTWARE — internal software error
		HTTP:     500, // Internal Server Error
		Tone:     ToneApologetic,
		Audience: AudienceOps,
		Message:  "An internal error occurred.",
		Why:      "An internal operation failed unexpectedly.",
		Fix:      "This is likely a bug. Please report it if the problem persists.",
	},
}

// String returns the lowercase name of the Family (e.g. "transient", "rejection").
func (f Family) String() string {
	if f.IsValid() {
		return familyData[f].Name
	}

	return strUnknown
}

// ParseFamily parses a family string, case-insensitive.
// Returns Transient for unrecognized values (fail-open for retry).
func ParseFamily(s string) Family {
	lower := strings.ToLower(s)
	for i, info := range familyData {
		if info.Name == lower {
			return Family(i)
		}
	}

	return Transient
}

// MarshalText implements encoding.TextMarshaler for YAML/JSON config.
func (f Family) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for YAML/JSON config.
// Unknown values are parsed as Transient (fail-open).
func (f *Family) UnmarshalText(
	text []byte,
) error {
	*f = ParseFamily(string(text))

	return nil
}

// IsRetryable reports whether operations that encounter this error should be retried.
func (f Family) IsRetryable() bool {
	return f == Transient
}

// IsValid reports whether the Family value is one of the six defined constants.
func (f Family) IsValid() bool {
	return f >= Rejection && f <= Orchestration
}

// Severity returns a total order across families for multi-error classification.
// Higher = worse. Used by Classify to pick the worst sub-error of an errors.Join
// result deterministically, independent of argument order.
//
//	Transient(1) < Rejection(2) < Conflict(3) < Infrastructure(4) < Orchestration(5) < Corruption(6)
//
// This preserves the fail-closed retry guarantee: if ANY sub-error is non-Transient
// (severity > 1), the joined error is non-Transient.
func (f Family) Severity() int {
	if f.IsValid() {
		return familyData[f].Severity
	}

	return 0
}

// ExitCode returns the BSD sysexits.h compatible exit code for this family.
func (f Family) ExitCode() int {
	if f.IsValid() {
		return familyData[f].Exit
	}

	return 70 // EX_SOFTWARE — internal software error
}

// HTTPStatus returns the recommended HTTP response status code for this family.
// Use at HTTP/REST boundaries to translate a classified error into a response code:
//
//	Rejection → 400, Conflict → 409, Transient → 503,
//	Corruption → 500, Infrastructure → 503, Orchestration → 500.
//
// Invalid families return 500 (Internal Server Error).
//
// # Rationale
//
// The mapping reflects "whose fault is it, and what can the client do?":
//
//   - Rejection → 400 (Bad Request): the client sent bad input, lacked
//     authorization, or asked for a missing resource. The client can fix this.
//   - Conflict → 409 (Conflict): the request conflicts with current state
//     (duplicate, version mismatch). The client must refresh and reconcile.
//   - Transient → 503 (Service Unavailable): a temporary, retryable failure.
//     The client SHOULD retry, ideally with backoff.
//   - Corruption → 500 (Internal Server Error): stored data is damaged — a
//     data-integrity break that is the server's problem, not the client's.
//     The client did nothing wrong and cannot fix it by retrying. (Earlier
//     revisions mapped this to 422; 500 is correct because 422 implies the
//     CLIENT submitted unprocessable data, which is not the case here.)
//   - Infrastructure → 503 (Service Unavailable): the system cannot serve the
//     request right now (closed, misconfigured, starting up). Retryable in
//     principle, though possibly requiring operator action.
//   - Orchestration → 500 (Internal Server Error): an internal coordination
//     failure (bug, misconfiguration). Not retryable. The client did nothing
//     wrong and cannot fix it by retrying.
//
// Corruption and Infrastructure are distinguished by severity and audience
// (see [Family.Severity] and [Family.DefaultAudience]), not by HTTP status.
func (f Family) HTTPStatus() int {
	if f.IsValid() {
		return familyData[f].HTTP
	}

	return 500
}

// DefaultMessage returns the default human-readable message for this family.
func (f Family) DefaultMessage() string {
	if f.IsValid() {
		return familyData[f].Message
	}

	return "An unexpected error occurred."
}

// DefaultWhy returns the default "why" explanation for this family.
func (f Family) DefaultWhy() string {
	if f.IsValid() {
		return familyData[f].Why
	}

	return ""
}

// DefaultFix returns the default fix suggestion for this family.
func (f Family) DefaultFix() string {
	if f.IsValid() {
		return familyData[f].Fix
	}

	return "Try again or contact support."
}

// Audience describes who should be notified about this error.
type Audience int

const (
	// AudienceUser indicates the error is the requester's concern.
	AudienceUser Audience = iota
	// AudienceOps indicates the error needs operational intervention.
	AudienceOps
	// AudienceAll indicates all parties should be notified.
	AudienceAll
)

// IsValid reports whether the Audience value is one of the three defined constants.
func (a Audience) IsValid() bool {
	return a >= AudienceUser && a <= AudienceAll
}

// String returns the lowercase name of the Audience (e.g. "user", "ops").
func (a Audience) String() string {
	if a.IsValid() {
		return audienceNames[a]
	}

	return strUnknown
}

// ParseAudience parses an audience string, case-insensitive.
// Returns AudienceUser for unrecognized values (safest default).
func ParseAudience(s string) Audience {
	lower := strings.ToLower(s)
	for a, name := range audienceNames {
		if name == lower {
			return a
		}
	}

	return AudienceUser
}

// MarshalText implements encoding.TextMarshaler for YAML/JSON config.
func (a Audience) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for YAML/JSON config.
// Unknown values are parsed as AudienceUser (safest default).
func (a *Audience) UnmarshalText(
	text []byte,
) error {
	*a = ParseAudience(string(text))

	return nil
}

var audienceNames = map[Audience]string{ //nolint:gochecknoglobals // Immutable lookup table.
	AudienceUser: strUser,
	AudienceOps:  strOps,
	AudienceAll:  strAll,
}

// Audience returns who should be notified about errors of this family.
func (f Family) Audience() Audience {
	if f.IsValid() {
		return familyData[f].Audience
	}

	return AudienceOps
}

// Tone is a presentation tone hint for error messages.
// Used by the presentation layer to choose appropriate language.
type Tone string

const (
	// ToneInstructional guides the user on how to fix their input.
	ToneInstructional Tone = "instructional"
	// ToneExplanatory explains what happened and why.
	ToneExplanatory Tone = "explanatory"
	// ToneReassuring reassures the user that the issue is temporary.
	ToneReassuring Tone = "reassuring"
	// ToneUrgent signals that immediate action is required.
	ToneUrgent Tone = "urgent"
	// ToneApologetic indicates the system is at fault.
	ToneApologetic Tone = "apologetic"
)

// Tone returns the presentation tone hint for this family.
func (f Family) Tone() Tone {
	if f.IsValid() {
		return familyData[f].Tone
	}

	return ToneApologetic
}
