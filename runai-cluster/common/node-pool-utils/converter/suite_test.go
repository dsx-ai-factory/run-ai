package converter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConverter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Converter Suite")
}
