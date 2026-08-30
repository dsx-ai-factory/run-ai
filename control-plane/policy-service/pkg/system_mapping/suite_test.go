package system_mapping_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSystemMapping(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "System Mapping Suite")
}
