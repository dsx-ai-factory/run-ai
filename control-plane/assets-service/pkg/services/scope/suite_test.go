package scope

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResponseScope(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scope suite")
}
