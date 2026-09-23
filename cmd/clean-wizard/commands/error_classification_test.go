package commands

import (
	"fmt"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-error-family/errorfamilytest"
	"github.com/stretchr/testify/assert"
)

func TestCommandSentinelClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantFam   errorfamily.Family
		wantCode  string
		retryable bool
	}{
		{
			name:      "ErrGitNotAvailable is Infrastructure (git missing is environmental)",
			err:       ErrGitNotAvailable,
			wantFam:   errorfamily.Infrastructure,
			wantCode:  "githistory.git_not_available",
			retryable: false,
		},
		{
			name:      "ErrNoGitRepositoriesFound is Rejection (user-supplied path)",
			err:       ErrNoGitRepositoriesFound,
			wantFam:   errorfamily.Rejection,
			wantCode:  "githistory.no_repositories",
			retryable: false,
		},
		{
			name:      "ErrSafetyChecksFailed is Conflict (repo state blocks operation)",
			err:       ErrSafetyChecksFailed,
			wantFam:   errorfamily.Conflict,
			wantCode:  "githistory.safety_checks_failed",
			retryable: false,
		},
		{
			name:      "ErrNotAGitRepository is Rejection (bad path input)",
			err:       ErrNotAGitRepository,
			wantFam:   errorfamily.Rejection,
			wantCode:  "githistory.not_a_git_repository",
			retryable: false,
		},
		{
			name:      "ErrProfileNotFound is Rejection",
			err:       ErrProfileNotFound,
			wantFam:   errorfamily.Rejection,
			wantCode:  "clean.profile_not_found",
			retryable: false,
		},
		{
			name:      "ErrProfileNoCleaners is Rejection",
			err:       ErrProfileNoCleaners,
			wantFam:   errorfamily.Rejection,
			wantCode:  "clean.profile_no_cleaners",
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorfamilytest.AssertFamily(t, tt.err, tt.wantFam)
			errorfamilytest.AssertCode(t, tt.err, tt.wantCode)
			errorfamilytest.AssertRetryable(t, tt.err, tt.retryable)
		})
	}
}

// TestClassifiedSentinelsSurviveContextWrapping verifies that plain fmt.Errorf
// wraps that only add context still classify from the sentinel via the unwrap
// chain — the reason context-only wraps stay bare instead of family-wrapped.
func TestClassifiedSentinelsSurviveContextWrapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantFam errorfamily.Family
	}{
		{
			name:    "context wrap of ErrGitNotAvailable stays Infrastructure",
			err:     fmt.Errorf("repoPath=%v: %w", "/tmp/repo", ErrGitNotAvailable),
			wantFam: errorfamily.Infrastructure,
		},
		{
			name:    "context wrap of ErrSafetyChecksFailed stays Conflict",
			err:     fmt.Errorf("repoPath=%v: %w", "/tmp/repo", ErrSafetyChecksFailed),
			wantFam: errorfamily.Conflict,
		},
		{
			name:    "context wrap of ErrProfileNotFound stays Rejection",
			err:     fmt.Errorf("%w: %q", ErrProfileNotFound, "daily"),
			wantFam: errorfamily.Rejection,
		},
		{
			name:    "errors.Is still matches through the context wrap",
			err:     fmt.Errorf("ctx: %w", ErrGitNotAvailable),
			wantFam: errorfamily.Infrastructure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorfamilytest.AssertFamily(t, tt.err, tt.wantFam)
		})
	}
}

func TestCommandSentinelExitCodes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 69, errorfamily.ExitCode(ErrGitNotAvailable)) // EX_UNAVAILABLE: Infrastructure
	assert.Equal(t, 1, errorfamily.ExitCode(ErrNoGitRepositoriesFound))
	assert.Equal(t, 1, errorfamily.ExitCode(ErrSafetyChecksFailed)) // Conflict maps to 1
	assert.NotErrorIs(t, ErrGitNotAvailable, ErrNotAGitRepository)
}
