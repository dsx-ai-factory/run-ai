package resources

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

const (
	RegistryTestValue = "gcr.io/run-ai-prod"
)

func TestOperand(t *testing.T) {
	t.Setenv(common.DefaultRegistryEnvVarName, RegistryTestValue)

	RegisterFailHandler(Fail)
	RunSpecs(t, "researcher service")
}
