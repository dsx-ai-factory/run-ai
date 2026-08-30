package workload_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestWorkloadController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Controller Integration Tests Suite")
}
