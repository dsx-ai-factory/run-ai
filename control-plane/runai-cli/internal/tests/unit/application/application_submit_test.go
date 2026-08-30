package application_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
)

const (
	ApplicationSubmitTestFilePath = "application-submit"
	chartURL                      = "https://charts.example.com/grafana-11.0.0.tgz"
)

var _ = Describe("Application Submit Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, ApplicationSubmitTestFilePath)
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

	Context("application submit validation", Ordered, func() {
		It("submit with a chart URL creates the application", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application submit my-app -p %s --chart-url %s", clients.ProjectOneName, chartURL), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Creating application my-app..."))
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("runai application describe my-app -p %s", clients.ProjectOneName)))

			request := clients.LastAIApplicationCreationRequest
			Expect(request.Name).To(Equal("my-app"))
			Expect(request.ProjectId).To(Equal(clients.ProjectOneID))
			Expect(request.Spec.ChartRef.GetChartURL()).To(Equal(chartURL))
			Expect(request.Spec.Values).To(BeNil())
			Expect(request.Spec.Auth).To(BeNil())
		})

		It("submit with values file and set overrides merges them into spec.values", func() {
			valuesPath := filepath.Join(GinkgoT().TempDir(), "values.yaml")
			Expect(os.WriteFile(valuesPath, []byte("admin:\n  password: fromfile\nreplicas: 2\n"), 0o600)).To(Succeed())

			_, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-values -p %s --chart-url %s --values %s --set admin.password=secret123",
				clients.ProjectOneName, chartURL, valuesPath), envMap)
			Expect(err).To(BeNil())

			values := clients.LastAIApplicationCreationRequest.Spec.Values
			Expect(values).To(HaveKeyWithValue("replicas", 2))
			Expect(values["admin"]).To(HaveKeyWithValue("password", "secret123"))
		})

		It("submit with a credential resolves it to an asset id", func() {
			_, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-cred -p %s --chart-url %s --credential %s",
				clients.ProjectOneName, chartURL, clients.CredentialName), envMap)
			Expect(err).To(BeNil())

			request := clients.LastAIApplicationCreationRequest
			Expect(request.Spec.Auth).NotTo(BeNil())
			Expect(request.Spec.Auth.GetAssetId()).To(Equal(clients.CredentialAssetID))
		})

		It("submit with an unknown credential fails with a not-found error", func() {
			output, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-badcred -p %s --chart-url %s --credential no-such-cred",
				clients.ProjectOneName, chartURL), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(`credential "no-such-cred" not found`))
		})

		It("submit with an ambiguous credential name fails", func() {
			output, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-dupcred -p %s --chart-url %s --credential %s",
				clients.ProjectOneName, chartURL, clients.CredentialNameAmbiguous), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("multiple credentials named %q found", clients.CredentialNameAmbiguous)))
		})

		It("submit with a non-NGC credential kind fails as not found", func() {
			output, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-dockercred -p %s --chart-url %s --credential %s",
				clients.ProjectOneName, chartURL, clients.CredentialNameDockerRegistry), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("credential %q not found", clients.CredentialNameDockerRegistry)))
		})

		It("submit without --chart-url fails as a required flag", func() {
			_, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-nochart -p %s", clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("chart-url"))
		})

		It("submit with an invalid values file fails on YAML validation", func() {
			valuesPath := filepath.Join(GinkgoT().TempDir(), "bad.yaml")
			Expect(os.WriteFile(valuesPath, []byte("a:\n\tb: [broken\n"), 0o600)).To(Succeed())

			output, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-badvals -p %s --chart-url %s --values %s",
				clients.ProjectOneName, chartURL, valuesPath), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("is not valid YAML"))
		})

		It("submit with an unknown project fails", func() {
			output, err := cli.RunCommand(fmt.Sprintf(
				"application submit my-app-badproj -p p300 --chart-url %s", chartURL), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("p300"))
		})
	})
})
