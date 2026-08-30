package factory_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Factory Suite")
}
