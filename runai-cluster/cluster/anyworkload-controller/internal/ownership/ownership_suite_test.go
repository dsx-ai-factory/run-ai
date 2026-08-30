package ownership_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOwnership(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Legacy Workload Ownership Suite")
}
