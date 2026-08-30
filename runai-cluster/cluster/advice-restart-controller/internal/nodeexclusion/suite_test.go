package nodeexclusion

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodeExclusion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Exclusion Suite")
}
