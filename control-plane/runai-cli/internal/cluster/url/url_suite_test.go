package url_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUrl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Url Suite")
}
