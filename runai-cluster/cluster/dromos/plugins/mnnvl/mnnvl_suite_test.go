package mnnvl

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMnnvl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MNNVL Plugin Suite")
}
