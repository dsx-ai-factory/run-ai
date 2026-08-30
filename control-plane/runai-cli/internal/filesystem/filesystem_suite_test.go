package filesystem

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFileSystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "File system suite")
}
