package finalizers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFinalizers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Finalizers Suite")
}
