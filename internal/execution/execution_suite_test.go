package execution_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// TestExecutionBDDSuite runs the Ginkgo behavior specs for the execution layer.
// The specs use in-memory cleaner doubles only, so they run in every mode
// (including -short) without external tools.
func TestExecutionBDDSuite(t *testing.T) {
	t.Parallel()

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Execution Layer Behavior Suite")
}
