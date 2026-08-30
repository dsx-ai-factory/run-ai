package distributed_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCrds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distributed Extension Suite")
}
