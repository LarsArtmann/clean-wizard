package projectexecutables

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestProjectExecutablesSuite(t *testing.T) {
	t.Parallel()
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ProjectExecutables Suite")
}
