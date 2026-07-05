package cleaner

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain"
	"github.com/LarsArtmann/clean-wizard/internal/result"
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
type NotAvailableError struct {
	CleanerName string
	Reason      string
}

func (e *NotAvailableError) Error() string {
	if e.Reason != "" {
		return e.CleanerName + " not available: " + e.Reason
	}
	return e.CleanerName + " not available"
}

// IsNotAvailableError reports whether err represents a cleaner that is not
// installed or not applicable on the current system. It checks for the typed
// *NotAvailableError first, then falls back to keyword matching for errors
// that originate from exec.LookPath or the OS.
func IsNotAvailableError(err error) bool {
	if err == nil {
		return false
	}

	// Fast path: typed sentinel
	var nae *NotAvailableError
	if errors.As(err, &nae) {
		return true
	}

	// Fallback: keyword match for ad-hoc errors from exec/OS
	msg := strings.ToLower(err.Error())
	for _, keyword := range notAvailableKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}

	return false
}

var notAvailableKeywords = []string{
	"not available",
	"not found",
	"not installed",
	"command not found",
	"no such file or directory",
}
