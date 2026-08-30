package distributedinference_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDistributedInference(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DistributedInference Suite")
}
