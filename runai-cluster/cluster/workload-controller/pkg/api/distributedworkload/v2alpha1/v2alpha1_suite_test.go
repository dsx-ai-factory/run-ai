package v2alpha1_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV2alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2alpha1 Suite")
}
