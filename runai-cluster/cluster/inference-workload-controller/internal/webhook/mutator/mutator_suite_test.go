package mutator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMutator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mutator Suite")
}
