package cluster

import (
	"fmt"
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Cluster Test")
}

const (
	TestFilePath        = "cluster"
	BadStringFormatting = "%!w("
)

var _ = Describe("Cluster Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)

	})

	BeforeEach(func() {
		cluster.Clear()
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Cluster command validation", Ordered, func() {

		It("cluster list", func() {
			output, err := cli.RunCommand("cluster list", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(cli.ClusterOneName))
			Expect(output).To(ContainSubstring(cli.ClusterOneDomain))
		})

		It("cluster set by name", func() {
			output, err := cli.RunCommand(fmt.Sprintf("cluster set %s", cli.ClusterOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Cluster %s set successfully\n", cli.ClusterOneName)))
		})

		It("cluster set by uuid", func() {
			output, err := cli.RunCommand(fmt.Sprintf("cluster set --id=%s", cli.ClusterOneUuid), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Cluster %s set successfully\n", cli.ClusterOneName)))
		})

		It("cluster set by name quiet mode", func() {
			output, err := cli.RunCommand(fmt.Sprintf("cluster set %s --quiet", cli.ClusterOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(BeEmpty())
		})
	})
})
