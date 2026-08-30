package mutators_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMutators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mutators Suite")
}
