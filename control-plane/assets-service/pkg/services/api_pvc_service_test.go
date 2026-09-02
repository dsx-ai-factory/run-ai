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
	"strings"
	"time"

	"github.com/google/uuid"
	datavolumes_service "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/datavolumes-service"
	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	dvapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	dvclientmock "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/mocks"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const (
	urlPVC = "/api/v1/asset/datasource/pvc"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type PVCTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewPVCTestFactory(router chi.Router) *PVCTestFactory {
	return &PVCTestFactory{router: router}
}

func (c *PVCTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.PvcListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *PVCTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.PvcCreationRequest{
		Meta: creationMeta,
		Spec: api.PvcAssetSpec{
			ClaimName: convutils.Ref(strings.ToLower(testutils.RandString(5))),
			Path:      convutils.Ref(testutils.RandString(10)),
			ReadOnly:  convutils.Ref(false),
			ClaimInfo: &api.ClaimInfo{
				StorageClass: convutils.Ref(strings.ToLower(testutils.RandString(8))),
				Size:         convutils.Ref("10M"),
			},
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *PVCTestFactory) marshalUpdateReq([]byte) string {
	request := api.PvcUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: &api.PvcFieldsUpdatable{
			Path: convutils.Ref(testutils.RandString(10)),
			/* At this point we are supporting limited fields in update
			               ClaimName: convutils.Ref(testutils.RandString(5)),
			   			ReadOnly:  convutils.Ref(false),
			   			ClaimInfo: &api.ClaimInfo{
			   				StorageClass: convutils.Ref(testutils.RandString(8)),
			   				Size:         convutils.Ref("10M"),
			   			},
			*/
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *PVCTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createPVCAsset(testing.BuildRandomSystemPVCAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *PVCTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createPVCAsset(testing.BuildRandomTenantPVCAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *PVCTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createPVCAsset(testing.BuildRandomClusterPVCAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *PVCTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *PVCTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createPVCAsset(testing.BuildRandomDepartmentPVCAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *PVCTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.PvcAssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *PVCTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = append(request.Assets.Datasources, api.AssetIdAndKind{
		Kind: api.ASSETKIND_PVC,
		Id:   assetMeta.Id,
	})
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func createPVCAsset(pvcCreation api.PvcCreationRequest, router chi.Router, header *http.Header) (api.PvcAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvcCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted), createHttpResp.Body.String())

	var resp api.PvcAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestPVCAsset", func() {
	var randoms, anotherCluster *testutils.RandomValues

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		randoms = testutils.NewRandomValues()
		anotherCluster = randoms.CloneWithDifferentCluster()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherCluster)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherCluster))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
			WithCluster(anotherCluster.ClusterId, testing.BuildDisplayedCluster(anotherCluster, "2.22.0", false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	Context("storage class not allowed", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)

			mock := k8sObjectsTracker.NewSimpleMockClient()

			mock.WithStorageClassConfig(*createRequest.Spec.ClaimInfo.StorageClass, k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
				Name:    *createRequest.Spec.ClaimInfo.StorageClass,
				Permissions: *k8sObjectsApi.NewNullableStorageClassPermissions(&k8sObjectsApi.StorageClassPermissions{
					AllowedForAssets: *k8sObjectsApi.NewNullableBool(convutils.Ref(false)),
				}),
			})

			k8sObjectsTracker.SetMock(mock)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("storage class %s is excluded from creating PVCs",
				convutils.Deref(createRequest.Spec.ClaimInfo.StorageClass, ""))))
		})

		It("should return status ok when autoDelete set to true", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Meta.AutoDelete = utils.Ref(true)

			mock := k8sObjectsTracker.NewSimpleMockClient()

			mock.WithStorageClassConfig(*createRequest.Spec.ClaimInfo.StorageClass, k8sObjectsApi.StorageClassConfiguration{
				ScopeId: randoms.ClusterId,
				Name:    *createRequest.Spec.ClaimInfo.StorageClass,
				Permissions: *k8sObjectsApi.NewNullableStorageClassPermissions(&k8sObjectsApi.StorageClassPermissions{
					AllowedForAssets:            *k8sObjectsApi.NewNullableBool(convutils.Ref(false)),
					AllowedForPersistentVolumes: *k8sObjectsApi.NewNullableBool(convutils.Ref(true)),
				}),
			})

			k8sObjectsTracker.SetMock(mock)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("size pattern is invalid", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ClaimInfo = &api.ClaimInfo{
				Size: convutils.Ref("111XXX"),
			}
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body).To(ContainSubstring("must comply to the pattern"))
		})
	})

	Context("existing pvc not in the cluster", func() {
		It("should fail on precondition", func() {
			request := testing.BuildRandomProjectPVCAsset(randoms)
			request.Spec.ExistingPvc = convutils.Ref(true)
			request.Spec.ClaimInfo = nil

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(request))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("of project scope could not be found in the cluster"))
		})
	})

	Context("existing pvc scope different than cluster resource", func() {
		It("should fail on precondition", func() {
			request := testing.BuildRandomProjectPVCAsset(randoms)
			request.Spec.ExistingPvc = convutils.Ref(true)
			request.Spec.ClaimInfo = nil

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithPvc(randoms.ClusterId, &request, nil))

			// we need to test with old cluster, one that supports non-project scoped pvc
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18.0", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

			request.Meta.Scope = api.SCOPE_CLUSTER
			request.Meta.ProjectId = nil
			request.Meta.ClusterId = convutils.Ref(randoms.ClusterId)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(request))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("of cluster scope could not be found in the cluster"))
		})
	})

	Context("existing pvc scope is asset originated", func() {
		It("should fail on precondition", func() {
			request := testing.BuildRandomProjectPVCAsset(randoms)

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithPvc(randoms.ClusterId, &request, nil))

			request.Spec.ExistingPvc = convutils.Ref(true)
			request.Spec.ClaimInfo = nil

			// we need to test with old cluster, one that supports non-project scoped pvc
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18.0", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(request))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("because it is linked to an existingPvc=false asset"))
		})
	})

	Context("delete pvc asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			pvc, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_PVC,
					Id:   pvc.Meta.Id,
				},
			}
			template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested pvc asset is being used by template '%s'", template.Meta.Name)))
		})
	})

	Context("delete pvc asset used by datavolume(s)", func() {
		It("should fail on precondition", func() {
			pvc, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			dvClientMock := &dvclientmock.IDatavolumesClient{}
			datavolumes_service.SetDatavolumesMockClient(dvClientMock)
			mockResp := dvapi.DatavolumeNoSharedScopes{
				Id:            "dv-id",
				Name:          "datavolume-1",
				Description:   utils.Ref("description"),
				ProjectId:     "4500001",
				ClusterId:     pvc.Meta.ClusterId,
				CreatedAt:     utils.Ref(time.Now().AddDate(0, 0, -1)),
				CreatedBy:     utils.Ref("tester1"),
				UpdatedAt:     utils.Ref(time.Now()),
				UpdatedBy:     utils.Ref("tester2"),
				OriginPvcName: pvc.Meta.Name,
				Status: dvapi.DatavolumeStatus{
					Phase:        dvapi.DATAVOLUMEPHASE_READY,
					PhaseMessage: utils.Ref("Ready for action"),
				},
			}

			filters := buildFiltersFromQueryParams(&query_types.QueryParams{
				ClusterId: pvc.Meta.ClusterId,
			})

			dvClientMock.
				On("GetDataVolumesNotPaginated", mock.Anything, dvapi.DATAVOLUMEREQUESTTYPE_ORIGINATED,
					"", filters, "", "").
				Return([]dvapi.DatavolumeNoSharedScopes{mockResp}, nil)

			// Attempt to delete the PVC
			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)

			// Validate the precondition failure
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested pvc asset is being used by datavolumes '%s'", mockResp.Name)))

			// Verify that the mock was called as expected
			dvClientMock.AssertExpectations(GinkgoT())

			// Restore Datavolume client mock to default behavior after test modification
			setupDatavolumeClientWithGetDataVolumesResponse()
		})
	})

	Context("existing pvc + claimInfo", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ExistingPvc = convutils.Ref(true)
			createRequest.Spec.ClaimInfo = &api.ClaimInfo{
				Size: convutils.Ref("10M"),
			}
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("claimInfo should not be provided for existingPvc"))
		})
	})

	Context("existing pvc + data sharing", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ExistingPvc = convutils.Ref(true)
			createRequest.Spec.DataSharing = convutils.Ref(true)
			createRequest.Spec.ClaimInfo = nil

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("dataSharing cannot be enabled when existingPvc is true"))
		})
	})

	Context("validate Ephemeral AutoDelete", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.Ephemeral = convutils.Ref(true)
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("autoDelete is mandatory when ephemeral is true"))
		})
	})

	Context("new pvc + no claimInfo", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ClaimInfo = nil
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("claimInfo is mandatory"))
		})
	})

	Context("new pvc + no claimName", func() {
		It("should return bad request", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ClaimName = nil
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("claimName is mandatory when existingPvc = false"))
		})
	})

	Context("existing pvc + no claimName", func() {
		It("should return status ok when existingPvc is true and no claimName is provided", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createRequest.Spec.ExistingPvc = convutils.Ref(true)
			createRequest.Spec.ClaimName = nil
			createRequest.Spec.ClaimInfo = nil
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("pvc events", func() {
		It("should return only creating event", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var pvc api.PvcAsset
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &pvc)
			Expect(respJsonErr).To(BeNil())

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/history", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var events api.PvcHistoryResponse
			respJsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &events)
			Expect(respJsonErr).To(BeNil())
			Expect(len(events.Records)).To(Equal(1))
			Expect(events.Records[0].Spec.PhaseUpdate).To(Not(BeNil()))
			Expect(events.Records[0].Spec.PhaseUpdate.Phase).To(Equal("Creating"))
		})

		It("should return only creating and deleting event", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var pvc api.PvcAsset
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &pvc)
			Expect(respJsonErr).To(BeNil())

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusAccepted))

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/history", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var events api.PvcHistoryResponse
			respJsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &events)
			Expect(respJsonErr).To(BeNil())
			Expect(len(events.Records)).To(Equal(2))
			Expect(events.Records[0].Spec.PhaseUpdate).To(Not(BeNil()))
			Expect(events.Records[0].Spec.PhaseUpdate.Phase).To(Equal("Creating"))
			Expect(events.Records[1].Spec.PhaseUpdate).To(Not(BeNil()))
			Expect(events.Records[1].Spec.PhaseUpdate.Phase).To(Equal("Deleting"))

		})

		It("should return only the specific asset events", func() {
			createRequest := testing.BuildRandomProjectPVCAsset(randoms)
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var pvc api.PvcAsset
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &pvc)
			Expect(respJsonErr).To(BeNil())

			createRequest = testing.BuildRandomProjectPVCAsset(randoms)
			createHttpReq, err = http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var pvc2 api.PvcAsset
			respJsonErr = json.Unmarshal(createHttpResp.Body.Bytes(), &pvc2)
			Expect(respJsonErr).To(BeNil())

			resourceId := uuid.New().String()
			resourceId2 := uuid.New().String()
			syncReq := api.AssetResourcesIdsSyncRequest{
				Items: []api.AssetResourcesIdsItem{
					{
						AssetId:    pvc.Meta.Id,
						ResourceId: resourceId,
					},
					{
						AssetId:    pvc2.Meta.Id,
						ResourceId: resourceId2,
					},
				},
			}
			syncReqBody, err := json.Marshal(syncReq)
			Expect(err).To(BeNil())

			syncHttpReq, err := http.NewRequest("POST", "/api/v1/asset/sync/asset-resources-ids", strings.NewReader(string(syncReqBody)))
			Expect(err).To(BeNil())
			syncHttpReq.Header = *headerForAdmin(randoms)

			syncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(syncHttpResp, syncHttpReq)
			Expect(syncHttpResp.Code).To(Equal(http.StatusOK))

			eventId := uuid.New().String()
			eventId2 := uuid.New().String()
			eventSyncReq := api.EventsSyncRequest{
				Events: []api.EventSyncFields{
					{
						Id:        eventId,
						ClusterId: randoms.ClusterId,
						CreatedAt: time.Now(),
						InvolvedObject: api.InvolvedObject{
							Uid:  resourceId,
							Kind: "PersistentVolumeClaim",
							Name: "test-pvc",
						},
						Reason:  "TestReason",
						Message: "Test event message",
						Type:    "Normal",
						Source:  "test-source",
					},
					{
						Id:        eventId2,
						ClusterId: randoms.ClusterId,
						CreatedAt: time.Now(),
						InvolvedObject: api.InvolvedObject{
							Uid:  resourceId2,
							Kind: "PersistentVolumeClaim",
							Name: "test-pvc-2",
						},
						Reason:  "TestReason",
						Message: "Test event message",
						Type:    "Normal",
						Source:  "test-source",
					},
				},
			}
			eventSyncReqBody, err := json.Marshal(eventSyncReq)
			Expect(err).To(BeNil())

			eventSyncHttpReq, err := http.NewRequest("POST", "/api/v1/asset/sync/events", strings.NewReader(string(eventSyncReqBody)))
			Expect(err).To(BeNil())
			eventSyncHttpReq.Header = *headerForAdmin(randoms)

			eventSyncHttpResp := httptest.NewRecorder()
			router.ServeHTTP(eventSyncHttpResp, eventSyncHttpReq)
			Expect(eventSyncHttpResp.Code).To(Equal(http.StatusOK))

			// Now, fetch the PVC history and expect to see the event
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/history", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var events api.PvcHistoryResponse
			respJsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &events)
			Expect(respJsonErr).To(BeNil())
			Expect(len(events.Records)).To(Equal(2))
			Expect(events.Records[1].Spec.Event.Id).To(Equal(eventId))
		})

		It("should fail to fetch PVC history if cluster version is too low (2.22)", func() {
			createRequest := testing.BuildRandomClusterPVCAsset(anotherCluster)
			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(createRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(anotherCluster)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var pvc api.PvcAsset
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &pvc)
			Expect(respJsonErr).To(BeNil())

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/history", urlPVC, pvc.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(anotherCluster)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))

			var errResp api.Error
			respJsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &errResp)
			Expect(respJsonErr).To(BeNil())
			Expect(errResp.Message).To(ContainSubstring("PVC Events are not supported in this cluster version"))
		})

	})

	Context("claim name uniqueness", func() {
		It("cluster scope PVC should reject duplicate claim names in same scope branch", func() {
			claimName := "unique-claim-" + testutils.RandString(10)

			// Create first PVC at cluster scope
			pvc1 := testing.BuildRandomClusterPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-asset-1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Test 1: Cluster -> Cluster (same cluster)
			pvc2 := testing.BuildRandomClusterPVCAsset(randoms)
			pvc2.Meta.Name = "pvc-asset-2-cluster"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 2: Cluster -> Department (same cluster)
			pvc3 := testing.BuildRandomDepartmentPVCAsset(randoms)
			pvc3.Meta.Name = "pvc-asset-3-department"
			pvc3.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq2, err2 := http.NewRequest("POST", urlPVC, makePayload(pvc3))
			Expect(err2).To(BeNil())
			createHttpReq2.Header = *headerForAdmin(randoms)

			createHttpResp2 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp2, createHttpReq2)
			Expect(createHttpResp2.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp2.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 3: Cluster -> Project (same cluster)
			pvc4 := testing.BuildRandomProjectPVCAsset(randoms)
			pvc4.Meta.Name = "pvc-asset-4-project"
			pvc4.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq3, err3 := http.NewRequest("POST", urlPVC, makePayload(pvc4))
			Expect(err3).To(BeNil())
			createHttpReq3.Header = *headerForAdmin(randoms)

			createHttpResp3 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp3, createHttpReq3)
			Expect(createHttpResp3.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp3.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))
		})

		It("department scope PVC should reject duplicate claim names in same scope branch", func() {
			claimName := "unique-claim-" + testutils.RandString(10)

			// Create first PVC at department scope
			pvc1 := testing.BuildRandomDepartmentPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-asset-1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Test 1: Department -> Cluster (same cluster)
			pvc2 := testing.BuildRandomClusterPVCAsset(randoms)
			pvc2.Meta.Name = "pvc-asset-2-cluster"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 2: Department -> Department (same department)
			pvc3 := testing.BuildRandomDepartmentPVCAsset(randoms)
			pvc3.Meta.Name = "pvc-asset-3-department"
			pvc3.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq2, err2 := http.NewRequest("POST", urlPVC, makePayload(pvc3))
			Expect(err2).To(BeNil())
			createHttpReq2.Header = *headerForAdmin(randoms)

			createHttpResp2 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp2, createHttpReq2)
			Expect(createHttpResp2.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp2.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 3: Department -> Project (same department)
			pvc4 := testing.BuildRandomProjectPVCAsset(randoms)
			pvc4.Meta.Name = "pvc-asset-4-project"
			pvc4.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq3, err3 := http.NewRequest("POST", urlPVC, makePayload(pvc4))
			Expect(err3).To(BeNil())
			createHttpReq3.Header = *headerForAdmin(randoms)

			createHttpResp3 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp3, createHttpReq3)
			Expect(createHttpResp3.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp3.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))
		})

		It("project scope PVC should reject duplicate claim names in same scope branch", func() {
			claimName := "unique-claim-" + testutils.RandString(10)

			// Create first PVC at project scope
			pvc1 := testing.BuildRandomProjectPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-asset-1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Test 1: Project -> Cluster (same cluster)
			pvc2 := testing.BuildRandomClusterPVCAsset(randoms)
			pvc2.Meta.Name = "pvc-asset-2-cluster"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 2: Project -> Department (same department)
			pvc3 := testing.BuildRandomDepartmentPVCAsset(randoms)
			pvc3.Meta.Name = "pvc-asset-3-department"
			pvc3.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq2, err2 := http.NewRequest("POST", urlPVC, makePayload(pvc3))
			Expect(err2).To(BeNil())
			createHttpReq2.Header = *headerForAdmin(randoms)

			createHttpResp2 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp2, createHttpReq2)
			Expect(createHttpResp2.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp2.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))

			// Test 3: Project -> Project (same project)
			pvc4 := testing.BuildRandomProjectPVCAsset(randoms)
			pvc4.Meta.Name = "pvc-asset-4-project"
			pvc4.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq3, err3 := http.NewRequest("POST", urlPVC, makePayload(pvc4))
			Expect(err3).To(BeNil())
			createHttpReq3.Header = *headerForAdmin(randoms)

			createHttpResp3 := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp3, createHttpReq3)
			Expect(createHttpResp3.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp3.Body.String()).To(ContainSubstring("already exists in the scope hierarchy"))
		})

		It("should reject claim name that already exists in k8s cluster", func() {
			claimName := "existing-k8s-pvc-" + testutils.RandString(10)

			// Create a mock PVC creation request to use with the mock
			mockPvcReq := testing.BuildRandomProjectPVCAsset(randoms)
			mockPvcReq.Spec.ClaimName = convutils.Ref(claimName)

			// Mock k8s tracker to return an existing PVC with this claim name
			mockTracker := k8sObjectsTracker.NewSimpleMockClient()
			mockTracker = mockTracker.WithPvc(randoms.ClusterId, &mockPvcReq, nil)
			k8sObjectsTracker.SetMock(mockTracker)

			// Try to create a PVC asset with the same claim name
			pvcReq := testing.BuildRandomProjectPVCAsset(randoms)
			pvcReq.Meta.Name = "new-pvc-asset"
			pvcReq.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvcReq))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusConflict))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("already exists in the project scope in the cluster"))

			// Reset mock to avoid affecting other tests
			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())
		})

		It("should allow renaming a PVC asset whose claim already exists in the cluster", func() {
			pvcReq := testing.BuildRandomProjectPVCAsset(randoms)
			created, _, _ := createPVCAsset(pvcReq, router, headerForAdmin(randoms))

			mockTracker := k8sObjectsTracker.NewSimpleMockClient()
			mockTracker = mockTracker.WithPvc(randoms.ClusterId, &pvcReq, nil)
			k8sObjectsTracker.SetMock(mockTracker)

			updateReq := api.PvcUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: created.Meta.Name + "-renamed",
				},
				Spec: &api.PvcFieldsUpdatable{
					Path: pvcReq.Spec.Path,
				},
			}

			updateHttpReq, err := http.NewRequest("PUT",
				fmt.Sprintf("%s/%s", urlPVC, created.Meta.Id), makePayload(updateReq))
			Expect(err).To(BeNil())
			updateHttpReq.Header = *headerForAdmin(randoms)

			updateHttpResp := httptest.NewRecorder()
			router.ServeHTTP(updateHttpResp, updateHttpReq)
			Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted), updateHttpResp.Body.String())

			k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())
		})

		It("should allow same claim name in different clusters", func() {
			claimName := "shared-claim-" + testutils.RandString(10)

			// Create PVC in first cluster at cluster scope
			pvc1 := testing.BuildRandomClusterPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-cluster1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Create PVC in second cluster with same claim name - should succeed
			pvc2 := testing.BuildRandomClusterPVCAsset(anotherCluster)
			pvc2.Meta.Name = "pvc-cluster2"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(anotherCluster)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("should allow same claim name at department scope in different clusters", func() {
			claimName := "shared-claim-" + testutils.RandString(10)

			// Create PVC in first cluster at department scope
			pvc1 := testing.BuildRandomDepartmentPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-dept1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Create PVC in second cluster with same claim name - should succeed
			pvc2 := testing.BuildRandomDepartmentPVCAsset(anotherCluster)
			pvc2.Meta.Name = "pvc-dept2"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(anotherCluster)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("should allow same claim name at project scope in different clusters", func() {
			claimName := "shared-claim-" + testutils.RandString(10)

			// Create PVC in first cluster at project scope
			pvc1 := testing.BuildRandomProjectPVCAsset(randoms)
			pvc1.Meta.Name = "pvc-project1"
			pvc1.Spec.ClaimName = convutils.Ref(claimName)
			createPVCAsset(pvc1, router, headerForAdmin(randoms))

			// Create PVC in second cluster with same claim name - should succeed
			pvc2 := testing.BuildRandomProjectPVCAsset(anotherCluster)
			pvc2.Meta.Name = "pvc-project2"
			pvc2.Spec.ClaimName = convutils.Ref(claimName)

			createHttpReq, err := http.NewRequest("POST", urlPVC, makePayload(pvc2))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(anotherCluster)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})
})
