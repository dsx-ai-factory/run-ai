package legacycleanup

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLegacyWorkloadCleanup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Legacy Workload Cleanup Suite")
}
