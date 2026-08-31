/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

const (
	urlDatasource = "/api/v1/asset/datasource"
)

type ExpectedVsActualInt struct {
	Expected int
	Actual   int
}

var _ = Describe("datasources specific tests", func() {
	router := CreateRouter()

	var policyMockUpdateTime time.Time

	randoms := testutils.NewRandomValues()
	differentProject := randoms.CloneWithDifferentProject()

	var nfsAsset api.NfsCreationRequest
	var pvcAsset api.PvcCreationRequest

	BeforeEach(func() {
		dbTesting.CleanTables()

		setupDatavolumeClientWithGetDataVolumesResponse()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject))
		authorization.SetMock(authMock)

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		configMapCreReq := testing.BuildRandomProjectConfigMapAsset(randoms)
		pvcAsset = testing.BuildRandomProjectPVCAsset(randoms)

		createConfigMapAsset(configMapCreReq, router, headerForAdmin(randoms), http.StatusAccepted)

		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
			WithPvc(randoms.ClusterId, &pvcAsset, nil).
			WithConfigMap(randoms.ClusterId, &configMapCreReq, nil))

		createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
		createComputeAsset(buildRandomProjectCompute(randoms), router, headerForAdmin(randoms))
		createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))
		createHostPathAsset(buildRandomProjectHostPathAsset(differentProject), router, headerForAdmin(randoms))
		createGitAsset(buildRandomProjectGit(randoms, nil), router, headerForAdmin(randoms))
		createS3Asset(buildRandomProjectS3(randoms, nil), router, headerForAdmin(randoms))
		createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, nil), router, headerForAdmin(randoms))

		nfsAsset = buildRandomTenantNFSAsset()
		createNFSAsset(nfsAsset, router, headerForAdmin(randoms))

		createPVCAsset(pvcAsset, router, headerForAdmin(randoms))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
	})

	Context("list data sources with various parameters", func() {
		It("should give the correct set", func() {
			//
			//    list all datasources
			//
			listHttpReq, err := http.NewRequest("GET", urlDatasource, nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())

			validateDataSources("list all", dsListResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_ENVIRONMENT:   {Expected: 0},
				api.ASSETKIND_COMPUTE:       {Expected: 0},
				api.ASSETKIND_HOST_PATH:     {Expected: 2},
				api.ASSETKIND_NFS:           {Expected: 1},
				api.ASSETKIND_PVC:           {Expected: 1},
				api.ASSETKIND_GIT:           {Expected: 1},
				api.ASSETKIND_S3:            {Expected: 1},
				api.ASSETKIND_CONFIG_MAP:    {Expected: 1},
				api.ASSETKIND_SECRET_VOLUME: {Expected: 1},
			})

			//
			//    list project X + tenant
			//
			listHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?%s", urlDatasource, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp = httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())
			validateDataSources("project + tenants", dsListResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_ENVIRONMENT:   {Expected: 0},
				api.ASSETKIND_COMPUTE:       {Expected: 0},
				api.ASSETKIND_HOST_PATH:     {Expected: 1},
				api.ASSETKIND_NFS:           {Expected: 1},
				api.ASSETKIND_PVC:           {Expected: 1},
				api.ASSETKIND_GIT:           {Expected: 1},
				api.ASSETKIND_S3:            {Expected: 1},
				api.ASSETKIND_CONFIG_MAP:    {Expected: 1},
				api.ASSETKIND_SECRET_VOLUME: {Expected: 1},
			})
		})
	})

	Context("list data sources with incorrect query params", func() {
		It("should raise error", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=xxxyyyzzz", urlDatasource), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())
			Expect(listHttpResp.Body.String()).To(ContainSubstring("invalid value provided for scope parameter"))
		})
	})

	Context("test usage info", func() {
		DescribeTable("usage info should be correct", func(workloadType string) {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			pvcCreReq := testing.BuildRandomProjectPVCAsset(randoms)
			configMapCreReq := testing.BuildRandomProjectConfigMapAsset(randoms)

			configMapAsset, _, _ := createConfigMapAsset(configMapCreReq, router, headerForAdmin(randoms), http.StatusAccepted)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithConfigMap(randoms.ClusterId, &configMapCreReq, nil))

			volumeAsset, _, _ := createHostPathAsset(buildRandomTenantHostPathAsset(), router, headerForAdmin(randoms))
			pvcAsset, _, _ := createPVCAsset(pvcCreReq, router, headerForAdmin(randoms))
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithConfigMap(randoms.ClusterId, &configMapCreReq, nil).
				WithPvc(randoms.ClusterId, &pvcCreReq, nil))

			nfsAsset, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, nil), router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, nil), router, headerForAdmin(randoms))
			secretVolumeAsset, _, _ := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, nil), router, headerForAdmin(randoms))

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

			createWorkloadRequest := []api.WorkloadCreationRequest{
				{
					Name:      "pvc-nfs-s3",
					ClusterId: randoms.ClusterId,
					ProjectId: int32(randoms.ProjectId),
					Namespace: fmt.Sprintf("runai-proj-%d", randoms.ProjectId),
					Assets: api.AssetsIds{
						Environment: envAsset.Meta.Id,
						Datasources: []api.AssetIdAndKind{
							{Id: pvcAsset.Meta.Id, Kind: pvcAsset.Meta.Kind},
							{Id: nfsAsset.Meta.Id, Kind: nfsAsset.Meta.Kind},
							{Id: s3Asset.Meta.Id, Kind: s3Asset.Meta.Kind},
						},
					},
				},
				{
					Name:      "volume-nfs-git",
					ClusterId: randoms.ClusterId,
					ProjectId: int32(randoms.ProjectId),
					Namespace: fmt.Sprintf("runai-proj-%d", randoms.ProjectId),
					Assets: api.AssetsIds{
						Environment: envAsset.Meta.Id,
						Datasources: []api.AssetIdAndKind{
							{Id: volumeAsset.Meta.Id, Kind: volumeAsset.Meta.Kind},
							{Id: nfsAsset.Meta.Id, Kind: nfsAsset.Meta.Kind},
							{Id: gitAsset.Meta.Id, Kind: gitAsset.Meta.Kind},
						},
					},
				},
				{
					Name:      "configmap-secret-volume",
					ClusterId: randoms.ClusterId,
					ProjectId: int32(randoms.ProjectId),
					Namespace: fmt.Sprintf("runai-proj-%d", randoms.ProjectId),
					Assets: api.AssetsIds{
						Environment: envAsset.Meta.Id,
						Datasources: []api.AssetIdAndKind{
							{Id: secretVolumeAsset.Meta.Id, Kind: secretVolumeAsset.Meta.Kind},
							{Id: configMapAsset.Meta.Id, Kind: configMapAsset.Meta.Kind},
						},
					},
				},
			}

			for _, req := range createWorkloadRequest {
				sort.Slice(req.Assets.Datasources, func(i, j int) bool {
					return req.Assets.Datasources[i].Id < req.Assets.Datasources[j].Id
				})

				createWsPayload, err := json.Marshal(req)
				Expect(err).To(BeNil())

				createWsHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], strings.NewReader(string(createWsPayload)))
				Expect(err).To(BeNil())
				createWsHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, createWsHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))
			}

			//  list with usage, currently not used by any workload/template
			listDsHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", urlDatasource), nil)
			Expect(err).To(BeNil())
			listDsHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, listDsHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var listPayload api.DatasourceListResponse
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &listPayload)
			Expect(err).To(BeNil())

			actualUsage := map[api.AssetKind]datastructs.Set[string]{}

			for _, entry := range listPayload.Entries {
				kind := entry.Meta.Kind
				if actualUsage[kind] == nil {
					actualUsage[kind] = datastructs.NewSet[string]()
				}
				if entry.UsedBy != nil {
					for _, ws := range entry.UsedBy.Workspaces {
						actualUsage[kind].Add(ws.Name)
					}
					for _, training := range entry.UsedBy.Trainings {
						actualUsage[kind].Add(training.Name)
					}
					for _, template := range entry.UsedBy.Templates {
						actualUsage[kind].Add(template.Name)
					}
				}
			}

			expectedUsage := map[api.AssetKind]datastructs.Set[string]{
				api.ASSETKIND_HOST_PATH:     {"volume-nfs-git": false},
				api.ASSETKIND_NFS:           {"volume-nfs-git": false, "pvc-nfs-s3": false},
				api.ASSETKIND_PVC:           {"pvc-nfs-s3": false},
				api.ASSETKIND_GIT:           {"volume-nfs-git": false},
				api.ASSETKIND_S3:            {"pvc-nfs-s3": false},
				api.ASSETKIND_CONFIG_MAP:    {"configmap-secret-volume": false},
				api.ASSETKIND_SECRET_VOLUME: {"configmap-secret-volume": false},
			}

			Expect(actualUsage).To(Equal(expectedUsage))
		}, forWorkspaces, forTrainings)
	})

	Context("test policy compliance info - positive", func() {
		DescribeTable("compliance info should be correct", func(workloadType string) {
			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
				},
			})

			//  list with compliance info
			listDsHttpReq, err := http.NewRequest("GET", fmt.Sprintf(
				"%s?complyToProject=%d&complyToWorkloadType=Training", urlDatasource, randoms.ProjectId), nil)
			Expect(err).To(BeNil())
			listDsHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listDsHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var listPayload api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &listPayload)
			Expect(err).To(BeNil())
			Expect(len(listPayload.Entries)).To(BeNumerically(">", 0))

			for _, entry := range listPayload.Entries {
				switch entry.Meta.Kind {
				case api.ASSETKIND_PVC, api.ASSETKIND_CONFIG_MAP:
					Expect(entry.Compliance.Compliance).To(Equal(false))
					Expect(entry.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
						{
							Details: "This data source can't be selected yet, because it is being synchronized to the cluster.",
						},
					}))
				default:
					Expect(entry.Compliance.Compliance).To(Equal(true), entry.Meta.Name, entry.Meta.Kind)
				}
			}
		}, forWorkspaces)
	})

	Context("test policy compliance info - negative", func() {
		DescribeTable("compliance info should be correct", func(workloadType string) {
			trainingPolicy := policyService.TrainingPolicyDefaultsAndRulesV2{
				Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
					Storage: *policyService.NewNullableSupersetRulesAllOfStorage(&policyService.SupersetRulesAllOfStorage{
						Nfs: *policyService.NewNullableNfssRules(&policyService.NfssRules{
							Attributes: *policyService.NewNullableNfsRules(&policyService.NfsRules{
								Path: *policyService.NewNullableStringRules(&policyService.StringRules{
									Options: []policyService.StringOption{
										{Value: "xxx"},
										{Value: "yyy"},
									},
								}),
							}),
						}),
						Pvc: *policyService.NewNullablePvcsRules(&policyService.PvcsRules{
							Attributes: *policyService.NewNullablePvcRules(&policyService.PvcRules{
								Path: *policyService.NewNullableStringRules(&policyService.StringRules{
									Options: []policyService.StringOption{
										{Value: "aaa"},
										{Value: "bbb"},
									},
								}),
							}),
						}),
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt:        time.Now(),
					EffectiveTrainingPolicyV2: flat_workloads.ConvertToTrainingPolicy(&trainingPolicy),
				},
			})

			//  list with compliance info
			listDsHttpReq, err := http.NewRequest("GET", fmt.Sprintf(
				"%s?complyToProject=%d&complyToWorkloadType=Training", urlDatasource, randoms.ProjectId), nil)
			Expect(err).To(BeNil())
			listDsHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listDsHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var listPayload api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &listPayload)
			Expect(err).To(BeNil())
			Expect(len(listPayload.Entries)).To(BeNumerically(">", 0))

			for _, entry := range listPayload.Entries {
				if entry.Meta.Kind == api.ASSETKIND_PVC {
					Expect(entry.Compliance.Compliance).To(Equal(false))
					Expect(entry.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
						{
							Field:   convutils.Ref("storage.pvc.attributes.path"),
							Details: "the value of this field must be one of: aaa, bbb",
							Rule:    convutils.Ref(api.POLICYRULEENUM_OPTIONS),
						},
					}))
				} else if entry.Meta.Kind == api.ASSETKIND_NFS {
					Expect(entry.Compliance.Compliance).To(Equal(false))
					Expect(entry.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
						{
							Field:   convutils.Ref("storage.nfs.attributes.path"),
							Details: "the value of this field must be one of: xxx, yyy",
							Rule:    convutils.Ref(api.POLICYRULEENUM_OPTIONS),
						},
					}))
				} else if entry.Meta.Kind == api.ASSETKIND_CONFIG_MAP {
					Expect(entry.Compliance.Compliance).To(Equal(false))
					Expect(entry.Compliance.Reason).To(Equal([]api.ComplianceInfoReason{
						{
							Details: "This data source can't be selected yet, because it is being synchronized to the cluster.",
						},
					}))
				} else {
					Expect(entry.Compliance.Compliance).To(Equal(true), entry.Meta.Name, entry.Meta.Kind)
				}
			}
		}, forWorkspaces)
	})

	Context("test status info - positive", func() {
		DescribeTable("status info should be correct", func(workloadType string) {
			listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?statusInfo=true", urlDatasource), nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			listAllHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listAllHttpResp, listAllHttpReq)
			Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

			var listResponse GenericListResponse
			err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
			Expect(err).To(BeNil())

			noIssues := 0

			for _, entry := range listResponse.Entries {
				if entry.Status != nil {
					if entry.Status.Status == api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND {
						noIssues += 1
					} else {
						Expect(entry.Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_CREATING))
					}
				}
			}

			Expect(noIssues).To(Equal(1)) // only secret volume
		}, forWorkspaces)
	})

	Context("test type parameter - positive tests", func() {
		It("should filter datasources by type parameter", func() {
			// Test filtering by each valid type
			types := []struct {
				typeParam    string
				expectedKind api.AssetKind
			}{
				{"hostPath", api.ASSETKIND_HOST_PATH},
				{"nfs", api.ASSETKIND_NFS},
				{"pvc", api.ASSETKIND_PVC},
				{"git", api.ASSETKIND_GIT},
				{"s3", api.ASSETKIND_S3},
				{"configMap", api.ASSETKIND_CONFIG_MAP},
				{"secretVolume", api.ASSETKIND_SECRET_VOLUME},
			}

			for _, tt := range types {
				listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=%s", urlDatasource, tt.typeParam), nil)
				Expect(err).To(BeNil())
				listHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var dsListResp api.DatasourceListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
				Expect(err).To(BeNil())

				// Verify all returned entries match the requested type
				Expect(len(dsListResp.Entries)).To(BeNumerically(">", 0), fmt.Sprintf("Expected at least one %s datasource", tt.typeParam))
				for _, entry := range dsListResp.Entries {
					Expect(entry.Meta.Kind).To(Equal(tt.expectedKind), fmt.Sprintf("Expected only %s datasources when filtering by type=%s", tt.expectedKind, tt.typeParam))
				}
			}
		})

		It("should return only hostPath datasources when type=hostPath", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=hostPath", urlDatasource), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())

			validateDataSources("type=hostPath", dsListResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_HOST_PATH: {Expected: 2},
			})
		})

		It("should return only nfs datasources when type=nfs", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=nfs", urlDatasource), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())

			validateDataSources("type=nfs", dsListResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_NFS: {Expected: 1},
			})
		})

		It("should combine type parameter with other filters like projectId", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=hostPath&%s", urlDatasource, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())

			// Should only return hostPath datasources for the specific project
			validateDataSources("type=hostPath with projectId", dsListResp.Entries, map[api.AssetKind]ExpectedVsActualInt{
				api.ASSETKIND_HOST_PATH: {Expected: 1},
			})
		})

		It("should return empty list when filtering by type with no matching datasources", func() {
			// First, clean the tables and create only non-hostPath datasources
			dbTesting.CleanTables()
			createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))

			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=hostPath", urlDatasource), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())

			Expect(len(dsListResp.Entries)).To(Equal(0))
		})
	})

	Context("test type parameter - negative tests", func() {
		It("should reject type=dataVolume with includeDataVolumes=false", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=dataVolume", urlDatasource), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)

			// When type=dataVolume is specified without includeDataVolumes=true,
			// the service should return OK with empty results
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var dsListResp api.DatasourceListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &dsListResp)
			Expect(err).To(BeNil())
			Expect(len(dsListResp.Entries)).To(Equal(0))
		})

		It("should reject type parameter combined with includeDataVolumes when type is not dataVolume", func() {
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?type=hostPath&dataVolumes=true&%s", urlDatasource, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)

			// When includeDataVolumes=true is used with type other than dataVolume, it should fail
			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("The parameters includeDataVolumes and type are mutually exclusive"))
		})
	})

	Context("test includeDescendants param", func() {
		var anotherCluster *testutils.RandomValues
		var anotherProj *testutils.RandomValues

		BeforeEach(func() {
			dbTesting.CleanTables()

			anotherProj = randoms.CloneWithDifferentProject()
			anotherCluster = randoms.CloneWithDifferentCluster()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherProj, anotherCluster)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherProj, anotherCluster))
			authorization.SetMock(authMock)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18", false)).
				WithCluster(anotherCluster.ClusterId, testing.BuildDisplayedCluster(anotherCluster, "2.18", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		It("should return correct datasource assets with includeDescendants flag", func() {
			// Create datasource assets at different scope levels
			tenantNfs, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))
			clusterHostPath, _, _ := createHostPathAsset(buildRandomClusterHostPathAsset(randoms), router, headerForAdmin(randoms))
			anotherClusterHostPath, _, _ := createHostPathAsset(buildRandomClusterHostPathAsset(anotherCluster), router, headerForAdmin(anotherCluster))
			departmentGit, _, _ := createGitAsset(buildRandomDepartmentGit(randoms, nil), router, headerForAdmin(randoms))
			projectS3, _, _ := createS3Asset(buildRandomProjectS3(randoms, nil), router, headerForAdmin(randoms))
			anotherProjectS3, _, _ := createS3Asset(buildRandomProjectS3(anotherProj, nil), router, headerForAdmin(anotherProj))

			// Test with no params - expect to get all datasource assets
			resp := fetchDatasourceAssets(router, randoms, urlDatasource)
			expectDatasourceAssets(resp.Entries, []string{tenantNfs.Meta.Id, clusterHostPath.Meta.Id, anotherClusterHostPath.Meta.Id, departmentGit.Meta.Id, projectS3.Meta.Id, anotherProjectS3.Meta.Id})

			// Test with cluster id - expect to get the cluster asset and all assets that are above (1 tenant)
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s", urlDatasource, randoms.ClusterId))
			expectDatasourceAssets(resp.Entries, []string{clusterHostPath.Meta.Id, tenantNfs.Meta.Id})

			// Test with cluster id and includeDescendants - expect to get the cluster asset, all assets that are above (1 tenant) and all that are below (1 department, 2 projects)
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&includeDescendants=true", urlDatasource, randoms.ClusterId))
			expectDatasourceAssets(resp.Entries, []string{clusterHostPath.Meta.Id, tenantNfs.Meta.Id, departmentGit.Meta.Id, projectS3.Meta.Id, anotherProjectS3.Meta.Id})

			// Test with cluster id and scope - expect to get just the cluster asset
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&scope=cluster", urlDatasource, randoms.ClusterId))
			expectDatasourceAssets(resp.Entries, []string{clusterHostPath.Meta.Id})

			// Test with department id - expect to get the department asset and all assets that are above (1 cluster, 1 tenant)
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s", urlDatasource, randoms.DepartmentId))
			expectDatasourceAssets(resp.Entries, []string{departmentGit.Meta.Id, clusterHostPath.Meta.Id, tenantNfs.Meta.Id})

			// Test with department id and scope - expect to get just the department asset
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&scope=department", urlDatasource, randoms.DepartmentId))
			expectDatasourceAssets(resp.Entries, []string{departmentGit.Meta.Id})

			// Test with department id and includeDescendants - expect to get the department asset, all assets that are above (1 cluster, 1 tenant) and no descendants (department doesn't support includeDescendants yet)
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&includeDescendants=true", urlDatasource, randoms.DepartmentId))
			expectDatasourceAssets(resp.Entries, []string{clusterHostPath.Meta.Id, tenantNfs.Meta.Id, departmentGit.Meta.Id})

			// Test with project id - expect to get the project asset and all assets that are above (1 department, 1 cluster, 1 tenant)
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?projectId=%v", urlDatasource, randoms.ProjectId))
			expectDatasourceAssets(resp.Entries, []string{projectS3.Meta.Id, departmentGit.Meta.Id, clusterHostPath.Meta.Id, tenantNfs.Meta.Id})

			// Test with project id and scope - expect to get just the project asset
			resp = fetchDatasourceAssets(router, randoms, fmt.Sprintf("%s?projectId=%v&scope=project", urlDatasource, randoms.ProjectId))
			expectDatasourceAssets(resp.Entries, []string{projectS3.Meta.Id})
		})

		It("should reject includeDescendants with scope parameter", func() {
			// Test that includeDescendants and scope are mutually exclusive
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?clusterId=%s&scope=cluster&includeDescendants=true", urlDatasource, randoms.ClusterId), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)

			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("includeDescendants and scope are mutual exclusive"))
		})
	})
})

func fetchDatasourceAssets(router chi.Router, randoms *testutils.RandomValues, url string) api.DatasourceListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	listHttpReq.Header = *headerForAdmin(randoms)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK))

	var resp api.DatasourceListResponse
	respJsonErr := json.Unmarshal(listHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp
}

func expectDatasourceAssets(entries []api.DatasourceListResponseEntry, expectedIds []string) {
	Expect(len(entries)).To(Equal(len(expectedIds)), "Expected %d datasource assets but got %d", len(expectedIds), len(entries))

	actualIds := make([]string, 0, len(entries))
	for _, entry := range entries {
		actualIds = append(actualIds, entry.Meta.Id)
	}

	for _, expectedId := range expectedIds {
		Expect(actualIds).To(ContainElement(expectedId))
	}
}

func validateDataSources(desc string, entries []api.DatasourceListResponseEntry, counts map[api.AssetKind]ExpectedVsActualInt) {
	// Check actual assets in the response
	for _, entry := range entries {
		if countsOfAsset, ok := counts[entry.Meta.Kind]; ok {
			countsOfAsset.Actual += 1
			counts[entry.Meta.Kind] = countsOfAsset
		} else {
			counts[entry.Meta.Kind] = ExpectedVsActualInt{Actual: 1}
		}
	}

	// Do the comparison
	for kind, counts := range counts {
		Expect(counts.Actual).To(Equal(counts.Expected), fmt.Sprintf("%s: for kind %s", desc, kind))
	}
}
