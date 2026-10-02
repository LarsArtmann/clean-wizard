// Package report maps workflow outcomes onto go-output/daghtml documents
// (ADR-0002). The mapping is pure: nodes carry status color, a human tooltip
// ("freed X | items | duration"), and the failure dot; edges stay empty
// because cleaners run in parallel with no dependencies.
package report

import (
	"fmt"
	"io"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/format"
	"github.com/larsartmann/go-output/daghtml"
)

// Node colors from the daghtml theme (graph.css custom properties).
const (
	colorSucceeded = "var(--success)"
	colorSkipped   = "var(--info)"
	colorFailed    = "var(--error)"
)

// tooltipFieldSep matches daghtml's visual convention for multi-field tooltips.
const tooltipFieldSep = " | "

// WorkflowResultToDAG converts a workflow outcome into the daghtml model,
// preserving registration order so reports are deterministic.
func WorkflowResultToDAG(wr *execution.WorkflowResult) daghtml.DAG {
	nodes := make([]daghtml.Node, 0, len(wr.Steps))

	for _, step := range wr.Steps {
		nodes = append(nodes, stepToNode(step))
	}

	return daghtml.DAG{Nodes: nodes, Edges: []daghtml.Edge{}}
}

func stepToNode(step execution.StepResult) daghtml.Node {
	node := daghtml.Node{
		ID:    step.Name,
		Label: step.Name,
		Color: colorSucceeded,
	}

	switch step.Status() {
	case execution.StepStatusSucceeded:
		node.Color = colorSucceeded
		node.Tooltip = successTooltip(step)
	case execution.StepStatusSkipped:
		node.Color = colorSkipped
		node.Tooltip = "skipped: " + step.Err.Error()
	case execution.StepStatusFailed:
		node.Color = colorFailed
		node.Error = true
		node.Tooltip = "failed: " + step.Err.Error()
	}

	return node
}

func successTooltip(step execution.StepResult) string {
	return fmt.Sprintf(
		"freed %s%s%d items%s%s",
		format.Bytes(int64(step.Clean.SizeEstimate.Value())),
		tooltipFieldSep,
		step.Clean.ItemsRemoved,
		tooltipFieldSep,
		format.Duration(step.Duration),
	)
}

// WorkflowSummary builds the footer line: outcome counts plus total freed.
func WorkflowSummary(wr *execution.WorkflowResult) string {
	return fmt.Sprintf(
		"%d cleaners | %d succeeded | %d skipped | %d failed | freed %s in %s",
		len(wr.Steps),
		len(wr.Succeeded()),
		len(wr.Skipped()),
		len(wr.Failed()),
		format.Bytes(int64(wr.TotalBytesFreed)),
		format.Duration(wr.Duration),
	)
}

// WriteWorkflowHTML renders a complete, self-contained HTML report to w.
// subtitle describes the run (mode, dry-run flag, timestamp) on the page header.
func WriteWorkflowHTML(w io.Writer, wr *execution.WorkflowResult, subtitle string) error {
	dag := WorkflowResultToDAG(wr)

	return daghtml.Write(w, dag,
		daghtml.WithTitle("Clean Wizard"),
		daghtml.WithSubtitle(subtitle),
		daghtml.WithFooter(WorkflowSummary(wr)),
	)
}
