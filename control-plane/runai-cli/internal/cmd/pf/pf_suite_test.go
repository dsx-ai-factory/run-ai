package pf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pf Suite")
}
