package defaulter

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDefaulter(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Defaulter")
}
