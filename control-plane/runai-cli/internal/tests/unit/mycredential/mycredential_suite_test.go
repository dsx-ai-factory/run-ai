package mycredential

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMyCredential(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "mycredential CLI tests")
}
