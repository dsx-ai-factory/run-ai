package manipulate

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestManipulate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Type Manipulate Suite")
}
