package dinf_to_lws

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDinfToLWSComposerSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dinf_to_lws composer tests suite")
}
