package resources

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	suite = "Runtime Rules"
)

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}
