package resources

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKAIOperand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KAI Operand Suite")
}
