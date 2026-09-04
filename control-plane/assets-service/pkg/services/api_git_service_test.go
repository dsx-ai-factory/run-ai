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
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/google/uuid"

	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const (
	urlGit = "/api/v1/asset/datasource/git"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type GitTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewGitTestFactory(router chi.Router) *GitTestFactory {
	return &GitTestFactory{router: router}
}

func (c *GitTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.GitListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *GitTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.GitCreationRequest{
		Meta: creationMeta,
		Spec: api.GitAssetSpec{
			Repository: convutils.Ref(testutils.RandString(10)),
			Path:       convutils.Ref(testutils.RandString(10)),
			Branch:     convutils.Ref(testutils.RandString(8)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *GitTestFactory) marshalUpdateReq([]byte) string {
	request := api.GitUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: api.GitAssetSpec{
			Repository: convutils.Ref(testutils.RandString(10)),
			Path:       convutils.Ref(testutils.RandString(10)),
			Branch:     convutils.Ref(testutils.RandString(8)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *GitTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGitAsset(buildRandomSystemGit(nil), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GitTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGitAsset(buildRandomTenantGit(nil), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GitTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGitAsset(buildRandomClusterGit(randoms, nil), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GitTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGitAsset(buildRandomProjectGit(randoms, nil), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GitTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGitAsset(buildRandomDepartmentGit(randoms, nil), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GitTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.GitAssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *GitTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = []api.AssetIdAndKind{
		{
			Id:   assetMeta.Id,
			Kind: api.ASSETKIND_GIT,
		},
	}
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomGitHelper(passwordKeyId *string) api.GitCreationRequest {
	return api.GitCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.GitAssetSpec{
			Repository:   convutils.Ref(testutils.RandString(10)),
			Path:         convutils.Ref(testutils.RandString(10)),
			Revision:     convutils.Ref(testutils.RandString(7)),
			CredentialId: passwordKeyId,
		},
	}
}

func buildRandomProjectGit(randoms *testutils.RandomValues, passwordKeyId *string) api.GitCreationRequest {
	ans := buildRandomGitHelper(passwordKeyId)
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentGit(randoms *testutils.RandomValues, passwordKeyId *string) api.GitCreationRequest {
	ans := buildRandomGitHelper(passwordKeyId)
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantGit(passwordKeyId *string) api.GitCreationRequest {
	ans := buildRandomGitHelper(passwordKeyId)
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterGit(randoms *testutils.RandomValues, passwordKeyId *string) api.GitCreationRequest {
	ans := buildRandomGitHelper(passwordKeyId)
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemGit(passwordKeyId *string) api.GitCreationRequest {
	ans := buildRandomGitHelper(passwordKeyId)
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createGitAsset(gitCreation api.GitCreationRequest, router chi.Router, header *http.Header) (api.GitAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlGit, makePayload(gitCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.GitAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("git specific tests", func() {
	var randoms *testutils.RandomValues
	var differentProject *testutils.RandomValues

	notExist := uuid.NewString()

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		randoms = testutils.NewRandomValues()
		differentProject = randoms.CloneWithDifferentProject()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})
	})

	Context("create git with password from different project", func() {
		It("should fail, both should be from the same project", func() {
			password, _, _ := createPassword(&randoms.ClusterId, buildRandomProjectPassword(differentProject), router, headerForAdmin(randoms))

			createPayload := buildRandomProjectGit(randoms, &password.Meta.Id)
			createHttpReq, err := http.NewRequest("POST", urlPassword, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("create git with password which does not exist", func() {
		It("should fail, both should be from the same project", func() {
			createPayload := buildRandomProjectGit(randoms, convutils.Ref(notExist))
			createHttpReq, err := http.NewRequest("POST", urlGit, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("create global scope unauthenticated git", func() {
		It("should succeed, public git are allowed to be global", func() {
			createPayload := buildRandomTenantGit(nil)
			createHttpReq, err := http.NewRequest("POST", urlGit, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("create git with branch and revision", func() {
		It("should fail, branch and revision can not be passed together", func() {
			createPayload := buildRandomProjectGit(randoms, convutils.Ref(notExist))
			createPayload.Spec.Branch = convutils.Ref("branch")
			createHttpReq, err := http.NewRequest("POST", urlGit, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("create git with branch and revision", func() {
		It("should succeed, branch can be sent without revision", func() {
			createPayload := buildRandomProjectGit(randoms, nil)
			createPayload.Spec.Branch = convutils.Ref("branch")
			createPayload.Spec.Revision = nil
			createHttpReq, err := http.NewRequest("POST", urlGit, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("delete git asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			git, _, _ := createGitAsset(buildRandomProjectGit(randoms, nil), router, headerForAdmin(randoms))

			var i int
			for i = 0; i < 3; i++ {
				templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
				templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
					{
						Kind: api.ASSETKIND_GIT,
						Id:   git.Meta.Id,
					},
				}
				createTemplateAsset(templatePayload, router, headerForAdmin(randoms))
			}

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGit, git.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring("the requested git asset is being used by 3 templates"))
		})
	})

	Context("update git's spec", func() {
		It("should allow changing all fields", func() {
			origAsset := buildRandomTenantGit(nil)

			origAsset.Spec = api.GitAssetSpec{
				Repository: convutils.Ref("repo-orig"),
				Branch:     convutils.Ref("branch-orig"),
				Path:       convutils.Ref("/orig/path"),
			}

			createdAsset, _, _ := createGitAsset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, urlGit, createdAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.GitUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: api.GitAssetSpec{
					Repository: convutils.Ref("repo-updated"),
					Revision:   convutils.Ref("switched-to-revision"),
					Path:       convutils.Ref("/orig/path/updated"),
				},
			}

			updateAssetSpec(randoms, router, urlGit, createdAsset.Meta.Id, &updateRequest)

			//
			//  validate that the update changed the fields correctly
			//
			validateAssetSpec(randoms, router, urlGit, createdAsset.Meta.Id, &updateRequest.Spec)
		})
	})
})
