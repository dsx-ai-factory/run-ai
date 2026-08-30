package credential

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCredential(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "credential CLI tests")
}
