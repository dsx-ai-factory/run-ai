package application_test

import (
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
)

const ApplicationDeleteTestFilePath = "application-delete"

var _ = Describe("Application Delete Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, ApplicationDeleteTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.MockAIApplication(mockCtrl)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Pre command validation", Ordered, func() {
		It("login validation", func() {
			err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})
	})

	Context("application delete validation", Ordered, func() {
		It("delete with -y succeeds and skips the prompt", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p %s -y", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("Application %q deleted successfully.", clients.AppName01)))
		})

		It("delete with --force succeeds and skips the prompt", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p %s --force", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("Application %q deleted successfully.", clients.AppName01)))
		})

		It("delete of a missing application fails with a not-found error", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p %s -y", clients.AppNameMissing, clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("application %q not found in project %q", clients.AppNameMissing, clients.ProjectOneName)))
		})

		It("delete of an application with actionStatus deleting fails as already being deleted", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p %s -y", clients.AppNameDeleting, clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("application %q is already being deleted", clients.AppNameDeleting)))
		})

		It("delete of an application with desiredPhase deleted fails as already being deleted", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p %s -y", clients.AppNameDesiredDeleted, clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("application %q is already being deleted", clients.AppNameDesiredDeleted)))
		})

		It("delete with an unknown project fails", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -p p300 -y", clients.AppName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("p300"))
		})

		It("delete without a project fails with a clear message", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application delete %s -y", clients.AppName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("project"))
		})

		It("delete without an application name fails", func() {
			_, err := cli.RunCommand("application delete -y", envMap)
			Expect(err).ToNot(BeNil())
		})
	})
})
