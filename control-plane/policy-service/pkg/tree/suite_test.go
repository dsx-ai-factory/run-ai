package tree

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
)

func TestTree(t *testing.T) {
	RegisterFailHandler(Fail)

	config.MockConfig()

	RunSpecs(t, "tree suite")
}
