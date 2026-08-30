package assets

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAssets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Assets Suite")
}
