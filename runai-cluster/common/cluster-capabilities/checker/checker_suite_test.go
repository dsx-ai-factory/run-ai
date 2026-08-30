package checker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestChecker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Checker Suite")
}
