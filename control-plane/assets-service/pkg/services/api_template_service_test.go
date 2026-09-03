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

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const (
	urlTemplate = "/api/v1/asset/workload-template"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type TemplateTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewTemplateTestFactory(router chi.Router) *TemplateTestFactory {
	return &TemplateTestFactory{router: router}
}

func (c *TemplateTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.WorkloadTemplateListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func randomWorkloadAssets() api.AssetsIds {
	result := api.AssetsIds{
		Environment: "env_" + testutils.RandString(5),
		Compute:     convutils.Ref("compute_" + testutils.RandString(5)),
		Datasources: []api.AssetIdAndKind{
			{Kind: api.ASSETKIND_HOST_PATH, Id: "hp_" + testutils.RandString(5)},
			{Kind: api.ASSETKIND_NFS, Id: "nfs_" + testutils.RandString(5)},
			{Kind: api.ASSETKIND_GIT, Id: "git_" + testutils.RandString(5)},
			{Kind: api.ASSETKIND_GIT, Id: "git_" + testutils.RandString(5)},
			{Kind: api.ASSETKIND_PVC, Id: "pvc_" + testutils.RandString(5)},
			{Kind: api.ASSETKIND_S3, Id: "s3_" + testutils.RandString(5)},
		},
	}
	sort.Slice(result.Datasources, func(i, j int) bool {
		return result.Datasources[i].Id < result.Datasources[j].Id
	})
	return result
}

func randomSpecificEnv() *api.SpecificRunParams {
	return &api.SpecificRunParams{
		Command: convutils.Ref(testutils.RandString(12)),
		Args:    convutils.Ref(testutils.RandString(12)),
		EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
			{
				Name:  convutils.Ref(testutils.RandString(5)),
				Value: convutils.Ref(testutils.RandString(8)),
			},
			{
				Name:  convutils.Ref(testutils.RandString(5)),
				Value: convutils.Ref(testutils.RandString(8)),
			},
		},
	}
}

func (c *TemplateTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.WorkloadTemplateCreationRequest{
		Meta: creationMeta,
		Spec: api.SpecificRunCreationFields{
			Assets:      randomWorkloadAssets(),
			SpecificEnv: randomSpecificEnv(),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *TemplateTestFactory) marshalUpdateReq(createdPayload []byte) string {
	var createdSpec api.WorkloadTemplate
	err := json.Unmarshal(createdPayload, &createdSpec)
	Expect(err).To(BeNil())

	request := api.WorkloadTemplateUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.SpecificRunCreationFields{
			Assets:      dbmodel.AssetsRefToAssetIds(&createdSpec.Spec.Assets),
			SpecificEnv: randomSpecificEnv(),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *TemplateTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	envAsset, _, _ := createEnvironmentAsset(buildRandomSystemEnvironmentAsset(randoms), c.router, header)
	resp, payload, checkSum := createTemplateAsset(buildRandomSystemTemplate(randoms, envAsset.Meta.Id), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *TemplateTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), c.router, header)
	resp, payload, checkSum := createTemplateAsset(buildRandomTenantTemplate(randoms, envAsset.Meta.Id), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *TemplateTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	envAsset, _, _ := createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), c.router, header)
	resp, payload, checkSum := createTemplateAsset(buildRandomClusterTemplate(randoms, envAsset.Meta.Id), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *TemplateTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	envAsset, _, _ := createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), c.router, header)
	resp, payload, checkSum := createTemplateAsset(buildRandomDepartmentTemplate(randoms, envAsset.Meta.Id), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *TemplateTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	envAsset, _, _ := createEnvironmentAsset(buildRandomProjectEnvironmentAsset(randoms), c.router, header)
	resp, payload, checkSum := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *TemplateTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.WorkloadTemplate]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *TemplateTestFactory) addToWorkloadCreationReq(_ *api.WorkloadCreationRequest, _ *api.AssetMeta) bool {
	return false // workspace template asset cannot be part of a workspace creation request
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomTemplate(scope api.Scope, randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	request := api.WorkloadTemplateCreationRequest{
		Meta: api.AssetCreationRequest{
			Name:  testutils.RandString(5),
			Scope: scope,
		},
		Spec: api.SpecificRunCreationFields{
			Assets: api.AssetsIds{
				Environment: envAssetId,
			},
			SpecificEnv: randomSpecificEnv(),
		},
	}

	if scope == api.SCOPE_CLUSTER {
		request.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	}
	if scope == api.SCOPE_DEPARTMENT {
		request.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	}
	if scope == api.SCOPE_PROJECT {
		request.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	}

	for _, asset := range assets {
		if asset.Kind == api.ASSETKIND_ENVIRONMENT {
			request.Spec.Assets.Environment = asset.Id
		} else if asset.Kind == api.ASSETKIND_COMPUTE {
			request.Spec.Assets.Compute = convutils.Ref(asset.Id)
		} else {
			if request.Spec.Assets.Datasources == nil {
				request.Spec.Assets.Datasources = make([]api.AssetIdAndKind, 0)
			}
			request.Spec.Assets.Datasources = append(request.Spec.Assets.Datasources, api.AssetIdAndKind{
				Kind: asset.Kind,
				Id:   asset.Id,
			})
		}
	}

	return request
}

func buildRandomProjectTemplate(randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	return buildRandomTemplate(api.SCOPE_PROJECT, randoms, envAssetId, assets...)
}

func buildRandomDepartmentTemplate(randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	return buildRandomTemplate(api.SCOPE_DEPARTMENT, randoms, envAssetId, assets...)
}

func buildRandomClusterTemplate(randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	return buildRandomTemplate(api.SCOPE_CLUSTER, randoms, envAssetId, assets...)
}

func buildRandomTenantTemplate(randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	return buildRandomTemplate(api.SCOPE_TENANT, randoms, envAssetId, assets...)
}

func buildRandomSystemTemplate(randoms *testutils.RandomValues, envAssetId string, assets ...api.AssetMeta) api.WorkloadTemplateCreationRequest {
	return buildRandomTemplate(api.SCOPE_SYSTEM, randoms, envAssetId, assets...)
}

func createTemplateAsset(request api.WorkloadTemplateCreationRequest, router chi.Router, header *http.Header) (api.WorkloadTemplate, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(request))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.WorkloadTemplate
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestTemplate", func() {
	var randoms *testutils.RandomValues

	router := CreateRouter()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		dbTesting.CleanTables()
		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		auth.TestingAllowCreateSystemScopedAssets = true
	})

	Context("attempt to create template in system, tenant, department and project scope", func() {
		It("should be accepted or rejected based on the asset scope", func() {
			envAssetSystem, _, _ := createEnvironmentAsset(buildRandomSystemEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAssetTenant, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAssetCluster, _, _ := createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAssetDept, _, _ := createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAssetProj, _, _ := createEnvironmentAsset(buildRandomProjectEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			checkCreateTemplate(randoms, router, api.SCOPE_SYSTEM, envAssetSystem.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_SYSTEM, envAssetTenant.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_SYSTEM, envAssetCluster.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_SYSTEM, envAssetDept.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_SYSTEM, envAssetProj.Meta, http.StatusBadRequest)

			checkCreateTemplate(randoms, router, api.SCOPE_TENANT, envAssetSystem.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_TENANT, envAssetTenant.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_TENANT, envAssetCluster.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_TENANT, envAssetDept.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_TENANT, envAssetProj.Meta, http.StatusBadRequest)

			checkCreateTemplate(randoms, router, api.SCOPE_CLUSTER, envAssetSystem.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_CLUSTER, envAssetTenant.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_CLUSTER, envAssetCluster.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_CLUSTER, envAssetDept.Meta, http.StatusBadRequest)
			checkCreateTemplate(randoms, router, api.SCOPE_CLUSTER, envAssetProj.Meta, http.StatusBadRequest)

			checkCreateTemplate(randoms, router, api.SCOPE_DEPARTMENT, envAssetSystem.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_DEPARTMENT, envAssetTenant.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_DEPARTMENT, envAssetCluster.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_DEPARTMENT, envAssetDept.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_DEPARTMENT, envAssetProj.Meta, http.StatusBadRequest)

			checkCreateTemplate(randoms, router, api.SCOPE_PROJECT, envAssetSystem.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_PROJECT, envAssetTenant.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_PROJECT, envAssetCluster.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_PROJECT, envAssetDept.Meta, http.StatusAccepted)
			checkCreateTemplate(randoms, router, api.SCOPE_PROJECT, envAssetProj.Meta, http.StatusAccepted)
		})
	})

	Context("specific env connections", func() {
		It("must have names that appear in env asset", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			createPayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			createPayload.Spec.SpecificEnv = &api.SpecificRunParams{
				Connections: &[]api.SpecificRunConnectionInfo{
					{
						Name: "not_to_be_found",
					},
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createPayload))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("connection named 'not_to_be_found' is not defined in the environment"))
		})
	})

	Context("assets contained in template should exist", func() {
		It("should fail with bad request", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			createPayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			createPayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_GIT,
					Id:   uuid.NewString(),
				},
			}

			createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createPayload))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("not associated with the requested pro"))
		})
	})

	Context("list template types", func() {
		It("should succeed", func() {
			var arrayOfResponses []api.WorkloadTemplate
			randomTenantEnvironmentAsset1 := buildRandomTenantEnvironmentAsset(randoms)
			envAsset1WST := api.WorkloadSupportedTypes{
				Training:    convutils.Ref(true),
				Workspace:   convutils.Ref(false),
				Inference:   convutils.Ref(false),
				Distributed: nil,
			}
			randomTenantEnvironmentAsset1.Meta.WorkloadSupportedTypes = &envAsset1WST

			// envAsset1 is training type environment
			var workloadTemplate1 api.WorkloadTemplate
			envAsset1, _, _ := createEnvironmentAsset(randomTenantEnvironmentAsset1, router, headerForAdmin(randoms))
			createHttpResp, err := buildAndCreateTemplate(router, randoms, envAsset1)
			err = json.Unmarshal(createHttpResp.Body.Bytes(), &workloadTemplate1)
			Expect(err).To(BeNil())
			arrayOfResponses = append(arrayOfResponses, workloadTemplate1)

			randomTenantEnvironmentAsset2 := buildRandomTenantEnvironmentAsset(randoms)
			envAsset2WST := api.WorkloadSupportedTypes{
				Training:    convutils.Ref(false),
				Workspace:   convutils.Ref(true),
				Inference:   convutils.Ref(false),
				Distributed: nil,
			}
			randomTenantEnvironmentAsset2.Meta.WorkloadSupportedTypes = &envAsset2WST
			// envAsset2 is workspace type environment
			var workloadTemplate2 api.WorkloadTemplate
			envAsset2, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			createHttpResp, err = buildAndCreateTemplate(router, randoms, envAsset2)
			err = json.Unmarshal(createHttpResp.Body.Bytes(), &workloadTemplate2)
			Expect(err).To(BeNil())
			arrayOfResponses = append(arrayOfResponses, workloadTemplate2)

			createHttpReq, err := http.NewRequest("GET", urlTemplate, nil)
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)
			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			workloadTemplateListResponse := api.WorkloadTemplateListResponse{}

			err = json.Unmarshal(createHttpResp.Body.Bytes(), &workloadTemplateListResponse)
			Expect(err).To(BeNil())

			// sort workloadTemplateListResponse by name
			sort.Slice(workloadTemplateListResponse.Entries[:], func(i, j int) bool {
				return workloadTemplateListResponse.Entries[i].Meta.Name < workloadTemplateListResponse.Entries[j].Meta.Name
			})

			// sort arrayOfResponse by name
			sort.Slice(arrayOfResponses[:], func(i, j int) bool {
				return arrayOfResponses[i].Meta.Name < arrayOfResponses[j].Meta.Name
			})

			Expect(len(workloadTemplateListResponse.Entries)).To(Equal(2))
			for i := 0; i < len(workloadTemplateListResponse.Entries); i++ {
				// same name and same WorkloadSupportedTypes
				Expect(workloadTemplateListResponse.Entries[i].Meta.WorkloadSupportedTypes).To(Equal(arrayOfResponses[i].Meta.WorkloadSupportedTypes))
				Expect(workloadTemplateListResponse.Entries[i].Meta.Name).To(Equal(arrayOfResponses[i].Meta.Name))
			}
		})
	})

	Context("update template content", func() {
		It("should succeed", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			pvcAsset1, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))
			pvcAsset2, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 1: create a template with single asset
			///////////////////////////////////////////////////////////////////////////////////////
			createRequest := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createRequest))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			var createdTemplate api.WorkloadTemplate
			err = json.Unmarshal(createHttpResp.Body.Bytes(), &createdTemplate)
			Expect(err).To(BeNil())
			Expect(createRequest.Spec.Assets).To(Equal(dbmodel.AssetsRefToAssetIds(&createdTemplate.Spec.Assets)))

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 2: add two PVC assets
			///////////////////////////////////////////////////////////////////////////////////////
			updateRequest := buildUpdateRequest(&createRequest)
			updateRequest.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_PVC,
					Id:   pvcAsset1.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_PVC,
					Id:   pvcAsset2.Meta.Id,
				},
			}
			sort.Slice(updateRequest.Spec.Assets.Datasources, func(i, j int) bool {
				return updateRequest.Spec.Assets.Datasources[i].Id < updateRequest.Spec.Assets.Datasources[j].Id
			})
			updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlTemplate, createdTemplate.Meta.Id), makePayload(updateRequest))
			updateHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			updateHttpResp := httptest.NewRecorder()
			router.ServeHTTP(updateHttpResp, updateHttpReq)
			Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))

			var updatedTemplate api.WorkloadTemplate
			err = json.Unmarshal(updateHttpResp.Body.Bytes(), &updatedTemplate)
			Expect(err).To(BeNil())
			Expect(dbmodel.AssetsRefToAssetIds(&updatedTemplate.Spec.Assets)).To(Equal(updateRequest.Spec.Assets))

			///////////////////////////////////////////////////////////////////////////////////////
			//  Part 3: remove one PVC assets
			///////////////////////////////////////////////////////////////////////////////////////
			updateRequest.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_PVC,
					Id:   pvcAsset2.Meta.Id,
				},
			}
			updateHttpReq, err = http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlTemplate, createdTemplate.Meta.Id), makePayload(updateRequest))
			updateHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			updateHttpResp = httptest.NewRecorder()
			router.ServeHTTP(updateHttpResp, updateHttpReq)
			Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))

			err = json.Unmarshal(updateHttpResp.Body.Bytes(), &updatedTemplate)
			Expect(err).To(BeNil())
			Expect(dbmodel.AssetsRefToAssetIds(&updatedTemplate.Spec.Assets)).To(Equal(updateRequest.Spec.Assets))
		})
	})

	Context("show tools contained in environment assets", func() {
		It("should show tools correctly", func() {
			env1 := buildRandomTenantEnvironmentAsset(randoms)
			env1.Spec.Connections = []api.Connection{
				{Name: "my-custom", InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_CUSTOM, ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL, ContainerPort: 8080,
				}},
				{Name: "my-mlflow-1", InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_MLFLOW, ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL, ContainerPort: 8080,
				}},
				{Name: "my-mlflow-2", InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_MLFLOW, ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL, ContainerPort: 8080,
				}},
				{Name: "my-tensor", InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_TENSORBOARD, ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL, ContainerPort: 8080,
				}},
				{Name: "my-mlflow-3", InternalToolInfo: &api.InternalToolInfo{
					ToolType: api.INTERNALTOOLTYPE_MLFLOW, ConnectionType: api.INTERNALCONNECTIONTYPE_EXTERNAL_URL, ContainerPort: 8080,
				}},
			}
			envAsset1, _, _ := createEnvironmentAsset(env1, router, headerForAdmin(randoms))

			createPayload := buildRandomProjectTemplate(randoms, envAsset1.Meta.Id)

			createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createPayload))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			listHttpReq, err := http.NewRequest("GET", urlTemplate, nil)
			listHttpReq.Header = *headerForNonAdmin(randoms)
			Expect(err).To(BeNil())

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)
			Expect(listHttpResp.Code).To(Equal(http.StatusOK))

			var listResponse api.WorkloadTemplateListResponse
			err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResponse)
			Expect(err).To(BeNil())
			Expect(len(listResponse.Entries)).To(Equal(1))
			expectedToolTypes := []api.ToolType{
				api.TOOLTYPE_CUSTOM,
				api.TOOLTYPE_MLFLOW,
				api.TOOLTYPE_TENSORBOARD,
			}
			Expect(*listResponse.Entries[0].Spec.Assets.Environment.ToolTypes).To(Equal(expectedToolTypes))
		})
	})
})

func buildAndCreateTemplate(router chi.Router, randoms *testutils.RandomValues, envAsset1 api.EnvironmentAsset) (*httptest.ResponseRecorder, error) {
	createRequest := buildRandomProjectTemplate(randoms, envAsset1.Meta.Id)
	createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createRequest))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(randoms)
	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
	return createHttpResp, err
}

func buildUpdateRequest(createRequest *api.WorkloadTemplateCreationRequest) api.WorkloadTemplateUpdateRequest {
	return api.WorkloadTemplateUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name: createRequest.Meta.Name,
		},
		Spec: api.SpecificRunCreationFields{
			Assets: createRequest.Spec.Assets,
		},
	}
}

func checkCreateTemplate(randoms *testutils.RandomValues, router chi.Router, templateScope api.Scope, envAsset api.AssetMeta, expectedStatus int) {
	text := fmt.Sprintf("creating %s scope template + %s scope asset -> expected %d", templateScope, envAsset.Scope, expectedStatus)

	var createPayload api.WorkloadTemplateCreationRequest
	switch templateScope {
	case api.SCOPE_SYSTEM:
		createPayload = buildRandomSystemTemplate(randoms, envAsset.Id)
	case api.SCOPE_TENANT:
		createPayload = buildRandomTenantTemplate(randoms, envAsset.Id)
	case api.SCOPE_CLUSTER:
		createPayload = buildRandomClusterTemplate(randoms, envAsset.Id)
	case api.SCOPE_DEPARTMENT:
		createPayload = buildRandomDepartmentTemplate(randoms, envAsset.Id)
	default:
		createPayload = buildRandomProjectTemplate(randoms, envAsset.Id)
	}

	createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(createPayload))
	createHttpReq.Header = *headerForAdmin(randoms)
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(expectedStatus), text)
}
