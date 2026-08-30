package translation

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompliance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Translation Suite")
}
