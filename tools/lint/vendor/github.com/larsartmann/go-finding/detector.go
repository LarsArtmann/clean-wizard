package finding

import (
	"context"
	"fmt"
	"os/exec"
)

// Detector is the interface implemented by tools that can find issues.
type Detector interface {
	// Name returns the detector's name.
	Name() string
	// Detect runs the detector and returns findings.
	Detect(ctx context.Context) ([]Finding, error)
}

// DetectorFunc is an adapter to use ordinary functions as Detectors.
type DetectorFunc func(ctx context.Context) ([]Finding, error)

// Detect implements Detector.
func (f DetectorFunc) Detect(ctx context.Context) ([]Finding, error) {
	return f(ctx)
}

// Name implements Detector. Returns "" — use NamedDetectorFunc for a named detector.
//
//nolint:revive // receiver unused by design — method exists only to satisfy Detector interface
func (f DetectorFunc) Name() string {
	return ""
}

// NamedDetectorFunc returns a Detector with the given name wrapping the provided function.
//
//nolint:ireturn // intentional: factory function returns interface for polymorphism
func NamedDetectorFunc(name string, fn DetectorFunc) Detector {
	return &namedDetector{name: name, fn: fn}
}

type namedDetector struct {
	name string
	fn   DetectorFunc
}

func (n *namedDetector) Detect(ctx context.Context) ([]Finding, error) {
	return n.fn(ctx)
}

func (n *namedDetector) Name() string {
	return n.name
}

// CheckBinary verifies that a binary exists in the system PATH.
// Returns the full path to the binary on success, or a wrapped NewIOError
// if the binary is not found. This standardizes the "run CLI tool → parse JSON"
// pattern that 4+ consumers independently implement.
func CheckBinary(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", NewIOError(fmt.Sprintf("binary %q not found in PATH", name), err)
	}

	return path, nil
}

// RunCmd executes a command with the given context and arguments, returning
// stdout output. Returns a wrapped NewIOError if the command fails.
// Use this with CheckBinary for the standard "run CLI tool → parse JSON" pattern.
func RunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: name is caller-controlled, not user input

	output, err := cmd.Output()
	if err != nil {
		return nil, NewIOError(
			fmt.Sprintf("command %q failed", name),
			err,
		)
	}

	return output, nil
}
