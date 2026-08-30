package workload_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Suite")
}
