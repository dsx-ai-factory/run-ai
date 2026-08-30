package external_workload_integrator

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExternalWorkloads(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "External Workload Integrator Integration Tests Suite")
}
