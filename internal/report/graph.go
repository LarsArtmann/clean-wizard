package report

import (
	"fmt"
	"io"

	"github.com/larsartmann/go-output"
	"github.com/larsartmann/go-output/graph"
)

// Supported graph formats for the pipeline preview.
const (
	GraphFormatMermaid = "mermaid"
	GraphFormatDOT     = "dot"
)

// PipelineGraph builds the cleaner-pipeline preview graph: one node per
// selected cleaner, no edges (cleaners run in parallel with no dependencies).
func PipelineGraph(names []string) output.Graph {
	builder := output.NewGraphBuilder()

	nodes := make([]output.GraphNode, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, output.GraphNode{
			ID:    output.NewBrandedID[output.GraphNodeIDBrand](name),
			Label: output.NewBrandedID[output.GraphNodeLabelBrand](name),
		})
	}

	builder.SetNodes(nodes)
	builder.SetEdges([]output.GraphEdge{})

	return builder.Build()
}

// IsSupportedGraphFormat reports whether format is a renderable pipeline format.
func IsSupportedGraphFormat(format string) bool {
	return format == GraphFormatMermaid || format == GraphFormatDOT
}

// WritePipelineGraph renders the pipeline preview in the requested format.
// An unsupported format is a programming error at the CLI boundary; the
// command layer validates with IsSupportedGraphFormat first.
func WritePipelineGraph(w io.Writer, names []string, format string) error {
	g := PipelineGraph(names)

	switch format {
	case GraphFormatMermaid:
		return graph.WriteMermaid(w, g, graph.WithCodeFence(true))
	case GraphFormatDOT:
		return graph.WriteDOT(w, g)
	default:
		return fmt.Errorf("unsupported graph format %q (use mermaid or dot)", format)
	}
}
