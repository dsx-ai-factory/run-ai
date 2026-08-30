package nodepool_controller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodePoolController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodePool Controller Suite")
}
