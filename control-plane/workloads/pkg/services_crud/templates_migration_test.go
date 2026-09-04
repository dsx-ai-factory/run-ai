/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"strings"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("template validations", Serial, Ordered, func() {
	var randoms *testutils.RandomValues

	mockHybridOn := &bc_mocks.TenantSettingsClient{}
	mockHybridOff := &bc_mocks.TenantSettingsClient{}

	var router chi.Router

	BeforeEach(func() {
		testing.CleanTables()

		randoms = testutils.NewRandomValues()

		mockHybridOn.On("GetHybridSubmissionEnabled", mock.Anything).Return(true, nil)
		mockHybridOff.On("GetHybridSubmissionEnabled", mock.Anything).Return(false, nil)

		router = createRouterInt(mockHybridOff)

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authorization.SetMock(authMock)
		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		assets.SetMock(assets.NewSimpleMock())
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	// The shared process-wide tenant id (see testutils.NewRandomValues) means any
	// leftover template rows would leak into other suites' template counts.
	AfterEach(func() {
		testing.CleanTables()
	})

	Context("convert old templates", func() {
		It("should create new templates correctly", func() {

			//
			//   create assets
			//
			envAsset := assetsApi.NewAssetMeta("env", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			envAsset.ClusterId.Set(&randoms.ClusterId)

			computeAsset := assetsApi.NewAssetMeta("compute", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_COMPUTE, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			computeAsset.ClusterId.Set(&randoms.ClusterId)

			hostPathAsset := assetsApi.NewAssetMeta("host-path", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_HOST_PATH, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			hostPathAsset.ClusterId.Set(&randoms.ClusterId)

			gitAsset := assetsApi.NewAssetMeta("git", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_GIT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			gitAsset.ClusterId.Set(&randoms.ClusterId)

			//
			//   template ws only - all assets
			//
			tempWsAllAssets := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("ws-only", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
						Compute: *assetsApi.NewNullableAssetsRefCompute(&assetsApi.AssetsRefCompute{
							Id: computeAsset.Id,
						}),
						Datasources: []assetsApi.DatasourceRef{
							{
								Id:   hostPathAsset.Id,
								Kind: assetsApi.ASSETKIND_HOST_PATH,
							},
							{
								Id:   gitAsset.Id,
								Kind: assetsApi.ASSETKIND_GIT,
							},
						},
					},
				},
			}
			tempWsAllAssets.Meta.ClusterId.Set(&randoms.ClusterId)
			tempWsAllAssets.Meta.DepartmentId.Set(&randoms.DepartmentId)
			tempWsAllAssets.Meta.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
				Workspace:   *assetsApi.NewNullableBool(convutils.Ref(true)),
				Distributed: *assetsApi.NewNullableBool(convutils.Ref(true)),
			})

			//
			//   template ws and training - env + git
			//
			tempProjWsAndTrain := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("proj-ws-and-train", assetsApi.SCOPE_PROJECT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
						Datasources: []assetsApi.DatasourceRef{
							{
								Id:   gitAsset.Id,
								Kind: assetsApi.ASSETKIND_GIT,
							},
						},
					},
				},
			}
			tempProjWsAndTrain.Meta.ClusterId.Set(&randoms.ClusterId)
			tempProjWsAndTrain.Meta.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			//
			//   template for distributed only
			//
			tempDistributed := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("dist-only", assetsApi.SCOPE_PROJECT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
						Datasources: []assetsApi.DatasourceRef{
							{
								Id:   gitAsset.Id,
								Kind: assetsApi.ASSETKIND_GIT,
							},
						},
					},
				},
			}
			tempDistributed.Meta.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
				Distributed: *assetsApi.NewNullableBool(convutils.Ref(true)),
			})
			tempDistributed.Meta.ClusterId.Set(&randoms.ClusterId)
			tempDistributed.Meta.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			//
			//   provide their flat representation (not needed for the migration but might help us check the final result
			//
			envSpec := assetsApi.SupersetSpec{
				Image:         *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
				CreateHomeDir: *assetsApi.NewNullableBool(convutils.Ref(true)),
			}

			computeSpec := assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(convutils.Ref(assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(3))),
				})),
			}

			hostPathSpec := assetsApi.SupersetSpec{
				Storage: *assetsApi.NewNullableSupersetSpecAllOfStorage(convutils.Ref(assetsApi.SupersetSpecAllOfStorage{
					HostPath: []assetsApi.HostPathInstance{
						{
							Path: *assetsApi.NewNullableString(convutils.Ref("/host/path/1")),
						},
					},
				})),
			}

			gitSpec := assetsApi.SupersetSpec{
				Storage: *assetsApi.NewNullableSupersetSpecAllOfStorage(convutils.Ref(assetsApi.SupersetSpecAllOfStorage{
					Git: []assetsApi.GitInstance{
						{
							Path: *assetsApi.NewNullableString(convutils.Ref("/git/path/1")),
						},
					},
				})),
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &envSpec).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, computeAsset.Id, &computeSpec).
				WithAssetMeta(hostPathAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, hostPathAsset.Id, &hostPathSpec).
				WithAssetMeta(gitAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, gitAsset.Id, &gitSpec).
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_TRAINING, envAsset.Id, &envSpec).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_TRAINING, computeAsset.Id, &computeSpec).
				WithAssetMeta(hostPathAsset).WithMapping(api.WORKLOADTYPE_TRAINING, hostPathAsset.Id, &hostPathSpec).
				WithAssetMeta(gitAsset).WithMapping(api.WORKLOADTYPE_TRAINING, gitAsset.Id, &gitSpec).
				WithTemplate(tempWsAllAssets).
				WithTemplate(tempProjWsAndTrain).
				WithTemplate(tempDistributed))

			//
			//   now let's migrate
			//
			ctx := testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			err := MigrateTemplates(ctx)
			Expect(err).To(BeNil())

			expected := make(map[string]*api.SupersetTemplate)

			expected["ws-only-workspace"] = &api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:         "ws-only-workspace",
					WorkloadType: Workspace,
					ScopeType:    api.SCOPETYPE_DEPARTMENT,
					ScopeId:      randoms.DepartmentId,
					CreatedBy:    testing.TestUser,
					UpdatedBy:    convutils.Ref(testing.TestUser),
					ClusterId:    convutils.Ref(randoms.ClusterId),
					DepartmentId: convutils.Ref(randoms.DepartmentId),
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(envAsset.Id),
					Compute:     convutils.Ref(computeAsset.Id),
					Datasources: []string{
						gitAsset.Id,
						hostPathAsset.Id,
					},
				},
				CombinedSpec: api.SupersetSpec{
					CreateHomeDir: convutils.Ref(true),
					Image:         convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
					},
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						HostPath: &[]api.HostPathInstance{
							{
								Name: convutils.Ref("templt-xxx"),
								Path: convutils.Ref("/host/path/1"),
							},
						},
						Git: &[]api.GitInstance{
							{
								Name: convutils.Ref("templt-xxx"),
								Path: convutils.Ref("/git/path/1"),
							},
						},
					},
				},
				Runnable: api.RunnableInfo{
					IsRunnable: true,
				},
			}

			expected["proj-ws-and-train-workspace"] = &api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:         "proj-ws-and-train-workspace",
					WorkloadType: Workspace,
					ScopeType:    api.SCOPETYPE_PROJECT,
					ScopeId:      randoms.ProjectIdStr,
					CreatedBy:    testing.TestUser,
					UpdatedBy:    convutils.Ref(testing.TestUser),
					ClusterId:    convutils.Ref(randoms.ClusterId),
					DepartmentId: convutils.Ref(randoms.DepartmentId),
					ProjectId:    convutils.Ref(int32(randoms.ProjectId)),
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(envAsset.Id),
					Datasources: []string{
						gitAsset.Id,
					},
				},
				CombinedSpec: api.SupersetSpec{
					CreateHomeDir: convutils.Ref(true),
					Image:         convutils.Ref("ubuntu"),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Git: &[]api.GitInstance{
							{
								Name: convutils.Ref("templt-xxx"),
								Path: convutils.Ref("/git/path/1"),
							},
						},
					},
				},
				Runnable: api.RunnableInfo{
					IsRunnable: true,
				},
			}

			var forTraining api.SupersetTemplate
			errCopy := convutils.CopySpecs(&forTraining, expected["proj-ws-and-train-workspace"])
			Expect(errCopy).To(BeNil())

			forTraining.Meta.Name = "proj-ws-and-train-training"
			forTraining.Meta.WorkloadType = api.WORKLOADTYPE_TRAINING
			expected["proj-ws-and-train-training"] = &forTraining

			//
			//   list the templates
			//
			result := listTemplatesForTest(router, randoms, "")
			Expect(len(result.Entries)).To(Equal(3))

			for _, listedTemplate := range result.Entries {
				actualTemplate, getErr := templates.GetTemplate(ctx, &listedTemplate.Meta.WorkloadType, listedTemplate.Meta.Id)
				Expect(getErr).To(BeNil())

				expectedTemplate := expected[listedTemplate.Meta.Name]
				Expect(expectedTemplate).NotTo(BeNil(), "not found expected for", listedTemplate.Meta.Id)
				Expect(actualTemplate.template).NotTo(BeNil())

				// avoid flaky diffs
				actualTemplate.template.Meta.Id = ""
				actualTemplate.template.Meta.CreatedAt = time.Time{}
				actualTemplate.template.Meta.UpdatedAt = nil

				expectedDataSources := datastructs.NewSet(expectedTemplate.Assets.Datasources...)
				actualDataSources := datastructs.NewSet(actualTemplate.template.Assets.Datasources...)
				expectedTemplate.Assets.Datasources = nil
				actualTemplate.template.Assets.Datasources = nil

				if actualTemplate.template.CombinedSpec.Storage != nil {
					if actualTemplate.template.CombinedSpec.Storage.Git != nil {
						for _, git := range *actualTemplate.template.CombinedSpec.Storage.Git {
							fixName(git.Name)
						}
					}

					if actualTemplate.template.CombinedSpec.Storage.HostPath != nil {
						for _, hostPath := range *actualTemplate.template.CombinedSpec.Storage.HostPath {
							fixName(hostPath.Name)
						}
					}
				}

				testutils.Expect(*actualTemplate.template).ToEqual(expectedTemplate)
				Expect(actualDataSources).To(Equal(expectedDataSources))
			}
		})
	})

	Context("auto-migration mechanism", func() {
		It("should migrate upon first time user enables hubrid submission", func() {
			//
			//   create assets
			//
			envAsset := assetsApi.NewAssetMeta("env-1", assetsApi.SCOPE_CLUSTER, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			envAsset.ClusterId.Set(&randoms.ClusterId)

			//
			//   create templates
			//
			tempWsOnly := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("ws-only", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
					},
				},
			}
			tempWsOnly.Meta.ClusterId.Set(&randoms.ClusterId)
			tempWsOnly.Meta.DepartmentId.Set(&randoms.DepartmentId)
			tempWsOnly.Meta.WorkloadSupportedTypes = *assetsApi.NewNullableWorkloadSupportedTypes(&assetsApi.WorkloadSupportedTypes{
				Workspace: *assetsApi.NewNullableBool(convutils.Ref(true)),
			})

			tempBoth := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("both", assetsApi.SCOPE_PROJECT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
					},
				},
			}
			tempBoth.Meta.ClusterId.Set(&randoms.ClusterId)
			tempBoth.Meta.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			envSpec := assetsApi.SupersetSpec{
				Image:         *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
				CreateHomeDir: *assetsApi.NewNullableBool(convutils.Ref(true)),
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &envSpec).
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_TRAINING, envAsset.Id, &envSpec).
				WithTemplate(tempWsOnly).
				WithTemplate(tempBoth))

			//
			//   hybrid submission off
			//
			router1 := createRouterInt(mockHybridOff)
			result := listTemplatesForTest(router1, randoms, "")
			Expect(len(result.Entries)).To(Equal(0))

			//
			//  hybrid submission on
			//
			router2 := createRouterInt(mockHybridOn)
			result = listTemplatesForTest(router2, randoms, "")
			Expect(len(result.Entries)).To(Equal(3))

			expected := datastructs.NewSet([]string{
				"both-workspace",
				"both-training",
				"ws-only-workspace",
			}...)

			actual := datastructs.NewSet[string]()

			for _, entry := range result.Entries {
				actual.Add(entry.Meta.Name)
			}

			Expect(actual).To(Equal(expected))

			//
			//  user now adds another assets based template
			//
			tempAnother := assetsApi.WorkloadTemplate{
				Meta: *assetsApi.NewAssetMeta("another", assetsApi.SCOPE_PROJECT, uuid.NewString(),
					assetsApi.ASSETKIND_WORKLOAD_TEMPLATE, testing.TestUser, time.Now(), testing.TestUser, time.Now()),
				Spec: assetsApi.SpecificRunInfoFields{
					Assets: assetsApi.AssetsRef{
						Environment: assetsApi.EnvironmentAssetRef{
							Id: envAsset.Id,
						},
					},
				},
			}
			tempAnother.Meta.ClusterId.Set(&randoms.ClusterId)
			tempAnother.Meta.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &envSpec).
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_TRAINING, envAsset.Id, &envSpec).
				WithTemplate(tempWsOnly).
				WithTemplate(tempBoth).
				WithTemplate(tempAnother))

			//
			//  list again - there is no change in the templates
			//
			result = listTemplatesForTest(router2, randoms, "")
			Expect(len(result.Entries)).To(Equal(3))
		})
	})
})

func fixName(name *string) {
	if name == nil {
		return
	}
	*name = strings.Split(*name, "-")[0] + "-xxx"
}
