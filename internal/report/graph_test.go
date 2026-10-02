package report_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LarsArtmann/clean-wizard/internal/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pipelineNames() []string {
	return []string{"nix", "homebrew", "docker", "cargo"}
}

func TestPipelineGraph_OrderAndParallelShape(t *testing.T) {
	t.Parallel()

	g := report.PipelineGraph(pipelineNames())

	names := make([]string, 0, len(g.Nodes()))
	for _, node := range g.Nodes() {
		names = append(names, node.ID())
	}

	assert.Equal(t, pipelineNames(), names, "nodes follow selection order")
	assert.Empty(t, g.Edges(), "parallel pipeline has no edges")
}

func TestIsSupportedGraphFormat(t *testing.T) {
	t.Parallel()

	assert.True(t, report.IsSupportedGraphFormat("mermaid"))
	assert.True(t, report.IsSupportedGraphFormat("dot"))
	assert.False(t, report.IsSupportedGraphFormat("svg"))
	assert.False(t, report.IsSupportedGraphFormat(""))
}

func TestWritePipelineGraph_Golden(t *testing.T) {
	t.Parallel()

	for _, format := range []string{report.GraphFormatMermaid, report.GraphFormatDOT} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			require.NoError(t, report.WritePipelineGraph(&buf, pipelineNames(), format))

			got := buf.String()
			goldenPath := filepath.Join("testdata", "pipeline-"+format+".golden")

			if *updateGolden {
				require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o600))
			}

			want, err := os.ReadFile(goldenPath)
			require.NoError(t, err, "golden file missing; run with -update-golden to create")

			assert.Equal(t, string(want), got, "pipeline %s render drifted from golden", format)
		})
	}
}

func TestWritePipelineGraph_UnsupportedFormatErrors(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := report.WritePipelineGraph(&buf, pipelineNames(), "svg")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported graph format")
}
