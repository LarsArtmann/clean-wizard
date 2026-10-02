package commands

import (
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/execution"
	"github.com/LarsArtmann/clean-wizard/internal/report"
	errorfamily "github.com/larsartmann/go-error-family"
)

// writeReportFile renders the workflow outcome as a self-contained HTML DAG
// report at path. Errors are classified per the CLI convention: the file
// cannot be created → Rejection (matches the config-save precedent), the
// render itself fails → Corruption.
func writeReportFile(path string, wr *execution.WorkflowResult, subtitle, command string) error {
	file, err := os.Create(path)
	if err != nil {
		return errorfamily.WrapRejectionf(err, command+".report_write", "cannot create report file %q", path)
	}
	defer file.Close()

	if err := report.WriteWorkflowHTML(file, wr, subtitle); err != nil {
		return errorfamily.WrapCorruption(err, command+".report_render", "failed to render HTML report")
	}

	return nil
}
