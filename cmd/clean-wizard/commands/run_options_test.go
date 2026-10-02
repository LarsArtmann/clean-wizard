package commands

import (
	"testing"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/progress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProgressRequested_Gate verifies the ADR-0002 gating contract at its CLI
// seam: machine-output modes (--json/--sarif) can never activate live progress,
// which is what guarantees zero progress bytes in their streams.
func TestProgressRequested_Gate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		flag          bool
		machineOutput bool
		want          bool
	}{
		{name: "flag off stays off", flag: false, machineOutput: false, want: false},
		{name: "json mode blocks progress", flag: true, machineOutput: true, want: false},
		{name: "sarif mode blocks progress", flag: true, machineOutput: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, progressRequested(tt.flag, tt.machineOutput))
		})
	}
}

// TestProgressRequested_NonTTYWriterIsDisabled asserts the CI/non-interactive
// half of the gate: under `go test`, stdout is a pipe, so Enabled must be
// false even with the flag set and no machine output.
func TestProgressRequested_NonTTYWriterIsDisabled(t *testing.T) {
	t.Parallel()

	assert.False(t, progressRequested(true, false),
		"go test stdout is never a TTY; progress must stay disabled")
	assert.False(t, progress.Enabled(nil), "nil writer must be reported disabled")}

// fakeProgressEmitter is a minimal non-nil emitter stand-in for option wiring.
type fakeProgressEmitter struct{}

func (fakeProgressEmitter) WorkflowStarted(string)               {}
func (fakeProgressEmitter) ActivityRegistered(string)            {}
func (fakeProgressEmitter) ActivityStarted(string)               {}
func (fakeProgressEmitter) ActivityRetrying(string, int, string) {}
func (fakeProgressEmitter) ActivityCompleted(string, time.Duration) {
}
func (fakeProgressEmitter) ActivityFailed(string, error, time.Duration) {
}
func (fakeProgressEmitter) ActivitySkipped(string, string) {}
func (fakeProgressEmitter) Note(string)                    {}
func (fakeProgressEmitter) WorkflowFinished(error)         {}
func (fakeProgressEmitter) Finish()                        {}

// TestBuildRunOptions_EmitterAccepted verifies both emitter states build a
// valid option list; the option's semantics (no-op vs emitting) are owned and
// tested by the execution layer.
func TestBuildRunOptions_EmitterAccepted(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		emitter     execution.ProgressEmitter
		wantOptions int
	}{
		{name: "nil emitter yields no options", emitter: nil, wantOptions: 0},
		{name: "non-nil emitter yields exactly the progress option", emitter: fakeProgressEmitter{}, wantOptions: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts, err := buildRunOptions(false, 0, 0, "", tc.emitter)
			require.NoError(t, err)
			assert.Len(t, opts, tc.wantOptions)
		})
	}
}
