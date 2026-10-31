package errorfamily

import (
	"context"
	"errors"
	"log/slog"
)

// LogError logs an error with structured fields derived from its classification:
// family, code, retryable, exit_code, and every error-context key (prefixed with "context.").
//
// Severity mapping follows the behavioral family:
//   - Transient (retryable) → slog.LevelWarn, since these are expected to self-heal.
//   - All other families → slog.LevelError.
//
// A nil error is a no-op. Pass a nil logger to use [slog.Default].
//
//	logger := slog.Default()
//	if err := run(); err != nil {
//	    errorfamily.LogError(err, logger)
//	}
func LogError(err error, logger *slog.Logger) {
	LogErrorContext(context.Background(), err, logger)
}

// LogErrorContext is the context-accepting variant of [LogError], for
// cancellation propagation and trace correlation in instrumented services.
func LogErrorContext(ctx context.Context, err error, logger *slog.Logger) {
	if err == nil {
		return
	}

	if logger == nil {
		logger = slog.Default()
	}

	family := Classify(err)
	logErrorInternal(ctx, err, logger, family, ExitCode(err))
}

// logErrorInternal is the shared emission path for [LogError] and the Logger
// hook in [HandleErrorWithContext]. It accepts a pre-computed family and exit
// code so that HandleError does not re-classify when logging alongside human
// output.
func logErrorInternal(
	ctx context.Context,
	err error,
	logger *slog.Logger,
	family Family,
	exitCode int,
) {
	attrs := []slog.Attr{
		slog.String("family", family.String()),
		slog.String("code", Code(err)),
		slog.Bool("retryable", family.IsRetryable()),
		slog.Int("exit_code", exitCode),
	}

	if contextual, ok := errors.AsType[Contextual](err); ok {
		for k, v := range contextual.ErrorContext() {
			attrs = append(attrs, slog.String("context."+k, v))
		}
	}

	level := slog.LevelError
	if family.IsRetryable() {
		level = slog.LevelWarn
	}

	logger.LogAttrs(ctx, level, err.Error(), attrs...)
}
