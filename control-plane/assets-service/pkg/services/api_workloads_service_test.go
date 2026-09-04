/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlWorkspace        = "/api/v1/workspace"
	urlWorkspaceV2      = "/api/v2/workspace"
	urlTraining         = "/api/v1/training"
	urlTrainingV2       = "/api/v2/training"
	urlInference        = "/api/v1/inference"
	urlInferenceV2      = "/api/v2/inference"
	urlModelInference   = "/api/v1/model-inference"
	urlModelInferenceV2 = "/api/v2/model-inference"

	urlNameAvailability = "/api/v1/asset/name-availability"

	urlModelHuggingFaceInfo = "/api/v2/huggingface-model/info-request"
)

var urlForCreatingWorkload = map[string]string{
	auth.WorkloadTypeWorkspace:   urlWorkspace,
	auth.WorkloadTypeTraining:    urlTraining,
	auth.WorkloadTypeDistributed: urlTraining,
	auth.WorkloadTypeInference:   urlInference,
}

var urlV1toV2 = map[string]string{
	urlWorkspace: urlWorkspaceV2,
	urlTraining:  urlTrainingV2,
	urlInference: urlInferenceV2,
}

var (
	forWorkspaces  = Entry(auth.WorkloadTypeWorkspace, auth.WorkloadTypeWorkspace)
	forTrainings   = Entry(auth.WorkloadTypeTraining, auth.WorkloadTypeTraining)
	forDistributed = Entry(auth.WorkloadTypeDistributed, auth.WorkloadTypeDistributed)
)

var forInference = Entry(auth.WorkloadTypeInference, auth.WorkloadTypeInference)

// createWorkload creates a workload in a project and returns its ID.
func createWorkload(r *testutils.RandomValues, router chi.Router, envAsset api.EnvironmentAsset, wlType string) string {
	reqPayload := randomWorkloadCreationRequest(r, envAsset.Meta.Id)
	createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[wlType], makePayload(reqPayload))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(r)

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusOK))

	var inference api.Inference
	jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
	Expect(jsonErr).To(BeNil())

	return inference.Meta.Id
}

// expectWorkloadStatus checks the status of a workload
func expectWorkloadStatus(r *testutils.RandomValues, wlType string, router chi.Router, targetId string, expectedStatus int) {
	getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlForCreatingWorkload[wlType], targetId), nil)
	Expect(err).To(BeNil())
	getHttpReq.Header = *headerForAdmin(r)

	getHttpResp := httptest.NewRecorder()
	router.ServeHTTP(getHttpResp, getHttpReq)
	Expect(getHttpResp.Code).To(Equal(expectedStatus))
}

var _ = Describe("Test Workspace, Training and Inference Services", func() {
	var randoms *testutils.RandomValues
	var differentProject *testutils.RandomValues
	var differentDepartment *testutils.RandomValues
	var policyMockUpdateTime time.Time

	var envAsset api.EnvironmentAsset
	var envAssetInference api.EnvironmentAsset

	router := CreateRouter()

	notExist := uuid.NewString()

	BeforeEach(func() {
		dbTesting.CleanTables()

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		randoms = testutils.NewRandomValues()
		differentProject = randoms.CloneWithDifferentProject()
		differentDepartment = differentProject.CloneWithDifferentProject().CloneWithDifferentDepartment()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject, differentDepartment)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject, differentDepartment))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		auth.TestingAllowCreateSystemScopedAssets = true

		envAssetCreationReq := buildRandomTenantEnvironmentAsset(randoms)
		envAsset, _, _ = createEnvironmentAsset(envAssetCreationReq, router, headerForAdmin(randoms))
		addServingPortToEnvironmentCreationRequest(&envAssetCreationReq)
		envAssetCreationReq.Meta.Name = testutils.RandString(5)
		envAssetInference, _, _ = createEnvironmentAsset(envAssetCreationReq, router, headerForAdmin(randoms))
	})

	Context("trying to create workload with no permission", func() {
		DescribeTable("should return forbidden error", func(workloadType string) {
			reqPayload := randomWorkloadCreationRequest(randoms, getEnvAssetIdByType(workloadType, &envAsset, &envAssetInference))
			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerWithMissingInfo(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("Workload contains asset from different project", func() {
		DescribeTable("should fail with asset not found in project error", func(workloadType string) {
			nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(differentProject), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, getEnvAssetIdByType(workloadType, &envAsset, &envAssetInference))
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body).To(ContainSubstring("not associated with the requested project"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("Workload contains asset from different department", func() {
		DescribeTable("should fail with asset not found in project error", func(workloadType string) {
			nfsAsset, _, _ := createNFSAsset(buildRandomDepartmentNFSAsset(differentDepartment), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, getEnvAssetIdByType(workloadType, &envAsset, &envAssetInference))
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body).To(ContainSubstring("not associated with the requested project"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("incorrect asset type", func() {
		DescribeTable("should fail validation", func(workloadType string) {
			nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, getEnvAssetIdByType(workloadType, &envAsset, &envAssetInference))
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("nfs rather than git"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("workloadVolume asset in inference workload (old cluster)", func() {
		It("should fail validation", func() {
			// simulate an incompatible cluster version, where workload volumes are not supported
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.19", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

			pvcCreationRequest := testing.BuildRandomProjectPVCAsset(randoms)
			pvcCreationRequest.Meta.AutoDelete = convutils.Ref(true)
			pvcAsset, _, _ := createPVCAsset(pvcCreationRequest, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.WorkloadVolumes = []string{
				pvcAsset.Meta.Id,
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[auth.WorkloadTypeInference], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("inference workload cannot use workload volumes in this cluster version"))
		})
	})

	Context("workloadVolume asset in inference workload - ephemeral", func() {
		It("should fail validation", func() {
			pvcCreationRequest := testing.BuildRandomProjectPVCAsset(randoms)
			pvcCreationRequest.Meta.AutoDelete = convutils.Ref(true)
			pvcCreationRequest.Spec.Ephemeral = convutils.Ref(true) // this is not allowed
			pvcAsset, _, _ := createPVCAsset(pvcCreationRequest, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.WorkloadVolumes = []string{
				pvcAsset.Meta.Id,
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[auth.WorkloadTypeInference], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("inference workload cannot use ephemeral workload volumes"))
		})
	})

	Context("workloadVolume asset in inference workload - valid", func() {
		It("should succeed", func() {
			pvcCreationRequest := testing.BuildRandomProjectPVCAsset(randoms)
			pvcCreationRequest.Meta.AutoDelete = convutils.Ref(true)
			pvcAsset, _, _ := createPVCAsset(pvcCreationRequest, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.WorkloadVolumes = []string{
				pvcAsset.Meta.Id,
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[auth.WorkloadTypeInference], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Spec.Assets.WorkloadVolumes).To(HaveLen(1))
			Expect(inference.Spec.Assets.WorkloadVolumes[0]).To(Equal(pvcAsset.Meta.Id))
			field, _ := getNestedField(inference.Workload, "spec", "pvcs", "items", "pvc", "value", "claimName", "value")
			Expect(field).To(Equal(*pvcAsset.Spec.ClaimName))
		})
	})

	Context("Handle project deletion", func() {
		It("basic flow", func() {
			nonTargetId := createWorkload(randoms, router, envAssetInference, auth.WorkloadTypeInference)
			targetId := createWorkload(differentProject, router, envAssetInference, auth.WorkloadTypeInference)

			// verify that workloads were created
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, nonTargetId, http.StatusOK)
			expectWorkloadStatus(differentProject, auth.WorkloadTypeInference, router, targetId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, differentProject.TenantIdStr)
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId: differentProject.ProjectIdStr,
			})
			Expect(err).To(BeNil())

			// verify that only the target was deleted
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, nonTargetId, http.StatusOK)
			expectWorkloadStatus(differentProject, auth.WorkloadTypeInference, router, targetId, http.StatusNotFound)
		})
	})

	Context("Handle cluster deletion", func() {
		It("basic flow", func() {
			workloadId := createWorkload(randoms, router, envAssetInference, auth.WorkloadTypeInference)

			// verify that workloads were created
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, workloadId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: randoms.ClusterId,
			})
			Expect(err).To(BeNil())

			// verify that thw workload was deleted
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, workloadId, http.StatusNotFound)
		})
	})

	Context("Handle tenant deletion", func() {
		It("basic flow", func() {
			workloadId := createWorkload(randoms, router, envAssetInference, auth.WorkloadTypeInference)

			// verify that workloads were created
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, workloadId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			})
			Expect(err).To(BeNil())

			// verify that the workload was deleted
			expectWorkloadStatus(randoms, auth.WorkloadTypeInference, router, workloadId, http.StatusNotFound)
		})
	})

	Context("s3 asset in inference workload", func() {
		It("should fail validation", func() {
			accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_S3,
					Id:   s3Asset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[auth.WorkloadTypeInference], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("inference workload cannot use data-source asset of type s3"))
		})
	})

	Context("specific env connection not found", func() {
		DescribeTable("must have names that appear in env asset", func(workloadType string) {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAssetInference, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, getEnvAssetIdByType(workloadType, &envAsset, &envAssetInference))

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name: "not_to_be_found",
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("connection named 'not_to_be_found' is not defined in the environment"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("specific env connections", func() {
		DescribeTable("private connection - tensorflow (ok)", func(workloadType string) {
			envAssetPayload := buildRandomTenantEnvironmentAsset(randoms)
			envAssetPayload.Spec.Connections = []api.Connection{
				{
					Name:       "notebook-a",
					IsExternal: false,
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_TENSORBOARD,
						ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
						ContainerPort:  8080,
						ExternalUrlInfo: &api.ExternalUrlInfo{
							IsCustomUrl: true,
							ExternalUrl: convutils.Ref("http://url.com/"),
						},
					},
				},
			}

			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name:        "notebook-a",
						ExternalUrl: convutils.Ref("http://url.com/"),
						AuthorizedUsers: &[]string{
							"test@run.ai",
							"another-user@run.ai",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("specific env connections", func() {
		DescribeTable("private connection - jupyter (ok)", func(workloadType string) {
			envAssetPayload := buildRandomTenantEnvironmentAsset(randoms)
			envAssetPayload.Spec.Connections = []api.Connection{
				{
					Name:       "notebook-a",
					IsExternal: false,
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
						ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
						ContainerPort:  8080,
						ExternalUrlInfo: &api.ExternalUrlInfo{
							IsCustomUrl: true,
							ExternalUrl: convutils.Ref("http://url.com/"),
						},
					},
				},
			}

			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name:        "notebook-a",
						ExternalUrl: convutils.Ref("http://url.com/"),
						AuthorizedUsers: &[]string{
							"test@run.ai",
							"another-user@run.ai",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("specific env connections", func() {
		DescribeTable("private connection - jupyter - non url", func(workloadType string) {
			envAssetPayload := buildRandomTenantEnvironmentAsset(randoms)
			envAssetPayload.Spec.Connections = []api.Connection{
				{
					Name:       "notebook-a",
					IsExternal: false,
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
						ConnectionType: api.INTERNALCONNECTIONTYPE_NODE_PORT,
						ContainerPort:  8080,
						NodePortInfo: &api.NodePortInfo{
							IsCustomPort: false,
						},
					},
				},
			}

			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name: "notebook-a",
						AuthorizedUsers: &[]string{
							"test@run.ai",
							"another-user@run.ai",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("specific env autoscaling", func() {
		DescribeTable("must be nil for non-inference workloads", func(workloadType string) {
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(1)),
					MaxReplicas: convutils.Ref(int32(1)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("it is not permitted to specify autoScaling values for workload type other than Inference"))
		}, forWorkspaces, forTrainings)
	})

	Context("specific env autoscaling", func() {
		DescribeTable("must be not nil value for inference workloads that was created with value", func(workloadType string) {
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(1)),
					MaxReplicas: convutils.Ref(int32(1)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var inference api.Inference
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			inferenceId := inference.Meta.Id

			///////////////////////////////////////////////////////////////////////////////////////
			// get the inference content
			///////////////////////////////////////////////////////////////////////////////////////
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlForCreatingWorkload[workloadType], inferenceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualInference api.Inference
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualInference)
			Expect(jsonErr).To(BeNil())
			Expect(actualInference.Spec.SpecificEnv.AutoScaling).To(Equal(reqPayload.SpecificEnv.AutoScaling))
		}, forInference)
	})

	Context("valid create request", func() {
		DescribeTable("should succeed with status ok", func(workloadType string) {
			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 1: create the workspace
			//          assets are taken from all valid scopes: system, tenant, dept, project
			///////////////////////////////////////////////////////////////////////////////////////
			environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			pvc1CreReq := testing.BuildRandomProjectPVCAsset(randoms)
			pvc2CreReq := testing.BuildRandomProjectPVCAsset(randoms)
			pvc2CreReq.Spec.ExistingPvc = convutils.Ref(true)
			pvc2CreReq.Spec.ClaimInfo = nil
			passwordCreReq := buildRandomProjectPassword(randoms)
			accessKeyCreReq := buildRandomProjectAccessKey(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithPassword(randoms.ClusterId, &passwordCreReq, nil).
				WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil).
				WithPvc(randoms.ClusterId, &pvc2CreReq, nil))

			pvcAsset, _, _ := createPVCAsset(pvc1CreReq, router, headerForAdmin(randoms))
			pvcAssetInference, _, _ := createPVCAsset(pvc2CreReq, router, headerForAdmin(randoms))
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithPassword(randoms.ClusterId, &passwordCreReq, nil).
				WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil).
				WithPvc(randoms.ClusterId, &pvc1CreReq, nil).
				WithPvc(randoms.ClusterId, &pvc2CreReq, nil))

			passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

			accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, accessKeyCreReq, router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

			reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id

			pvcAssetMeta := pvcAsset.Meta
			if workloadType == auth.WorkloadTypeInference {
				pvcAssetMeta = pvcAssetInference.Meta
			}
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   gitAsset.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_PVC,
					Id:   pvcAssetMeta.Id,
				},
			}

			// s3 is not supported in inference
			if workloadType != auth.WorkloadTypeInference {
				reqPayload.Assets.Datasources = append(reqPayload.Assets.Datasources, api.AssetIdAndKind{
					Kind: api.ASSETKIND_S3,
					Id:   s3Asset.Meta.Id,
				})
			}

			if workloadType == auth.WorkloadTypeDistributed {
				reqPayload.Distributed = &api.Distributed{
					NumWorkers:    1,
					DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Master: &api.DistMaster{
						Assets: api.AssetsIds{
							Environment: environmentAsset.Meta.Id,
							Datasources: []api.AssetIdAndKind{
								{
									Kind: api.ASSETKIND_S3,
									Id:   s3Asset.Meta.Id,
								},
							},
						},
					},
				}
			}
			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			workspaceId := workspace.Meta.Id

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 3: get the workspace content -> V2 APIs
			///////////////////////////////////////////////////////////////////////////////////////
			urlGetV1 := urlForCreatingWorkload[workloadType]
			urlGetV2 := urlV1toV2[urlGetV1]

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGetV2, workspaceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualWorkloadV2 api.TrainingV2
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualWorkloadV2)
			Expect(jsonErr).To(BeNil())

			expectedWorkloadV2 := api.TrainingV2{
				Meta: api.WorkloadMeta{
					Name:          reqPayload.Name,
					RequestedName: reqPayload.Name,
					WorkloadId:    workspaceId,
					ClusterId:     randoms.ClusterId,
					ProjectId:     fmt.Sprintf("%d", randoms.ProjectId),
					CreatedBy:     actualWorkloadV2.Meta.CreatedBy,
					CreatedAt:     actualWorkloadV2.Meta.CreatedAt,
				},
				Assets: api.AssetsRef{
					Environment: api.EnvironmentAssetRef{
						Id:   environmentAsset.Meta.Id,
						Name: environmentAsset.Meta.Name,
					},
					Compute: &api.AssetsRefCompute{
						Id:   computeAsset.Meta.Id,
						Name: computeAsset.Meta.Name,
					},
					Datasources: []api.DatasourceRef{
						metaToDatasourceRef(gitAsset.Meta),
						metaToDatasourceRef(pvcAssetMeta),
					},
				},
			}

			var convErr *api.Error
			expectedWorkloadV2.SpecificRunParams, convErr = genericToTrainingRunParams(reqPayload.SpecificEnv)
			Expect(convErr).To(BeNil())

			// s3 is not supported in inference
			if workloadType != auth.WorkloadTypeInference {
				expectedWorkloadV2.Assets.Datasources = append(expectedWorkloadV2.Assets.Datasources, metaToDatasourceRef(s3Asset.Meta))
			}

			if workloadType == auth.WorkloadTypeDistributed {
				expectedWorkloadV2.Distributed = &api.DistributedData{
					NumWorkers: 1,
					Framework:  convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Master: &api.DistributedMasterData{
						Assets: &api.AssetsRefOptional{
							Environment: api.EnvironmentAssetRef{
								Id:   environmentAsset.Meta.Id,
								Name: environmentAsset.Meta.Name,
							},
							Datasources: []api.DatasourceRef{
								metaToDatasourceRef(s3Asset.Meta),
							},
						},
					},
				}
			}
			sort.Slice(expectedWorkloadV2.Assets.Datasources, func(i, j int) bool {
				return expectedWorkloadV2.Assets.Datasources[i].Id < expectedWorkloadV2.Assets.Datasources[j].Id
			})
			Expect(actualWorkloadV2).To(Equal(expectedWorkloadV2))
		}, forWorkspaces, forTrainings, forInference, forDistributed)
	})

	Context("Distributed create request", func() {
		DescribeTable("valid request", func(workloadType string) {
			if workloadType != auth.WorkloadTypeDistributed {
				return
			}

			environmentAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			passwordCreReq := buildRandomProjectPassword(randoms)
			accessKeyCreReq := buildRandomProjectAccessKey(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithPassword(randoms.ClusterId, &passwordCreReq, nil).
				WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil))

			passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

			accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, accessKeyCreReq, router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   gitAsset.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_S3,
					Id:   s3Asset.Meta.Id,
				},
			}
			reqPayload.Distributed = &api.Distributed{
				NumWorkers:    2,
				DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
				Master: &api.DistMaster{
					Assets: api.AssetsIds{
						Environment: environmentAsset.Meta.Id,
						Datasources: []api.AssetIdAndKind{
							{
								Kind: api.ASSETKIND_S3,
								Id:   s3Asset.Meta.Id,
							},
						},
						Compute: &computeAsset.Meta.Id,
					},
				},
			}
			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForNonAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			workspaceId := workspace.Meta.Id

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 2: get the workspace content (V2)
			///////////////////////////////////////////////////////////////////////////////////////
			urlGetV1 := urlForCreatingWorkload[workloadType]
			urlGetV2 := urlV1toV2[urlGetV1]

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGetV2, workspaceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForNonAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualTraining api.TrainingV2
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualTraining)
			Expect(jsonErr).To(BeNil())

			currentTime := time.Now()
			actualTraining.Meta.CreatedAt = currentTime

			expectedTraining := api.TrainingV2{
				Meta: api.WorkloadMeta{
					WorkloadId:    workspaceId,
					ClusterId:     randoms.ClusterId,
					ProjectId:     strconv.Itoa(int(randoms.ProjectId)),
					Name:          reqPayload.Name,
					RequestedName: reqPayload.Name,
					CreatedBy:     actualTraining.Meta.CreatedBy,
					CreatedAt:     currentTime,
				},
				SpecificRunParams: &api.TrainingSpecificRunParams{
					Command: convutils.Ref("echo"),
					Args:    convutils.Ref("123"),
				},
				Assets: api.AssetsRef{
					Environment: api.EnvironmentAssetRef{
						Id:   environmentAsset.Meta.Id,
						Name: environmentAsset.Meta.Name,
					},
					Compute: &api.AssetsRefCompute{
						Id:   computeAsset.Meta.Id,
						Name: computeAsset.Meta.Name,
					},
					Datasources: []api.DatasourceRef{
						metaToDatasourceRef(gitAsset.Meta),
						metaToDatasourceRef(s3Asset.Meta),
					},
				},
				Distributed: &api.DistributedData{
					NumWorkers: 2,
					Framework:  convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Master: &api.DistributedMasterData{
						Assets: &api.AssetsRefOptional{
							Environment: api.EnvironmentAssetRef{
								Id:   environmentAsset.Meta.Id,
								Name: environmentAsset.Meta.Name,
							},
							Datasources: []api.DatasourceRef{
								metaToDatasourceRef(s3Asset.Meta),
							},
							Compute: &api.AssetsRefCompute{
								Id:   computeAsset.Meta.Id,
								Name: computeAsset.Meta.Name,
							},
						},
					},
				},
			}

			sort.Slice(expectedTraining.Assets.Datasources, func(i, j int) bool {
				return expectedTraining.Assets.Datasources[i].Id < expectedTraining.Assets.Datasources[j].Id
			})

			testutils.Expect(actualTraining).ToEqual(expectedTraining)
		}, forDistributed)

		DescribeTable("valid request with master==nil when NoMaster=True", func(workloadType string) {
			if workloadType != auth.WorkloadTypeDistributed {
				return
			}

			environmentAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			passwordCreReq := buildRandomProjectPassword(randoms)
			accessKeyCreReq := buildRandomProjectAccessKey(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithPassword(randoms.ClusterId, &passwordCreReq, nil).
				WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil))

			passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

			accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, accessKeyCreReq, router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   gitAsset.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_S3,
					Id:   s3Asset.Meta.Id,
				},
			}
			reqPayload.Distributed = &api.Distributed{
				NumWorkers:    1,
				DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				NoMaster:      convutils.Ref(true),
			}
			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForNonAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			workspaceId := workspace.Meta.Id

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 2: get the workspace content (v2)
			///////////////////////////////////////////////////////////////////////////////////////
			urlGetV1 := urlForCreatingWorkload[workloadType]
			urlGetV2 := urlV1toV2[urlGetV1]

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGetV2, workspaceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForNonAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualWorkload api.TrainingV2
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualWorkload)
			Expect(jsonErr).To(BeNil())

			currentTime := time.Now()
			actualWorkload.Meta.CreatedAt = currentTime

			expectedWorkload := api.TrainingV2{
				Meta: api.WorkloadMeta{
					WorkloadId:    workspaceId,
					ClusterId:     randoms.ClusterId,
					ProjectId:     strconv.Itoa(int(randoms.ProjectId)),
					Name:          reqPayload.Name,
					RequestedName: reqPayload.Name,
					CreatedBy:     actualWorkload.Meta.CreatedBy,
					CreatedAt:     currentTime,
				},
				SpecificRunParams: &api.TrainingSpecificRunParams{
					Command: actualWorkload.SpecificRunParams.Command,
					Args:    actualWorkload.SpecificRunParams.Args,
				},
				Assets: api.AssetsRef{
					Environment: api.EnvironmentAssetRef{
						Id:   environmentAsset.Meta.Id,
						Name: environmentAsset.Meta.Name,
					},
					Compute: &api.AssetsRefCompute{
						Id:   computeAsset.Meta.Id,
						Name: computeAsset.Meta.Name,
					},
					Datasources: []api.DatasourceRef{
						metaToDatasourceRef(gitAsset.Meta),
						metaToDatasourceRef(s3Asset.Meta),
					},
				},
				Distributed: &api.DistributedData{
					NumWorkers: 1,
					Framework:  convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				},
			}

			sort.Slice(expectedWorkload.Assets.Datasources, func(i, j int) bool {
				return expectedWorkload.Assets.Datasources[i].Id < expectedWorkload.Assets.Datasources[j].Id
			})
			testutils.Expect(actualWorkload).ToEqual(expectedWorkload)
		}, forDistributed)

		DescribeTable("should fail with master!=nil when NoMaster=True", func(workloadType string) {
			if workloadType != auth.WorkloadTypeDistributed {
				return
			}

			environmentAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id
			reqPayload.Distributed = &api.Distributed{
				NumWorkers:    2,
				NoMaster:      convutils.Ref(true),
				DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
				Master: &api.DistMaster{
					Assets: api.AssetsIds{
						Environment: environmentAsset.Meta.Id,
						Compute:     &computeAsset.Meta.Id,
					},
				},
			}
			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForNonAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("NoMaster is not available for the chosen framework"))
		}, forDistributed)

		DescribeTable("valid request when Master==nil will use worker spec", func(workloadType string) {
			if workloadType != auth.WorkloadTypeDistributed {
				return
			}

			environmentAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			passwordCreReq := buildRandomProjectPassword(randoms)
			accessKeyCreReq := buildRandomProjectAccessKey(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil).
				WithPassword(randoms.ClusterId, &passwordCreReq, nil))

			passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

			accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, accessKeyCreReq, router, headerForAdmin(randoms))
			s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   gitAsset.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_S3,
					Id:   s3Asset.Meta.Id,
				},
			}
			reqPayload.Distributed = &api.Distributed{
				NumWorkers:    2,
				DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
				Master: &api.DistMaster{
					Assets: api.AssetsIds{
						Environment: environmentAsset.Meta.Id,
						Datasources: []api.AssetIdAndKind{
							{
								Kind: api.ASSETKIND_S3,
								Id:   s3Asset.Meta.Id,
							},
						},
						Compute: &computeAsset.Meta.Id,
					},
				},
			}
			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForNonAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			workspaceId := workspace.Meta.Id

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 2: get the workspace content (V2)
			///////////////////////////////////////////////////////////////////////////////////////
			urlGetV1 := urlForCreatingWorkload[workloadType]
			urlGetV2 := urlV1toV2[urlGetV1]

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGetV2, workspaceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForNonAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualTraining api.TrainingV2
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualTraining)
			Expect(jsonErr).To(BeNil())

			currentTime := time.Now()
			actualTraining.Meta.CreatedAt = currentTime

			expectedWorkload := api.TrainingV2{
				Meta: api.WorkloadMeta{
					WorkloadId:    workspaceId,
					ClusterId:     randoms.ClusterId,
					ProjectId:     strconv.Itoa(int(randoms.ProjectId)),
					Name:          reqPayload.Name,
					RequestedName: reqPayload.Name,
					CreatedBy:     actualTraining.Meta.CreatedBy,
					CreatedAt:     currentTime,
				},
				SpecificRunParams: &api.TrainingSpecificRunParams{
					Command: reqPayload.SpecificEnv.Command,
					Args:    reqPayload.SpecificEnv.Args,
				},
				Assets: api.AssetsRef{
					Environment: api.EnvironmentAssetRef{
						Id:   environmentAsset.Meta.Id,
						Name: environmentAsset.Meta.Name,
					},
					Compute: &api.AssetsRefCompute{
						Id:   computeAsset.Meta.Id,
						Name: computeAsset.Meta.Name,
					},
					Datasources: []api.DatasourceRef{
						metaToDatasourceRef(gitAsset.Meta),
						metaToDatasourceRef(s3Asset.Meta),
					},
				},
				Distributed: &api.DistributedData{
					NumWorkers: 2,
					Framework:  convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
					Master: &api.DistributedMasterData{
						Assets: &api.AssetsRefOptional{
							Environment: api.EnvironmentAssetRef{
								Id:   environmentAsset.Meta.Id,
								Name: environmentAsset.Meta.Name,
							},
							Datasources: []api.DatasourceRef{
								metaToDatasourceRef(s3Asset.Meta),
							},
							Compute: &api.AssetsRefCompute{
								Id:   computeAsset.Meta.Id,
								Name: computeAsset.Meta.Name,
							},
						},
					},
				},
			}

			sort.Slice(expectedWorkload.Assets.Datasources, func(i, j int) bool {
				return expectedWorkload.Assets.Datasources[i].Id < expectedWorkload.Assets.Datasources[j].Id
			})

			testutils.Expect(actualTraining).ToEqual(expectedWorkload)
		}, forDistributed)
	})

	Context("get non existent workload", func() {
		DescribeTable("should fail with not found error", func(workloadType string) {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlForCreatingWorkload[workloadType], notExist), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("create workload in project the researcher has no access", func() {
		DescribeTable("should fail with not found error", func(workloadType string) {
			environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlForCreatingWorkload[workloadType], notExist), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("multiple environment variables in specific env - assets", func() {
		DescribeTable("should fail validation", func(workloadType string) {
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
					{Name: convutils.Ref("USER"), Value: convutils.Ref("root")},
					{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("duplicate variable name 'HOME'"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("uidGidSource VS runAsUid and runAsGid", func() {
		DescribeTable("uigGid=from image -> uid should not be provided", func(workloadType string) {
			userFromContainer := buildRandomTenantEnvironmentAsset(randoms)
			userFromContainer.Spec.UidGidSource = convutils.Ref(flat_workloads.UidGidFromTheImage)

			envAssetUserFromContainer, _, _ := createEnvironmentAsset(userFromContainer, router, headerForAdmin(randoms))

			// create workload

			reqCreationPayload := randomWorkloadCreationRequest(randoms, envAssetUserFromContainer.Meta.Id)
			reqCreationPayload.SpecificEnv = &api.SpecificRunParams{
				RunAsUid: convutils.Ref(int64(0)),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqCreationPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("it is not permitted to override the runAsUid from the environment asset"))
		}, forWorkspaces, forTrainings, forInference)

		DescribeTable("uidGidSource = image -> gid should not be provided", func(workloadType string) {
			userFromContainer := buildRandomTenantEnvironmentAsset(randoms)
			userFromContainer.Spec.UidGidSource = convutils.Ref(flat_workloads.UidGidFromTheImage)

			envAssetUserFromContainer, _, _ := createEnvironmentAsset(userFromContainer, router, headerForAdmin(randoms))

			reqCreationPayload := randomWorkloadCreationRequest(randoms, envAssetUserFromContainer.Meta.Id)
			reqCreationPayload.SpecificEnv = &api.SpecificRunParams{
				RunAsGid: convutils.Ref(int64(3)),
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqCreationPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("it is not permitted to override the runAsGid from the environment asset"))
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("created workload -> response", func() {
		DescribeTable("workload-composer should ignore serving port for non-inference workloads", func(workloadType string) {
			envAssetb := buildRandomTenantEnvironmentAsset(randoms)
			envAssetb.Spec.Connections = []api.Connection{
				{
					Name: "serving-port",
					InternalToolInfo: &api.InternalToolInfo{
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ContainerPort:  8080,
						ServingPortInfo: &api.ServingPortInfo{
							Protocol: convutils.Ref("http"),
						},
					},
				},
			}
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, envAssetb, router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
			///
			Expect(createHttpResp.Body.String()).NotTo(ContainSubstring("ServingPort"))
		}, forWorkspaces, forTrainings)
	})

	Context("created workload -> response", func() {
		DescribeTable("workload should contain the workspace/training and cluster id annotation", func(workloadType string) {
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workload dbmodel.GenericWorkload
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workload)
			Expect(jsonErr).To(BeNil())
			Expect(workload.Meta.Id).NotTo(Equal(""))
			Expect(workload.Workload).NotTo(BeNil())
			Expect(workload.Workload["metadata"]).NotTo(BeNil())

			payload := make(map[string]interface{})
			jsonErr = json.Unmarshal(createHttpResp.Body.Bytes(), &payload)
			Expect(jsonErr).To(BeNil())

			workloadCrd := payload["workload"].(map[string]interface{})
			Expect(workloadCrd).NotTo(BeNil())
			metadata := workloadCrd["metadata"].(map[string]interface{})
			Expect(metadata).NotTo(BeNil())

			annotations := metadata["annotations"].(map[string]interface{})
			Expect(annotations["clusterId"]).To(Equal(workload.Meta.ClusterId))
			Expect(annotations["workloadId"]).To(Equal(workload.Meta.Id))
			if workloadType == auth.WorkloadTypeWorkspace {
				Expect(annotations["workspaceId"]).To(Equal(workload.Meta.Id))
			}
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("created workload then delete by id", func() {
		DescribeTable("should delete workload by id successfully", func(workloadType string) {
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			var workload dbmodel.GenericWorkload
			jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workload)
			Expect(jsonErr).To(BeNil())

			payload := make(map[string]interface{})
			jsonErr = json.Unmarshal(createHttpResp.Body.Bytes(), &payload)
			Expect(jsonErr).To(BeNil())

			workloadCr := payload["workload"].(map[string]interface{})
			Expect(workloadCr).NotTo(BeNil())

			metadata := workloadCr["metadata"].(map[string]interface{})
			Expect(metadata).NotTo(BeNil())

			workloadId := metadata["annotations"].(map[string]interface{})["workloadId"]
			Expect(workloadId).To(Equal(workload.Meta.Id))

			// delete by id
			reqPayload = randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlForCreatingWorkload[workloadType], workloadId), makePayload(reqPayload))
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))
		}, forInference)

		DescribeTable("should delete fail when try to delete non-existing id", func(workloadType string) {
			envAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlForCreatingWorkload[workloadType], notExist), makePayload(reqPayload))
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring("cannot find Inference"))
		}, forInference)
	})

	Context("creating workload with assets -> updates assets UsageTimes", func() {
		DescribeTable("by workload type", func(workloadType string) {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets

			environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomProjectEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomProjectCompute(randoms), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)

			environmentsHttpReq, errEnvironmentList := http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlEnvironment, *environmentAsset.Meta.ProjectId, *environmentAsset.Meta.ProjectId, workloadType, *environmentAsset.Meta.ClusterId), nil)
			Expect(errEnvironmentList).To(BeNil())

			environmentList, envJsonErr := getEnvironmentsEntriesFromList(router, randoms, environmentsHttpReq)
			Expect(envJsonErr).To(BeNil())

			// There are 2 or 3 environments because one is being created in the BeforeEach and one only for inference, so here we are finding the relevant one
			Expect(len(environmentList)).To(BeNumerically(">=", 2))
			for _, env := range environmentList {
				if env.Meta.Id == environmentAsset.Meta.Id {
					Expect(env.Meta.Name).To(Equal(environmentAsset.Meta.Name))
					Expect(env.UsageTimes).To(BeNil())
				}
			}

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, *computeAsset.Meta.ProjectId, *computeAsset.Meta.ProjectId, workloadType, *computeAsset.Meta.ClusterId), nil)
			Expect(errComputeList).To(BeNil())

			computeList, computeJsonErr := getComputesEntriesFromList(router, randoms, computesHttpReq)

			Expect(computeJsonErr).To(BeNil())
			Expect(len(computeList)).To(Equal(1))
			Expect(computeList[0].Meta.Name).To(Equal(computeList[0].Meta.Name))
			Expect(computeList[0].UsageTimes).To(BeNil())

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 3: Creating workload with the recently created assets

			workloadReqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			workloadReqPayload.Assets.Environment = environmentAsset.Meta.Id
			workloadReqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			workloadReqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			workloadPreCreationTimestamp := time.Now().UnixNano()

			workloadCreateHttpResp := getWorkloadCreationResponse(router, randoms, workloadType, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			Expect(workspace.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(workspace.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			///////////////////////////////////////////////////////////////////////////////////////
			//STEP 4: checking the list for the existing usage info

			// checking computes list response
			computes, jsonComputesErr := getComputesEntriesFromList(router, randoms, computesHttpReq)
			Expect(jsonComputesErr).To(BeNil())

			Expect(len(computes)).To(Equal(1))
			Expect(computes[0].Meta.Name).To(Equal(computes[0].Meta.Name))
			Expect(computes[0].UsageTimes).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())

			// checking environments list response
			environments, jsonEnvsErr := getEnvironmentsEntriesFromList(router, randoms, environmentsHttpReq)
			Expect(jsonEnvsErr).To(BeNil())

			// There are 2 or 3 environments because one is being created in the BeforeEach and one only for inference, so here we are finding the relevant one
			Expect(len(environmentList)).To(BeNumerically(">=", 2))
			var environmentUsageTimePostFirstWokloadUsage *time.Time
			for _, env := range environments {
				if env.Meta.Id == environmentAsset.Meta.Id {
					Expect(env.Meta.Name).To(Equal(environmentAsset.Meta.Name))
					Expect(env.UsageTimes).NotTo(BeNil())
					Expect(env.UsageTimes.LastUsedByWorkload).NotTo(BeNil())
					Expect(env.UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())
					// saving the usage time of the environment for the check in the next step
					environmentUsageTimePostFirstWokloadUsage = env.UsageTimes.LastUsedByWorkload
				}
			}

			///////////////////////////////////////////////////////////////////////////////////////
			// Step 5: using the environment in a different workload creation and checking that its UsageTimes has been updated

			secondWorkloadReqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			secondWorkloadReqPayload.Assets.Environment = environmentAsset.Meta.Id
			secondWorkloadReqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("test"),
				Args:    convutils.Ref("x=4"),
			}

			secondWorkloadCreateHttpResp := getWorkloadCreationResponse(router, randoms, workloadType, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK))

			var secondWorkload api.Workspace
			secondWorkloadJsonErr := json.Unmarshal(secondWorkloadCreateHttpResp.Body.Bytes(), &secondWorkload)
			Expect(secondWorkloadJsonErr).To(BeNil())
			Expect(secondWorkload.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			secondEnvironments, secondJsonEnvsErr := getEnvironmentsEntriesFromList(router, randoms, environmentsHttpReq)
			Expect(secondJsonEnvsErr).To(BeNil())

			// There are 3 or 4 environments because one is being created in the BeforeEach and one only for inference, so here we are finding the relevant one
			Expect(len(environmentList)).To(BeNumerically(">=", 3))
			for _, env := range secondEnvironments {
				if env.Meta.Id == environmentAsset.Meta.Id {
					Expect(env.Meta.Name).To(Equal(environmentAsset.Meta.Name))
					Expect(env.UsageTimes).NotTo(BeNil())
					Expect(env.UsageTimes.LastUsedByWorkload).NotTo(BeNil())
					// checking that the new value is greater than the usage value of the previous workload
					Expect(env.UsageTimes.LastUsedByWorkload.UnixNano() > environmentUsageTimePostFirstWokloadUsage.UnixNano()).To(BeTrue())
				}
			}
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("creating workload with cluster scope assets -> updates assets UsageTimes", func() {
		DescribeTable("by workload type", func(workloadType string) {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets - system scope

			environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomClusterCompute(randoms), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)
			// (checking only compute asset, as previous tests already checked the environment asset, here we're
			//  focused on system scope difference)

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?complyToWorkloadType=%s", urlCompute, workloadType), nil)
			Expect(errComputeList).To(BeNil())

			computeList, computeJsonErr := getComputesEntriesFromList(router, randoms, computesHttpReq)

			Expect(computeJsonErr).To(BeNil())
			Expect(len(computeList)).To(Equal(1))
			Expect(computeList[0].Meta.Name).To(Equal(computeList[0].Meta.Name))
			Expect(computeList[0].UsageTimes).To(BeNil())

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 3: Creating workload with the recently created assets

			workloadReqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			workloadReqPayload.Assets.Environment = environmentAsset.Meta.Id
			workloadReqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			workloadReqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			workloadPreCreationTimestamp := time.Now().UnixNano()

			workloadCreateHttpResp := getWorkloadCreationResponse(router, randoms, workloadType, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			Expect(workspace.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(workspace.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			///////////////////////////////////////////////////////////////////////////////////////
			//STEP 4: checking the list for the existing usage info

			// checking computes list response
			computesHttpReq, errComputeList = http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, workspace.Meta.ProjectId, workspace.Meta.ProjectId, workloadType, workspace.Meta.ClusterId), nil)
			computes, jsonComputesErr := getComputesEntriesFromList(router, randoms, computesHttpReq)
			Expect(jsonComputesErr).To(BeNil())

			Expect(len(computes)).To(Equal(1))
			Expect(computes[0].Meta.Name).To(Equal(computes[0].Meta.Name))
			Expect(computes[0].UsageTimes).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("creating workload with system scope assets -> updates assets UsageTimes", func() {
		DescribeTable("by workload type", func(workloadType string) {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets - system scope

			environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomSystemCompute(), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)
			// (checking only compute asset, as previous tests already checked the environment asset, here we're
			//  focused on system scope difference)

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?complyToWorkloadType=%s", urlCompute, workloadType), nil)
			Expect(errComputeList).To(BeNil())

			computeList, computeJsonErr := getComputesEntriesFromList(router, randoms, computesHttpReq)

			Expect(computeJsonErr).To(BeNil())
			Expect(len(computeList)).To(Equal(1))
			Expect(computeList[0].Meta.Name).To(Equal(computeList[0].Meta.Name))
			Expect(computeList[0].UsageTimes).To(BeNil())

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 3: Creating workload with the recently created assets

			workloadReqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			workloadReqPayload.Assets.Environment = environmentAsset.Meta.Id
			workloadReqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			workloadReqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			workloadPreCreationTimestamp := time.Now().UnixNano()

			workloadCreateHttpResp := getWorkloadCreationResponse(router, randoms, workloadType, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var workspace api.Workspace
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &workspace)
			Expect(jsonErr).To(BeNil())
			Expect(workspace.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(workspace.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			///////////////////////////////////////////////////////////////////////////////////////
			//STEP 4: checking the list for the existing usage info

			// checking computes list response
			computesHttpReq, errComputeList = http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, workspace.Meta.ProjectId, workspace.Meta.ProjectId, workloadType, workspace.Meta.ClusterId), nil)
			computes, jsonComputesErr := getComputesEntriesFromList(router, randoms, computesHttpReq)
			Expect(jsonComputesErr).To(BeNil())

			Expect(len(computes)).To(Equal(1))
			Expect(computes[0].Meta.Name).To(Equal(computes[0].Meta.Name))
			Expect(computes[0].UsageTimes).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())
		}, forWorkspaces, forTrainings, forInference)
	})

	Context("name validation - mandatory parameters", func() {
		It("should give bad request error if they are missing", func() {
			Expect(assetNameValidation(router, randoms, "assetName=any&scopeType=tenant")).To(Equal(http.StatusBadRequest), "missing kind")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant")).To(Equal(http.StatusBadRequest), "missing name")
			Expect(assetNameValidation(router, randoms, "kind=environment&name=abc")).To(Equal(http.StatusBadRequest), "missing scope")
		})
	})

	Context("name validation - scope id is mandatory for all scope but system and tenant", func() {
		It("should give bad request error", func() {
			Expect(assetNameValidation(router, randoms, "kind=environment&assetName=any&scopeType=system")).To(
				Equal(http.StatusNoContent), "system + no scope id - ok")
			Expect(assetNameValidation(router, randoms, "kind=environment&assetName=any&scopeType=tenant")).To(
				Equal(http.StatusNoContent), "tenant + no scope id - ok")
			Expect(assetNameValidation(router, randoms, fmt.Sprintf("kind=environment&assetName=any&scopeType=cluster&scopeId=%s", uuid.NewString()))).To(
				Equal(http.StatusNoContent), "cluster + no scope id - ok")
			Expect(assetNameValidation(router, randoms, "kind=compute&assetName=tenant-asset&scopeType=department")).To(
				Equal(http.StatusBadRequest), "department + no scope id")
			Expect(assetNameValidation(router, randoms, "kind=compute&assetName=tenant-asset&scopeType=project")).To(
				Equal(http.StatusBadRequest), "project + no scope id")
		})
	})

	Context("name validation - system", func() {
		It("can be reused by another scope", func() {
			envAsset := buildRandomSystemEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
		})
	})

	Context("name validation - tenant", func() {
		It("single occurrence entire system", func() {
			envAsset := buildRandomTenantEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, fmt.Sprintf("kind=environment&scopeType=cluster&scopeId=%s"+sameName, uuid.NewString()))).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusConflict), "sameName")
		})
	})

	Context("name validation - cluster", func() {
		It("single occurrence entire system", func() {
			envAsset := buildRandomClusterEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScope := "&scopeId=" + fmt.Sprintf("%s", convutils.Deref(envAsset.Meta.ClusterId, ""))
			diffScope := "&scopeId=" + fmt.Sprintf("%s", uuid.NewString())

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+diffName)).To(
				Equal(http.StatusNoContent), "diffName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameScope+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameScope+diffName)).To(
				Equal(http.StatusNoContent), "diffName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+diffScope+sameName)).To(
				Equal(http.StatusNoContent), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department&scopeId="+randoms.DepartmentId+sameName)).To(
				Equal(http.StatusConflict), "sameName")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project&scopeId="+randoms.ProjectIdStr+sameName)).To(
				Equal(http.StatusConflict), "sameName")
		})
	})

	Context("name validation - department", func() {
		It("only different dept is accepted", func() {
			envAsset := buildRandomDepartmentEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScopeId := fmt.Sprintf("&scopeId=%s", randoms.DepartmentId)
			diffScopeId := fmt.Sprintf("&scopeId=%s", differentDepartment.DepartmentId)

			sameDepProjectId := fmt.Sprintf("&scopeId=%v", randoms.ProjectId)
			diffDepProjectId := fmt.Sprintf("&scopeId=%v", differentDepartment.ProjectId)

			clusterScopeId := "&scopeId=" + randoms.ClusterId

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+diffName+sameScopeId)).To(
				Equal(http.StatusNoContent), "diffName + sameScopeId")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "sameName + diffScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "system + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "tenant + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+clusterScopeId+sameName)).To(
				Equal(http.StatusConflict), "cluster + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=department"+sameName+sameScopeId)).To(
				Equal(http.StatusConflict), "department + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+sameDepProjectId)).To(
				Equal(http.StatusConflict), "project + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffDepProjectId)).To(
				Equal(http.StatusNoContent), "project + sameName + diffScopeId")
		})
	})

	Context("name validation - project", func() {
		It("only different project is accepted", func() {
			envAsset := buildRandomProjectEnvironmentAsset(randoms)
			envAsset.Meta.Name = "xyz"
			createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

			sameName := "&assetName=xyz"
			diffName := "&assetName=abc"

			sameScopeId := fmt.Sprintf("&scopeId=%s", randoms.ProjectIdStr)
			diffScopeId := fmt.Sprintf("&scopeId=%s", differentDepartment.ProjectIdStr)

			clusterScopeId := fmt.Sprintf("&scopeId=%s", randoms.ClusterId)

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+diffName+sameScopeId)).To(
				Equal(http.StatusNoContent), "diffName + sameScopeId")
			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "sameName + diffScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=system"+sameName)).To(
				Equal(http.StatusNoContent), "system + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=tenant"+sameName)).To(
				Equal(http.StatusConflict), "tenant + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=cluster"+sameName+clusterScopeId)).To(
				Equal(http.StatusConflict), "cluster + sameName")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+sameScopeId)).To(
				Equal(http.StatusConflict), "project + sameName + sameScopeId")

			Expect(assetNameValidation(router, randoms, "kind=environment&scopeType=project"+sameName+diffScopeId)).To(
				Equal(http.StatusNoContent), "project + sameName + diffScopeId")
		})
	})

	Context("assets override", func() {
		Context("valid create workload request", func() {
			DescribeTable("should override the container path value of specified data source", func(workloadType string) {
				///////////////////////////////////////////////////////////////////////////////////////
				//  Part 1: create the workload and assets
				///////////////////////////////////////////////////////////////////////////////////////
				environmentAsset := createEnvironmentAssetAnyWorkloadType(workloadType, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

				passwordCreReq := buildRandomProjectPassword(randoms)
				accessKeyCreReq := buildRandomProjectAccessKey(randoms)
				pvcCreReq := testing.BuildRandomProjectPVCAsset(randoms)
				pvcCreReqInf := testing.BuildRandomProjectPVCAsset(randoms)
				pvcCreReqInf.Spec.ExistingPvc = convutils.Ref(true)
				pvcCreReqInf.Spec.ClaimInfo = nil

				// At first, the k8s tracker isn't aware of the pvcCreReq claim
				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(randoms.ClusterId, &passwordCreReq, nil).
					WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil).
					WithPvc(randoms.ClusterId, &pvcCreReqInf, nil))

				pvcAsset, _, _ := createPVCAsset(pvcCreReq, router, headerForAdmin(randoms))
				// after creating the asset, it should be aware pvcCreReq claim so let's add it
				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(randoms.ClusterId, &passwordCreReq, nil).
					WithAccessKey(randoms.ClusterId, &accessKeyCreReq, nil).
					WithPvc(randoms.ClusterId, &pvcCreReqInf, nil).
					WithPvc(randoms.ClusterId, &pvcCreReq, nil))

				pvcAssetInference, _, _ := createPVCAsset(pvcCreReqInf, router, headerForAdmin(randoms))

				passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
				gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

				accessKeyAsset, _, _ := createAccessKey(&randoms.ClusterId, accessKeyCreReq, router, headerForAdmin(randoms))
				s3Asset, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKeyAsset.Meta.Id), router, headerForAdmin(randoms))

				nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(randoms), router, headerForAdmin(randoms))

				hostPathAsset, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

				reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)
				reqPayload.Assets.Environment = environmentAsset.Meta.Id

				overrides := &api.DataSourceOverrides{
					ContainerPath: convutils.Ref("/path/to/override"),
				}

				pvcAssetMeta := pvcAsset.Meta
				if workloadType == auth.WorkloadTypeInference {
					pvcAssetMeta = pvcAssetInference.Meta
				}
				reqPayload.Assets.Datasources = []api.AssetIdAndKind{
					{
						Kind:      api.ASSETKIND_GIT,
						Id:        gitAsset.Meta.Id,
						Overrides: overrides,
					},
					{
						Kind:      api.ASSETKIND_PVC,
						Id:        pvcAssetMeta.Id,
						Overrides: overrides,
					},
					{
						Kind:      api.ASSETKIND_NFS,
						Id:        nfsAsset.Meta.Id,
						Overrides: overrides,
					},
					{
						Kind:      api.ASSETKIND_HOST_PATH,
						Id:        hostPathAsset.Meta.Id,
						Overrides: overrides,
					},
				}

				// s3 is not supported in inference
				if workloadType != auth.WorkloadTypeInference {
					reqPayload.Assets.Datasources = append(reqPayload.Assets.Datasources, api.AssetIdAndKind{
						Kind:      api.ASSETKIND_S3,
						Id:        s3Asset.Meta.Id,
						Overrides: overrides,
					})
				}

				if workloadType == auth.WorkloadTypeDistributed {
					reqPayload.Distributed = &api.Distributed{
						NumWorkers:    1,
						DistFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
						Master: &api.DistMaster{
							Assets: api.AssetsIds{
								Environment: environmentAsset.Meta.Id,
								Datasources: []api.AssetIdAndKind{
									{
										Kind:      api.ASSETKIND_S3,
										Id:        s3Asset.Meta.Id,
										Overrides: overrides,
									},
								},
							},
						},
					}
				}
				sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
					return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
				})
				reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
				reqPayload.SpecificEnv = &api.SpecificRunParams{
					Command: convutils.Ref("echo"),
					Args:    convutils.Ref("123"),
				}

				createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

				var workspace api.Workspace
				jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workspace)
				Expect(jsonErr).To(BeNil())
				workspaceId := workspace.Meta.Id

				///////////////////////////////////////////////////////////////////////////////////////
				//  Part 3: get the workspace content -> V2 APIs
				///////////////////////////////////////////////////////////////////////////////////////
				urlGetV1 := urlForCreatingWorkload[workloadType]
				urlGetV2 := urlV1toV2[urlGetV1]

				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGetV2, workspaceId), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var actualWorkloadV2 api.TrainingV2
				jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualWorkloadV2)
				Expect(jsonErr).To(BeNil())

				expectedWorkloadV2 := api.TrainingV2{
					Meta: api.WorkloadMeta{
						Name:          reqPayload.Name,
						RequestedName: reqPayload.Name,
						WorkloadId:    workspaceId,
						ClusterId:     randoms.ClusterId,
						ProjectId:     fmt.Sprintf("%d", randoms.ProjectId),
						CreatedBy:     actualWorkloadV2.Meta.CreatedBy,
						CreatedAt:     actualWorkloadV2.Meta.CreatedAt,
					},
					Assets: api.AssetsRef{
						Environment: api.EnvironmentAssetRef{
							Id:   environmentAsset.Meta.Id,
							Name: environmentAsset.Meta.Name,
						},
						Compute: &api.AssetsRefCompute{
							Id:   computeAsset.Meta.Id,
							Name: computeAsset.Meta.Name,
						},
						Datasources: []api.DatasourceRef{
							getDatasourceRefWithOverrides(gitAsset.Meta, *overrides),
							getDatasourceRefWithOverrides(pvcAssetMeta, *overrides),
							getDatasourceRefWithOverrides(nfsAsset.Meta, *overrides),
							getDatasourceRefWithOverrides(hostPathAsset.Meta, *overrides),
						},
					},
				}

				var convErr *api.Error
				expectedWorkloadV2.SpecificRunParams, convErr = genericToTrainingRunParams(reqPayload.SpecificEnv)
				Expect(convErr).To(BeNil())

				// nfs, s3 and hostPath are not supported in inference
				if workloadType != auth.WorkloadTypeInference {
					expectedWorkloadV2.Assets.Datasources = append(expectedWorkloadV2.Assets.Datasources, getDatasourceRefWithOverrides(s3Asset.Meta, *overrides))
				}

				if workloadType == auth.WorkloadTypeDistributed {
					expectedWorkloadV2.Distributed = &api.DistributedData{
						NumWorkers: 1,
						Framework:  convutils.Ref(api.DISTRIBUTEDFRAMEWORK_MPI),
						Master: &api.DistributedMasterData{
							Assets: &api.AssetsRefOptional{
								Environment: api.EnvironmentAssetRef{
									Id:   environmentAsset.Meta.Id,
									Name: environmentAsset.Meta.Name,
								},
								Datasources: []api.DatasourceRef{
									getDatasourceRefWithOverrides(s3Asset.Meta, *overrides),
								},
							},
						},
					}
				}
				sort.Slice(expectedWorkloadV2.Assets.Datasources, func(i, j int) bool {
					return expectedWorkloadV2.Assets.Datasources[i].Id < expectedWorkloadV2.Assets.Datasources[j].Id
				})
				Expect(actualWorkloadV2).To(Equal(expectedWorkloadV2))
			}, forWorkspaces, forTrainings, forInference, forDistributed)
		})
	})

	Context("parallelism and completions", func() {
		DescribeTable("should fail validation if parallelism is used and workload type is not training", func(workloadType string) {
			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Parallelism: convutils.Ref(int32(2)),
			}

			if workloadType == auth.WorkloadTypeDistributed {
				reqPayload.Distributed = &api.Distributed{
					NumWorkers: 1,
				}
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			if workloadType != auth.WorkloadTypeTraining {
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("parallelism is only supported for standard training workloads"))
			} else {
				Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			}
		}, forWorkspaces, forTrainings, forInference, forDistributed)

		DescribeTable("should fail validation if completions is used and workload type is not training", func(workloadType string) {
			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Completions: convutils.Ref(int32(5)),
			}

			if workloadType == auth.WorkloadTypeDistributed {
				reqPayload.Distributed = &api.Distributed{
					NumWorkers: 1,
				}
			}

			createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			if workloadType != auth.WorkloadTypeTraining {
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("completions is only supported for standard training workloads"))
			} else {
				Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			}
		}, forWorkspaces, forTrainings, forInference, forDistributed)
	})

	Context("cluster sync helpers (updateWorkloadStatus, syncWorkloadStatus, deleteWorkload)", func() {
		It("updateWorkloadStatus sets workloads.current_state", func() {
			workloadID := createWorkload(randoms, router, envAsset, auth.WorkloadTypeWorkspace)
			ctx := adminContextForClusterSync(randoms)
			resp, err := updateWorkloadStatus(ctx, randoms.ClusterId, api.WorkloadStatusUpdate{
				Id:     workloadID,
				Status: string(api.WORKLOADSTATE_SUBMITTED),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			var rec dbmodel.WorkloadRec
			Expect(runaidb.DB().Actual().Where("id = ?", workloadID).First(&rec).Error).To(BeNil())
			Expect(string(rec.CurrentState)).To(Equal(string(api.WORKLOADSTATE_SUBMITTED)))
		})
		It("updateWorkloadStatus returns 404 when workload does not exist", func() {
			ctx := adminContextForClusterSync(randoms)
			resp, err := updateWorkloadStatus(ctx, randoms.ClusterId, api.WorkloadStatusUpdate{
				Id:     uuid.NewString(),
				Status: string(api.WORKLOADSTATE_SUBMITTED),
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
		It("syncWorkloadStatus updates multiple workspaces", func() {
			id1 := createWorkload(randoms, router, envAsset, auth.WorkloadTypeWorkspace)
			id2 := createWorkload(randoms, router, envAsset, auth.WorkloadTypeWorkspace)
			ctx := adminContextForClusterSync(randoms)
			resp, err := syncWorkloadStatus(ctx, randoms.ClusterId, []api.WorkloadStatusUpdate{
				{Id: id1, Status: string(api.WORKLOADSTATE_SUBMITTED)},
				{Id: id2, Status: string(api.WORKLOADSTATE_FAILED)},
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			var r1, r2 dbmodel.WorkloadRec
			Expect(runaidb.DB().Actual().Where("id = ?", id1).First(&r1).Error).To(BeNil())
			Expect(runaidb.DB().Actual().Where("id = ?", id2).First(&r2).Error).To(BeNil())
			Expect(string(r1.CurrentState)).To(Equal(string(api.WORKLOADSTATE_SUBMITTED)))
			Expect(string(r2.CurrentState)).To(Equal(string(api.WORKLOADSTATE_FAILED)))
		})
		It("deleteWorkload removes row and second delete is 404", func() {
			workloadID := createWorkload(randoms, router, envAsset, auth.WorkloadTypeWorkspace)
			ctx := adminContextForClusterSync(randoms)
			resp, err := deleteWorkload(ctx, randoms.ClusterId, workloadID)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNoContent))
			resp2, _ := deleteWorkload(ctx, randoms.ClusterId, workloadID)
			Expect(resp2.Code).To(Equal(http.StatusNotFound))
		})
	})
})

func assetNameValidation(router chi.Router, randoms *testutils.RandomValues, filters string) int {
	validateHttpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", urlNameAvailability, filters), nil)
	Expect(err).To(BeNil())

	validateHttpReq.Header = *headerForAdmin(randoms)

	validateHttpResp := httptest.NewRecorder()
	router.ServeHTTP(validateHttpResp, validateHttpReq)
	return validateHttpResp.Code
}

func getEnvAssetIdByType(workloadType string, envAsset *api.EnvironmentAsset, envAssetInference *api.EnvironmentAsset) string {
	if workloadType == auth.WorkloadTypeInference {
		return envAssetInference.Meta.Id
	}

	return envAsset.Meta.Id
}

func randomWorkloadCreationRequest(randoms *testutils.RandomValues, envAssetId string) *api.WorkloadCreationRequest {
	return &api.WorkloadCreationRequest{
		Name:      testutils.RandString(5),
		ProjectId: int32(randoms.ProjectId),
		ClusterId: randoms.ClusterId,
		Namespace: testutils.RandString(8),
		Assets: api.AssetsIds{
			Environment: envAssetId,
		},
	}
}

func metaToDatasourceRef(assetMeta api.AssetMeta) api.DatasourceRef {
	return api.DatasourceRef{
		Id:   assetMeta.Id,
		Name: assetMeta.Name,
		Kind: assetMeta.Kind,
	}
}

func getDatasourceRefWithOverrides(assetMeta api.AssetMeta, overrides api.DataSourceOverrides) api.DatasourceRef {
	return api.DatasourceRef{
		Id:        assetMeta.Id,
		Name:      assetMeta.Name,
		Kind:      assetMeta.Kind,
		Overrides: &overrides,
	}
}

func getEnvironmentsEntriesFromList(router chi.Router, randoms *testutils.RandomValues, httpRequest *http.Request) ([]api.EnvironmentAsset, error) {
	httpRequest.Header = *headerForAdmin(randoms)

	httpResp := httptest.NewRecorder()
	router.ServeHTTP(httpResp, httpRequest)

	var list api.EnvironmentListResponse
	jsonErr := json.Unmarshal(httpResp.Body.Bytes(), &list)

	return list.Entries, jsonErr
}

func getComputesEntriesFromList(router chi.Router, randoms *testutils.RandomValues, httpRequest *http.Request) ([]api.ComputeAsset, error) {
	httpRequest.Header = *headerForAdmin(randoms)

	httpResp := httptest.NewRecorder()
	router.ServeHTTP(httpResp, httpRequest)

	var list api.ComputeListResponse
	jsonErr := json.Unmarshal(httpResp.Body.Bytes(), &list)

	return list.Entries, jsonErr
}

func getWorkloadCreationResponse(router chi.Router, randoms *testutils.RandomValues, workloadType string, workloadReqPayload *api.WorkloadCreationRequest) *httptest.ResponseRecorder {
	workloadCreateHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(workloadReqPayload))
	Expect(err).To(BeNil())
	workloadCreateHttpReq.Header = *headerForAdmin(randoms)

	workloadCreateHttpResp := httptest.NewRecorder()
	router.ServeHTTP(workloadCreateHttpResp, workloadCreateHttpReq)
	return workloadCreateHttpResp
}

func adminContextForClusterSync(r *testutils.RandomValues) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, r.TenantIdStr)
	return context.WithValue(ctx, runaictx.CfSubjectID, "admin")
}

/* WAIT_FOR_THESE_TESTS_TO_MOVE_TO_E2E

/////////////////////////////////////////////////////////////////////////////////////
//  The following tests need to move to e2e because policy validation is no longer
//  done in assets service but in workloads service, so it will not be effective
//  to implement those tests internally in asset service
/////////////////////////////////////////////////////////////////////////////////////

type PayloadAndGoodFor struct {
	Payload interface{}
	GoodFor string
}

Context("attempt to create workload with non compliant asset", func() {
    It("should fail with conflict error", func() {
        environmentAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

        effectivePolicy := policyService.WorkspacePolicyDefaultsAndRulesV2{
            Rules: *policyService.NewNullableWorkspacePolicyRulesV2(&policyService.WorkspacePolicyRulesV2{
                Image: *policyService.NewNullableStringRules(&policyService.StringRules{
                    Options: []policyService.StringOption{
                        {
                            Value: "image-with-long-name",
                        },
                    },
                }),
            }),
        }

        policies.SetMock(policies.SimpleMockClientFactory{
            SimpleMockClient: policies.SimpleMockClient{
                EffectiveUpdatedAt:         policyMockUpdateTime,
                EffectiveWorkspacePolicyV2: &effectivePolicy,
            },
        })

        reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
        reqPayload.Assets.Environment = environmentAsset.Meta.Id

        createHttpReq, err := http.NewRequest("POST", urlWorkspace, makePayload(reqPayload))
        Expect(err).To(BeNil())
        createHttpReq.Header = *headerForAdmin(randoms)

        createHttpResp := httptest.NewRecorder()
        router.ServeHTTP(createHttpResp, createHttpReq)
        Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
        Expect(createHttpResp.Body.String()).To(ContainSubstring("does not comply with the policy: environment"))
    })
})

Context("attempt to create workload with non compliant specific run params", func() {
    It("should fail with conflict error", func() {
        workspacePolicy := policyApi.WorkspacePolicyDefaultsAndRulesV2{
            Rules: &policyApi.WorkspacePolicyRulesV2{
                Security: &policyApi.SecurityFieldsRules{
                    RunAsUid: &policyApi.IntegerRulesOptional{
                        Max: convutils.Ref(int32(40)),
                    },
                    RunAsGid: &policyApi.IntegerRulesOptional{
                        Min: convutils.Ref(int32(50)),
                    },
                },
            },
        }

        trainingPolicy := policyApi.TrainingPolicyDefaultsAndRulesV2{
            Rules: &policyApi.TrainingPolicyRulesV2{
                Security: &policyApi.SecurityFieldsRules{
                    RunAsUid: &policyApi.IntegerRulesOptional{
                        Min: convutils.Ref(int32(40)),
                    },
                    RunAsGid: &policyApi.IntegerRulesOptional{
                        Max: convutils.Ref(int32(50)),
                    },
                },
            },
        }

        policies.SetMock(policies.SimpleMockClientFactory{
            SimpleMockClient: policies.SimpleMockClient{
                EffectiveUpdatedAt:         policyMockUpdateTime,
                EffectiveWorkspacePolicyV2: workloads_v2.ConvertToWorkspacePolicy(&workspacePolicy),
                EffectiveTrainingPolicyV2:  workloads_v2.ConvertToTrainingPolicy(&trainingPolicy),
            },
        })

        goodForWorkspaceBadForTraining := buildRandomTenantEnvironmentAsset(randoms)
        goodForWorkspaceBadForTraining.Spec.UidGidSource = convutils.Ref(workloads_v2.CustomUidGid)
        goodForWorkspaceBadForTraining.Spec.RunAsUid = convutils.Ref(int32(39))
        goodForWorkspace, _, _ := createEnvironmentAsset(goodForWorkspaceBadForTraining, router, headerForAdmin(randoms))

        goodForTrainingBadForWorkspace := buildRandomTenantEnvironmentAsset(randoms)
        goodForTrainingBadForWorkspace.Spec.UidGidSource = convutils.Ref(workloads_v2.CustomUidGid)
        goodForTrainingBadForWorkspace.Spec.RunAsUid = convutils.Ref(int64(51))
        goodForTraining, _, _ := createEnvironmentAsset(goodForTrainingBadForWorkspace, router, headerForAdmin(randoms))

        payloadgoodForWorkspace := randomWorkloadCreationRequest(randoms, goodForWorkspace.Meta.Id)
        payloadGoodForTraining := randomWorkloadCreationRequest(randoms, goodForTraining.Meta.Id)

        req1 := PayloadAndGoodFor{Payload: payloadgoodForWorkspace, GoodFor: auth.WorkloadTypeWorkspace}
        req2 := PayloadAndGoodFor{Payload: payloadGoodForTraining, GoodFor: auth.WorkloadTypeTraining}

        for _, workloadType := range []string{auth.WorkloadTypeTraining, auth.WorkloadTypeWorkspace} {
            for _, req := range []PayloadAndGoodFor{req1, req2} {
                createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(req.Payload))
                Expect(err).To(BeNil())
                createHttpReq.Header = *headerForAdmin(randoms)

                createHttpResp := httptest.NewRecorder()
                router.ServeHTTP(createHttpResp, createHttpReq)

                if workloadType == req.GoodFor {
                    Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
                } else {
                    Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
                    Expect(createHttpResp.Body.String()).To(ContainSubstring("does not comply with the policy: environment"))
                }
            }
        }

    })
})

Context("validate imposed assets when creating workload", func() {
    It("if canAdd = false we cannot add assets of the same kind", func() {
        s3Asset1, _, _ := createS3Asset(buildRandomTenantS3(nil), router, headerForAdmin(randoms))
        s3Asset2, _, _ := createS3Asset(buildRandomTenantS3(nil), router, headerForAdmin(randoms))
        gitAsset1, _, _ := createGitAsset(buildRandomTenantGit(nil), router, headerForAdmin(randoms))
        gitAsset2, _, _ := createGitAsset(buildRandomTenantGit(nil), router, headerForAdmin(randoms))

        workspacePolicy := policyApi.WorkspacePolicyDefaultsAndRulesV2{
            Defaults: &policyApi.WorkspacePolicyDefaultsAndRulesV2Defaults{
                ImposedAssets: &policyApi.ImposedAssets{
                    Datasources: &[]string{
                        s3Asset1.Meta.Id,
                        gitAsset1.Meta.Id,
                    },
                },
            },
            Rules: &policyApi.WorkspacePolicyRulesV2{
                Storage: &policyApi.StorageFieldsRules{
                    S3: &policyApi.StorageFieldsRulesS3{
                        Blocked: convutils.Ref(true),
                    },
                },
            },
        }

        policies.SetMock(policies.SimpleMockClientFactory{
            SimpleMockClient: policies.SimpleMockClient{
                EffectiveUpdatedAt:         policyMockUpdateTime,
                EffectiveWorkspacePolicyV2: workloads_v2.ConvertToWorkspacePolicy(&workspacePolicy),
            },
        })

        reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
        reqPayload.Assets.Datasources = []api.AssetIdAndKind{
            {
                Kind: api.ASSETKIND_S3,
                Id:   s3Asset1.Meta.Id,
            },
            {
                Kind: api.ASSETKIND_S3,
                Id:   s3Asset2.Meta.Id,
            },
            {
                Kind: api.ASSETKIND_GIT,
                Id:   gitAsset1.Meta.Id,
            },
            {
                Kind: api.ASSETKIND_GIT,
                Id:   gitAsset2.Meta.Id,
            },
        }

        createHttpReq, err := http.NewRequest("POST", urlWorkspace, makePayload(reqPayload))
        Expect(err).To(BeNil())
        createHttpReq.Header = *headerForAdmin(randoms)

        createHttpResp := httptest.NewRecorder()
        router.ServeHTTP(createHttpResp, createHttpReq)
        Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
        Expect(createHttpResp.Body.String()).To(ContainSubstring("workload does not comply with the policy: s3 (1 errors)"))
    })
})

Context("Compliance info checking - workspace", func() {
    It("should be validated", func() {
        workspacePolicy := policyApi.WorkspacePolicyDefaultsAndRulesV2{
            Rules: &policyApi.WorkspacePolicyRulesV2{
                NodeType: &policyApi.StringRules{
                    Required: convutils.Ref(true),
                },
                NodePools: &policyApi.ArrayRules{
                    CanEdit: convutils.Ref(false),
                },
            },
            Defaults: &policyApi.WorkspacePolicyDefaultsAndRulesV2Defaults{
                NodePools: &[]string{
                    "a",
                    "b",
                },
            },
        }

        policyMockClient := policies.SimpleMockClientFactory{
            SimpleMockClient: policies.SimpleMockClient{
                EffectiveUpdatedAt:         policyMockUpdateTime,
                EffectiveWorkspacePolicyV2: ConvertToWorkspacePolicy(&workspacePolicy),
            },
        }

        policies.SetMock(&policyMockClient)

        effectivePolicy, err := GetEffectivePolicyOfProject(context.Background(), &policyMockClient, 1, auth.WorkloadTypeWorkspace)
        Expect(err).To(BeNil())

        goodRunParams := api.SpecificRunParams{
            Command:  convutils.Ref("ls -la"),
            RunAsUid: convutils.Ref(int64(-1)),
            NodeType: convutils.Ref("aaa"),
            NodePools: &[]string{
                "a",
                "b",
            },
        }

        badRunParams := api.SpecificRunParams{
            Command:  convutils.Ref("ls -la"),
            RunAsUid: convutils.Ref(int32(-1)),
            NodePools: &[]string{
                "b",
                "a",
            },
        }

        err = ValidateSpecific(&goodRunParams, effectivePolicy, auth.WorkloadTypeWorkspace)
        Expect(err).To(BeNil())

        err = ValidateSpecificRun(&badRunParams, effectivePolicy, auth.WorkloadTypeWorkspace)
        Expect(err).NotTo(BeNil())

        actualErrors := datastructs.NewSet[string](strings.Split(err.Message, "\n")...)

        expectedErrors := datastructs.NewSet[string](
            "workspace.nodePools: the administrator prohibited changing the value of this field",
            "workspace.nodeType: value must be provided for required fields",
        )

        Expect(actualErrors).To(Equal(expectedErrors))
    })
})
*/
