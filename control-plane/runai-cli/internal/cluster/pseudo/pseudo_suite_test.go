package pseudo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPseudo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pseudo Suite")
}
