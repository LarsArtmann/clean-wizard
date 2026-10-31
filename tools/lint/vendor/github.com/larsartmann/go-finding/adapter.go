package finding

import (
	"context"
	"fmt"
)

// ToolRunFunc executes an external tool and returns its raw output.
// The function receives the context for cancellation and timeout propagation.
type ToolRunFunc func(ctx context.Context) ([]byte, error)

// ToolAdapter converts output from an external tool into Findings.
// It implements the Detector interface, so it can be used directly
// with the pipeline package.
//
// Type parameter O is the tool-specific output type that JSON decodes into.
type ToolAdapter[O any] struct {
	name    string
	run     ToolRunFunc
	parse   func([]byte) (O, error)
	convert func(O) ([]Finding, error)
}

// NewToolAdapter creates a ToolAdapter that:
//  1. Runs the tool via run,
//  2. Parses raw bytes into type O via parse,
//  3. Converts O into Findings via convert.
//
// All three functions must be non-nil.
func NewToolAdapter[O any](
	name string,
	run ToolRunFunc,
	parse func([]byte) (O, error),
	convert func(O) ([]Finding, error),
) *ToolAdapter[O] {
	if run == nil {
		panic("finding: ToolAdapter run function must not be nil")
	}

	if parse == nil {
		panic("finding: ToolAdapter parse function must not be nil")
	}

	if convert == nil {
		panic("finding: ToolAdapter convert function must not be nil")
	}

	return &ToolAdapter[O]{
		name:    name,
		run:     run,
		parse:   parse,
		convert: convert,
	}
}

// Name returns the tool's name.
func (a *ToolAdapter[O]) Name() string { return a.name }

// Detect runs the tool, parses its output, and returns Findings.
// Returns nil (not an empty slice) when the tool produces no output.
func (a *ToolAdapter[O]) Detect(ctx context.Context) ([]Finding, error) {
	output, err := a.run(ctx)
	if err != nil {
		return nil, fmt.Errorf("tool %s: %w", a.name, err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	parsed, err := a.parse(output)
	if err != nil {
		return nil, fmt.Errorf("tool %s parse: %w", a.name, err)
	}

	findings, err := a.convert(parsed)
	if err != nil {
		return nil, fmt.Errorf("tool %s convert: %w", a.name, err)
	}

	return findings, nil
}
