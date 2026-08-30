package list_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComputeList(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute List Suite")
}
