package patching

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestScope(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "patching suite")
}
