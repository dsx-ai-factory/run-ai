package affinity

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAffinity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Affinity Utils Suite")
}
