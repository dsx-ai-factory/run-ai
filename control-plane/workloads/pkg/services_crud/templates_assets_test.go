/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("templates assets test", func() {
	var randoms *testutils.RandomValues

	router := CreateRouter()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		Expect(randoms).NotTo(BeNil())

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authorization.SetMock(authMock)
		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))

		assets.SetMock(assets.NewSimpleMock())
	})

	// The shared process-wide tenant id (see testutils.NewRandomValues) means any
	// leftover template rows would leak into other suites' template counts.
	AfterEach(func() {
		testing.CleanTables()
	})

	Context("collect assets", func() {
		It("should build the map correctly", func() {
			linkedAsset := make(mapAssetIdToReplicas)

			collectAssets(linkedAsset, &api.TemplateAssetsIdsOptional{
				Environment: convutils.Ref("env-w"),
				Compute:     convutils.Ref("compute-wm"),
				Datasources: []string{
					"git-w",
					"pvc-wm",
					"git-wm",
				},
			}, workerReplica)

			collectAssets(linkedAsset, &api.TemplateAssetsIdsOptional{
				Environment: convutils.Ref("env-m"),
				Compute:     convutils.Ref("compute-m"),
				Datasources: []string{
					"pvc-m",
					"git-wm",
				},
			}, masterReplica)

			collectAssets(linkedAsset, &api.TemplateAssetsIdsOptional{
				Compute: convutils.Ref("compute-wm"),
				Datasources: []string{
					"pvc-wm",
				},
			}, workerAndMaster)

			collectAssets(linkedAsset, &api.TemplateAssetsIdsOptional{
				Datasources: []string{
					"s3-wm",
				},
			}, workerAndMaster)

			collectAssets(linkedAsset, &api.TemplateAssetsIdsOptional{
				Compute: convutils.Ref("compute-m"),
			}, masterReplica)

			Expect(linkedAsset).To(Equal(mapAssetIdToReplicas{
				"env-w":      workerReplica.withCategory(models_crud.AssetCategoryEnvironment),
				"env-m":      masterReplica.withCategory(models_crud.AssetCategoryEnvironment),
				"compute-m":  masterReplica.withCategory(models_crud.AssetCategoryCompute),
				"compute-wm": workerAndMaster.withCategory(models_crud.AssetCategoryCompute),
				"git-w":      workerReplica.withCategory(models_crud.AssetCategoryDatasource),
				"git-wm":     workerAndMaster.withCategory(models_crud.AssetCategoryDatasource),
				"pvc-m":      masterReplica.withCategory(models_crud.AssetCategoryDatasource),
				"pvc-wm":     workerAndMaster.withCategory(models_crud.AssetCategoryDatasource),
				"s3-wm":      workerAndMaster.withCategory(models_crud.AssetCategoryDatasource),
			}))
		})

	})

	Context("asset hooks", func() {
		It("should update the templates correctly", func() {
			//
			//   create assets
			//
			envAsset := assetsApi.NewAssetMeta("env", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			envAsset.ClusterId.Set(&randoms.ClusterId)
			envAsset.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			computeAsset := assetsApi.NewAssetMeta("compute", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_COMPUTE, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			computeAsset.ClusterId.Set(&randoms.ClusterId)
			computeAsset.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

			hostPathAsset := assetsApi.NewAssetMeta("host-path", assetsApi.SCOPE_PROJECT, uuid.NewString(),
				assetsApi.ASSETKIND_HOST_PATH, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			hostPathAsset.ClusterId.Set(&randoms.ClusterId)
			hostPathAsset.ProjectId.Set(convutils.Ref(int32(randoms.ProjectId)))

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
							Path: *assetsApi.NewNullableString(convutils.Ref("/path1")),
						},
					},
				})),
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, envAsset.Id, &envSpec).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, computeAsset.Id, &computeSpec).
				WithAssetMeta(hostPathAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, hostPathAsset.Id, &hostPathSpec))

			//
			//  create two templates that uses these assets
			//
			projTemplate1 := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "dist-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Spec: api.SupersetSpec{
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						HostPath: &[]api.HostPathInstance{
							{
								Path: convutils.Ref("/another/worker/path"),
							},
						},
					},
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(envAsset.Id),
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Storage: &api.WorkspaceSpecSpecAllOfStorage{
							HostPath: &[]api.HostPathInstance{
								{
									Path: convutils.Ref("/another/master/path"),
								},
							},
						},
					},
					Assets: &api.TemplateAssetsIdsOptional{
						Environment: convutils.Ref(envAsset.Id),
						Compute:     convutils.Ref(computeAsset.Id),
						Datasources: []string{
							hostPathAsset.Id,
						},
					},
				},
			}

			projTemplate2 := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "dist-2",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Spec: api.SupersetSpec{
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(12)),
					},
				},
				Master: &api.TemplateMasterSection{
					Assets: &api.TemplateAssetsIdsOptional{
						Environment: convutils.Ref(envAsset.Id),
						Datasources: []string{
							hostPathAsset.Id,
						},
					},
				},
			}

			shouldNotChange := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "dist-3",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   randoms.ProjectIdStr,
				},
				Spec: api.SupersetSpec{
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(12)),
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(5)),
						},
					},
				},
			}

			url := fmt.Sprintf("%s/distributed", urlTemplates)

			for _, template := range []api.SupersetTemplate{projTemplate1, projTemplate2, shouldNotChange} {
				createReq, err := http.NewRequest("POST", url, composeCreationReq[api.DistributedTemplateCreationRequest](template))
				Expect(err).To(BeNil())
				createReq.Header = *headerForAdmin(randoms)

				createResp := httptest.NewRecorder()
				router.ServeHTTP(createResp, createReq)
				Expect(createResp.Code).To(Equal(http.StatusOK))
			}

			//
			//   change the spec of all three assets
			//
			envSpec = assetsApi.SupersetSpec{
				Image:         *assetsApi.NewNullableString(convutils.Ref("ubuntu-updated")),
				CreateHomeDir: *assetsApi.NewNullableBool(convutils.Ref(false)),
			}

			computeSpec = assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(convutils.Ref(assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(6))),
				})),
			}

			hostPathSpec = assetsApi.SupersetSpec{
				Storage: *assetsApi.NewNullableSupersetSpecAllOfStorage(convutils.Ref(assetsApi.SupersetSpecAllOfStorage{
					HostPath: []assetsApi.HostPathInstance{
						{
							Path: *assetsApi.NewNullableString(convutils.Ref("/path1/updated")),
						},
					},
				})),
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, envAsset.Id, &envSpec).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, computeAsset.Id, &computeSpec).
				WithAssetMeta(hostPathAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, hostPathAsset.Id, &hostPathSpec))

			//
			//   issue asset modify to all three assets
			//
			AssetModified(context.Background(), &events.AssetModified{
				TenantId: randoms.TenantIdStr,
				AssetId:  envAsset.Id,
			})

			AssetModified(context.Background(), &events.AssetModified{
				TenantId: randoms.TenantIdStr,
				AssetId:  computeAsset.Id,
			})

			AssetModified(context.Background(), &events.AssetModified{
				TenantId: randoms.TenantIdStr,
				AssetId:  hostPathAsset.Id,
			})

			//
			//  list the templates
			//
			listResult := listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==project&sortBy=name")
			Expect(len(listResult.Entries)).To(Equal(3))

			//
			//  validate master spec of template #1
			//
			expectedDest1Spec := api.SupersetSpec{
				// from environment
				Image:         convutils.Ref("ubuntu-updated"),
				CreateHomeDir: convutils.Ref(false),

				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Path: convutils.Ref("/another/worker/path"),
						},
					},
				},
			}

			expectedDest1MasterSpec := api.MasterSpec{
				// from environment
				Image:         convutils.Ref("ubuntu-updated"),
				CreateHomeDir: convutils.Ref(false),

				// from compute
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(6)),
				},

				// from host path asset
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Name: convutils.Ref("templt-0"),
							Path: convutils.Ref("/path1/updated"),
						},
						{
							Path: convutils.Ref("/another/master/path"),
						},
					},
				},
			}

			dest1 := listResult.Entries[0]
			testutils.Expect(dest1.CombinedSpec).ToEqual(expectedDest1Spec)
			testutils.Expect(dest1.CombinedMasterSpec).ToEqual(expectedDest1MasterSpec)
			testutils.Expect(dest1.Assets).ToEqual(projTemplate1.Assets)
			testutils.Expect(dest1.Master.Assets).ToEqual(projTemplate1.Master.Assets)

			//
			//  validate master spec of template #2 (only master changed)
			//
			expectedDest2Spec := api.MasterSpec{
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(12)),
				},
			}

			expectedDest2MasterSpec := api.MasterSpec{
				// from environment
				Image:         convutils.Ref("ubuntu-updated"),
				CreateHomeDir: convutils.Ref(false),

				// from host path asset
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Name: convutils.Ref("templt-0"),
							Path: convutils.Ref("/path1/updated"),
						},
					},
				},
			}

			dest2 := listResult.Entries[1]
			testutils.Expect(dest2.CombinedSpec).ToEqual(expectedDest2Spec)
			testutils.Expect(dest2.CombinedMasterSpec).ToEqual(expectedDest2MasterSpec)
			testutils.Expect(dest2.Assets).ToEqual(api.TemplateAssetsIdsOptional{})
			testutils.Expect(dest2.Master.Assets).ToEqual(projTemplate2.Master.Assets)

			// no change in dest3
			dest3 := listResult.Entries[2]
			testutils.Expect(dest3.CombinedSpec).ToEqual(shouldNotChange.Spec)
			testutils.Expect(dest3.CombinedMasterSpec).ToEqual(shouldNotChange.Master.Spec)

			//
			// delete the environment asset
			//
			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, computeAsset.Id, &computeSpec).
				WithAssetMeta(hostPathAsset).WithMapping(api.WORKLOADTYPE_DISTRIBUTED, hostPathAsset.Id, &hostPathSpec))

			AssetDeleted(context.Background(), &events.AssetDeleted{
				TenantId: randoms.TenantIdStr,
				AssetId:  envAsset.Id,
			})

			//
			//  list the templates
			//
			listResult = listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==project&sortBy=name")
			Expect(len(listResult.Entries)).To(Equal(3))

			//
			//  validate master spec of template #1
			//
			expectedDest1Spec = api.SupersetSpec{
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Path: convutils.Ref("/another/worker/path"),
						},
					},
				},
			}

			expectedDest1MasterSpec = api.MasterSpec{
				// from compute
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(6)),
				},

				// from host path asset
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Name: convutils.Ref("templt-0"),
							Path: convutils.Ref("/path1/updated"),
						},
						{
							Path: convutils.Ref("/another/master/path"),
						},
					},
				},
			}

			remainingAssets := api.TemplateAssetsIdsOptional{
				Compute: convutils.Ref(computeAsset.Id),
				Datasources: []string{
					hostPathAsset.Id,
				},
			}

			dest1 = listResult.Entries[0]
			testutils.Expect(dest1.CombinedSpec).ToEqual(expectedDest1Spec)
			testutils.Expect(dest1.CombinedMasterSpec).ToEqual(expectedDest1MasterSpec)
			testutils.Expect(dest1.Assets).ToEqual(api.TemplateAssetsIdsOptional{}) // no assets remains
			testutils.Expect(dest1.Master.Assets).ToEqual(remainingAssets)

			//
			//  validate master spec of template #2 (only master changed)
			//
			expectedDest2Spec = api.MasterSpec{
				Compute: &api.WorkspaceSpecSpecAllOfCompute{
					GpuDevicesRequest: convutils.Ref(int32(12)),
				},
			}

			expectedDest2MasterSpec = api.MasterSpec{
				// from host path asset
				Storage: &api.WorkspaceSpecSpecAllOfStorage{
					HostPath: &[]api.HostPathInstance{
						{
							Name: convutils.Ref("templt-0"),
							Path: convutils.Ref("/path1/updated"),
						},
					},
				},
			}

			dest2 = listResult.Entries[1]
			testutils.Expect(dest2.CombinedSpec).ToEqual(expectedDest2Spec)
			testutils.Expect(dest2.CombinedMasterSpec).ToEqual(expectedDest2MasterSpec)

			// no change in dest3
			dest3 = listResult.Entries[2]
			testutils.Expect(dest3.CombinedSpec).ToEqual(shouldNotChange.Spec)
			testutils.Expect(dest3.CombinedMasterSpec).ToEqual(shouldNotChange.Master.Spec)
		})
	})

	Context("asset hooks - workspace template", func() {
		It("should update the templates correctly", func() {
			//
			//   create assets
			//
			envAsset := assetsApi.NewAssetMeta("env", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
				assetsApi.ASSETKIND_ENVIRONMENT, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			envAsset.ClusterId.Set(&randoms.ClusterId)
			envAsset.DepartmentId.Set(convutils.Ref(randoms.DepartmentId))

			computeAsset := assetsApi.NewAssetMeta("compute", assetsApi.SCOPE_DEPARTMENT, uuid.NewString(),
				assetsApi.ASSETKIND_COMPUTE, testing.TestUser, time.Now(), testing.TestUser, time.Now())
			computeAsset.ClusterId.Set(&randoms.ClusterId)
			computeAsset.DepartmentId.Set(convutils.Ref(randoms.DepartmentId))

			envSpec := assetsApi.SupersetSpec{
				Image:         *assetsApi.NewNullableString(convutils.Ref("ubuntu")),
				CreateHomeDir: *assetsApi.NewNullableBool(convutils.Ref(true)),
			}

			computeSpec := assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(convutils.Ref(assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(3))),
				})),
			}

			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &envSpec).
				WithAssetMeta(computeAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, computeAsset.Id, &computeSpec))

			//
			//  create a template that uses these assets
			//
			projTemplate := api.SupersetTemplate{
				Meta: api.TemplateMetaFields{
					Name:      "ws-1",
					ScopeType: api.SCOPETYPE_DEPARTMENT,
					ScopeId:   randoms.DepartmentId,
				},
				Assets: &api.TemplateAssetsIdsOptional{
					Environment: convutils.Ref(envAsset.Id),
					Compute:     convutils.Ref(computeAsset.Id),
				},
			}

			url := fmt.Sprintf("%s/workspaces", urlTemplates)

			createReq, err := http.NewRequest("POST", url, composeCreationReq[api.WorkspaceTemplateCreationRequest](projTemplate))
			Expect(err).To(BeNil())
			createReq.Header = *headerForAdmin(randoms)

			createResp := httptest.NewRecorder()
			router.ServeHTTP(createResp, createReq)
			Expect(createResp.Code).To(Equal(http.StatusOK))

			//
			// delete the compute asset
			//
			assets.SetMock(assets.NewSimpleMock().
				WithAssetMeta(envAsset).WithMapping(api.WORKLOADTYPE_WORKSPACE, envAsset.Id, &envSpec))

			AssetDeleted(context.Background(), &events.AssetDeleted{
				TenantId: randoms.TenantIdStr,
				AssetId:  computeAsset.Id,
			})

			listResult := listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms, url, "scopeType==department&sortBy=name")
			Expect(len(listResult.Entries)).To(Equal(1))

			Expect(listResult.Entries[0].Assets).To(Equal(api.TemplateAssetsIds{
				Environment: convutils.Ref(envAsset.Id),
			}))
		})
	})

})
