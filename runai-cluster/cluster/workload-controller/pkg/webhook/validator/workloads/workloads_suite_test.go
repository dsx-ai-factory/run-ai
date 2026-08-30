package workloads_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkloads(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workloads Suite")
}
