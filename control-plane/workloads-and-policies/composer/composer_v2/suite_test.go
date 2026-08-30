package composer_v2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestWorkloadsComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fields composer tests suite")
}
