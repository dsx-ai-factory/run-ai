package iw_tw_workloads

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkloadsComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "iw_tw_workloads composer tests suite")
}
