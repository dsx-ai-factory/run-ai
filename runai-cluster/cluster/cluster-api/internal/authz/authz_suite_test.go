package authz

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuthz(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authz Suite")
}
