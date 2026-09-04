/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	"github.com/google/uuid"

	"net/http"
	"net/http/httptest"
	"strings"

	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/clusters-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clustersMocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlEnvironment = "/api/v1/asset/environment"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type EnvironmentTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewEnvironmentTestFactory(router chi.Router) *EnvironmentTestFactory {
	return &EnvironmentTestFactory{router: router}
}

func (c *EnvironmentTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.EnvironmentListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *EnvironmentTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.EnvironmentCreationRequest{
		Meta: creationMeta,
		Spec: generateEnvironmentAssetSpec(),
	}

	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *EnvironmentTestFactory) marshalUpdateReq([]byte) string {
	request := api.EnvironmentUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: generateEnvironmentAssetSpec(),
	}

	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *EnvironmentTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createEnvironmentAsset(generateEnvironmentCreationRequest(randoms, api.SCOPE_SYSTEM), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *EnvironmentTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createEnvironmentAsset(generateEnvironmentCreationRequest(randoms, api.SCOPE_TENANT), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *EnvironmentTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createEnvironmentAsset(generateEnvironmentCreationRequest(randoms, api.SCOPE_CLUSTER), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *EnvironmentTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createEnvironmentAsset(generateEnvironmentCreationRequest(randoms, api.SCOPE_PROJECT), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *EnvironmentTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createEnvironmentAsset(generateEnvironmentCreationRequest(randoms, api.SCOPE_DEPARTMENT), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *EnvironmentTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	gt := GenericTest[api.EnvironmentAssetSpec]{}
	gt.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *EnvironmentTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	request.Assets.Environment = assetMeta.Id
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func generateEnvironmentAssetSpec() api.EnvironmentAssetSpec {
	imagePullPolicy := api.ImagePullPolicy("Always")

	return api.EnvironmentAssetSpec{
		EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
			{
				Name:  convutils.Ref("key1"),
				Value: convutils.Ref("value1"),
			},
			{
				Name:  convutils.Ref("key2"),
				Value: convutils.Ref("value2"),
			},
		},
		Connections: []api.Connection{
			{
				Name: "my-jupyter-1",
				InternalToolInfo: &api.InternalToolInfo{
					ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
					ContainerPort:  8888,
					ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
				},
			},
		},
		Args:                     convutils.Ref(testutils.RandString(10)),
		Command:                  convutils.Ref(testutils.RandString(10)),
		Image:                    convutils.Ref(testutils.RandString(10)),
		ImagePullPolicy:          &imagePullPolicy,
		WorkingDir:               convutils.Ref(testutils.RandString(10)),
		HostIpc:                  convutils.Ref(true),
		HostNetwork:              convutils.Ref(true),
		CreateHomeDir:            convutils.Ref(true),
		AllowPrivilegeEscalation: convutils.Ref(true),
		UidGidSource:             convutils.Ref(flat_workloads.CustomUidGid),
		Capabilities:             &[]api.Capability{api.CAPABILITY_AUDIT_READ, api.CAPABILITY_CHOWN},
	}
}

func generateEnvironmentCreationRequest(randoms *testutils.RandomValues, scope api.Scope) api.EnvironmentCreationRequest {
	meta := api.AssetCreationRequest{
		Name:  testutils.RandString(5),
		Scope: scope,
		WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
			Workspace:     convutils.Ref(true),
			Training:      convutils.Ref(false),
			Inference:     convutils.Ref(false),
			Distributed:   convutils.Ref(false),
			DistFramework: nil,
		},
	}

	if scope == api.SCOPE_PROJECT {
		meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	}

	if scope == api.SCOPE_DEPARTMENT {
		meta.DepartmentId = &randoms.DepartmentId
	}

	if scope == api.SCOPE_CLUSTER {
		meta.ClusterId = &randoms.ClusterId
	}

	return api.EnvironmentCreationRequest{
		Meta: meta,
		Spec: generateEnvironmentAssetSpec(),
	}
}

func createEnvironmentAsset(environmentCreation api.EnvironmentCreationRequest, router chi.Router, header *http.Header) (api.EnvironmentAsset, []byte, []byte) {
	postPayload, err := json.Marshal(environmentCreation)
	Expect(err).To(BeNil())

	createHttpReq, err := http.NewRequest("POST", urlEnvironment, strings.NewReader(string(postPayload)))
	Expect(err).To(BeNil())
	createHttpReq.Header = *header

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)

	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted), createHttpResp.Body.String())

	var resp api.EnvironmentAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

func createEnvironmentAssetAnyWorkloadType(workloadType string, environmentCreation api.EnvironmentCreationRequest, router chi.Router, header *http.Header) api.EnvironmentAsset {
	if workloadType == auth.WorkloadTypeInference {
		addServingPortToEnvironmentCreationRequest(&environmentCreation)
	}

	envAsset, _, _ := createEnvironmentAsset(environmentCreation, router, header)
	return envAsset
}

func buildRandomEnvironmentAsset(scope api.Scope, randomsSet *testutils.RandomValues) api.EnvironmentCreationRequest {
	result := api.EnvironmentCreationRequest{
		Meta: api.AssetCreationRequest{
			Name:      testutils.RandString(5),
			Scope:     scope,
			ClusterId: nil,
			ProjectId: nil,
			WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
				Workspace:     convutils.Ref(true),
				Training:      convutils.Ref(false),
				Inference:     convutils.Ref(false),
				Distributed:   convutils.Ref(false),
				DistFramework: nil,
			},
		},
		Spec: api.EnvironmentAssetSpec{
			Image: convutils.Ref(testutils.RandString(5)),
		},
	}
	if scope == api.SCOPE_CLUSTER {
		result.Meta.ClusterId = convutils.Ref(randomsSet.ClusterId)
	}
	if scope == api.SCOPE_DEPARTMENT {
		result.Meta.DepartmentId = convutils.Ref(randomsSet.DepartmentId)
	}
	if scope == api.SCOPE_PROJECT {
		result.Meta.ProjectId = convutils.Ref(int32(randomsSet.ProjectId))
	}
	return result
}

func buildRandomSystemEnvironmentAsset(randoms *testutils.RandomValues) api.EnvironmentCreationRequest {
	return buildRandomEnvironmentAsset(api.SCOPE_SYSTEM, randoms)
}

func buildRandomTenantEnvironmentAsset(randoms *testutils.RandomValues) api.EnvironmentCreationRequest {
	return buildRandomEnvironmentAsset(api.SCOPE_TENANT, randoms)
}

func buildRandomClusterEnvironmentAsset(randoms *testutils.RandomValues) api.EnvironmentCreationRequest {
	return buildRandomEnvironmentAsset(api.SCOPE_CLUSTER, randoms)
}

func buildRandomDepartmentEnvironmentAsset(randoms *testutils.RandomValues) api.EnvironmentCreationRequest {
	return buildRandomEnvironmentAsset(api.SCOPE_DEPARTMENT, randoms)
}

func buildRandomProjectEnvironmentAsset(randoms *testutils.RandomValues) api.EnvironmentCreationRequest {
	return buildRandomEnvironmentAsset(api.SCOPE_PROJECT, randoms)
}

func addServingPortToEnvironmentCreationRequest(envAssetCreationReq *api.EnvironmentCreationRequest) {
	if envAssetCreationReq.Spec.Connections == nil {
		envAssetCreationReq.Spec.Connections = []api.Connection{}
	}

	envAssetCreationReq.Spec.Connections = append(envAssetCreationReq.Spec.Connections, api.Connection{
		Name: "serving-port",
		InternalToolInfo: &api.InternalToolInfo{
			ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
			ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
			ContainerPort:  8080,
		},
	})
}

func addEnvVarToEnvironmentCreationRequest(envAssetCreationReq *api.EnvironmentCreationRequest, envvar api.EnvironmentVariableOfAsset) {
	if envAssetCreationReq.Spec.EnvironmentVariables == nil {
		envAssetCreationReq.Spec.EnvironmentVariables = &[]api.EnvironmentVariableOfAsset{}
	}
	*envAssetCreationReq.Spec.EnvironmentVariables = append(*envAssetCreationReq.Spec.EnvironmentVariables, envvar)
}

func fetchEnvironments(router chi.Router, randoms *testutils.RandomValues, url string) api.EnvironmentListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	listHttpReq.Header = *headerForAdmin(randoms)
	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)
	Expect(listHttpResp.Code).To(Equal(http.StatusOK))
	var resp api.EnvironmentListResponse
	respJsonErr := json.Unmarshal(listHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp
}

var _ = Describe("environment specific tests", func() {
	var randoms *testutils.RandomValues
	var anotherProj *testutils.RandomValues
	var anotherCluster *testutils.RandomValues

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		randoms = testutils.NewRandomValues()
		anotherProj = randoms.CloneWithDifferentProject()
		anotherCluster = randoms.CloneWithDifferentCluster()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherProj, anotherCluster)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherProj, anotherCluster))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
			WithCluster(anotherCluster.ClusterId, testing.BuildDisplayedCluster(anotherCluster, testing.UnitTestDefaultClusterVersion, true)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

	})

	injectRoutingMethod := func(Id string, routingMethod string) {
		err := runaidb.DB().Actual().Model(&dbmodel.EnvironmentAssetRec{}).
			Where(fmt.Sprintf("%s = ?", dbtables.AssetIdColumnName), Id).Update(dbtables.EnvironmentColumnRoutingMethod, routingMethod).Error
		Expect(err).To(BeNil())
	}

	expectedEnvironments := func(expectedEnvironments []api.EnvironmentAsset, actualIds []string) {
		Expect(len(actualIds)).To(Equal(len(expectedEnvironments)))
		for _, expectedEnvironment := range expectedEnvironments {
			Expect(actualIds).To(ContainElement(expectedEnvironment.Meta.Id))
		}
	}

	Context("list environments without filter", Ordered, func() {
		var existedEnvironment api.EnvironmentAsset
		var agnosticEnvironment api.EnvironmentAsset
		var pathBaseEnvironment api.EnvironmentAsset
		var hostBaseEnvironment api.EnvironmentAsset

		BeforeEach(func() {
			config.Init()
			dbTesting.CleanTables()
			// injects environment assets with different routing methods on the tenant level
			existedEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			existedEnvironment, _, _ = createEnvironmentAsset(existedEnvironmentAsset, router, headerForAdmin(randoms))

			agnosticEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			agnosticEnvironment, _, _ = createEnvironmentAsset(agnosticEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(agnosticEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodAgnostic)

			pathBaseRoutingEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			pathBaseEnvironment, _, _ = createEnvironmentAsset(pathBaseRoutingEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(pathBaseEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodPathBase)

			hostBaseRoutingEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			hostBaseEnvironment, _, _ = createEnvironmentAsset(hostBaseRoutingEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(hostBaseEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodHostBase)

			clustersClient.GetClient()
		})

		AfterAll(func() {
			mockClient := &clustersMocks.IClusterClient{}
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 0},
					{Key: "false", Count: 0},
					{Key: "missing", Count: 0}}, nil)
			clustersClient.SetMock(mockClient)
		})

		It("list environment of a tenant without clusters", func() {

			mockClient := &clustersMocks.IClusterClient{}
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 0},
					{Key: "false", Count: 0},
					{Key: "missing", Count: 0}}, nil).Once()
			clustersClient.SetMock(mockClient)

			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})
		})

		It("list environment of a tenant with missing clusters routing method", func() {

			mockClient := &clustersMocks.IClusterClient{}
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 0},
					{Key: "false", Count: 0},
					{Key: "missing", Count: 4}}, nil).Once()
			clustersClient.SetMock(mockClient)

			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})
		})

		It("list environment of a tenant with there is a path base routing clusters", func() {

			clusterEnvironmentAssetRequest := buildRandomClusterEnvironmentAsset(randoms)
			clusterEnvironmentAsset, _, _ := createEnvironmentAsset(clusterEnvironmentAssetRequest, router, headerForAdmin(randoms))
			injectRoutingMethod(clusterEnvironmentAsset.Meta.Id, dbtables.EnvironmentColumnRoutingMethodPathBase)

			mockClient := &clustersMocks.IClusterClient{}
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 0},
					{Key: "false", Count: 1},
					{Key: "missing", Count: 0}}, nil).Once()
			clustersClient.SetMock(mockClient)

			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id, clusterEnvironmentAsset.Meta.Id})
		})

		It("list environment of a tenant with there is a host base routing clusters", func() {

			mockClient := &clustersMocks.IClusterClient{}
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 1},
					{Key: "false", Count: 0},
					{Key: "missing", Count: 0}}, nil).Once()
			clustersClient.SetMock(mockClient)

			clusterEnvironmentAssetRequest := buildRandomClusterEnvironmentAsset(anotherCluster)
			clusterEnvironmentAsset, _, _ := createEnvironmentAsset(clusterEnvironmentAssetRequest, router, headerForAdmin(anotherCluster))
			injectRoutingMethod(clusterEnvironmentAsset.Meta.Id, dbtables.EnvironmentColumnRoutingMethodHostBase)

			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id, clusterEnvironmentAsset.Meta.Id})
		})

		It("list environment of a tenant with there is a mix of routing methods clusters", func() {

			mockClient := &clustersMocks.IClusterClient{}
			clustersClient.SetMock(mockClient)
			mockClient.On("GetClustersCount", mock.Anything, "subdomain").
				Return([]openapi.ClustersCount{
					{Key: "true", Count: 1},
					{Key: "false", Count: 1},
					{Key: "missing", Count: 0}}, nil).Once()

			hostClusterEnvironmentAssetRequest := buildRandomClusterEnvironmentAsset(anotherCluster)
			hostClusterEnvironmentAsset, _, _ := createEnvironmentAsset(hostClusterEnvironmentAssetRequest, router, headerForAdmin(anotherCluster))
			injectRoutingMethod(hostClusterEnvironmentAsset.Meta.Id, dbtables.EnvironmentColumnRoutingMethodHostBase)

			pathClusterEnvironmentAssetRequest := buildRandomClusterEnvironmentAsset(anotherCluster)
			pathClusterEnvironmentAsset, _, _ := createEnvironmentAsset(pathClusterEnvironmentAssetRequest, router, headerForAdmin(anotherCluster))
			injectRoutingMethod(pathClusterEnvironmentAsset.Meta.Id, dbtables.EnvironmentColumnRoutingMethodPathBase)

			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id, hostClusterEnvironmentAsset.Meta.Id, pathClusterEnvironmentAsset.Meta.Id})
		})

	})

	Context("list inference with query param", func() {

		It("should success on precondition ", func() {
			randomTenantEnvironmentAssetBasic := buildRandomTenantEnvironmentAsset(randoms)
			randomTenantEnvironmentAssetBasic.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
				Workspace:   convutils.Ref(true),
				Training:    convutils.Ref(true),
				Inference:   convutils.Ref(false),
				Distributed: convutils.Ref(false),
			}
			_, _, _ = createEnvironmentAsset(randomTenantEnvironmentAssetBasic, router, headerForAdmin(randoms))

			randomTenantEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomTenantEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
				Workspace:     convutils.Ref(true),
				Training:      convutils.Ref(true),
				Inference:     convutils.Ref(false),
				Distributed:   convutils.Ref(true),
				DistFramework: convutils.Ref(string(api.DISTRIBUTEDFRAMEWORK_MPI)),
			}
			_, _, _ = createEnvironmentAsset(randomTenantEnvironmentAsset, router, headerForAdmin(randoms))

			randomTenantEnvironmentAssetForInference := buildRandomTenantEnvironmentAsset(randoms)
			randomTenantEnvironmentAssetForInference.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
				Workspace:   convutils.Ref(false),
				Training:    convutils.Ref(false),
				Inference:   convutils.Ref(true),
				Distributed: convutils.Ref(false),
			}
			_, _, _ = createEnvironmentAsset(randomTenantEnvironmentAssetForInference, router, headerForAdmin(randoms))

			// list compute assets
			// add query param to list only inference query param
			createHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?isInference=true", urlEnvironment), nil)
			Expect(err).To(BeNil())

			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			var resp api.EnvironmentListResponse
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(1))
		})

		It("filtring by routing method", func() {

			existedEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			existedEnvironment, _, _ := createEnvironmentAsset(existedEnvironmentAsset, router, headerForAdmin(randoms))

			agnosticEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			agnosticEnvironment, _, _ := createEnvironmentAsset(agnosticEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(agnosticEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodAgnostic)

			pathBaseRoutingEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			pathBaseEnvironment, _, _ := createEnvironmentAsset(pathBaseRoutingEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(pathBaseEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodPathBase)

			hostBaseRoutingEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			hostBaseEnvironment, _, _ := createEnvironmentAsset(hostBaseRoutingEnvironmentAsset, router, headerForAdmin(randoms))
			injectRoutingMethod(hostBaseEnvironment.Meta.Id, dbtables.EnvironmentColumnRoutingMethodHostBase)

			// Test with no params expect to get all assets
			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})

			// Test with a path base cluster id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?clusterId=%s", urlEnvironment, randoms.ClusterId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id})

			// Test with host base cluster id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?clusterId=%s", urlEnvironment, anotherCluster.ClusterId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})

			// Test with path base project id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?projectId=%v", urlEnvironment, randoms.ProjectId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id})

			// Test with path base project id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?projectId=%v", urlEnvironment, anotherCluster.ProjectId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})

			// Test with path base department id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?departmentId=%v", urlEnvironment, randoms.DepartmentId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, pathBaseEnvironment.Meta.Id})

			// Test with path base department id, expect to get only asset that supports the cluster routing method the tenant that support
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?departmentId=%v", urlEnvironment, anotherCluster.DepartmentId))
			expectedEnvironments(resp.Entries, []string{existedEnvironment.Meta.Id, agnosticEnvironment.Meta.Id, hostBaseEnvironment.Meta.Id})
		})

		It("test includeDescendants param", func() {

			// Create asset in tenant, cluster, another cluster, and project
			randomTenantEnvironmentAsset := buildRandomTenantEnvironmentAsset(randoms)
			randomTenantEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			tenantEnvironment, _, _ := createEnvironmentAsset(randomTenantEnvironmentAsset, router, headerForAdmin(randoms))
			randomClusterEnvironmentAsset := buildRandomClusterEnvironmentAsset(randoms)
			randomClusterEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			clusterEnvironment, _, _ := createEnvironmentAsset(randomClusterEnvironmentAsset, router, headerForAdmin(randoms))
			anotherCLusterEnvironmentAsset := buildRandomClusterEnvironmentAsset(anotherCluster)
			anotherCLusterEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			anotherClusterEnvironment, _, _ := createEnvironmentAsset(anotherCLusterEnvironmentAsset, router, headerForAdmin(anotherCluster))
			randomDepartmentEnvironmentAsset := buildRandomDepartmentEnvironmentAsset(randoms)
			randomDepartmentEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			departmentEnvironment, _, _ := createEnvironmentAsset(randomDepartmentEnvironmentAsset, router, headerForAdmin(randoms))
			randomProjectEnvironmentAsset := buildRandomProjectEnvironmentAsset(randoms)
			randomProjectEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			projectEnvironment, _, _ := createEnvironmentAsset(randomProjectEnvironmentAsset, router, headerForAdmin(randoms))
			anotherProjectEnvironmentAsset := buildRandomProjectEnvironmentAsset(anotherProj)
			anotherProjectEnvironmentAsset.Meta.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{Workspace: convutils.Ref(true)}
			anotherProjectEnvironment, _, _ := createEnvironmentAsset(anotherProjectEnvironmentAsset, router, headerForAdmin(anotherProj))

			// Test with no params expect to get all assets
			resp := fetchEnvironments(router, randoms, fmt.Sprintf("%s", urlEnvironment))
			expectedEnvironments(resp.Entries, []string{tenantEnvironment.Meta.Id, clusterEnvironment.Meta.Id, anotherClusterEnvironment.Meta.Id, departmentEnvironment.Meta.Id, projectEnvironment.Meta.Id, anotherProjectEnvironment.Meta.Id})

			// Test with cluster id, expect to get the cluster asset and all assets that are above (1 tenant)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?clusterId=%v", urlEnvironment, randoms.ClusterId))
			expectedEnvironments(resp.Entries, []string{clusterEnvironment.Meta.Id, tenantEnvironment.Meta.Id})

			// Test with cluster id and descendants, expect to get the cluster asset, all assets that are above (1 tenant) and all that is below (2 projects, 1 department)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?clusterId=%s&includeDescendants=true", urlEnvironment, randoms.ClusterId))
			expectedEnvironments(resp.Entries, []string{clusterEnvironment.Meta.Id, tenantEnvironment.Meta.Id, departmentEnvironment.Meta.Id, projectEnvironment.Meta.Id, anotherProjectEnvironment.Meta.Id})

			// Test with cluster id and scope, expect to get just the cluster asset
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?clusterId=%s&scope=cluster", urlEnvironment, randoms.ClusterId))
			expectedEnvironments(resp.Entries, []string{clusterEnvironment.Meta.Id})

			// Test with department id, expect to get the cluster asset and all assets that are above (1 cluster, 1 tenant)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?departmentId=%s", urlEnvironment, randoms.DepartmentId))
			expectedEnvironments(resp.Entries, []string{departmentEnvironment.Meta.Id, clusterEnvironment.Meta.Id, tenantEnvironment.Meta.Id})

			// Test with department id and scope, expect to get just the department asset
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?departmentId=%s&scope=department", urlEnvironment, randoms.DepartmentId))
			expectedEnvironments(resp.Entries, []string{departmentEnvironment.Meta.Id})

			// Test with department id and descendants, expect to get the department asset, all assets that are above (1 cluster, 1 tenant) and all that is below (2 projects)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?departmentId=%s&includeDescendants=true", urlEnvironment, randoms.DepartmentId))
			expectedEnvironments(resp.Entries, []string{clusterEnvironment.Meta.Id, tenantEnvironment.Meta.Id, departmentEnvironment.Meta.Id})

			// Test with project id, expect to get the project asset, all assets that are above (1 department, 1 cluster, 1 tenant)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?projectId=%v", urlEnvironment, randoms.ProjectId))
			expectedEnvironments(resp.Entries, []string{clusterEnvironment.Meta.Id, tenantEnvironment.Meta.Id, departmentEnvironment.Meta.Id, projectEnvironment.Meta.Id})

			// Test with project id, expect to get the project asset, all assets that are above (1 department, 1 cluster, 1 tenant)
			resp = fetchEnvironments(router, randoms, fmt.Sprintf("%s?projectId=%v&scope=project", urlEnvironment, randoms.ProjectId))
			expectedEnvironments(resp.Entries, []string{projectEnvironment.Meta.Id})

		})
	})

	Context("delete env asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			template, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, env.Meta.Id), router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlEnvironment, env.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring("the requested environment asset is being used by template '%s'", template.Meta.Name))
		})
	})

	Context("container port out of range", func() {
		It("should fail on validation ", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)
			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  -1,
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "ContainerPort: provided parameter is not respecting minimum value constraint")
		})
	})

	Context("multiple environment variables", func() {
		It("should fail on validation ", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)
			creationReq.Spec.EnvironmentVariables = &[]api.EnvironmentVariableOfAsset{
				{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
				{Name: convutils.Ref("HOME"), Value: convutils.Ref("/home/user")},
			}
			expectConnectionValidationError(creationReq, router, randoms, "duplicate variable name 'HOME'")
		})
	})

	Context("multiple tool name for connections", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
					},
				},
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "tool name 'mytool' is used by multiple connections, should be unique")
		})
	})

	Context("update asset's spec", func() {
		It("should allow changing all fields", func() {
			origAsset := buildRandomTenantEnvironmentAsset(randoms)

			origAsset.Spec = api.EnvironmentAssetSpec{
				Command:            convutils.Ref("command-orig"),
				Args:               convutils.Ref("args-orig"),
				Image:              convutils.Ref("image-orig"),
				SupplementalGroups: convutils.Ref("1,2,3"),

				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{
						Name:  convutils.Ref("var-0-orig"),
						Value: convutils.Ref("val-0-orig"),
					},
				},

				UidGidSource: convutils.Ref(api.UIDGIDSOURCE_CUSTOM),
				RunAsUid:     convutils.Ref(int64(0)),
				RunAsGid:     convutils.Ref(int64(1)),
			}

			createdEnvAsset, _, _ := createEnvironmentAsset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, urlEnvironment, createdEnvAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.EnvironmentUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: api.EnvironmentAssetSpec{
					Command:            convutils.Ref("command-updated"),
					Args:               convutils.Ref("args-updated"),
					Image:              convutils.Ref("image-updated"),
					SupplementalGroups: convutils.Ref("5,6,7"),

					EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
						{
							Name:  convutils.Ref("var-0-updated"),
							Value: convutils.Ref("val-0-updated"),
						},
					},

					UidGidSource: convutils.Ref(api.UIDGIDSOURCE_CUSTOM),
					RunAsUid:     convutils.Ref(int64(2)),
					RunAsGid:     convutils.Ref(int64(3)),
				},
			}

			updateAssetSpec(randoms, router, urlEnvironment, createdEnvAsset.Meta.Id, &updateRequest)

			//
			//  validate that the update changed the fields correctly
			//
			validateAssetSpec(randoms, router, urlEnvironment, createdEnvAsset.Meta.Id, &updateRequest.Spec)
		})
		Context("create asset with environment variables include description", func() {
			It("should allow ", func() {
				origAsset := buildRandomTenantEnvironmentAsset(randoms)

				origAsset.Spec = api.EnvironmentAssetSpec{
					Image: convutils.Ref("image-orig"),

					EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
						{
							Name:  convutils.Ref("var-0-orig"),
							Value: convutils.Ref("val-0-orig"),
						},
						{
							Name:        convutils.Ref("Home"),
							Value:       convutils.Ref("/home/user"),
							Description: convutils.Ref("Home directory"),
						},
					},
				}

				createdEnvAsset, _, _ := createEnvironmentAsset(origAsset, router, headerForAdmin(randoms))

				validateAssetSpec(randoms, router, urlEnvironment, createdEnvAsset.Meta.Id, &origAsset.Spec)
			})
		})
	})

	Context("serving port", func() {
		It("should fail on unknown protocol value ", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool-0",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
					},
				},
				{
					Name: "mytool-1",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						ServingPortInfo: &api.ServingPortInfo{
							Protocol: convutils.Ref("test"),
						},
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "unsupported value 'test' for field 'protocol'.")
		})
		It("should fail on multiple serving ports", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool-0",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						ServingPortInfo: &api.ServingPortInfo{
							Protocol: convutils.Ref(constants.ProtocolHTTP),
						},
					},
				},
				{
					Name: "mytool-1",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_SERVING_PORT,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						ServingPortInfo: &api.ServingPortInfo{
							Protocol: convutils.Ref(constants.ProtocolGRPC),
						},
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "only one serving port is allowed")
		})
	})

	Context("load balancer should not allow url", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						ExternalUrlInfo: &api.ExternalUrlInfo{
							IsCustomUrl: true,
						},
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "externalUrlInfo should not be provided for LoadBalancer connection")
		})
	})

	Context("load balancer should not allow external url", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						ExternalUrlInfo: &api.ExternalUrlInfo{
							ExternalUrl: convutils.Ref("http://ofer.com"),
						},
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "externalUrlInfo should not be provided for LoadBalancer connection")
		})
	})

	Context("environment asset is mandatory for creating workloads", func() {
		It("should fail if not selected", func() {
			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{}, &api.SpecificRunParams{})
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("environment asset is not selected, cannot create workload"))
		})
	})

	Context("url missing from specific env", func() {
		It("should fail validation", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
									ToolType:       api.INTERNALTOOLTYPE_PYCHARM,
									ExternalUrlInfo: &api.ExternalUrlInfo{
										IsCustomUrl: true,
									},
								},
							},
						},
					},
				},
			}
			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("url is mandatory for connection 'charmer'"))

			randomEnv := randomSpecificEnv()

			err = validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("url is mandatory for connection 'charmer'"))

			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name: "charmer",
				},
			}
			err = validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("url is mandatory for connection 'charmer'"))
		})
	})

	Context("port missing from specific env", func() {
		It("should fail validation", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_NODE_PORT,
									ToolType:       api.INTERNALTOOLTYPE_PYCHARM,
									NodePortInfo: &api.NodePortInfo{
										IsCustomPort: true,
									},
								},
							},
						},
					},
				},
			}
			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("port is mandatory for connection 'charmer'"))

			randomEnv := randomSpecificEnv()

			err = validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("port is mandatory for connection 'charmer'"))

			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name: "charmer",
				},
			}
			err = validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("port is mandatory for connection 'charmer'"))
		})
	})

	Context("environment with both authorizedUsers and authorizedGroups while creating workloads", func() {
		It("should fail if both authorizedUsers and authorizedGroups used", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
									ContainerPort:  8888,
									ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
								},
							},
						},
					},
				},
			}

			randomEnv := randomSpecificEnv()
			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name:             "charmer",
					AuthorizedGroups: &[]string{"group1"},
					AuthorizedUsers:  &[]string{"user1"},
					NodePort:         convutils.Ref(int32(12345)),
				},
			}

			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("authorizedUsers and authorizedGroups are mutually exclusive"))
		})

		It("should not fail if AuthorizedGroups used alone", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
									ContainerPort:  8888,
									ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
								},
							},
						},
					},
				},
			}

			randomEnv := randomSpecificEnv()
			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name:             "charmer",
					AuthorizedGroups: &[]string{"group1"},
					NodePort:         convutils.Ref(int32(12345)),
				},
			}

			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).To(BeNil())
		})

		It("should not fail if AuthorizedUsers used alone", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL,
									ContainerPort:  8888,
									ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
								},
							},
						},
					},
				},
			}

			randomEnv := randomSpecificEnv()
			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name:            "charmer",
					AuthorizedUsers: &[]string{"user1"},
					NodePort:        convutils.Ref(int32(12345)),
				},
			}

			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).To(BeNil())
		})
	})

	Context("same port used twice", func() {
		It("should fail validation", func() {
			assets := dbmodel.AssetsContainedInWorkload{
				AssetsEnvironment: &dbmodel.EnvironmentAssetRec{
					EnvironmentAssetSpec: api.EnvironmentAssetSpec{
						Connections: []api.Connection{
							{
								Name: "charmer",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_NODE_PORT,
									ToolType:       api.INTERNALTOOLTYPE_PYCHARM,
									ContainerPort:  8000,
									NodePortInfo: &api.NodePortInfo{
										IsCustomPort: true,
									},
								},
							},
							{
								Name: "notebook",
								InternalToolInfo: &api.InternalToolInfo{
									ConnectionType: api.INTERNALCONNECTIONTYPE_NODE_PORT,
									ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
									ContainerPort:  8000,
									NodePortInfo: &api.NodePortInfo{
										IsCustomPort: true,
									},
								},
							},
						},
					},
				},
			}

			randomEnv := randomSpecificEnv()
			randomEnv.Connections = &[]api.SpecificRunConnectionInfo{
				{
					Name:     "charmer",
					NodePort: convutils.Ref(int32(12345)),
				},
				{
					Name:     "notebook",
					NodePort: convutils.Ref(int32(12345)),
				},
			}
			err := validateSpecificEnv(auth.WorkloadTypeWorkspace, []dbmodel.AssetsContainedInWorkload{assets}, randomEnv)
			Expect(err).ToNot(BeNil())
			Expect(err.Message).To(ContainSubstring("node port 12345 is used by both 'notebook' and 'charmer' connections"))
		})
	})

	Context("load balancer should not allow custom port", func() {
		It("should fail on validation if not aligned", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
						NodePortInfo: &api.NodePortInfo{
							IsCustomPort: true,
						},
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "nodePortInfo should not be provided for LoadBalancer connection")
		})
	})

	Context("external tool should not allow internal info", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name:       "mytool",
					IsExternal: true,
					InternalToolInfo: &api.InternalToolInfo{
						ConnectionType: api.INTERNALCONNECTIONTYPE_LOAD_BALANCER,
						ToolType:       api.INTERNALTOOLTYPE_CUSTOM,
						ContainerPort:  8080,
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "InternalToolInfo should not be provided for external tools")
		})
	})

	Context("internal tool should not allow external info", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					ExternalToolInfo: &api.ExternalToolInfo{
						ToolType:    api.EXTERNALTOOLTYPE_WANDB,
						ExternalUrl: "http://tom.mefahed.com",
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "externalToolInfo should not be provided for internal tools")
		})
	})

	Context("internal tool should not allow external info", func() {
		It("should fail on validation", func() {
			creationReq := buildRandomTenantEnvironmentAsset(randoms)

			creationReq.Spec.Connections = []api.Connection{
				{
					Name: "mytool",
					ExternalToolInfo: &api.ExternalToolInfo{
						ToolType:    api.EXTERNALTOOLTYPE_COMET,
						ExternalUrl: "http://tom.mefahed.com",
					},
				},
			}
			expectConnectionValidationError(creationReq, router, randoms, "externalToolInfo should not be provided for internal tools")
		})
	})

	Context("same variable name twice", func() {
		It("should reject the request", func() {
			vars := []api.EnvironmentVariableOfAsset{
				{
					Name:  convutils.Ref("HOME"),
					Value: convutils.Ref("/tmp"),
				},
				{
					Name:  convutils.Ref("USER"),
					Value: convutils.Ref("myuser"),
				},
				{
					Name:  convutils.Ref("HOME"),
					Value: convutils.Ref("/home"),
				},
			}
			validationError := validateEnvironmentVars(&vars)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("duplicate variable name 'HOME'"))
		})
	})

	Context("value is not given/too much given", func() {
		It("should reject the request", func() {
			vars := []api.EnvironmentVariableOfAsset{
				{
					Name:  convutils.Ref("HOME"),
					Value: convutils.Ref("/home"),
					Credential: &api.EnvironmentVariableCredential{
						AssetId: convutils.Ref(uuid.NewString()),
						Key:     convutils.Ref("aaa"),
					},
				},
			}
			validationError := validateEnvironmentVars(&vars)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("environment variable 'HOME' must have only one of the following: value, credential, configMap"))
		})

		It("should reject the request", func() {
			vars := []api.EnvironmentVariableOfAsset{
				{
					Name: convutils.Ref("HOME"),
					ConfigMap: &api.EnvironmentVariableConfigMap{
						Name: convutils.Ref("aaa"),
						Key:  convutils.Ref("aaa"),
					},
					Credential: &api.EnvironmentVariableCredential{
						AssetId: convutils.Ref(uuid.NewString()),
						Key:     convutils.Ref("aaa"),
					},
				},
			}
			validationError := validateEnvironmentVars(&vars)
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("environment variable 'HOME' must have only one of the following: value, credential, configMap"))
		})
	})

	Context("runAsUid with uid/gid from image", func() {
		It("should reject the request", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource: convutils.Ref(flat_workloads.UidGidFromTheImage),
				RunAsUid:     convutils.Ref(int64(0)),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("runAsUid should not be provided when source of uid/gid is fromTheImage"))
		})
	})

	Context("runAsGid with source uid/gid from image", func() {
		It("should reject the request", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource: convutils.Ref(flat_workloads.UidGidFromTheImage),
				RunAsGid:     convutils.Ref(int64(3)),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("runAsGid should not be provided when source of uid/gid is fromTheImage"))
		})
	})

	Context("supplemental with source uid/gid from image", func() {
		It("should reject the request", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.UidGidFromTheImage),
				SupplementalGroups: convutils.Ref("5"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("supplementalGroups should not be provided when source of uid/gid is fromTheImage"))
		})
	})

	Context("override uid/gid in workload only for custom uid/gid", func() {
		It("should reject the request otherwise", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:              convutils.Ref(flat_workloads.UidGidFromTheImage),
				OverrideUidGidInWorkspace: true,
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("overrideUidGidInWorkspace should not be enabled when source of uid/gid is fromTheImage"))
		})
	})

	Context("supplemental groups validation", func() {
		It("it should allow both ; and , as separator", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,2;3;4,5"),
			})
			Expect(validationError).To(BeNil())
		})

		It("it should reject any other separator", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,2-3;4,5"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("invalid group id '2-3'"))
		})

		It("it should reject negative group ids", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,-6"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("invalid group id '-6': should be no less than zero"))
		})
	})

	Context("supplemental with source uid/gid from image", func() {
		It("should reject the request", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.UidGidFromTheImage),
				SupplementalGroups: convutils.Ref("5"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("supplementalGroups should not be provided when source of uid/gid is fromTheImage"))
		})
	})

	Context("negative uid/gid", func() {
		It("should reject the request", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource: convutils.Ref(flat_workloads.CustomUidGid),
				RunAsUid:     convutils.Ref(int64(-1)),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("runAsUid should be no less than 0"))

			validationError = ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource: convutils.Ref(flat_workloads.CustomUidGid),
				RunAsGid:     convutils.Ref(int64(-1)),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("runAsGid should be no less than 0"))
		})
	})

	Context("override uid/gid in workload only for custom uid/gid", func() {
		It("should reject the request otherwise", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:              convutils.Ref(flat_workloads.UidGidFromTheImage),
				OverrideUidGidInWorkspace: true,
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("overrideUidGidInWorkspace should not be enabled when source of uid/gid is fromTheImage"))
		})
	})

	Context("supplemental groups validation", func() {
		It("it should allow both ; and , as separator", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,2;3;4,5"),
			})
			Expect(validationError).To(BeNil())
		})

		It("it should reject any other separator", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,2-3;4,5"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("invalid group id '2-3'"))
		})

		It("it should reject negative group ids", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				UidGidSource:       convutils.Ref(flat_workloads.CustomUidGid),
				SupplementalGroups: convutils.Ref("1,-6"),
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("invalid group id '-6': should be no less than zero"))
		})

		It("unsupported capability", func() {
			validationError := ValidateEnvironmentRequestSpec(api.EnvironmentAssetSpec{
				Capabilities: &[]api.Capability{
					"CHOWN",
					"aaa",
				},
			})
			Expect(validationError).NotTo(BeNil())
			Expect(validationError.Message).To(ContainSubstring("unsupported value 'aaa' for field 'Capability'."))
		})
	})
})

func expectConnectionValidationError(creationReq api.EnvironmentCreationRequest, router chi.Router, randoms *testutils.RandomValues, errText string) {
	createHttpReq, err := http.NewRequest("POST", urlEnvironment, makePayload(creationReq))
	createHttpReq.Header = *headerForAdmin(randoms)
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

	Expect(respJsonErr).To(BeNil())
	Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
	Expect(createHttpResp.Body.String()).To(ContainSubstring(errText))
}
