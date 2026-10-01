package systemcache_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSystemCacheSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SystemCache Cleaner Suite")
}
