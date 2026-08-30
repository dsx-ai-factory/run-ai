package nodeaffinity

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNodeAffinity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodeAffinity Plugin Suite")
}
