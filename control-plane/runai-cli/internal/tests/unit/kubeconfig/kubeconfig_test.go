package kubeconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/k8s"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/siderolabs/go-kubeconfig"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI KubeConfig Test")
}

const (
	TestFilePath = "kubeconfig"
	KubeConfig   = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://127.0.0.1:6443
  name: docker-desktop
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://fox-demo.runailabs.com:6443
  name: fox-demo
contexts:
- context:
    cluster: docker-desktop
    user: docker-desktop
  name: docker-desktop
- context:
    cluster: fox-demo
    user: runai-authenticated-user
  name: fox-demo
current-context: fox-demo
kind: Config
preferences:
  colors: true
users:
- name: admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
- name: docker-desktop
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
- name: runai-authenticated-user
  user:
    auth-provider:
      config:
        airgapped: "true"
        auth-flow: remote-browser
        client-id: runai-cli
        idp-issuer-url: https://fox-demo.runailabs.com/auth/realms/runai
        realm: runai
        redirect-uri: https://fox-demo.runailabs.com/oauth-code
      name: oidc`
)

var _ = Describe("KubeConfig Test", Ordered, func() {
	var (
		configPath     string
		kubePath       string
		kubeFile       string
		err            error
		envMap         map[string]string
		mockKubeClient *mocks.MockKubeClienter
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		kubePath = filepath.Join(configPath, ".kube")
		kubeFile = filepath.Join(kubePath, "config")
		envMap = cli.CreateEnvMap(configPath)
		envMap["HOME"] = configPath
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		mockKubeClient = mocks.NewMockKubeClienter(mockCtrl)
	})

	BeforeEach(func() {
		mockCtrl := gomock.NewController(GinkgoT())
		mockKubeClient = mocks.NewMockKubeClienter(mockCtrl)
		k8s.SetClient(mockKubeClient)
	})

	Context("Pre login kubeconfig validation", func() {
		It("kubeconfig - need to login", func() {
			output, err := cli.RunCommand("kubeconfig set --bypass-validation", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(consts.ReLoginMsg))
		})
	})

	Context("Post login kubeconfig validation", func() {
		BeforeAll(func() {
			auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		})
		Context("Pre kubeconfig command validation", func() {
			It("login validation", func() {
				err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
				if err != nil {
					Fail("authentication failed, error: " + err.Error())
				}
			})
		})

		Context("KubeConfig command validation", Ordered, func() {
			It("kubeconfig set - no config - error no directory", func() {
				err := os.MkdirAll(kubePath, os.ModePerm)
				Expect(err).To(BeNil())
				output, err := cli.RunCommand("kubeconfig set --bypass-validation", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(ContainSubstring(".kube/config: no such file or directory"))
			})

			It("kubeconfig set - empty config - error kubeconfig empty", func() {
				f, err := os.Create(kubeFile)
				defer f.Close()
				Expect(err).To(BeNil())
				output, err := cli.RunCommand("kubeconfig set --bypass-validation", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(ContainSubstring("kubeconfig current context is empty"))
			})

			It("kubeconfig set - empty config - context is empty", func() {
				f, err := os.Create(kubeFile)
				Expect(err).To(BeNil())
				defer f.Close()
				output, err := cli.RunCommand("kubeconfig set --bypass-validation", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(ContainSubstring("kubeconfig current context is empty"))
			})

			It("kubeconfig set - success", func() {
				err = clients.ClusterSet(envMap)
				f, err := os.OpenFile(kubeFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				Expect(err).To(BeNil())
				defer f.Close()
				_, err = f.WriteString(KubeConfig)
				Expect(err).To(BeNil())
				output, err := cli.RunCommand("kubeconfig set --bypass-validation", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(consts.KubeConfigSetSuccessfully))
			})

			It("kubeconfig set - token validation", func() {
				merger, err := kubeconfig.Load(kubeFile)
				Expect(err).To(BeNil())
				configContext, ok := merger.Contexts[merger.CurrentContext]
				Expect(ok).To(BeTrue())
				Expect(configContext.AuthInfo).To(Equal("runai-cli"))
				ai, ok := merger.AuthInfos[configContext.AuthInfo]
				Expect(ok).To(BeTrue())
				Expect(ai.Token).To(BeEmpty())

				Expect(ai.Exec).NotTo(BeNil())
				Expect(ai.Exec.Command).To(Equal("runai"))
				Expect(ai.Exec.Args).To(ContainElements("auth", "get-token", "-o", "kubeconfig"))
			})

			It("kubeconfig set with cluster validation - match error", func() {
				mockKubeClient.EXPECT().GetServerUrl(gomock.Any()).Return("https://fox-demo.runailabs.com:6443").AnyTimes()

				err = clients.ClusterSet(envMap)
				f, err := os.OpenFile(kubeFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				Expect(err).To(BeNil())
				defer f.Close()
				_, err = f.WriteString(KubeConfig)
				Expect(err).To(BeNil())
				output, err := cli.RunCommand("kubeconfig set", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(HavePrefix("Error: Cluster domain mismatch\nConfigured cluster domain: http://localhost.runailabs.com/\nKubernetes cluster domain: https://fox-demo.runailabs.com:6443"))
				Expect(output).To(ContainSubstring("The cluster URL in your kubeconfig does not match the cluster URL configured in the Run:ai CLI"))
			})

			It("kubeconfig set with cluster validation - matches - success", func() {
				mockKubeClient.EXPECT().GetServerUrl(gomock.Any()).Return("http://localhost.runailabs.com:6443").AnyTimes()

				_ = clients.ClusterSet(envMap)
				f, err := os.OpenFile(kubeFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				Expect(err).To(BeNil())
				defer f.Close()
				_, err = f.WriteString(KubeConfig)
				Expect(err).To(BeNil())
				output, err := cli.RunCommand("kubeconfig set", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(consts.KubeConfigSetSuccessfully))
			})
		})
	})
})
