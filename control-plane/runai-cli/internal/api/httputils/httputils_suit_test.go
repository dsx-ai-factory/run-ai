package httputils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDescribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Utils Suite")
}
