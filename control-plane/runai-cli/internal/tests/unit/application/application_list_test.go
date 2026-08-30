package application_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
)

const ApplicationListTestFilePath = "application-list"

var _ = Describe("Application List Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, ApplicationListTestFilePath)
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

		It("cluster set", func() {
			err := clients.ClusterSet(envMap)
			if err != nil {
				Fail("cluster set failed, error: " + err.Error())
			}
		})
	})

	Context("application list validation", Ordered, func() {
		It("application list -A returns all apps", func() {
			output, err := cli.RunCommand("application list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.AppName01))
			Expect(output).To(ContainSubstring(clients.AppName02))
			Expect(output).To(ContainSubstring(clients.AppName03))
		})

		It("application list -A as json returns all apps with all fields", func() {
			output, err := cli.RunCommand("application list -A --json", envMap)
			Expect(err).To(BeNil())

			var data list.BriefData
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.AIApplications).To(HaveLen(3))
			Expect(data.NextToken).To(BeNil())

			Expect(data.AIApplications[0].Name).To(Equal(clients.AppName01))
			Expect(data.AIApplications[0].Type).To(Equal("inference"))
			Expect(data.AIApplications[0].ProjectName).To(Equal(clients.ProjectOneName))

			Expect(data.AIApplications[1].Name).To(Equal(clients.AppName02))
			Expect(data.AIApplications[1].Type).To(Equal("training"))
			Expect(data.AIApplications[1].ProjectName).To(Equal(clients.ProjectOneName))

			Expect(data.AIApplications[2].Name).To(Equal(clients.AppName03))
			Expect(data.AIApplications[2].Type).To(Equal("inference"))
			Expect(data.AIApplications[2].ProjectName).To(Equal(clients.ProjectTwoName))
		})

		It("application list -A and --project flags together should fail", func() {
			_, err := cli.RunCommand(fmt.Sprintf("application list -A --project=%s", clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("application list --project p301 returns only proj1 apps", func() {
			output, err := cli.RunCommand(fmt.Sprintf("application list --project=%s --json", clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())

			var data list.BriefData
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.AIApplications).To(HaveLen(2))
			for _, app := range data.AIApplications {
				Expect(app.ProjectName).To(Equal(clients.ProjectOneName))
			}
		})

		It("application list --project p302 returns only proj2 apps", func() {
			output, err := cli.RunCommand(fmt.Sprintf("application list --project=%s --json", clients.ProjectTwoName), envMap)
			Expect(err).To(BeNil())

			var data list.BriefData
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.AIApplications).To(HaveLen(1))
			Expect(data.AIApplications[0].Name).To(Equal(clients.AppName03))
		})

		It("application list --project unknown fails", func() {
			output, err := cli.RunCommand("application list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("p300"))
		})

		It("application list -A as yaml returns all apps", func() {
			output, err := cli.RunCommand("application list -A --yaml", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(fmt.Sprintf("name: %s", clients.AppName01)))
			Expect(output).To(ContainSubstring(fmt.Sprintf("name: %s", clients.AppName02)))
			Expect(output).To(ContainSubstring(fmt.Sprintf("name: %s", clients.AppName03)))
		})

		It("application list with default project uses config project", func() {
			err := clients.ProjectSet(clients.ProjectOneName, envMap)
			if err != nil {
				Fail("project set failed: " + err.Error())
			}
			defer func() {
				_ = clients.ClusterSet(envMap)
			}()

			output, err := cli.RunCommand("application list", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.AppName01))
			Expect(output).To(ContainSubstring(clients.AppName02))
		})

		It("application list missing project fails with clear message", func() {
			output, err := cli.RunCommand("application list", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("project"))
		})

		It("application list -A --columns invalid-column fails with invalid column error", func() {
			output, err := cli.RunCommand("application list -A --columns invalid-column", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("invalid-column"))
		})

		It("application list --set-default-columns --columns invalid-column fails early and persists nothing", func() {
			before := settings.GetInstance().GetAIApplicationColumns()

			output, err := cli.RunCommand("application list -A --columns invalid-column --set-default-columns", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("invalid-column"))

			Expect(settings.GetInstance().GetAIApplicationColumns()).To(Equal(before))
		})
	})
})
