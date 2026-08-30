package submit_test

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	cptemplatemocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/template"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("Submit with Template Flag Tests", Ordered, func() {
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

		// Setup mocks
		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		// Setup for template mocks
		templates := map[string]string{
			"valid-template":    "template-id-123",
			"training-template": "template-id-456",
		}
		mockLister := cptemplatemocks.NewMockTemplateLister(mockCtrl)
		mockDescribeLister := cptemplatemocks.NewMockTemplateDescribeLister(mockCtrl)

		mockLister.EXPECT().
			List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*workloads_openapi.TemplatesListResponse, error) {
				var templateName string
				for _, filter := range filterBy {
					if len(filter) > 6 && filter[:6] == "name==" {
						templateName = filter[6:]
						break
					}
				}

				if templateId, exists := templates[templateName]; exists {
					return &workloads_openapi.TemplatesListResponse{
						Entries: []workloads_openapi.SupersetTemplate{
							{
								Meta: workloads_openapi.TemplateMetaFields{
									Id:        templateId,
									Name:      templateName,
									ScopeType: workloads_openapi.SCOPETYPE_PROJECT,
								},
							},
						},
					}, nil
				}

				return nil, fmt.Errorf("template '%s' does not exist", templateName)
			}).AnyTimes()

		mockDescribeLister.EXPECT().
			List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*workloads_openapi.TemplatesListResponse, error) {
				var templateName string
				for _, filter := range filterBy {
					if len(filter) > 6 && filter[:6] == "name==" {
						templateName = filter[6:]
						break
					}
				}

				if templateId, exists := templates[templateName]; exists {
					return &workloads_openapi.TemplatesListResponse{
						Entries: []workloads_openapi.SupersetTemplate{
							{
								Meta: workloads_openapi.TemplateMetaFields{
									Id:        templateId,
									Name:      templateName,
									ScopeType: workloads_openapi.SCOPETYPE_PROJECT,
								},
							},
						},
					}, nil
				}

				return nil, fmt.Errorf("template '%s' does not exist", templateName)
			}).AnyTimes()

		// Inject the mock into the template package
		template.SetClient(mockDescribeLister)

		clients.NewWorkloadSubmitMocker(mockCtrl, &clients.Training{}, []string{"workload-with-template", "workload-no-template"})
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

	Describe("Template flag registration for all workload types", func() {
		workloadTypes := []struct {
			name    string
			command string
		}{
			{"training", "training submit -h"},
			{"workspace", "workspace submit -h"},
			{"inference", "inference submit -h"},
			{"pytorch", "training pytorch submit -h"},
			{"mpi", "training mpi submit -h"},
			{"tensorflow", "training tf submit -h"},
			{"xgboost", "training xgboost submit -h"},
		}

		for _, wl := range workloadTypes {
			wl := wl
			It(fmt.Sprintf("should show --template flag in help for %s", wl.name), func() {
				output, err := cli.RunCommand(wl.command, envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring("--template"),
					fmt.Sprintf("Command '%s' should have --template flag", wl.command))
			})
		}
	})

	Describe("Template submission tests", func() {
		Context("submit with valid template", func() {
			It("should successfully submit workload with valid template", func() {
				output, err := cli.RunCommand("training submit workload-with-template -p p301 --template valid-template --image ubuntu", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring("Creating training workload-with-template"))
			})
		})

		Context("Submit with non-existent template", func() {
			It("should return error when template does not exist", func() {
				output, err := cli.RunCommand("training submit test-fail -p p301 --template non-existent-template --image ubuntu", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(ContainSubstring("template 'non-existent-template' does not exist"))
			})
		})

		Context("Submit with empty template", func() {
			It("should return error when template flag is set with empty value", func() {
				output, err := cli.RunCommandStringSplit(context.TODO(), []string{"training", "submit", "test-empty-template", "-p", "p301", "--template", "", "--image", "ubuntu"}, envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(ContainSubstring("flag --template requires a non-empty value"))
			})
		})
	})
})
