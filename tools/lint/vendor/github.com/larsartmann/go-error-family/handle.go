package errorfamily

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Error code constants to satisfy goconst linter.
const (
	codeFileNotFound      = "file.not_found"
	codePermissionDenied  = "permission.denied"
	codeDBTimeout         = "db.timeout"
	codeDBConnection      = "db.connection"
	codeDBError           = "db.error"
	codeConfigInvalid     = "config.invalid"
	codeConfigNotFound    = "config.not_found"
	codeConflict          = "conflict"
	codeValidation        = "validation"
	codeTimeout           = "timeout"
	codeConnectionRefused = "connection.refused"
	codeGitError          = "git.error"
)

// HandleResult contains the full output of handling an error at the CLI boundary.
type HandleResult struct {
	// ExitCode is the process exit code derived from the error's Family.
	ExitCode int
	// Message is the human-readable error message.
	Message string
	// SuggestedFix is an actionable suggestion for resolving the error.
	SuggestedFix string
}

// HandleConfig controls how HandleError processes an error at the CLI boundary.
type HandleConfig struct {
	// Output is where human-readable messages are written. Defaults to os.Stderr.
	Output io.Writer

	// Registry provides classification sentinels and message templates.
	// If nil, DefaultRegistry is used. Set this for test isolation or
	// scoped error handling within a single binary.
	Registry *Registry

	// TemplateOverride overrides the default message template for a specific error code.
	// Takes precedence over registered templates.
	// map[errorCode]MessageTemplate
	TemplateOverride map[string]MessageTemplate

	// DiagnosticFunc runs diagnostics for the error. If nil, no diagnostics run.
	// When set, diagnostics run automatically — no separate enable flag needed.
	DiagnosticFunc DiagnosticFunc

	// OnDiagnosed is called after diagnostics complete, before exit.
	// Receives the error and diagnostic findings. Useful for logging/metrics.
	OnDiagnosed func(err error, findings []DiagnosticFinding)

	// Logger, when set, receives a structured log entry (family, code,
	// retryable, exit_code, context.*) in the same call as human output —
	// classifying once, logging once. Nil (default) skips structured logging,
	// preserving the original behavior.
	//
	// This is the recommended hook for systemd/journald observability:
	// every HandleError call emits a self-contained structured record without
	// a separate LogError call.
	Logger *slog.Logger
}

// DiagnosticFunc runs diagnostics for an error and returns results.
// The diagnose.Runner satisfies this via its Run method.
type DiagnosticFunc func(ctx context.Context, err error) []DiagnosticFinding

// DiagnosticFinding is a minimal result type for the CLI boundary.
// Avoids importing diagnose while preserving type safety.
type DiagnosticFinding struct {
	// RuleName identifies which diagnostic rule produced this finding.
	RuleName string
	// Status is the check outcome: "healthy", "degraded", "failed", or "unknown".
	Status string
	// Summary is a human-readable explanation of the finding.
	Summary string
	// SuggestedFix is an actionable resolution suggested by the rule.
	SuggestedFix string
	// Confidence is 0.0–1.0 indicating how likely this explains the error.
	Confidence float64
}

// MessageTemplate defines the Wix-style presentation for an error code.
// Based on the Wix UX framework: What / Why / Fix / WayOut.
type MessageTemplate struct {
	// What describes what happened. Supports {key} placeholders from error context.
	What string
	// Why explains why it happened. Supports {key} placeholders.
	Why string
	// Fix suggests how to resolve the error. Supports {key} placeholders.
	Fix string
	// WayOut provides an escape hatch or alternative action. Supports {key} placeholders.
	WayOut string
}

// HandleError is the CLI boundary handler — the meta service.
//
// Call this exactly once at the top of main():
//
//	func main() {
//	    if err := run(); err != nil {
//	        os.Exit(errorfamily.HandleError(err))
//	    }
//	}
//
// It:
//  1. Classifies the error (Family → exit code)
//  2. Extracts Code and Context for specific messaging
//  3. Optionally runs diagnostic rules
//  4. Formats a Wix-quality message for the user
//  5. Writes to stderr and returns the exit code
func HandleError(err error) int {
	return HandleErrorWithConfig(err, HandleConfig{})
}

// HandleErrorWithConfig is the configurable version of HandleError.
func HandleErrorWithConfig(err error, cfg HandleConfig) int {
	return HandleErrorWithContext(context.Background(), err, cfg)
}

// HandleErrorWithContext handles an error with caller-provided context for
// cancellation and diagnostic propagation. This is the preferred entry point
// when the caller has a context.Context available.
//
// All other HandleError variants delegate to this function.
func HandleErrorWithContext(ctx context.Context, err error, cfg HandleConfig) int {
	if err == nil {
		return 0
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	reg := cfg.Registry
	if reg == nil {
		reg = DefaultRegistry
	}

	family := reg.Classify(err)
	exitCode := resolveExitCode(err, family)

	code := extractCode(err)
	errCtx := extractContext(err)

	if cfg.DiagnosticFunc != nil {
		findings := cfg.DiagnosticFunc(ctx, err)
		if cfg.OnDiagnosed != nil {
			cfg.OnDiagnosed(err, findings)
		}
	}

	if cfg.Logger != nil {
		logErrorInternal(ctx, err, cfg.Logger, family, exitCode)
	}

	message := renderCLI(code, errCtx, family, cfg, reg)

	_, _ = fmt.Fprintln(
		cfg.Output,
		message,
	) //nolint:legacyerrors // best-effort CLI output: HandleError returns an exit code, a failed write has no recovery path

	return exitCode
}

// HandleErrorDetailed returns a structured result without writing output.
// Useful for HTTP handlers, gRPC interceptors, and programmatic consumers.
//
// Uses the same template resolution chain as HandleError: registered templates,
// built-in defaults, and family fallbacks.
func HandleErrorDetailed(err error) *HandleResult {
	return HandleErrorDetailedWithConfig(err, HandleConfig{})
}

// HandleErrorDetailedWithConfig returns a structured result with template overrides.
func HandleErrorDetailedWithConfig(err error, cfg HandleConfig) *HandleResult {
	if err == nil {
		return &HandleResult{ExitCode: 0}
	}

	reg := cfg.Registry
	if reg == nil {
		reg = DefaultRegistry
	}

	family := reg.Classify(err)
	code := extractCode(err)
	errCtx := extractContext(err)

	result := &HandleResult{
		ExitCode: resolveExitCode(err, family),
		Message:  renderCLI(code, errCtx, family, cfg, reg),
	}

	if !family.IsRetryable() {
		result.SuggestedFix = resolveSuggestedFix(code, errCtx, cfg, reg, family)
	}

	return result
}

// resolveExitCode checks for a custom [ExitCoder] override on the error chain
// first, then falls back to the family-based default.
func resolveExitCode(err error, family Family) int {
	if ec, ok := errors.AsType[ExitCoder](err); ok {
		if code := ec.ExitCode(); code != 0 {
			return code
		}
	}

	return family.ExitCode()
}

func extractCode(err error) string {
	return Code(err)
}

func extractContext(err error) map[string]string {
	if contextual, ok := errors.AsType[Contextual](err); ok {
		return contextual.ErrorContext()
	}

	return map[string]string{}
}

// resolveTemplate walks the shared template-resolution chain (per-call override →
// registry → built-in default) and returns the first matching template.
// renderCLI and resolveSuggestedFix both build on this so the resolution order
// can never diverge between message rendering and fix suggestion.
func resolveTemplate(code string, cfg HandleConfig, reg *Registry) (MessageTemplate, bool) {
	if cfg.TemplateOverride != nil {
		if tmpl, ok := cfg.TemplateOverride[code]; ok {
			return tmpl, true
		}
	}

	if tmpl, ok := reg.lookupTemplate(code); ok {
		return tmpl, true
	}

	if tmpl, ok := lookupDefault(code); ok {
		return tmpl, true
	}

	return MessageTemplate{}, false
}

func renderCLI(
	code string,
	context map[string]string,
	family Family,
	cfg HandleConfig,
	reg *Registry,
) string {
	if tmpl, ok := resolveTemplate(code, cfg, reg); ok {
		return applyTemplate(tmpl, context, family)
	}

	return renderMessage(code, context, family)
}

// resolveSuggestedFix finds the Fix field from the resolved template
// (override → registry → built-in default), falling back to the family default.
// Only called for non-retryable errors. A template is treated as a cohesive
// unit: its Fix belongs with its What/Why rather than being mixed across sources.
func resolveSuggestedFix(
	code string,
	errCtx map[string]string,
	cfg HandleConfig,
	reg *Registry,
	family Family,
) string {
	if tmpl, ok := resolveTemplate(code, cfg, reg); ok && tmpl.Fix != "" {
		return applyContext(tmpl.Fix, errCtx)
	}

	return family.DefaultFix()
}

func renderMessage(code string, context map[string]string, family Family) string {
	if code == "" {
		return family.DefaultMessage()
	}

	// Code-specific template from defaults.
	if tmpl, ok := lookupDefault(code); ok {
		return applyTemplate(tmpl, context, family)
	}

	// Family fallback with code as header.
	var parts []string

	parts = append(parts, "Error: "+code)
	if why := family.DefaultWhy(); why != "" {
		parts = append(parts, why)
	}

	if fix := family.DefaultFix(); fix != "" {
		parts = append(parts, fix)
	}

	return strings.Join(parts, "\n")
}

func applyTemplate(tmpl MessageTemplate, context map[string]string, family Family) string {
	var parts []string
	if tmpl.What != "" {
		parts = append(parts, applyContext(tmpl.What, context))
	}

	if why := tmpl.Why; why != "" {
		parts = append(parts, applyContext(why, context))
	} else if why := family.DefaultWhy(); why != "" {
		parts = append(parts, why)
	}

	if tmpl.Fix != "" {
		parts = append(parts, applyContext(tmpl.Fix, context))
	}

	if tmpl.WayOut != "" {
		parts = append(parts, applyContext(tmpl.WayOut, context))
	}

	return strings.Join(parts, "\n")
}

func applyContext(template string, context map[string]string) string {
	s := template
	for k, v := range context {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}

	return s
}

func lookupDefault(code string) (MessageTemplate, bool) {
	tmpl, ok := defaultMessages[strings.ToLower(code)]

	return tmpl, ok
}

// defaultMessages maps error codes (lowercase) to human-readable messages.
// Codes are matched exactly — no substring matching.
var defaultMessages = map[string]MessageTemplate{ //nolint:gochecknoglobals // Immutable default message templates.
	codeFileNotFound: {
		What: "A required resource was not found.",
		Fix:  "Check that the path and resource name are correct.",
	},
	codePermissionDenied: {
		What: "Permission was denied.",
		Fix:  "Check file permissions or run with appropriate privileges.",
	},
	codeDBTimeout: {
		What: "The database operation timed out.",
		Fix:  "Increase the timeout or check system resources.",
	},
	codeDBConnection: {
		What: "Could not establish a database connection.",
		Fix:  "Check that the database is running and reachable.",
	},
	codeDBError: {
		What: "A database operation failed.",
		Fix:  "Check the database logs for details.",
	},
	codeConfigInvalid: {
		What: "There is a configuration issue.",
		Fix:  "Review your configuration file for errors.",
	},
	codeConfigNotFound: {
		What: "A configuration file was not found.",
		Fix:  "Check that the config file path is correct.",
	},
	codeConflict:   {What: "A conflict was detected.", Fix: msgRefreshData},
	codeValidation: {What: "Validation failed.", Fix: msgCheckInput},
	codeTimeout: {
		What: "The operation timed out.",
		Fix:  "Increase the timeout or check system resources.",
	},
	codeConnectionRefused: {
		What: "Could not establish a connection.",
		Fix:  "Check that the target service is running.",
	},
	codeGitError: {What: "A git operation failed.", Fix: "Check the git repository state."},
}

// RegisterTemplate adds a MessageTemplate for a specific error code.
// Thread-safe. Overrides any existing template for the same code.
//
// Delegates to [DefaultRegistry]. For scoped registration, use
// [Registry.RegisterTemplate] on a custom Registry.
func RegisterTemplate(code string, tmpl MessageTemplate) {
	DefaultRegistry.RegisterTemplate(code, tmpl)
}

// UnregisterTemplate removes a previously registered template.
// Thread-safe. No-op if the code has no registered template.
func UnregisterTemplate(code string) {
	DefaultRegistry.UnregisterTemplate(code)
}

// TemplateForCode resolves a [MessageTemplate] for an error code, checking
// registered templates on [DefaultRegistry] first, then built-in defaults.
// Returns (zero, false) when no template exists for the code.
//
// Delegates to [DefaultRegistry]. For scoped lookups, use
// [Registry.TemplateForCode] on a custom Registry.
func TemplateForCode(code string) (MessageTemplate, bool) {
	return DefaultRegistry.TemplateForCode(code)
}
