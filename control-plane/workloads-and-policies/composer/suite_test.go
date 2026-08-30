package composer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "composer tests suite")
}
