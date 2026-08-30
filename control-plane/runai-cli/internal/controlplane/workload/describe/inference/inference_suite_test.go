package inference_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInference(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inference Suite")
}
