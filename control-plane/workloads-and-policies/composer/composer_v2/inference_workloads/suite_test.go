package inference_workloads

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInferenceComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "inference_workloads composer tests suite")
}
