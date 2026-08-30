package builders_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuilders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builders suite")
}
