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
	urlInference        = "/api/v1/inference"
	urlInferenceV2      = "/api/v2/inference"
	urlModelInference   = "/api/v1/model-inference"
	urlModelInferenceV2 = "/api/v2/model-inference"

	urlNameAvailability = "/api/v1/asset/name-availability"

	urlModelHuggingFaceInfo = "/api/v2/huggingface-model/info-request"
)

var (
	forWorkspaces = Entry(auth.WorkloadTypeWorkspace, auth.WorkloadTypeWorkspace)
	forTrainings  = Entry(auth.WorkloadTypeTraining, auth.WorkloadTypeTraining)
)

func createInference(r *testutils.RandomValues, router chi.Router, envAsset api.EnvironmentAsset) string {
	reqPayload := randomWorkloadCreationRequest(r, envAsset.Meta.Id)
	createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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

// expectInferenceStatus checks whether the inference is available.
func expectInferenceStatus(r *testutils.RandomValues, router chi.Router, targetId string, expectedStatus int) {
	getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInference, targetId), nil)
	Expect(err).To(BeNil())
	getHttpReq.Header = *headerForAdmin(r)

	getHttpResp := httptest.NewRecorder()
	router.ServeHTTP(getHttpResp, getHttpReq)
	Expect(getHttpResp.Code).To(Equal(expectedStatus))
}

// seedLegacyWorkload supplies historical rows to the readers that outlive the asset submission API.
func seedLegacyWorkload(r *testutils.RandomValues, workloadType, name string, assets ...api.AssetMeta) string {
	GinkgoHelper()
	rec := dbmodel.WorkloadRec{
		WorkloadType: workloadType,
		WorkloadMetaV1: api.WorkloadMetaV1{
			Id:        uuid.NewString(),
			Name:      name,
			TenantId:  int32(r.TenantId),
			ProjectId: int32(r.ProjectId),
			ClusterId: r.ClusterId,
		},
	}
	db := runaidb.DB().Actual()
	Expect(db.Create(&rec).Error).To(Succeed())
	for _, asset := range assets {
		link := dbmodel.WorkloadsAssetsRec{
			WorkloadId: rec.Id,
			AssetId:    asset.Id,
			AssetKind:  asset.Kind,
			AutoDelete: convutils.Deref(asset.AutoDelete, false),
		}
		Expect(db.Create(&link).Error).To(Succeed())
	}
	return rec.Id
}

var _ = Describe("Inference submission and historical workload records", func() {
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
		It("should return forbidden error", func() {
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerWithMissingInfo(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("Workload contains asset from different project", func() {
		It("should fail with asset not found in project error", func() {
			nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(differentProject), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body).To(ContainSubstring("not associated with the requested project"))
		})
	})

	Context("Workload contains asset from different department", func() {
		It("should fail with asset not found in project error", func() {
			nfsAsset, _, _ := createNFSAsset(buildRandomDepartmentNFSAsset(differentDepartment), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body).To(ContainSubstring("not associated with the requested project"))
		})
	})

	Context("incorrect asset type", func() {
		It("should fail validation", func() {
			nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)
			reqPayload.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   nfsAsset.Meta.Id,
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("nfs rather than git"))
		})
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

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
			nonTargetId := createInference(randoms, router, envAssetInference)
			targetId := createInference(differentProject, router, envAssetInference)

			// verify that workloads were created
			expectInferenceStatus(randoms, router, nonTargetId, http.StatusOK)
			expectInferenceStatus(differentProject, router, targetId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, differentProject.TenantIdStr)
			err := HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId: differentProject.ProjectIdStr,
			})
			Expect(err).To(BeNil())

			// verify that only the target was deleted
			expectInferenceStatus(randoms, router, nonTargetId, http.StatusOK)
			expectInferenceStatus(differentProject, router, targetId, http.StatusNotFound)
		})
	})

	Context("Handle cluster deletion", func() {
		It("basic flow", func() {
			workloadId := createInference(randoms, router, envAssetInference)

			// verify that workloads were created
			expectInferenceStatus(randoms, router, workloadId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: randoms.ClusterId,
			})
			Expect(err).To(BeNil())

			// verify that thw workload was deleted
			expectInferenceStatus(randoms, router, workloadId, http.StatusNotFound)
		})
	})

	Context("Handle tenant deletion", func() {
		It("basic flow", func() {
			workloadId := createInference(randoms, router, envAssetInference)

			// verify that workloads were created
			expectInferenceStatus(randoms, router, workloadId, http.StatusOK)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err := HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			})
			Expect(err).To(BeNil())

			// verify that the workload was deleted
			expectInferenceStatus(randoms, router, workloadId, http.StatusNotFound)
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

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("inference workload cannot use data-source asset of type s3"))
		})
	})

	Context("specific env connection not found", func() {
		It("must have names that appear in env asset", func() {
			envAssetInference, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAssetInference.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name: "not_to_be_found",
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("connection named 'not_to_be_found' is not defined in the environment"))
		})
	})

	Context("specific env connections", func() {
		It("private connection - tensorflow (ok)", func() {
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

			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name:        "notebook-a",
						ExternalUrl: convutils.Ref("http://url.com/"),
						AuthorizedUsers: &[]string{
							"test@example.com",
							"another-user@example.com",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
		})
	})

	Context("specific env connections", func() {
		It("private connection - jupyter (ok)", func() {
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

			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name:        "notebook-a",
						ExternalUrl: convutils.Ref("http://url.com/"),
						AuthorizedUsers: &[]string{
							"test@example.com",
							"another-user@example.com",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())
		})
	})

	Context("specific env connections", func() {
		It("private connection - jupyter - non url", func() {
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

			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, envAssetPayload, router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name: "notebook-a",
						AuthorizedUsers: &[]string{
							"test@example.com",
							"another-user@example.com",
						},
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("specific env autoscaling", func() {
		It("must be not nil value for inference workloads that was created with value", func() {
			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
				AutoScaling: &api.SpecificRunAutoScalingAutoScaling{
					MinReplicas: convutils.Ref(int32(1)),
					MaxReplicas: convutils.Ref(int32(1)),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInference, inferenceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualInference api.Inference
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualInference)
			Expect(jsonErr).To(BeNil())
			Expect(actualInference.Spec.SpecificEnv.AutoScaling).To(Equal(reqPayload.SpecificEnv.AutoScaling))
		})
	})

	Context("valid create request", func() {
		It("should succeed with status ok", func() {
			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 1: create the inference
			//          assets are taken from all valid scopes: system, tenant, dept, project
			///////////////////////////////////////////////////////////////////////////////////////
			environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

			pvc2CreReq := testing.BuildRandomProjectPVCAsset(randoms)
			pvc2CreReq.Spec.ExistingPvc = convutils.Ref(true)
			pvc2CreReq.Spec.ClaimInfo = nil
			passwordCreReq := buildRandomProjectPassword(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
				WithPassword(randoms.ClusterId, &passwordCreReq, nil).
				WithPvc(randoms.ClusterId, &pvc2CreReq, nil))

			pvcAssetInference, _, _ := createPVCAsset(pvc2CreReq, router, headerForAdmin(randoms))

			passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
			gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

			reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)
			reqPayload.Assets.Environment = environmentAsset.Meta.Id

			pvcAssetMeta := pvcAssetInference.Meta
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

			sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
				return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
			})
			reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("echo"),
				Args:    convutils.Ref("123"),
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
			//  Part 3: get the inference content -> V2 APIs
			///////////////////////////////////////////////////////////////////////////////////////

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInferenceV2, inferenceId), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualWorkloadV2 api.InferenceV2
			jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualWorkloadV2)
			Expect(jsonErr).To(BeNil())

			expectedWorkloadV2 := api.InferenceV2{
				Meta: api.WorkloadMeta{
					Name:          reqPayload.Name,
					RequestedName: reqPayload.Name,
					WorkloadId:    inferenceId,
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
			expectedWorkloadV2.SpecificRunParams, convErr = genericToInferenceRunParams(reqPayload.SpecificEnv)
			Expect(convErr).To(BeNil())

			sort.Slice(expectedWorkloadV2.Assets.Datasources, func(i, j int) bool {
				return expectedWorkloadV2.Assets.Datasources[i].Id < expectedWorkloadV2.Assets.Datasources[j].Id
			})
			Expect(actualWorkloadV2).To(Equal(expectedWorkloadV2))
		})
	})

	Context("get non existent workload", func() {
		It("should fail with not found error", func() {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInference, notExist), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("create workload in project the researcher has no access", func() {
		It("should fail with not found error", func() {
			environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInference, notExist), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("multiple environment variables in specific env - assets", func() {
		It("should fail validation", func() {
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
					{Name: convutils.Ref("USER"), Value: convutils.Ref("root")},
					{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("duplicate variable name 'HOME'"))
		})
	})

	Context("uidGidSource VS runAsUid and runAsGid", func() {
		It("uigGid=from image -> uid should not be provided", func() {
			userFromContainer := buildRandomTenantEnvironmentAsset(randoms)
			userFromContainer.Spec.UidGidSource = convutils.Ref(flat_workloads.UidGidFromTheImage)

			envAssetUserFromContainer, _, _ := createEnvironmentAsset(userFromContainer, router, headerForAdmin(randoms))

			// create workload

			reqCreationPayload := randomWorkloadCreationRequest(randoms, envAssetUserFromContainer.Meta.Id)
			reqCreationPayload.SpecificEnv = &api.SpecificRunParams{
				RunAsUid: convutils.Ref(int64(0)),
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqCreationPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("it is not permitted to override the runAsUid from the environment asset"))
		})

		It("uidGidSource = image -> gid should not be provided", func() {
			userFromContainer := buildRandomTenantEnvironmentAsset(randoms)
			userFromContainer.Spec.UidGidSource = convutils.Ref(flat_workloads.UidGidFromTheImage)

			envAssetUserFromContainer, _, _ := createEnvironmentAsset(userFromContainer, router, headerForAdmin(randoms))

			reqCreationPayload := randomWorkloadCreationRequest(randoms, envAssetUserFromContainer.Meta.Id)
			reqCreationPayload.SpecificEnv = &api.SpecificRunParams{
				RunAsGid: convutils.Ref(int64(3)),
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqCreationPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("it is not permitted to override the runAsGid from the environment asset"))
		})
	})

	Context("created workload -> response", func() {
		It("workload should contain the workload and cluster ID annotations", func() {
			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
		})
	})

	Context("created workload then delete by id", func() {
		It("should delete workload by id successfully", func() {
			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlInference, workloadId), makePayload(reqPayload))
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("should delete fail when try to delete non-existing id", func() {
			envAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlInference, notExist), makePayload(reqPayload))
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring("cannot find Inference"))
		})
	})

	Context("creating workload with assets -> updates assets UsageTimes", func() {
		It("by workload type", func() {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets

			environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomProjectEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomProjectCompute(randoms), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)

			environmentsHttpReq, errEnvironmentList := http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlEnvironment, *environmentAsset.Meta.ProjectId, *environmentAsset.Meta.ProjectId, auth.WorkloadTypeInference, *environmentAsset.Meta.ClusterId), nil)
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

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, *computeAsset.Meta.ProjectId, *computeAsset.Meta.ProjectId, auth.WorkloadTypeInference, *computeAsset.Meta.ClusterId), nil)
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

			workloadCreateHttpResp := getInferenceCreationResponse(router, randoms, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var inference api.Inference
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(inference.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

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

			secondWorkloadCreateHttpResp := getInferenceCreationResponse(router, randoms, secondWorkloadReqPayload)
			Expect(secondWorkloadCreateHttpResp.Code).To(Equal(http.StatusOK))

			var secondWorkload api.Inference
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
		})
	})

	Context("creating workload with cluster scope assets -> updates assets UsageTimes", func() {
		It("by workload type", func() {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets - system scope

			environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomClusterCompute(randoms), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)
			// (checking only compute asset, as previous tests already checked the environment asset, here we're
			//  focused on system scope difference)

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?complyToWorkloadType=%s", urlCompute, auth.WorkloadTypeInference), nil)
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

			workloadCreateHttpResp := getInferenceCreationResponse(router, randoms, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var inference api.Inference
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(inference.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			///////////////////////////////////////////////////////////////////////////////////////
			//STEP 4: checking the list for the existing usage info

			// checking computes list response
			computesHttpReq, errComputeList = http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, inference.Meta.ProjectId, inference.Meta.ProjectId, auth.WorkloadTypeInference, inference.Meta.ClusterId), nil)
			computes, jsonComputesErr := getComputesEntriesFromList(router, randoms, computesHttpReq)
			Expect(jsonComputesErr).To(BeNil())

			Expect(len(computes)).To(Equal(1))
			Expect(computes[0].Meta.Name).To(Equal(computes[0].Meta.Name))
			Expect(computes[0].UsageTimes).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())
		})
	})

	Context("creating workload with system scope assets -> updates assets UsageTimes", func() {
		It("by workload type", func() {
			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 1: Creating compute and environment assets - system scope

			environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomSystemCompute(), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			// STEP 2: Checking the list for the missing usage info (no workload is currently using these assets)
			// (checking only compute asset, as previous tests already checked the environment asset, here we're
			//  focused on system scope difference)

			computesHttpReq, errComputeList := http.NewRequest("GET", fmt.Sprintf("%s?complyToWorkloadType=%s", urlCompute, auth.WorkloadTypeInference), nil)
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

			workloadCreateHttpResp := getInferenceCreationResponse(router, randoms, workloadReqPayload)
			Expect(workloadCreateHttpResp.Code).To(Equal(http.StatusOK), workloadCreateHttpResp.Body.String())

			var inference api.Inference
			jsonErr := json.Unmarshal(workloadCreateHttpResp.Body.Bytes(), &inference)
			Expect(jsonErr).To(BeNil())
			Expect(inference.Spec.Assets.Compute.Id).To(Equal(computeAsset.Meta.Id))
			Expect(inference.Spec.Assets.Environment.Id).To(Equal(environmentAsset.Meta.Id))

			///////////////////////////////////////////////////////////////////////////////////////
			//STEP 4: checking the list for the existing usage info

			// checking computes list response
			computesHttpReq, errComputeList = http.NewRequest("GET", fmt.Sprintf("%s?projectId=%d&complyToProject=%d&complyToWorkloadType=%s&clusterId=%s", urlCompute, inference.Meta.ProjectId, inference.Meta.ProjectId, auth.WorkloadTypeInference, inference.Meta.ClusterId), nil)
			computes, jsonComputesErr := getComputesEntriesFromList(router, randoms, computesHttpReq)
			Expect(jsonComputesErr).To(BeNil())

			Expect(len(computes)).To(Equal(1))
			Expect(computes[0].Meta.Name).To(Equal(computes[0].Meta.Name))
			Expect(computes[0].UsageTimes).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload).NotTo(BeNil())
			Expect(computes[0].UsageTimes.LastUsedByWorkload.UnixNano() >= workloadPreCreationTimestamp).To(BeTrue())
		})
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
			It("should override the container path value of specified data source", func() {
				///////////////////////////////////////////////////////////////////////////////////////
				//  Part 1: create the workload and assets
				///////////////////////////////////////////////////////////////////////////////////////
				environmentAsset := createEnvironmentAssetAnyWorkloadType(auth.WorkloadTypeInference, buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				computeAsset, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))

				passwordCreReq := buildRandomProjectPassword(randoms)

				pvcCreReqInf := testing.BuildRandomProjectPVCAsset(randoms)
				pvcCreReqInf.Spec.ExistingPvc = convutils.Ref(true)
				pvcCreReqInf.Spec.ClaimInfo = nil

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithPassword(randoms.ClusterId, &passwordCreReq, nil).
					WithPvc(randoms.ClusterId, &pvcCreReqInf, nil))

				pvcAssetInference, _, _ := createPVCAsset(pvcCreReqInf, router, headerForAdmin(randoms))

				passwordAsset, _, _ := createPassword(&randoms.ClusterId, passwordCreReq, router, headerForAdmin(randoms))
				gitAsset, _, _ := createGitAsset(buildRandomProjectGit(randoms, &passwordAsset.Meta.Id), router, headerForAdmin(randoms))

				nfsAsset, _, _ := createNFSAsset(buildRandomProjectNFSAsset(randoms), router, headerForAdmin(randoms))

				hostPathAsset, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

				reqPayload := randomWorkloadCreationRequest(randoms, environmentAsset.Meta.Id)
				reqPayload.Assets.Environment = environmentAsset.Meta.Id

				overrides := &api.DataSourceOverrides{
					ContainerPath: convutils.Ref("/path/to/override"),
				}

				pvcAssetMeta := pvcAssetInference.Meta
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

				sort.Slice(reqPayload.Assets.Datasources, func(i, j int) bool {
					return reqPayload.Assets.Datasources[i].Id < reqPayload.Assets.Datasources[j].Id
				})
				reqPayload.Assets.Compute = convutils.Ref(computeAsset.Meta.Id)
				reqPayload.SpecificEnv = &api.SpecificRunParams{
					Command: convutils.Ref("echo"),
					Args:    convutils.Ref("123"),
				}

				createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
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
				//  Part 3: get the inference content -> V2 APIs
				///////////////////////////////////////////////////////////////////////////////////////

				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlInferenceV2, inferenceId), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var actualWorkloadV2 api.InferenceV2
				jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualWorkloadV2)
				Expect(jsonErr).To(BeNil())

				expectedWorkloadV2 := api.InferenceV2{
					Meta: api.WorkloadMeta{
						Name:          reqPayload.Name,
						RequestedName: reqPayload.Name,
						WorkloadId:    inferenceId,
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
				expectedWorkloadV2.SpecificRunParams, convErr = genericToInferenceRunParams(reqPayload.SpecificEnv)
				Expect(convErr).To(BeNil())

				sort.Slice(expectedWorkloadV2.Assets.Datasources, func(i, j int) bool {
					return expectedWorkloadV2.Assets.Datasources[i].Id < expectedWorkloadV2.Assets.Datasources[j].Id
				})
				Expect(actualWorkloadV2).To(Equal(expectedWorkloadV2))
			})
		})
	})

	Context("parallelism and completions", func() {
		It("should fail validation if parallelism is used and workload type is not training", func() {
			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Parallelism: convutils.Ref(int32(2)),
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("parallelism is only supported for standard training workloads"))

		})

		It("should fail validation if completions is used and workload type is not training", func() {
			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
			reqPayload.SpecificEnv = &api.SpecificRunParams{
				Completions: convutils.Ref(int32(5)),
			}

			createHttpReq, err := http.NewRequest("POST", urlInference, makePayload(reqPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("completions is only supported for standard training workloads"))

		})
	})

	Context("cluster sync helpers (updateWorkloadStatus, syncWorkloadStatus, deleteWorkload)", func() {
		It("updateWorkloadStatus sets workloads.current_state", func() {
			workloadID := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
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
			id1 := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
			id2 := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
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
			workloadID := seedLegacyWorkload(randoms, auth.WorkloadTypeWorkspace, testutils.RandString(8), envAsset.Meta)
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

func getInferenceCreationResponse(router chi.Router, randoms *testutils.RandomValues, workloadReqPayload *api.WorkloadCreationRequest) *httptest.ResponseRecorder {
	workloadCreateHttpReq, err := http.NewRequest("POST", urlInference, makePayload(workloadReqPayload))
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
