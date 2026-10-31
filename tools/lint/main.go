// Command lint runs the published go-humanize-linter over the repository so
// the H007 (manual byte parsing) check is reproducible in CI without /tmp
// tool state. Exit codes follow the linter contract: 0 clean, 1
// high-confidence findings, 2 only medium/low findings.
package main

import (
	"context"
	"fmt"
	"os"

	finding "github.com/larsartmann/go-finding"
	humanizelint "github.com/larsartmann/go-humanize-linter"
	"github.com/larsartmann/go-linter-sdk"
)

func main() {
	registry := humanizelint.DefaultRegistry()

	report, err := registry.Run(context.Background(), ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "humanize-lint: run failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(linter.ExitCodeByConfidence(report, finding.ConfidenceHigh))
}
