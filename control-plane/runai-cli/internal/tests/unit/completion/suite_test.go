package completion

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompletion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI autocompletion Tests")
}
