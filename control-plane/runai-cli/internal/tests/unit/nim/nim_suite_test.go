package nim_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNim(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI NIM Tests Suite")
}
