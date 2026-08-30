package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommonScreenSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Screen Suite")
}
