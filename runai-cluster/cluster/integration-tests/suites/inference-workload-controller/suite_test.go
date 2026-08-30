package inference_workload_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestInferenceWorkloadController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inference Workload Controller Integration Tests Suite")
}
