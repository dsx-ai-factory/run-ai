package describe_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComputeDescribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute Describe Suite")
}
