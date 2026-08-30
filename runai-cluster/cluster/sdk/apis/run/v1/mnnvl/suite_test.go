package mnnvl

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMNNVLs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "run.ai/v1 MNNVL Types")
}
