package runaijob_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunaijob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runaijob Suite")
}
