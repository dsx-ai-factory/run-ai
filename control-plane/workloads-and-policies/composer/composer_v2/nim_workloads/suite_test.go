package nim_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestNimServiceComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "nim service composer tests suite")
}
