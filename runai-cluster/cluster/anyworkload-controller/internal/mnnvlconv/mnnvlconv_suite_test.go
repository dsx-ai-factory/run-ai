package mnnvlconv

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMnnvlconv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mnnvlconv Suite")
}
