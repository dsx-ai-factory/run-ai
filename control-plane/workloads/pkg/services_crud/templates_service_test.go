package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"

	"github.com/google/uuid"
	bc_mocks "github.com/run-ai/runai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/stretchr/testify/mock"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/org_unit"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/control-plane/workloads/pkg/testing_crud"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

const (
	testHost     = "http://localhost"
	urlTemplates = "/api/v1/workload-templates"
)

type testPerWorkloadType struct {
	nick               string
	workloadType       api.WorkloadType
	url                string
	composeCreationReq func(genericTemplate api.SupersetTemplate) *strings.Reader
}

type templateCreationReq interface {
	api.WorkspaceTemplateCreationRequest | api.TrainingTemplateCreationRequest | api.DistributedTemplateCreationRequest | api.InferenceTemplateCreationRequest | api.DistributedInferenceTemplateCreationRequest
}

func composeCreationReq[T templateCreationReq](genericTemplate api.SupersetTemplate) *strings.Reader {
	var req T

	asJson, err := json.Marshal(genericTemplate)
	Expect(err).To(BeNil())

	err = json.Unmarshal(asJson, &req)
	Expect(err).To(BeNil())

	return makePayload(req)
}

var tests = []TableEntry{
	Entry("workspaces", testPerWorkloadType{
		nick:               "ws",
		workloadType:       api.WORKLOADTYPE_WORKSPACE,
		url:                fmt.Sprintf("%s%s/workspaces", testHost, urlTemplates),
		composeCreationReq: composeCreationReq[api.WorkspaceTemplateCreationRequest],
	}),
	Entry("trainings", testPerWorkloadType{
		nick:               "tr",
		workloadType:       api.WORKLOADTYPE_TRAINING,
		url:                fmt.Sprintf("%s%s/trainings", testHost, urlTemplates),
		composeCreationReq: composeCreationReq[api.TrainingTemplateCreationRequest],
	}),
	Entry("distributed", testPerWorkloadType{
		nick:               "dist",
		workloadType:       api.WORKLOADTYPE_DISTRIBUTED,
		url:                fmt.Sprintf("%s%s/distributed", testHost, urlTemplates),
		composeCreationReq: composeCreationReq[api.DistributedTemplateCreationRequest],
	}),
	Entry("inference", testPerWorkloadType{
		nick:               "inf",
		workloadType:       api.WORKLOADTYPE_INFERENCE,
		url:                fmt.Sprintf("%s%s/inferences", testHost, urlTemplates),
		composeCreationReq: composeCreationReq[api.InferenceTemplateCreationRequest],
	}),
	Entry("distributed-inference", testPerWorkloadType{
		nick:               "dinf",
		workloadType:       api.WORKLOADTYPE_DISTRIBUTED_INFERENCE,
		url:                fmt.Sprintf("%s%s/distributed-inferences", testHost, urlTemplates),
		composeCreationReq: composeCreationReq[api.DistributedInferenceTemplateCreationRequest],
	}),
}

var _ = Describe("template validations", Serial, func() {
	var randoms *testutils.RandomValues
	var forbiddenProject *testutils.RandomValues
	var anotherAllowedProject *testutils.RandomValues

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	var authMock *testing_crud.RandomsBasedAuthMockClient

	var router chi.Router

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		forbiddenProject = randoms.CloneWithDifferentProject().Forbidden()
		anotherAllowedProject = randoms.CloneWithDifferentCluster().CloneWithDifferentDepartment().CloneWithDifferentProject().Allowed()

		authMock = testing_crud.NewRandomsBasedAuthMockClient().
			With(randoms).
			With(forbiddenProject).
			With(anotherAllowedProject)

		authorization.SetMock(authMock)
		orgunit.SetMock(authMock.OrgUnitMock())

		mockTenantSettingsClient.On("GetHybridSubmissionEnabled", mock.Anything).Return(false, nil)

		router = createRouterInt(mockTenantSettingsClient)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)).WithRandoms(randoms))

		assets.SetMock(assets.NewSimpleMock())

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("tenant scope: create -> get -> delete", func() {
		DescribeTable("affected in the list", func(testInfo testPerWorkloadType) {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        fmt.Sprintf("my-%s-template", testInfo.nick),
					Description: convutils.Ref("my-description"),
				},
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					WorkingDir:           convutils.Ref("/home/worker"),
				},
				Master: &api.TemplateMasterSection{ // for dist training
					Spec: &api.MasterTemplateSpec{
						WorkingDir: convutils.Ref("/home/master"),
					},
				},
			}

			// Distributed inference has a different structure with Leader/Worker
			if testInfo.workloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE {
				toCreate.Spec.Leader = &api.DistributedInferenceLeaderWorkerSpec{
					WorkingDir: convutils.Ref("/home/leader"),
				}
			}

			createReq, err := http.NewRequest("POST", testInfo.url, testInfo.composeCreationReq(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   list with filter of workload type
			//
			urlByType := fmt.Sprintf("filterBy=workloadType==%s", testInfo.workloadType)

			countResult := countTemplatesForTest(router, randoms, urlByType)
			Expect(countResult.Count).To(BeNumerically("==", int64(1)))

			listResult := listTemplatesForTest(router, randoms, urlByType)
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			Expect(listedTemp.Meta.Name).To(Equal(toCreate.Meta.Name))
			Expect(*listedTemp.Meta.Description).To(Equal(*toCreate.Meta.Description))
			Expect(listedTemp.Meta.ScopeType).To(Equal(toCreate.Meta.ScopeType))
			Expect(listedTemp.Meta.ScopeId).To(Equal(randoms.TenantIdStr))
			Expect(listedTemp.Meta.ClusterId).To(BeNil())
			Expect(listedTemp.Meta.DepartmentId).To(BeNil())
			Expect(listedTemp.Meta.ProjectId).To(BeNil())
			Expect(listedTemp.Meta.CreatedAt.String()).NotTo(Equal(""))
			Expect(listedTemp.Meta.UpdatedAt.String()).To(Equal(listedTemp.Meta.CreatedAt.String()))
			Expect(listedTemp.Meta.CreatedBy).To(Equal("admin"))
			Expect(*listedTemp.Meta.UpdatedBy).To(Equal("admin"))

			Expect(listedTemp.Runnable.IsRunnable).To(Equal(false))
			Expect(*listedTemp.Runnable.Issue).To(ContainSubstring("image is not provided"))

			//
			//  get should return the proper workload
			//
			getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testInfo.url, listedTemp.Meta.Id), nil)
			Expect(err).To(BeNil())
			getReq.Header = *headerForAdmin(randoms)

			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq)
			Expect(getResp.Code).To(Equal(http.StatusOK))

			var loadedTemp api.SupersetTemplate
			jsonErr := json.Unmarshal(getResp.Body.Bytes(), &loadedTemp)
			Expect(jsonErr).To(BeNil())
			Expect(loadedTemp.Meta).To(Equal(listedTemp.Meta))

			if testInfo.workloadType == api.WORKLOADTYPE_DISTRIBUTED {
				Expect(loadedTemp.Spec.WorkingDir).To(Equal(toCreate.Spec.WorkingDir))
				Expect(loadedTemp.CombinedSpec.WorkingDir).To(Equal(toCreate.Spec.WorkingDir))
				Expect(loadedTemp.Spec).To(Equal(toCreate.Spec))
				Expect(*loadedTemp.Master.Spec).To(Equal(*toCreate.Master.Spec))
				Expect(*loadedTemp.CombinedMasterSpec).To(Equal(*toCreate.Master.Spec))
			} else if testInfo.workloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE {
				// Distributed inference has Leader/Worker instead of top-level WorkingDir
				Expect(loadedTemp.Spec.Leader).NotTo(BeNil())
				Expect(loadedTemp.Spec.Leader.WorkingDir).To(Equal(toCreate.Spec.Leader.WorkingDir))
			} else {
				Expect(loadedTemp.Spec.WorkingDir).To(Equal(toCreate.Spec.WorkingDir))
				Expect(loadedTemp.CombinedSpec.WorkingDir).To(Equal(toCreate.Spec.WorkingDir))
			}

			Expect(loadedTemp.Runnable).To(Equal(listedTemp.Runnable))
		}, tests)
	})

	Context("system preset deletion", func() {
		It("authorizes the deletion for the requesting tenant", func() {
			templateId := createTestTemplate(
				0,
				"system-template-to-delete",
				api.SCOPETYPE_SYSTEM,
				"",
				nil,
				nil,
				nil,
			)

			deleteReq, err := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("%s/workspaces/%s", urlTemplates, templateId),
				nil,
			)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(randoms)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusNoContent), deleteResp.Body.String())

			var preset models_crud.TemplatesPresets
			result := dao_utils.DB().
				Where("preset_template_id = ? AND tenant_id = ?", templateId, randoms.TenantId).
				First(&preset)
			Expect(result.Error).To(BeNil())
			Expect(preset.IsDeleted).To(BeTrue())
		})
	})

	Context("system preset updates", func() {
		It("rejects public updates explicitly", func() {
			templateId := createTestTemplate(
				0,
				"system-template-to-update",
				api.SCOPETYPE_SYSTEM,
				"",
				nil,
				nil,
				nil,
			)

			newName := "updated-system-template"
			patchReq, err := http.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("%s/workspaces/%s", urlTemplates, templateId),
				makePayload(api.WorkspaceTemplatePatchRequest{
					Meta: &api.TemplatePatchMetaFields{
						Name: &newName,
					},
				}),
			)
			Expect(err).To(BeNil())
			patchReq.Header = *headerForAdmin(randoms)

			patchResp := httptest.NewRecorder()
			router.ServeHTTP(patchResp, patchReq)

			Expect(patchResp.Code).To(Equal(http.StatusForbidden), patchResp.Body.String())
			Expect(patchResp.Body.String()).To(ContainSubstring("system templates cannot be updated"))

			rec, dbErr := dao_crud.GetTemplateById(context.Background(), randoms.TenantId, templateId, nil)
			Expect(dbErr).To(BeNil())
			Expect(rec.Name).To(Equal("system-template-to-update"))
		})

		It("allows preset synchronization updates", func() {
			templateId := createTestTemplate(
				0,
				"system-template-before-preset-sync",
				api.SCOPETYPE_SYSTEM,
				"",
				nil,
				nil,
				nil,
			)

			preset := jupyterLabOneGpu
			preset.template.Meta.Name = "system-template-after-preset-sync"
			preset.template.Meta.Description = convutils.Ref("updated by preset synchronization")

			service := TemplatesService{}
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
			service.updatePreset(
				ctx,
				templateId,
				dao_crud.MakeTemplateNameAndType(preset.template.Meta.Name, string(api.WORKLOADTYPE_WORKSPACE)),
				preset,
			)

			rec, dbErr := dao_crud.GetTemplateById(context.Background(), randoms.TenantId, templateId, nil)
			Expect(dbErr).To(BeNil())
			Expect(rec.Name).To(Equal(preset.template.Meta.Name))
			Expect(rec.Description).To(Equal(preset.template.Meta.Description))
		})
	})

	Context("forbidden fields", func() {
		It("should return bad request", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_TENANT,
					Name:      "temp-with-credentials",
				},
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					EnvironmentVariables: convutils.Ref([]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("ok"),
							Value: convutils.Ref("ok-value"),
						},
						{
							Name: convutils.Ref("not-good"),
							UserCredential: convutils.Ref(api.EnvironmentVariableUserCredential{
								Name: convutils.Ref("secret-var"),
								Key:  convutils.Ref("secret-key"),
							}),
						},
					}),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						WorkingDir: convutils.Ref("/home/master"),
					},
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/distributed", testHost, urlTemplates),
				composeCreationReq[api.DistributedTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("not-good: cannot use user credentials as the value source of an environmentVariable within a template"))
		})
	})

	Context("forbidden fields - git", func() {
		It("should return bad request", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_TENANT,
					Name:      "temp-with-credentials",
				},
				Spec: api.SupersetSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					EnvironmentVariables: convutils.Ref([]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("ok"),
							Value: convutils.Ref("ok-value"),
						},
					}),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Storage: &api.WorkspaceSpecSpecAllOfStorage{
							Git: convutils.Ref([]api.GitInstance{
								{
									Name:           convutils.Ref("good"),
									Repository:     convutils.Ref("repo"),
									PasswordSecret: convutils.Ref(uuid.NewString()),
								},
								{
									Name:       convutils.Ref("not-good"),
									Repository: convutils.Ref("secret-repo"),
									SecretRef: &api.GitSecretRef{
										Name:                 "my-cred",
										AuthenticationMethod: api.GITAUTHENTICATIONMETHOD_SSH_KEY,
									},
								},
							}),
						},
					},
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/distributed", testHost, urlTemplates),
				composeCreationReq[api.DistributedTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("not-good: cannot reference user credentials within a template"))
		})
	})

	Context("template scopes", func() {
		It("cluster scope", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   randoms.ClusterId,
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   list with filter of workload type
			//
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url, "scopeType==cluster")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			Expect(listedTemp.Meta.ScopeType).To(Equal(toCreate.Meta.ScopeType))
			Expect(listedTemp.Meta.ScopeId).To(Equal(randoms.ClusterId))
			Expect(*listedTemp.Meta.ClusterId).To(Equal(randoms.ClusterId))
			Expect(listedTemp.Meta.DepartmentId).To(BeNil())
			Expect(listedTemp.Meta.ProjectId).To(BeNil())
		})

		It("department scope", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   list with filter of workload type
			//
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url, "scopeType==department")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			Expect(listedTemp.Meta.ScopeType).To(Equal(toCreate.Meta.ScopeType))
			Expect(listedTemp.Meta.ScopeId).To(Equal(randoms.DepartmentId))
			Expect(*listedTemp.Meta.ClusterId).To(Equal(randoms.ClusterId))
			Expect(*listedTemp.Meta.DepartmentId).To(Equal(randoms.DepartmentId))
			Expect(listedTemp.Meta.ProjectId).To(BeNil())
		})

		It("project scope", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   list with filter of workload type
			//
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url, "scopeType==project")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			Expect(listedTemp.Meta.ScopeType).To(Equal(toCreate.Meta.ScopeType))
			Expect(listedTemp.Meta.ScopeId).To(Equal(randoms.ProjectIdStr))
			Expect(*listedTemp.Meta.ClusterId).To(Equal(randoms.ClusterId))
			Expect(*listedTemp.Meta.DepartmentId).To(Equal(randoms.DepartmentId))
			Expect(uint(*listedTemp.Meta.ProjectId)).To(Equal(randoms.ProjectId))
		})

		It("include ancestors", func() {
			sameTenant := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_TENANT,
					ScopeId:   randoms.TenantIdStr,
					Name:      fmt.Sprintf("t"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameCluster := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   randoms.ClusterId,
					Name:      fmt.Sprintf("c"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameDept := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      fmt.Sprintf("d"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameProj := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
					Name:      fmt.Sprintf("p"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			anotherCluster := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   anotherAllowedProject.ClusterId,
					Name:      fmt.Sprintf("xc"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			anotherDept := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   anotherAllowedProject.DepartmentId,
					Name:      fmt.Sprintf("xd"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			anotherProj := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   anotherAllowedProject.ProjectIdStr,
					Name:      fmt.Sprintf("xp"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}
			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			for _, temp := range []api.SupersetTemplate{
				sameTenant,
				sameCluster,
				sameProj,
				sameDept,
				anotherCluster,
				anotherDept,
				anotherProj,
			} {
				createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](temp))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			//
			//   list with filter of specific project
			//
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url,
				fmt.Sprintf("projectId==%s", randoms.ProjectIdStr))
			expectAncestorsTest(listResult, []api.SupersetTemplate{sameProj}, "expecting only sameProject")

			countResult := countTemplatesForTest(router, randoms, fmt.Sprintf("filterBy=projectId==%s", randoms.ProjectIdStr))
			Expect(countResult.Count).To(BeNumerically("==", 1))

			//
			//   project + ancestors
			//
			listResult = listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url,
				fmt.Sprintf("projectId==%s&includeAncestors=true", randoms.ProjectIdStr))
			expectAncestorsTest(listResult, []api.SupersetTemplate{sameProj, sameDept, sameCluster, sameTenant}, "expecting sameProject + sameDept + sameClust + tenant")

			countResult = countTemplatesForTest(router, randoms, fmt.Sprintf("filterBy=projectId==%s&includeAncestors=true", randoms.ProjectIdStr))
			Expect(countResult.Count).To(BeNumerically("==", 4))

			//
			//  wrong combination of parameters
			//
			listReq, err := http.NewRequest("GET", fmt.Sprintf("%s?includeAncestors=true&filterBy=clusterId==%s,projectId==%s", url, randoms.ClusterId, randoms.ProjectIdStr), nil)
			Expect(err).To(BeNil())
			listReq.Header = *headerForAdmin(randoms)

			listResp := httptest.NewRecorder()
			router.ServeHTTP(listResp, listReq)
			Expect(listResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listResp.Body.String()).To(ContainSubstring("cannot use both projectId and clusterId filters with includeAncestors=true"))
		})

		It("include ancestors with clusterId", func() {
			sameTenant := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_TENANT,
					ScopeId:   randoms.TenantIdStr,
					Name:      fmt.Sprintf("t"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameCluster := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   randoms.ClusterId,
					Name:      fmt.Sprintf("c"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameDept := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      fmt.Sprintf("d"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			sameProj := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
					Name:      fmt.Sprintf("p"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			anotherCluster := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   anotherAllowedProject.ClusterId,
					Name:      fmt.Sprintf("xc"),
				},
				Spec: api.SupersetSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			for _, temp := range []api.SupersetTemplate{
				sameTenant,
				sameCluster,
				sameDept,
				sameProj,
				anotherCluster,
			} {
				createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](temp))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			//
			//   list with filter of specific cluster
			//
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url,
				fmt.Sprintf("clusterId==%s", randoms.ClusterId))
			expectAncestorsTest(listResult, []api.SupersetTemplate{sameCluster, sameDept, sameProj}, "expecting cluster, department, and project templates for the cluster")

			countResult := countTemplatesForTest(router, randoms, fmt.Sprintf("filterBy=clusterId==%s", randoms.ClusterId))
			Expect(countResult.Count).To(BeNumerically("==", 3))

			//
			//   cluster + ancestors (system, tenant, department, and project)
			//
			listResult = listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url,
				fmt.Sprintf("clusterId==%s&includeAncestors=true", randoms.ClusterId))
			// We expect sameCluster + sameDept + sameProj + sameTenant + any existing system templates
			// The exact count depends on how many system templates exist
			Expect(len(listResult.Entries)).To(BeNumerically(">=", 4), "expecting at least sameCluster + sameDept + sameProj + sameTenant")

			countResult = countTemplatesForTest(router, randoms, fmt.Sprintf("filterBy=clusterId==%s&includeAncestors=true", randoms.ClusterId))
			Expect(countResult.Count).To(BeNumerically(">=", 4), "expecting at least sameCluster + sameDept + sameProj + sameTenant")

			//
			//   wrong combination of parameters
			//
			listReq, err := http.NewRequest("GET", fmt.Sprintf("%s?includeAncestors=true&filterBy=clusterId==%s,projectId==%s", url, randoms.ClusterId, randoms.ProjectIdStr), nil)
			Expect(err).To(BeNil())
			listReq.Header = *headerForAdmin(randoms)

			listResp := httptest.NewRecorder()
			router.ServeHTTP(listResp, listReq)
			Expect(listResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listResp.Body.String()).To(ContainSubstring("cannot use both projectId and clusterId filters with includeAncestors=true"))
		})
	})

	Context("name availability", func() {
		It("should not allow to create in ancestor tenant with the same name", func() {
			// Creating a template with the name "my-template"
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      "my-template",
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}
			url := fmt.Sprintf("%s/workspaces", urlTemplates)
			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)
			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Checking if the name is available
			urlNameAvailability := fmt.Sprintf("%s/name-availability?templateName=my-template&scopeType=tenant&scopeId=%s", urlTemplates, randoms.TenantIdStr)
			nameAvailabilityReq, err := http.NewRequest("GET", urlNameAvailability, nil)
			Expect(err).To(BeNil())
			nameAvailabilityReq.Header = *headerForAdmin(randoms)
			nameAvailabilityResp := httptest.NewRecorder()
			router.ServeHTTP(nameAvailabilityResp, nameAvailabilityReq)
			Expect(nameAvailabilityResp.Code).To(Equal(http.StatusConflict))
			Expect(nameAvailabilityResp.Body.String()).To(ContainSubstring("the given template name is not available"))
		})

		It("should not allow to create in ancestor cluster with the same name", func() {
			// Creating a template with the name "my-template"
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      "my-template",
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}
			url := fmt.Sprintf("%s/workspaces", urlTemplates)
			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)
			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Checking if the name is available
			urlNameAvailability := fmt.Sprintf("%s/name-availability?templateName=my-template&scopeType=cluster&scopeId=%s", urlTemplates, randoms.ClusterId)
			nameAvailabilityReq, err := http.NewRequest("GET", urlNameAvailability, nil)
			Expect(err).To(BeNil())
			nameAvailabilityReq.Header = *headerForAdmin(randoms)
			nameAvailabilityResp := httptest.NewRecorder()
			router.ServeHTTP(nameAvailabilityResp, nameAvailabilityReq)
			Expect(nameAvailabilityResp.Code).To(Equal(http.StatusConflict))
			Expect(nameAvailabilityResp.Body.String()).To(ContainSubstring("the given template name is not available"))
		})

		It("should not allow to create in chile project with the same name", func() {
			// Creating a template with the name "my-template"
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      "my-template",
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}
			url := fmt.Sprintf("%s/workspaces", urlTemplates)
			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)
			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Checking if the name is available
			urlNameAvailability := fmt.Sprintf("%s/name-availability?templateName=my-template&scopeType=project&scopeId=%s", urlTemplates, randoms.ProjectIdStr)
			nameAvailabilityReq, err := http.NewRequest("GET", urlNameAvailability, nil)
			Expect(err).To(BeNil())
			nameAvailabilityReq.Header = *headerForAdmin(randoms)
			nameAvailabilityResp := httptest.NewRecorder()
			router.ServeHTTP(nameAvailabilityResp, nameAvailabilityReq)
			Expect(nameAvailabilityResp.Code).To(Equal(http.StatusConflict))
			Expect(nameAvailabilityResp.Body.String()).To(ContainSubstring("the given template name is not available"))
		})

		It("should allow to create in child project with different name", func() {
			// Creating a template with the name "my-template"
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
					Name:      "my-template",
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}
			url := fmt.Sprintf("%s/workspaces", urlTemplates)
			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)
			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Checking if the name is available
			urlNameAvailability := fmt.Sprintf("%s/name-availability?templateName=his-template&scopeType=project&scopeId=%s", urlTemplates, randoms.ProjectIdStr)
			nameAvailabilityReq, err := http.NewRequest("GET", urlNameAvailability, nil)
			Expect(err).To(BeNil())
			nameAvailabilityReq.Header = *headerForAdmin(randoms)
			nameAvailabilityResp := httptest.NewRecorder()
			router.ServeHTTP(nameAvailabilityResp, nameAvailabilityReq)
			Expect(nameAvailabilityResp.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("pagination", func() {
		It("next/limit are ok", func() {
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   randoms.ClusterId,
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			//
			//    create 5 templates
			//
			for i := 1; i <= 5; i++ {
				toCreate.Meta.Name = fmt.Sprintf("my-template-%d", i)

				url := fmt.Sprintf("%s/workspaces", urlTemplates)

				createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			//
			//   list all of them, descending order of names
			//
			urlSortNameDesc := "sortBy=name&sortOrder=desc"

			countResult := countTemplatesForTest(router, randoms, urlSortNameDesc)
			Expect(countResult.Count).To(BeNumerically("==", 5))

			listResult := listTemplatesForTest(router, randoms, urlSortNameDesc)

			Expect(listResult.Pagination.NextOffset).To(BeNumerically("==", 0))
			Expect(listResult.Pagination.Limit).To(BeNumerically("==", 50))
			Expect(listResult.Pagination.Offset).To(BeNumerically("==", 0))

			Expect(len(listResult.Entries)).To(Equal(5))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("my-template-5"))
			Expect(listResult.Entries[1].Meta.Name).To(Equal("my-template-4"))
			Expect(listResult.Entries[2].Meta.Name).To(Equal("my-template-3"))
			Expect(listResult.Entries[3].Meta.Name).To(Equal("my-template-2"))
			Expect(listResult.Entries[4].Meta.Name).To(Equal("my-template-1"))

			//
			//   list 3 from offset 1 (should return my-template-4;3;2
			//
			listResult = listTemplatesForTest(router, randoms, "sortBy=name&sortOrder=desc&offset=1&limit=3")

			Expect(listResult.Pagination.NextOffset).To(BeNumerically("==", 4))
			Expect(listResult.Pagination.Limit).To(BeNumerically("==", 3))
			Expect(listResult.Pagination.Offset).To(BeNumerically("==", 1))

			Expect(len(listResult.Entries)).To(Equal(3))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("my-template-4"))
			Expect(listResult.Entries[1].Meta.Name).To(Equal("my-template-3"))
			Expect(listResult.Entries[2].Meta.Name).To(Equal("my-template-2"))

			//
			//   list 3 from offset 4 (should return my-template-1
			//
			listResult = listTemplatesForTest(router, randoms, "sortBy=name&sortOrder=desc&offset=4&limit=3")

			Expect(listResult.Pagination.NextOffset).To(BeNumerically("==", 0))
			Expect(listResult.Pagination.Limit).To(BeNumerically("==", 3))
			Expect(listResult.Pagination.Offset).To(BeNumerically("==", 4))

			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("my-template-1"))
		})
	})

	//WAIT_FOR_OFER -> need to do the same test with "patch" once we support cleanup of fields

	Context("distributed update from worker/master to worker + same as", func() {
		It("should change properly", func() {
			toCreate := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.DistributedTemplateSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Image:                convutils.Ref("ubuntu"),
					NumWorkers:           convutils.Ref(int32(4)),
					SegmentSize:          convutils.Ref(int32(2)),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
			}

			url := fmt.Sprintf("%s/distributed", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   specs are ok
			//
			listResult := listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==cluster")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			testutils.Expect(listedTemp.Spec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.CombinedSpec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.Master.Spec).ToEqual(toCreate.Master.Spec)
			testutils.Expect(listedTemp.CombinedMasterSpec).ToEqual(toCreate.Master.Spec)

			Expect(listedTemp.Runnable.IsRunnable).To(Equal(false))
			Expect(*listedTemp.Runnable.Issue).To(Equal("master.image is not provided"))

			//
			//  change to master same as worker
			//
			toUpdate := api.DistributedTemplateUpdateRequest{
				Meta: api.TemplateUpdateMetaFields{
					Name: toCreate.Meta.Name,
				},
				Spec:                   toCreate.Spec,
				Master:                 nil,
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			urlById := fmt.Sprintf("%s/distributed/%s", urlTemplates, listedTemp.Meta.Id)

			putReq, err := http.NewRequest("PUT", urlById, makePayload(toUpdate))
			Expect(err).To(BeNil())
			putReq.Header = *headerForAdmin(randoms)

			putResp := httptest.NewRecorder()
			router.ServeHTTP(putResp, putReq)
			Expect(putResp.Code).To(Equal(http.StatusOK))

			//
			//   template becomes runnable
			//
			listResult = listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==cluster")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp = listResult.Entries[0]

			testutils.Expect(listedTemp.Spec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.CombinedSpec).ToEqual(toCreate.Spec)
			Expect(listedTemp.Master).To(BeNil())
			Expect(listedTemp.CombinedMasterSpec).To(BeNil())
			Expect(*listedTemp.MasterSpecSameAsWorker).To(Equal(true))

			Expect(listedTemp.Runnable.IsRunnable).To(Equal(true))
		})
	})

	Context("distributed patch", func() {
		It("add master section - good", func() {
			toCreate := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.DistributedTemplateSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Image:                convutils.Ref("ubuntu-worker"),
				},
			}

			url := fmt.Sprintf("%s/distributed", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//   specs are ok
			//
			listResult := listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==cluster")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			testutils.Expect(listedTemp.Spec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.CombinedSpec).ToEqual(toCreate.Spec)
			Expect(listedTemp.Master).To(BeNil())
			Expect(listedTemp.CombinedMasterSpec).To(BeNil())

			Expect(listedTemp.Runnable.IsRunnable).To(Equal(false))
			Expect(*listedTemp.Runnable.Issue).To(Equal("master is not provided for MPI framework"))

			//
			//  add master section
			//
			toPatch := api.DistributedTemplatePatchRequest{
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu-master"),
					},
				},
			}

			urlById := fmt.Sprintf("%s/distributed/%s", urlTemplates, listedTemp.Meta.Id)

			patchReq, err := http.NewRequest("PATCH", urlById, makePayload(toPatch))
			Expect(err).To(BeNil())
			patchReq.Header = *headerForAdmin(randoms)

			patchResp := httptest.NewRecorder()
			router.ServeHTTP(patchResp, patchReq)
			Expect(patchResp.Code).To(Equal(http.StatusOK))

			//
			//   template becomes runnable
			//
			listResult = listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==cluster")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp = listResult.Entries[0]

			testutils.Expect(listedTemp.Spec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.CombinedSpec).ToEqual(toCreate.Spec)
			testutils.Expect(listedTemp.Master.Spec).ToEqual(toPatch.Master.Spec)
			testutils.Expect(listedTemp.CombinedMasterSpec).ToEqual(toPatch.Master.Spec)

			Expect(listedTemp.Runnable.IsRunnable).To(Equal(true))

			//
			//  change to JAX is forbidden (due to the master)
			//
			toPatch = api.DistributedTemplatePatchRequest{
				Spec: api.DistributedTemplateSpec{
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_JAX),
				},
			}

			patchReq, err = http.NewRequest("PATCH", urlById, makePayload(toPatch))
			Expect(err).To(BeNil())
			patchReq.Header = *headerForAdmin(randoms)

			patchResp = httptest.NewRecorder()
			router.ServeHTTP(patchResp, patchReq)
			Expect(patchResp.Code).To(Equal(http.StatusBadRequest))
			Expect(patchResp.Body.String()).To(ContainSubstring("master should not be provided"))

			//
			//  attempt to patch it as workspace should fail
			//
			urlByIdWrong := fmt.Sprintf("%s/workspaces/%s", urlTemplates, listedTemp.Meta.Id)

			toPatchAsWorkspace := api.WorkspaceTemplatePatchRequest{}

			patchReq, err = http.NewRequest("PATCH", urlByIdWrong, makePayload(toPatchAsWorkspace))
			Expect(err).To(BeNil())
			patchReq.Header = *headerForAdmin(randoms)

			patchResp = httptest.NewRecorder()
			router.ServeHTTP(patchResp, patchReq)
			Expect(patchResp.Code).To(Equal(http.StatusNotFound))
			Expect(patchResp.Body.String()).To(ContainSubstring("Workspace template"))
			Expect(patchResp.Body.String()).To(ContainSubstring("does not exist"))
		})
	})

	Context("list with limited scopes", func() {
		It("should be filtered", func() {
			service := TemplatesService{}

			toCreate1 := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			toCreate2 := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(forbiddenProject.ProjectIdStr),
					Name:      fmt.Sprintf("my-template"),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate1))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			createReq, err = http.NewRequest("POST", url, makePayload(toCreate2))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp = httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			listResult, listErr := service.ListTemplatesUnified(ctx, nil, 0, 0, "", "", []string{"scopeType==project"}, "", nil, false)
			Expect(listErr).To(BeNil())
			Expect(len(listResult.Entries)).To(Equal(2))

			//
			//   allow access to projects only, not departments and so on - only single template match
			//
			authMock = testing_crud.NewRandomsBasedAuthMockClient().
				WithSpecificProject(randoms).
				WithSpecificProject(forbiddenProject)

			orgunit.SetMock(authMock.OrgUnitMock())

			ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			listResult, listErr = service.ListTemplatesUnified(ctx, nil, 0, 0, "", "", []string{"scopeType==project"}, "", nil, false)
			Expect(listErr).To(BeNil())
			Expect(len(listResult.Entries)).To(Equal(1))
		})
	})

	Context("validate department scope admin vs cluster template", func() {
		It("should be able to read it", func() {
			service := TemplatesService{}

			//
			//  create the template with all necessary permissions
			//
			toCreate := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("cluster-template"),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("ubuntu"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			//  now allow only department access
			//
			authMock = testing_crud.NewRandomsBasedAuthMockClient().WithDepartment(randoms)
			orgunit.SetMock(authMock.OrgUnitMock())

			//
			//   should see it in list templates
			//
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			resp, err := service.ListWorkspaceTemplates(ctx, 0, 50, "", "", []string{}, "", false, 0)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			payload := resp.Body.(*UnifiedListResponse)
			Expect(payload.Entries).To(HaveLen(1))

			id := payload.Entries[0].Meta.Id

			//
			//   should see it in get by id
			//
			resp, err = service.GetTemplateById(ctx, id)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
		})
	})

	Context("free-text search", func() {
		It("should search across all text fields", func() {
			toCreate := []api.SupersetTemplate{
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Python Training Template",
						Description: convutils.Ref("Template for Python ML training"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
						Image:      convutils.Ref("python:3.9"),
					},
				},
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "TensorFlow Inference",
						Description: convutils.Ref("Template for TensorFlow model serving"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
						Image:      convutils.Ref("tensorflow/serving"),
					},
				},
			}

			// Create test templates
			for _, template := range toCreate {
				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s/workspaces", urlTemplates), composeCreationReq[api.WorkspaceTemplateCreationRequest](template))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			// Test search by name
			listResult := listTemplatesForTest(router, randoms, "search=Python")
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("Python Training Template"))

			// Test search by description
			listResult = listTemplatesForTest(router, randoms, "search=TensorFlow")
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("TensorFlow Inference"))

			// Test search by multiple terms
			listResult = listTemplatesForTest(router, randoms, "search=Python & Training")
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("Python Training Template"))

			// Test search with no matches
			listResult = listTemplatesForTest(router, randoms, "search=NonExistent")
			Expect(len(listResult.Entries)).To(Equal(0))

			// Test count with search
			countResult := countTemplatesForTest(router, randoms, "search=Python")
			Expect(countResult.Count).To(Equal(int64(1)))
		})
	})

	Context("template search", func() {
		It("should search by name", func() {
			// Create a template with a specific name
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "Python Training Template",
					Description: convutils.Ref("Template for Python ML training"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/trainings", testHost, urlTemplates), composeCreationReq[api.TrainingTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Search by name
			searchQuery := "search=Python"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("Python Training Template"))
		})

		It("should search by description", func() {
			// Create a template with a specific description
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "TensorFlow Template",
					Description: convutils.Ref("Template for TensorFlow model serving"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/inferences", testHost, urlTemplates), composeCreationReq[api.InferenceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Search by description
			searchQuery := "search=TensorFlow"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("TensorFlow Template"))
		})

		It("should search by workload type", func() {
			// Create a template with a specific workload type
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "Distributed Training",
					Description: convutils.Ref("Template for distributed training"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/distributed", testHost, urlTemplates), composeCreationReq[api.DistributedTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Search by workload type
			searchQuery := "search=Distributed"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("Distributed Training"))
		})

		It("should search with multiple terms", func() {
			// Create a template with specific terms
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "Python ML Template",
					Description: convutils.Ref("Template for Python machine learning"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/trainings", testHost, urlTemplates), composeCreationReq[api.TrainingTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Search with multiple terms
			searchQuery := "search=Python & ML"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(1))
			Expect(listResult.Entries[0].Meta.Name).To(Equal("Python ML Template"))
		})

		It("should return no results for non-existent search", func() {
			// Search for non-existent term
			searchQuery := "search=NonExistent"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(0))
		})

		It("should count templates correctly with search", func() {
			// Create a template
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "Python Template",
					Description: convutils.Ref("Template for Python"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/trainings", testHost, urlTemplates), composeCreationReq[api.TrainingTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Count with search
			searchQuery := "search=Python"
			countResult := countTemplatesForTest(router, randoms, searchQuery)
			Expect(countResult.Count).To(BeNumerically("==", 1))
		})

		It("should find multiple templates with common term", func() {
			// Create multiple templates with common term
			templates := []api.SupersetTemplate{
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Python Training Template",
						Description: convutils.Ref("Template for Python ML training"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Python Inference Template",
						Description: convutils.Ref("Template for Python model serving"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Python Workspace Template",
						Description: convutils.Ref("Template for Python development"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
			}

			// Create all templates
			for _, template := range templates {
				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/workspaces", testHost, urlTemplates), composeCreationReq[api.WorkspaceTemplateCreationRequest](template))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			// Search for common term
			searchQuery := "search=Python"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(3))

			// Verify all results contain the search term
			for _, entry := range listResult.Entries {
				Expect(entry.Meta.Name).To(ContainSubstring("Python"))
			}

			// Count should match
			countResult := countTemplatesForTest(router, randoms, searchQuery)
			Expect(countResult.Count).To(Equal(int64(3)))
		})

		It("should find templates with partial matches", func() {
			// Create templates with similar terms
			templates := []api.SupersetTemplate{
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Deep Learning Template",
						Description: convutils.Ref("Template for deep learning"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Machine Learning Template",
						Description: convutils.Ref("Template for machine learning"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
			}

			// Create all templates
			for _, template := range templates {
				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/trainings", testHost, urlTemplates), composeCreationReq[api.TrainingTemplateCreationRequest](template))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			// Search for partial term
			searchQuery := "search=Learning"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(2))

			// Verify all results contain the search term
			for _, entry := range listResult.Entries {
				Expect(entry.Meta.Name).To(ContainSubstring("Learning"))
			}

			// Count should match
			countResult := countTemplatesForTest(router, randoms, searchQuery)
			Expect(countResult.Count).To(Equal(int64(2)))
		})

		It("should find templates with terms in different fields", func() {
			// Create templates with terms in different fields
			templates := []api.SupersetTemplate{
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Training Template",
						Description: convutils.Ref("Template for Python development"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
				{
					Meta: api.TemplateMetaFields{
						ScopeType:   api.SCOPETYPE_TENANT,
						Name:        "Python Template",
						Description: convutils.Ref("Template for training"),
					},
					Spec: api.SupersetSpec{
						WorkingDir: convutils.Ref("/home/worker"),
					},
				},
			}

			// Create all templates
			for _, template := range templates {
				createReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s/trainings", testHost, urlTemplates), composeCreationReq[api.TrainingTemplateCreationRequest](template))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			// Search for terms that appear in different fields
			searchQuery := "search=Python & Training"
			listResult := listTemplatesForTest(router, randoms, searchQuery)
			Expect(len(listResult.Entries)).To(Equal(2))

			// Verify all results contain at least one of the search terms
			for _, entry := range listResult.Entries {
				Expect(entry.Meta.Name).To(Or(
					ContainSubstring("Python"),
					ContainSubstring("Training"),
				))
			}

			// Count should match
			countResult := countTemplatesForTest(router, randoms, searchQuery)
			Expect(countResult.Count).To(Equal(int64(2)))
		})

		It("should find templates by created_by field", func() {
			// Create a template with a specific creator
			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType:   api.SCOPETYPE_TENANT,
					Name:        "test-template",
					Description: convutils.Ref("test description"),
				},
				Spec: api.SupersetSpec{
					WorkingDir: convutils.Ref("/home/worker"),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			// Create template as admin
			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			// Search for template by created_by field
			listResult := listTemplatesForTest(router, randoms, "search=admin")
			Expect(len(listResult.Entries)).To(Equal(1))
			listedTemp := listResult.Entries[0]

			Expect(listedTemp.Meta.Name).To(Equal(toCreate.Meta.Name))
			Expect(listedTemp.Meta.CreatedBy).To(Equal("admin"))

			// Verify that searching for a non-existent creator returns no results
			listResult = listTemplatesForTest(router, randoms, "search=nonexistent")
			Expect(len(listResult.Entries)).To(Equal(0))
		})
	})

	Context("asset validation - existence of assets", func() {
		It("should allow only existing assets", func() {
			withEnvAsset := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "with-env",
					ScopeType: api.SCOPETYPE_TENANT,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(uuid.NewString()),
				},
			}

			withComputeAsset := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "with-compute",
					ScopeType: api.SCOPETYPE_TENANT,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Compute: convutils.Ref(uuid.NewString()),
				},
			}

			withDataSource := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "with-data-source",
					ScopeType: api.SCOPETYPE_TENANT,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Datasources: []string{
						uuid.NewString(),
					},
				},
			}

			for _, temp := range []api.SupersetTemplate{withEnvAsset, withComputeAsset, withDataSource} {
				for _, wsType := range []string{"workspaces", "trainings", "distributed", "inferences"} {
					whatWeCheck := fmt.Sprintf("%s template %s", wsType, temp.Meta.Name)

					url := fmt.Sprintf("%s/%s", urlTemplates, wsType)

					createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](temp))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(randoms)

					createResp := httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq)
					Expect(createResp.Code).To(Equal(http.StatusNotFound), whatWeCheck)
					Expect(createResp.Body.String()).To(ContainSubstring("asset not found"), whatWeCheck)
				}
			}
		})

		It("should allow only existing assets - master spec", func() {
			withEnvAsset := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "with-env",
					ScopeType: api.SCOPETYPE_TENANT,
				},
				Master: &api.TemplateMasterSection{
					Assets: &api.TemplateAssetsIdsOptional{
						Environment: convutils.Ref(uuid.NewString()),
					},
				},
			}

			url := fmt.Sprintf("%s/distributed", urlTemplates)

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.DistributedTemplateCreationRequest](withEnvAsset))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusNotFound))
			Expect(createResp.Body.String()).To(ContainSubstring("asset not found"))
		})
	})

	Context("asset validation - scopes", func() {
		It("can only get from same scope or above", func() {
			//
			//   create assets in all levels
			//
			systemAsset := assetsApi.NewAssetMeta("system-env", assetsApi.SCOPE_SYSTEM, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

			tenantAsset := assetsApi.NewAssetMeta("tenant-env", assetsApi.SCOPE_TENANT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

			clusterAsset := assetsApi.NewAssetMeta("cluster-env", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			clusterAsset.ClusterId.Set(&randoms.ClusterId)

			deptAsset := assetsApi.NewAssetMeta("department-env", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			deptAsset.ClusterId.Set(&randoms.ClusterId)
			deptAsset.DepartmentId.Set(&randoms.DepartmentId)

			projAsset := assetsApi.NewAssetMeta("project-env", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			projAsset.ClusterId.Set(&randoms.ClusterId)
			projAsset.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(systemAsset).WithMapping(api.WORKLOADTYPE_TRAINING, systemAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(tenantAsset).WithMapping(api.WORKLOADTYPE_TRAINING, tenantAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(clusterAsset).WithMapping(api.WORKLOADTYPE_TRAINING, clusterAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(deptAsset).WithMapping(api.WORKLOADTYPE_TRAINING, deptAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(projAsset).WithMapping(api.WORKLOADTYPE_TRAINING, projAsset.Id, &assetsApi.SupersetSpec{}))

			scopeToLevel := map[string]int{
				string(assetsApi.SCOPE_SYSTEM):     0,
				string(assetsApi.SCOPE_TENANT):     1,
				string(assetsApi.SCOPE_CLUSTER):    2,
				string(assetsApi.SCOPE_DEPARTMENT): 3,
				string(assetsApi.SCOPE_PROJECT):    4,
			}

			//
			//   templates from all levels
			//
			tenantTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "tenant",
					ScopeType: api.SCOPETYPE_TENANT,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(uuid.NewString()),
				},
			}

			clusterTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "cluster",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   randoms.ClusterId,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(uuid.NewString()),
				},
			}

			deptTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "dept",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(uuid.NewString()),
				},
			}

			projTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "dept",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(uuid.NewString()),
				},
			}

			for _, template := range []api.SupersetTemplate{tenantTemplate, clusterTemplate, deptTemplate, projTemplate} {
				for _, assetMeta := range []*assetsApi.AssetMeta{systemAsset, tenantAsset, clusterAsset, deptAsset, projAsset} {

					whatWeCheck := fmt.Sprintf("template scope %s VS asset scope %s", template.Meta.ScopeType, assetMeta.Scope)

					url := fmt.Sprintf("%s/trainings", urlTemplates)

					// inject the asset id and make template name unique
					template.Meta.Name = fmt.Sprintf("template_%s_asset_%s", template.Meta.ScopeType, assetMeta.Scope)
					template.Assets.Environment = convutils.Ref(assetMeta.Id)

					createReq, err := http.NewRequest("POST", url, composeCreationReq[api.TrainingTemplateCreationRequest](template))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(randoms)

					createResp := httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq)

					levelOfTemplate := scopeToLevel[string(template.Meta.ScopeType)]
					levelOfAsset := scopeToLevel[string(assetMeta.Scope)]

					if levelOfAsset <= levelOfTemplate {
						Expect(createResp.Code).To(Equal(http.StatusOK), whatWeCheck)
					} else {
						Expect(createResp.Code).To(Equal(http.StatusBadRequest), whatWeCheck)
						Expect(createResp.Body.String()).To(ContainSubstring(fmt.Sprintf("template of scope %s cannot link to an asset", template.Meta.ScopeType)))
					}
				}
			}
		})
	})

	Context("asset validation - ancestors", func() {
		It("can only get from its ancestors", func() {
			sameCluster := assetsApi.NewAssetMeta("same-cluster", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			sameCluster.ClusterId.Set(&randoms.ClusterId)

			differentCluster := assetsApi.NewAssetMeta("different-cluster", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			differentCluster.ClusterId.Set(convutils.Ref(randoms.CloneWithDifferentCluster().ClusterId))

			sameDept := assetsApi.NewAssetMeta("same-dept", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			sameDept.ClusterId.Set(&randoms.ClusterId)
			sameDept.DepartmentId.Set(&randoms.DepartmentId)

			differentDept := assetsApi.NewAssetMeta("different-dept", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			differentDept.ClusterId.Set(&randoms.ClusterId)
			differentDept.DepartmentId.Set(&randoms.CloneWithDifferentDepartment().DepartmentId)

			sameProj := assetsApi.NewAssetMeta("same-proj", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			sameProj.ClusterId.Set(&randoms.ClusterId)
			sameProj.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			differentProj := assetsApi.NewAssetMeta("different-proj", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			differentProj.ClusterId.Set(&randoms.ClusterId)
			differentProj.ProjectId.Set(convutils.Ref(int32(randoms.CloneWithDifferentProject().ProjectId)))

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(sameCluster).WithMapping(api.WORKLOADTYPE_TRAINING, sameCluster.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(sameDept).WithMapping(api.WORKLOADTYPE_TRAINING, sameDept.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(sameProj).WithMapping(api.WORKLOADTYPE_TRAINING, sameProj.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(differentCluster).WithMapping(api.WORKLOADTYPE_TRAINING, differentCluster.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(differentDept).WithMapping(api.WORKLOADTYPE_TRAINING, differentDept.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(differentProj).WithMapping(api.WORKLOADTYPE_TRAINING, differentProj.Id, &assetsApi.SupersetSpec{}))

			for _, assetMeta := range []*assetsApi.AssetMeta{sameCluster, sameDept, sameProj, differentCluster, differentDept, differentProj} {

				whatWeCheck := fmt.Sprintf("template link to %s", assetMeta.Name)

				url := fmt.Sprintf("%s/trainings", urlTemplates)

				projTemplate := api.SupersetTemplate{
					Meta: api.TemplateMetaFields{
						Name:      fmt.Sprintf("link-to-%s", assetMeta.Name),
						ScopeType: api.SCOPETYPE_PROJECT,
						ScopeId:   randoms.ProjectIdStr,
					},
					Assets: &api.TemplateAssetsIdsOptional{
						Environment: convutils.Ref(assetMeta.Id),
					},
				}

				createReq, err := http.NewRequest("POST", url, composeCreationReq[api.TrainingTemplateCreationRequest](projTemplate))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)

				if strings.HasPrefix(assetMeta.Name, "same") {
					Expect(createResp.Code).To(Equal(http.StatusOK), whatWeCheck)
				} else {
					Expect(createResp.Code).To(Equal(http.StatusBadRequest), whatWeCheck)
					Expect(createResp.Body.String()).To(ContainSubstring(fmt.Sprintf("cannot link to an asset of %s", assetMeta.Scope)))
				}
			}
		})
	})

	Context("asset validation - asset kind", func() {
		It("the stated kind need to match the asset kind", func() {
			envAsset := assetsApi.NewAssetMeta("env-asset", assetsApi.SCOPE_TENANT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

			computeAsset := assetsApi.NewAssetMeta("compute-asset", assetsApi.SCOPE_TENANT, uuid.NewString(),
				assetsApi.ASSETKIND_COMPUTE, testing.TestUser, time.Now(), testing.TestUser, time.Now())

			pvcAsset := assetsApi.NewAssetMeta("pvc-asset", assetsApi.SCOPE_TENANT, uuid.NewString(),
				assetsApi.ASSETKIND_PVC, testing.TestUser, time.Now(), testing.TestUser, time.Now())

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, computeAsset.Id, &assetsApi.SupersetSpec{}).
				WithAssetMeta(pvcAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, pvcAsset.Id, &assetsApi.SupersetSpec{}))

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			toCreate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "env-vs-compute",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(computeAsset.Id),
				},
			}

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("kind of asset"))

			toCreate = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "compute-vs-pvc",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Compute: convutils.Ref(pvcAsset.Id),
				},
			}

			createReq, err = http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp = httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("kind of asset"))

			toCreate = api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "pvc-vs-env",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Datasources: []string{
						envAsset.Id,
					},
				},
			}

			createReq, err = http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp = httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createResp.Body.String()).To(ContainSubstring("kind of asset"))
		})

		Context("asset validation - supported workloads", func() {
			It("can only link to assets supporting the workload type", func() {
				workspaceOnly := assetsApi.NewAssetMeta("Workspace", assetsApi.SCOPE_TENANT, uuid.NewString(),
					assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

				workspaceOnly.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Workspace:   *assetsApi.NewNullableBool(convutils.Ref(true)),
					Training:    *assetsApi.NewNullableBool(convutils.Ref(false)),
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(false)),
					Inference:   *assetsApi.NewNullableBool(convutils.Ref(false)),
				})

				trainingOnly := assetsApi.NewAssetMeta("Training", assetsApi.SCOPE_TENANT, uuid.NewString(),
					assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

				trainingOnly.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Workspace:   *assetsApi.NewNullableBool(convutils.Ref(false)),
					Training:    *assetsApi.NewNullableBool(convutils.Ref(true)),
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(false)),
					Inference:   *assetsApi.NewNullableBool(convutils.Ref(false)),
				})

				distributedOnly := assetsApi.NewAssetMeta("Distributed", assetsApi.SCOPE_TENANT, uuid.NewString(),
					assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

				distributedOnly.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Workspace:   *assetsApi.NewNullableBool(convutils.Ref(false)),
					Training:    *assetsApi.NewNullableBool(convutils.Ref(false)),
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(true)),
					Inference:   *assetsApi.NewNullableBool(convutils.Ref(false)),
				})

				inferenceOnly := assetsApi.NewAssetMeta("Inference", assetsApi.SCOPE_TENANT, uuid.NewString(),
					assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

				inferenceOnly.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
					Workspace:   *assetsApi.NewNullableBool(convutils.Ref(false)),
					Training:    *assetsApi.NewNullableBool(convutils.Ref(false)),
					Distributed: *assetsApi.NewNullableBool(convutils.Ref(false)),
					Inference:   *assetsApi.NewNullableBool(convutils.Ref(true)),
				})

				workloadTypes := []assetsApi.WorkloadType{
					assetsApi.WORKLOADTYPE_WORKSPACE,
					assetsApi.WORKLOADTYPE_TRAINING,
					assetsApi.WORKLOADTYPE_DISTRIBUTED,
					assetsApi.WORKLOADTYPE_INFERENCE,
				}

				assets.SetMock(assets.NewSimpleMock().
					WithAssetMeta(workspaceOnly).WithMapping(api.WORKLOADTYPE_WORKSPACE, workspaceOnly.Id, &assetsApi.SupersetSpec{}).
					WithAssetMeta(trainingOnly).WithMapping(api.WORKLOADTYPE_TRAINING, trainingOnly.Id, &assetsApi.SupersetSpec{}).
					WithAssetMeta(distributedOnly).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, distributedOnly.Id, &assetsApi.SupersetSpec{}).
					WithAssetMeta(inferenceOnly).WithMapping(api.WORKLOADTYPE_INFERENCE, inferenceOnly.Id, &assetsApi.SupersetSpec{}))

				for _, workloadType := range workloadTypes {
					for _, assetSupportingOneType := range []*assetsApi.AssetMeta{workspaceOnly, trainingOnly, distributedOnly, inferenceOnly} {
						whatWeCheck := fmt.Sprintf("template type %s with asset type %s", workloadType, assetSupportingOneType.Name)

						// url is in plural lower-case except for distributed - no s.
						typeUrl := strings.ToLower(string(workloadType))
						if workloadType != assetsApi.WORKLOADTYPE_DISTRIBUTED {
							typeUrl += "s"
						}

						url := fmt.Sprintf("%s/%s", urlTemplates, typeUrl)

						template := api.SupersetTemplate{
							Meta: api.TemplateMetaFields{
								Name:      fmt.Sprintf("%s-vs-%s", workloadType, assetSupportingOneType.Name),
								ScopeType: api.SCOPETYPE_TENANT,
							},
							Assets: &api.TemplateAssetsIdsOptional{
								Environment: convutils.Ref(assetSupportingOneType.Id),
							},
						}

						createReq, err := http.NewRequest("POST", url, composeCreationReq[api.TrainingTemplateCreationRequest](template))
						Expect(err).To(BeNil())
						createReq.Header = *headerForAdmin(randoms)

						createResp := httptest.NewRecorder()
						router.ServeHTTP(createResp, createReq)

						if string(workloadType) == assetSupportingOneType.Name {
							Expect(createResp.Code).To(Equal(http.StatusOK), whatWeCheck)
						} else {
							Expect(createResp.Code).To(Equal(http.StatusBadRequest), whatWeCheck)
							Expect(createResp.Body.String()).To(ContainSubstring(fmt.Sprintf("is not supported for workload of type %s", workloadType)))
						}
					}
				}

			})

			Context("asset validation - fields conflict - worker", func() {
				It("should be rejected", func() {
					envAsset := assetsApi.NewAssetMeta("env", assetsApi.SCOPE_TENANT, uuid.NewString(),
						assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())

					envAsFlat := &assetsApi.SupersetSpec{
						Security: *assetsApi.NewNullableSupersetSpecAllOfSecurity(&assetsApi.SupersetSpecAllOfSecurity{
							HostIpc: *assetsApi.NewNullableBool(convutils.Ref(true)),
						}),
					}

					computeAsset := assetsApi.NewAssetMeta("compute", assetsApi.SCOPE_TENANT, uuid.NewString(),
						assetsApi.ASSETKIND_COMPUTE, testing.TestUser, time.Now(), testing.TestUser, time.Now())

					computeAsFlat := &assetsApi.SupersetSpec{
						Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(&assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
							GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(5))),
						}),
					}

					assets.SetMock(assets.NewSimpleMock().
						WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, computeAsset.Id, computeAsFlat).
						WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, envAsset.Id, envAsFlat))

					url := fmt.Sprintf("%s/distributed", urlTemplates)

					//
					//   case 1: compute asset VS compute field in worker
					//
					template := api.SupersetTemplate{
						Meta: api.TemplateMetaFields{
							Name:      fmt.Sprintf("conflict-1"),
							ScopeType: api.SCOPETYPE_TENANT,
						},
						Spec: api.SupersetSpec{
							Compute: &api.WorkspaceSpecSpecAllOfCompute{
								GpuDevicesRequest: convutils.Ref(int32(5)),
							},
						},
						Assets: &api.TemplateAssetsIdsOptional{
							Environment: convutils.Ref(envAsset.Id),
							Compute:     convutils.Ref(computeAsset.Id),
						},
					}

					createReq, err := http.NewRequest("POST", url, composeCreationReq[api.DistributedTemplateCreationRequest](template))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(randoms)

					createResp := httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq)

					Expect(createResp.Code).To(Equal(http.StatusBadRequest))
					Expect(createResp.Body.String()).To(ContainSubstring("compute fields are not allowed in the spec of a template that is linked to a compute asset"))

					//
					//   case 2 env asset VS env field in worker
					//
					template = api.SupersetTemplate{
						Meta: api.TemplateMetaFields{
							Name:      fmt.Sprintf("conflict-1"),
							ScopeType: api.SCOPETYPE_TENANT,
						},
						Spec: api.SupersetSpec{
							Security: &api.WorkspaceSpecSpecAllOfSecurity{
								HostNetwork: convutils.Ref(true),
							},
						},
						Assets: &api.TemplateAssetsIdsOptional{
							Environment: convutils.Ref(envAsset.Id),
							Compute:     convutils.Ref(computeAsset.Id),
						},
					}

					createReq, err = http.NewRequest("POST", url, composeCreationReq[api.DistributedTemplateCreationRequest](template))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(randoms)

					createResp = httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq)

					Expect(createResp.Code).To(Equal(http.StatusBadRequest))
					Expect(createResp.Body.String()).To(ContainSubstring("field security.hostNetwork is not allowed in the spec of a template that is linked to an environment asset"))

					//
					//   case 3 env asset VS env field in master
					//
					template = api.SupersetTemplate{
						Meta: api.TemplateMetaFields{
							Name:      fmt.Sprintf("conflict-1"),
							ScopeType: api.SCOPETYPE_TENANT,
						},
						Spec: api.SupersetSpec{
							Security: &api.WorkspaceSpecSpecAllOfSecurity{
								HostNetwork: convutils.Ref(true),
							},
						},
						Assets: &api.TemplateAssetsIdsOptional{
							Compute: convutils.Ref(computeAsset.Id),
						},
						Master: &api.TemplateMasterSection{
							Spec: &api.MasterTemplateSpec{
								Image: convutils.Ref("ubuntu"),
							},
							Assets: &api.TemplateAssetsIdsOptional{
								Environment: convutils.Ref(envAsset.Id),
							},
						},
					}

					createReq, err = http.NewRequest("POST", url, composeCreationReq[api.DistributedTemplateCreationRequest](template))
					Expect(err).To(BeNil())
					createReq.Header = *headerForAdmin(randoms)

					createResp = httptest.NewRecorder()
					router.ServeHTTP(createResp, createReq)

					Expect(createResp.Code).To(Equal(http.StatusBadRequest))
					Expect(createResp.Body.String()).To(ContainSubstring("field image is not allowed in the spec of a template that is linked to an environment asset"))
				})

			})

		})

	})

	Context("distributed inference specific validations", func() {
		It("should create template with comprehensive leader and worker specs", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType:   api.SCOPETYPE_CLUSTER,
					ScopeId:     convutils.Ref(randoms.ClusterId),
					Name:        fmt.Sprintf("comprehensive-dist-inf-template"),
					Description: convutils.Ref("Template for distributed inference serving"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(2)),
					Workers:  convutils.Ref(int32(3)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8081)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("leader-image:v1"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(1)),
							CpuCoreRequest:    convutils.Ref(float64(4)),
							CpuMemoryRequest:  convutils.Ref("4Gi"),
						},
					},
					Worker: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("worker-image:v1"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(1)),
							CpuCoreRequest:    convutils.Ref(float64(2)),
							CpuMemoryRequest:  convutils.Ref("2Gi"),
						},
					},
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK), "Response body: %s", createResp.Body.String())

			var created api.DistributedInferenceTemplate
			jsonErr := json.Unmarshal(createResp.Body.Bytes(), &created)
			Expect(jsonErr).To(BeNil())

			Expect(created.Meta.Name).To(Equal("comprehensive-dist-inf-template"))
			Expect(*created.Spec.Replicas).To(Equal(int32(2)))
			Expect(*created.Spec.Workers).To(Equal(int32(3)))
			Expect(created.Spec.Leader).NotTo(BeNil())
			Expect(created.Spec.Worker).NotTo(BeNil())
			Expect(*created.Spec.Leader.Image).To(Equal("leader-image:v1"))
			Expect(*created.Spec.Worker.Image).To(Equal("worker-image:v1"))
			Expect(*created.Spec.ServingPort.Port).To(Equal(int32(8081)))
			Expect(*created.Spec.ServingPort.Protocol).To(Equal(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP))
		})

		It("should accept leader-only configuration with zero workers", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType:   api.SCOPETYPE_CLUSTER,
					ScopeId:     convutils.Ref(randoms.ClusterId),
					Name:        fmt.Sprintf("leader-only-dist-inf-template"),
					Description: convutils.Ref("Template for single-pod distributed inference"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(1)),
					Workers:  convutils.Ref(int32(0)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("leader-image:v2"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(2)),
							CpuMemoryLimit:    convutils.Ref("4Gi"),
						},
						WorkingDir: convutils.Ref("/app"),
					},
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.DistributedInferenceTemplate
			json.Unmarshal(createResp.Body.Bytes(), &created)

			Expect(created.Meta.Name).To(Equal("leader-only-dist-inf-template"))
			Expect(*created.Spec.Workers).To(Equal(int32(0)))
			Expect(*created.Spec.Replicas).To(Equal(int32(1)))
			Expect(*created.Spec.Leader.Image).To(Equal("leader-image:v2"))
			Expect(*created.Spec.Leader.Compute.GpuDevicesRequest).To(Equal(int32(2)))
			Expect(*created.Spec.Leader.WorkingDir).To(Equal("/app"))
		})

		It("should create template with custom port", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType:   api.SCOPETYPE_CLUSTER,
					ScopeId:     convutils.Ref(randoms.ClusterId),
					Name:        fmt.Sprintf("custom-port-dist-inf-template"),
					Description: convutils.Ref("Template with custom port"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(1)),
					Workers:  convutils.Ref(int32(2)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(9090)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("server:v1"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							CpuCoreRequest: convutils.Ref(float64(2)),
						},
					},
					Worker: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("worker:v1"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							CpuCoreRequest: convutils.Ref(float64(1)),
						},
					},
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, err := http.NewRequest("POST", url, makePayload(toCreate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.DistributedInferenceTemplate
			json.Unmarshal(createResp.Body.Bytes(), &created)

			Expect(created.Meta.Name).To(Equal("custom-port-dist-inf-template"))
			Expect(created.Spec.ServingPort).NotTo(BeNil())
			Expect(*created.Spec.ServingPort.Port).To(Equal(int32(9090)))
			Expect(*created.Spec.ServingPort.Protocol).To(Equal(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP))
		})

		It("should mark template as not runnable without image", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("incomplete-dist-inf-template"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(2)),
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						WorkingDir: convutils.Ref("/app"),
						// Image missing
					},
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, _ := http.NewRequest("POST", url, makePayload(toCreate))
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.DistributedInferenceTemplate
			json.Unmarshal(createResp.Body.Bytes(), &created)

			Expect(created.Runnable.IsRunnable).To(BeFalse())
			Expect(*created.Runnable.Issue).To(Equal("leader.image is not provided"))
		})

		It("should mark template as not runnable without servingPort", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("no-port-dist-inf-template"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(2)),
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("complete-image:v1"),
					},
					// servingPort missing
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, _ := http.NewRequest("POST", url, makePayload(toCreate))
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.DistributedInferenceTemplate
			json.Unmarshal(createResp.Body.Bytes(), &created)

			Expect(created.Runnable.IsRunnable).To(BeFalse())
			Expect(*created.Runnable.Issue).To(Equal("servingPort is not provided"))
		})

		It("should mark complete template as runnable", func() {
			toCreate := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
					Name:      fmt.Sprintf("complete-dist-inf-template"),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(2)),
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(8080)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("complete-image:v1"),
					},
				},
			}

			url := fmt.Sprintf("%s/distributed-inferences", urlTemplates)

			createReq, _ := http.NewRequest("POST", url, makePayload(toCreate))
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.DistributedInferenceTemplate
			json.Unmarshal(createResp.Body.Bytes(), &created)

			Expect(created.Runnable.IsRunnable).To(BeTrue())
		})
	})

	Context("routing method filtering", func() {
		injectRoutingMethod := func(templateId string, routingMethod string) {
			err := dao_utils.DB().Model(&models_crud.Template{}).
				Where(fmt.Sprintf("%s = ?", models_crud.TemplatesColumnId), templateId).
				Update(models_crud.TemplatesColumnRoutingMethod, routingMethod).Error
			Expect(err).To(BeNil())
		}

		createTemplate := func(router chi.Router, randoms *testutils.RandomValues, name string) string {
			url := fmt.Sprintf("%s/workspaces", urlTemplates)
			temp := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					ScopeType: api.SCOPETYPE_TENANT,
					Name:      name,
				},
				Spec: api.SupersetSpec{Image: convutils.Ref("ubuntu")},
			}

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](temp))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			var created api.WorkspaceTemplate
			Expect(json.Unmarshal(createResp.Body.Bytes(), &created)).To(Succeed())
			return created.Meta.Id
		}

		collectNames := func(entries []api.WorkspaceTemplatesListResponseItem) datastructs.Set[string] {
			names := datastructs.NewSet[string]()
			for _, entry := range entries {
				names.Add(entry.Meta.Name)
			}
			return names
		}

		It("filtering by routing method", func() {
			pathClusterId := randoms.ClusterId
			hostClusterId := anotherAllowedProject.ClusterId

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(pathClusterId, testing.BuildDisplayedClusterWithSubdomain(pathClusterId, testing.TestDefaultClusterVersion, false)).
				WithCluster(hostClusterId, testing.BuildDisplayedClusterWithSubdomain(hostClusterId, testing.TestDefaultClusterVersion, true)))
			checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

			existedId := createTemplate(router, randoms, "existed-ws")
			agnosticId := createTemplate(router, randoms, "agnostic-ws")
			injectRoutingMethod(agnosticId, models_crud.TemplatesColumnRoutingMethodAgnostic)
			pathId := createTemplate(router, randoms, "path-only-ws")
			injectRoutingMethod(pathId, models_crud.TemplatesColumnRoutingMethodPathBase)
			hostId := createTemplate(router, randoms, "host-only-ws")
			injectRoutingMethod(hostId, models_crud.TemplatesColumnRoutingMethodHostBase)

			_ = existedId

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			// No cluster filter → all templates returned
			listResult := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, randoms, url,
				fmt.Sprintf("workloadType==%s", api.WORKLOADTYPE_WORKSPACE))
			names := collectNames(listResult.Entries)
			Expect(names.Has("existed-ws")).To(BeTrue(), "existed template should be visible without cluster filter")
			Expect(names.Has("agnostic-ws")).To(BeTrue(), "agnostic template should be visible without cluster filter")
			Expect(names.Has("path-only-ws")).To(BeTrue(), "path-base template should be visible without cluster filter")
			Expect(names.Has("host-only-ws")).To(BeTrue(), "host-base template should be visible without cluster filter")

			// Path-base cluster → path-base, agnostic, and empty routing method templates
			listReq, err := http.NewRequest("GET", fmt.Sprintf("%s?filterBy=clusterId==%s&includeAncestors=true", url, pathClusterId), nil)
			Expect(err).To(BeNil())
			listReq.Header = *headerForAdmin(randoms)
			listResp := httptest.NewRecorder()
			router.ServeHTTP(listResp, listReq)
			Expect(listResp.Code).To(Equal(http.StatusOK))
			var pathResult api.WorkspaceTemplatesListResponse
			Expect(json.Unmarshal(listResp.Body.Bytes(), &pathResult)).To(Succeed())
			names = collectNames(pathResult.Entries)
			Expect(names.Has("existed-ws")).To(BeTrue(), "empty routing_method template should be visible on path cluster")
			Expect(names.Has("agnostic-ws")).To(BeTrue(), "agnostic template should be visible on path cluster")
			Expect(names.Has("path-only-ws")).To(BeTrue(), "path-base template should be visible on path cluster")
			Expect(names.Has("host-only-ws")).To(BeFalse(), "host-base template should NOT be visible on path cluster")

			// Host-base cluster → host-base, agnostic, and empty routing method templates
			listReq, err = http.NewRequest("GET", fmt.Sprintf("%s?filterBy=clusterId==%s&includeAncestors=true", url, hostClusterId), nil)
			Expect(err).To(BeNil())
			listReq.Header = *headerForAdmin(randoms)
			listResp = httptest.NewRecorder()
			router.ServeHTTP(listResp, listReq)
			Expect(listResp.Code).To(Equal(http.StatusOK))
			var hostResult api.WorkspaceTemplatesListResponse
			Expect(json.Unmarshal(listResp.Body.Bytes(), &hostResult)).To(Succeed())
			names = collectNames(hostResult.Entries)
			Expect(names.Has("existed-ws")).To(BeTrue(), "empty routing_method template should be visible on host cluster")
			Expect(names.Has("agnostic-ws")).To(BeTrue(), "agnostic template should be visible on host cluster")
			Expect(names.Has("host-only-ws")).To(BeTrue(), "host-base template should be visible on host cluster")
			Expect(names.Has("path-only-ws")).To(BeFalse(), "path-base template should NOT be visible on host cluster")
		})
	})
})

func countTemplatesForTest(router chi.Router, randoms *testutils.RandomValues, queryString string) *api.CountResponse {
	url := fmt.Sprintf("%s/count?%s", urlTemplates, queryString)

	listReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	listReq.Header = *headerForAdmin(randoms)

	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	Expect(listResp.Code).To(Equal(http.StatusOK))

	var result api.CountResponse
	jsonErr := json.Unmarshal(listResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

func createTemplateForTest[T templateCreationReq](router chi.Router, url string, req *T, randoms *testutils.RandomValues) {
	createReq, err := http.NewRequest("POST", url, makePayload(req))
	Expect(err).To(BeNil())
	createReq.Header = *headerForAdmin(randoms)

	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	Expect(createResp.Code).To(Equal(http.StatusOK))
}

func listTemplatesForTest(router chi.Router, randoms *testutils.RandomValues, queryString string) *api.TemplatesListResponse {
	listReq, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", urlTemplates, queryString), nil)
	Expect(err).To(BeNil())
	listReq.Header = *headerForAdmin(randoms)

	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	Expect(listResp.Code).To(Equal(http.StatusOK))

	var result api.TemplatesListResponse
	jsonErr := json.Unmarshal(listResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

func listTemplatesForTestType[T any](router chi.Router, randoms *testutils.RandomValues, url string, queryString string) *T {
	listReq, err := http.NewRequest("GET", fmt.Sprintf("%s?filterBy=%s", url, queryString), nil)
	Expect(err).To(BeNil())
	listReq.Header = *headerForAdmin(randoms)

	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	Expect(listResp.Code).To(Equal(http.StatusOK))

	var result T
	jsonErr := json.Unmarshal(listResp.Body.Bytes(), &result)
	Expect(jsonErr).To(BeNil())

	return &result
}

func expectAncestorsTest(listResult *api.WorkspaceTemplatesListResponse, expected []api.SupersetTemplate, desc string) {
	Expect(listResult).NotTo(BeNil(), desc)
	Expect(len(listResult.Entries)).To(Equal(len(expected)), desc)

	expectedNames := datastructs.NewSet[string]()

	for _, entry := range expected {
		expectedNames.Add(entry.Meta.Name)
	}

	actualNames := datastructs.NewSet[string]()

	for _, entry := range listResult.Entries {
		actualNames.Add(entry.Meta.Name)
	}

	Expect(expectedNames).To(Equal(actualNames), desc)
}
