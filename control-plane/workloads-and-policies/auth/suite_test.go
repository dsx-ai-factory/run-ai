package auth

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAuthUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "auth utils suite")
}
