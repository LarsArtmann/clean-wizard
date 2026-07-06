package cleaner

import (
	"context"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
	errorfamily "github.com/larsartmann/go-error-family"
)

// CleanerBase holds the common fields shared by all cleaner implementations.
type CleanerBase struct {
	verbose bool
	dryRun  bool
}

// NewCleanerBase creates a CleanerBase with the given settings.
func NewCleanerBase(verbose, dryRun bool) CleanerBase {
	return CleanerBase{verbose: verbose, dryRun: dryRun}
}

// GetVerbose returns the verbose setting.
func (cb CleanerBase) GetVerbose() bool { return cb.verbose }

// GetDryRun returns the dryRun setting.
func (cb CleanerBase) GetDryRun() bool { return cb.dryRun }

// Cleaner defines the interface for all cleaner implementations.
type Cleaner interface {
	// Name returns the unique identifier for this cleaner.
	// Used for result tracking and registry operations.
	Name() string

	// Type returns the operation type for this cleaner.
	// Used for categorization and filtering.
	Type() domain.OperationType

	// Clean executes the cleaning operation and returns the result.
	Clean(ctx context.Context) result.Result[domain.CleanResult]

	// IsAvailable checks if the cleaner can run on this system.
	IsAvailable(ctx context.Context) bool

	// Scan scans for cleanable items and returns them as scan items.
	// This is used for dry-run estimation and preview functionality.
	Scan(ctx context.Context) result.Result[[]domain.ScanItem]
}

// NixStoreSizer defines the interface for cleaners that can report store size.
type NixStoreSizer interface {
	// GetStoreSize returns the size of the Nix store in bytes.
	// Returns 0 if Nix is not available.
	GetStoreSize(ctx context.Context) int64
}

// AgeBasedCleaner extends Cleaner with age-based filtering capabilities.
// Cleaners that implement this interface can filter items by age during cleanup.
type AgeBasedCleaner interface {
	Cleaner

	// SetMaxAge sets the maximum age for items to be considered for cleanup.
	// Items older than this duration will be cleaned.
	SetMaxAge(duration time.Duration)

	// GetMaxAge returns the current maximum age setting.
	GetMaxAge() time.Duration

	// SupportsAgeFiltering returns true if this cleaner supports age-based filtering.
	SupportsAgeFiltering() bool
}

// NotAvailableError marks errors that indicate a cleaner's prerequisite (e.g. a
// package manager binary) is not present on the system. The execution layer uses
// this to classify a step as "skipped" rather than "failed".
//
// Implements errorfamily.Classified (Infrastructure) and errorfamily.Coded so
// that errorfamily.Classify and errorfamily.IsRetryable work without any
// keyword matching or registration.
type NotAvailableError struct {
	CleanerName string
	Reason      string
	Code        string
}

func (e *NotAvailableError) Error() string {
	if e.Reason != "" {
		return e.CleanerName + " not available: " + e.Reason
	}
	return e.CleanerName + " not available"
}

// ErrorCode returns the machine-readable code for this error type.
// If Code is set (e.g. "cleaner.cargo.not_installed"), it takes precedence;
// otherwise the generic "cleaner.not_available" is returned.
func (e *NotAvailableError) ErrorCode() string {
	if e.Code != "" {
		return e.Code
	}
	return "cleaner.not_available"
}

// ErrorFamily classifies this as Infrastructure — the system cannot serve
// because a prerequisite is missing. Infrastructure is not retryable.
func (e *NotAvailableError) ErrorFamily() errorfamily.Family {
	return errorfamily.Infrastructure
}

// NewNotAvailableError constructs a NotAvailableError with a per-cleaner
// diagnostic code (e.g. "cleaner.cargo.not_available"). Using this factory
// ensures code consistency across all cleaner call sites.
func NewNotAvailableError(cleanerName, reason string) *NotAvailableError {
	return &NotAvailableError{
		CleanerName: cleanerName,
		Reason:      reason,
		Code:        "cleaner." + cleanerName + ".not_available",
	}
}

// IsNotAvailableError reports whether err represents a cleaner that is not
// installed or not applicable on the current system. Uses errorfamily.Classify
// so that both typed *NotAvailableError and registered sentinels (e.g.
// exec.ErrNotFound) are detected without keyword matching.
func IsNotAvailableError(err error) bool {
	if err == nil {
		return false
	}
	return errorfamily.Classify(err) == errorfamily.Infrastructure
}
