package platform_identifier

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFlavors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flavors Suite")
}
