/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm Repos API tests", Ordered, func() {
	helmRepoIds := map[string]string{}

	var helmReposApiService api.HelmReposAPIServicer
	var ctx context.Context

	BeforeEach(func() {
		ctx = SetContext(TestTenant, TestUser)
		helmReposApiService = services.NewHelmReposAPIService()
		authorization.SetMock(&SimpleMockFactory{})
	})

	AfterEach(func() {
		for helmRepoId := range helmRepoIds {
			cleanupCtx := SetContext(TestTenant, TestUser)
			_ = dao.DeleteHelmRepo(cleanupCtx, helmRepoId)
		}
		helmRepoIds = map[string]string{}
	})

	Context("create helm repo", func() {
		var (
			request api.HelmRepoCreationRequest
		)

		BeforeEach(func() {
			request = api.HelmRepoCreationRequest{
				Name:      "test-helm-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}
		})

		It("create sanity", func() {
			now := time.Now()
			time.Sleep(1 * time.Millisecond)

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))

			created := res.Body.(api.HelmRepo)
			helmRepoIds[created.Id] = TestTenant

			validateHelmRepo(created, request)
			Expect(created.CreatedAt.After(now)).To(BeTrue())
			Expect(created.UpdatedAt.After(now)).To(BeTrue())
		})

		It("create with different name", func() {
			request.Name = "another-helm-repo"
			request.Url = "https://another.charts.com"

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))

			created := res.Body.(api.HelmRepo)
			helmRepoIds[created.Id] = TestTenant

			Expect(created.Name).To(Equal(request.Name))
			Expect(created.Url).To(Equal(request.Url))
		})

		It("create private repo", func() {
			request.Name = "private-helm-repo"
			request.Url = "https://private.charts.example.com"
			request.IsPrivate = true

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))

			created := res.Body.(api.HelmRepo)
			helmRepoIds[created.Id] = TestTenant
			validateHelmRepo(created, request)
			Expect(created.IsPrivate).To(BeTrue())
		})

		It("create with missing subjectId should fail", func() {
			ctx = context.Background()
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))

			ctx = context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)
			res, err = helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("create with conflict: multiple helm repos with same name", func() {
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			created := res.Body.(api.HelmRepo)
			helmRepoIds[created.Id] = TestTenant

			request.Url = "https://different.charts.com"
			res, err = helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo err should be nil")
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("already exists"))
		})

		It("create with project scope type should fail", func() {
			request.ScopeType = api.SCOPETYPE_PROJECT
			request.ScopeId = "1"
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported for helm repos"))
		})

		It("create with system scope should fail", func() {
			request.ScopeType = api.SCOPETYPE_SYSTEM
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("system scope is not supported"))
		})

		It("create with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("create with ftp URL should fail", func() {
			request.Url = "ftp://charts.example.com"
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})

		It("create with loopback URL should fail", func() {
			request.Url = "https://127.0.0.1"
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})

		It("create with private IP URL should fail", func() {
			request.Url = "https://10.0.0.1/repo"
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})

		It("create with cloud metadata URL should fail", func() {
			request.Url = "http://169.254.169.254/latest/meta-data"
			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})
	})

	Context("get helm repo", func() {
		var (
			testHelmRepoId string
			request        api.HelmRepoCreationRequest
		)

		BeforeEach(func() {
			request = api.HelmRepoCreationRequest{
				Name:      "test-helm-repo-get",
				Url:       "https://charts.get.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testHelmRepoId = res.Body.(api.HelmRepo).Id
			helmRepoIds[testHelmRepoId] = TestTenant
		})

		It("get sanity", func() {
			res, err := helmReposApiService.GetHelmRepoById(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			helmRepo := res.Body.(api.HelmRepo)
			validateHelmRepo(helmRepo, request)
		})

		It("get with invalid helm repo id", func() {
			res, err := helmReposApiService.GetHelmRepoById(ctx, "invalid-uuid")
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("get with non-existent helm repo id", func() {
			nonExistentId := uuid.NewString()
			res, err := helmReposApiService.GetHelmRepoById(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("get with different tenant - should not find", func() {
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err := helmReposApiService.GetHelmRepoById(ctx2, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("list helm repos", func() {
		var (
			testHelmRepoIds []string
		)

		BeforeEach(func() {
			testHelmRepoIds = []string{}

			for i := 0; i < 3; i++ {
				request := api.HelmRepoCreationRequest{
					Name:      "list-helm-repo-" + uuid.NewString()[:8],
					Url:       "https://charts" + uuid.NewString()[:8] + ".example.com",
					ScopeType: api.SCOPETYPE_TENANT,
					ScopeId:   TestTenant,
				}

				res, err := helmReposApiService.CreateHelmRepo(ctx, request)
				Expect(err).To(BeNil(), "Create helm repo should succeed")
				Expect(res.Code).To(Equal(http.StatusOK))
				created := res.Body.(api.HelmRepo)
				testHelmRepoIds = append(testHelmRepoIds, created.Id)
				helmRepoIds[created.Id] = TestTenant
			}
		})

		It("list sanity", func() {
			res, err := helmReposApiService.ListHelmRepos(ctx)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			response := res.Body.(api.HelmRepoListResponse)
			Expect(len(response.HelmRepos)).To(BeNumerically(">=", 3))

			foundIds := make(map[string]bool)
			for _, helmRepo := range response.HelmRepos {
				foundIds[helmRepo.Id] = true
			}

			for _, id := range testHelmRepoIds {
				Expect(foundIds[id]).To(BeTrue(), "Helm repo %s should be in the list", id)
			}
		})

		It("list with different tenant - should not see other tenant helm repos", func() {
			ctx2 := SetContext(TestTenant2, TestUser)

			res, err := helmReposApiService.ListHelmRepos(ctx2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			response := res.Body.(api.HelmRepoListResponse)

			foundIds := make(map[string]bool)
			for _, helmRepo := range response.HelmRepos {
				foundIds[helmRepo.Id] = true
			}

			for _, id := range testHelmRepoIds {
				Expect(foundIds[id]).To(BeFalse(), "Helm repo %s should NOT be in the list for different tenant", id)
			}
		})
	})

	Context("update helm repo", func() {
		var (
			testHelmRepoId string
			createRequest  api.HelmRepoCreationRequest
		)

		BeforeEach(func() {
			createRequest = api.HelmRepoCreationRequest{
				Name:      "test-helm-repo-update",
				Url:       "https://charts.update.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := helmReposApiService.CreateHelmRepo(ctx, createRequest)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testHelmRepoId = res.Body.(api.HelmRepo).Id
			helmRepoIds[testHelmRepoId] = TestTenant
		})

		It("update name sanity", func() {
			time.Sleep(1 * time.Millisecond)
			ctx = SetContext(TestTenant, TestUser2)

			updateRequest := api.HelmRepoUpdateRequest{
				Name: "updated-helm-repo-name",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.HelmRepo)
			Expect(updated.Name).To(Equal(updateRequest.Name))
			Expect(updated.Url).To(Equal(createRequest.Url))
			Expect(updated.UpdatedBy).To(Equal(TestUser2))
			Expect(updated.UpdatedAt.After(updated.CreatedAt)).To(BeTrue())
		})

		It("update url sanity", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Url: "https://updated.charts.com",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.HelmRepo)
			Expect(updated.Url).To(Equal(updateRequest.Url))
			Expect(updated.Name).To(Equal(createRequest.Name))
		})

		It("update both name and url", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Name: "fully-updated-helm-repo",
				Url:  "https://fully.updated.charts.com",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.HelmRepo)
			Expect(updated.Name).To(Equal(updateRequest.Name))
			Expect(updated.Url).To(Equal(updateRequest.Url))
		})

		It("update isPrivate", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				IsPrivate: func() *bool {
					v := true
					return &v
				}(),
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.HelmRepo)
			Expect(updated.IsPrivate).To(BeTrue())

			updateRequest.IsPrivate = func() *bool {
				v := false
				return &v
			}()
			res, err = helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated = res.Body.(api.HelmRepo)
			Expect(updated.IsPrivate).To(BeFalse())
		})

		It("update with invalid helm repo id", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Name: "new-name",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, "invalid-uuid", updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("update with non-existent helm repo id", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Name: "new-name",
			}

			nonExistentId := uuid.NewString()
			res, err := helmReposApiService.UpdateHelmRepo(ctx, nonExistentId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("update with conflicting name", func() {
			anotherRequest := api.HelmRepoCreationRequest{
				Name:      "another-helm-repo-for-conflict",
				Url:       "https://another.charts.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := helmReposApiService.CreateHelmRepo(ctx, anotherRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			another := res.Body.(api.HelmRepo)
			helmRepoIds[another.Id] = TestTenant

			updateRequest := api.HelmRepoUpdateRequest{
				Name: anotherRequest.Name,
			}

			res, err = helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("already exists"))
		})

		It("update with same name (no change) should succeed", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Name: createRequest.Name,
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("update with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			updateRequest := api.HelmRepoUpdateRequest{
				Name: "new-name",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("update with missing subject id", func() {
			ctx = context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)

			updateRequest := api.HelmRepoUpdateRequest{
				Name: "new-name",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("update with ftp URL should fail", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Url: "ftp://evil.example.com",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})

		It("update with private IP URL should fail", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Url: "https://192.168.1.1/repo",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})

		It("update with cloud metadata URL should fail", func() {
			updateRequest := api.HelmRepoUpdateRequest{
				Url: "http://169.254.169.254/latest/meta-data",
			}

			res, err := helmReposApiService.UpdateHelmRepo(ctx, testHelmRepoId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo URL"))
		})
	})

	Context("delete helm repo", func() {
		var (
			testHelmRepoId  string
			testHelmRepoId2 string
		)

		BeforeEach(func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-helm-repo-delete-" + uuid.NewString()[:8],
				Url:       "https://charts.delete1-" + uuid.NewString()[:8] + ".example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testHelmRepoId = res.Body.(api.HelmRepo).Id
			helmRepoIds[testHelmRepoId] = TestTenant

			request.Name = "test-helm-repo-delete-2-" + uuid.NewString()[:8]
			request.Url = "https://charts.delete2-" + uuid.NewString()[:8] + ".example.com"
			res, err = helmReposApiService.CreateHelmRepo(ctx, request)
			Expect(err).To(BeNil(), "Create helm repo should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testHelmRepoId2 = res.Body.(api.HelmRepo).Id
			helmRepoIds[testHelmRepoId2] = TestTenant
		})

		It("delete sanity", func() {
			res, err := helmReposApiService.DeleteHelmRepo(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))
			delete(helmRepoIds, testHelmRepoId)

			res, err = helmReposApiService.GetHelmRepoById(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			res, err = helmReposApiService.GetHelmRepoById(ctx, testHelmRepoId2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete with invalid helm repo id", func() {
			res, err := helmReposApiService.DeleteHelmRepo(ctx, "invalid-uuid")
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("delete with non-existent helm repo id", func() {
			nonExistentId := uuid.NewString()
			res, err := helmReposApiService.DeleteHelmRepo(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("delete with different tenant - should not find", func() {
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err := helmReposApiService.DeleteHelmRepo(ctx2, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			res, err = helmReposApiService.GetHelmRepoById(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			res, err := helmReposApiService.DeleteHelmRepo(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))

			authorization.SetMock(&SimpleMockFactory{})
			res, err = helmReposApiService.GetHelmRepoById(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete twice should fail on second attempt", func() {
			res, err := helmReposApiService.DeleteHelmRepo(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))
			delete(helmRepoIds, testHelmRepoId)

			res, err = helmReposApiService.DeleteHelmRepo(ctx, testHelmRepoId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})
	})
})

func validateHelmRepo(actual api.HelmRepo, requested api.HelmRepoCreationRequest) {
	_, err := uuid.Parse(actual.Id)
	Expect(err).To(BeNil(), "Helm repo ID should be a valid UUID")
	Expect(actual.CreatedAt).ToNot(BeZero(), "CreatedAt should not be zero")
	Expect(actual.UpdatedAt).ToNot(BeZero(), "UpdatedAt should not be zero")
	Expect(actual.Name).To(Equal(requested.Name), "Name")
	Expect(actual.Url).To(Equal(requested.Url), "Url")
	Expect(actual.IsPrivate).To(Equal(requested.IsPrivate), "IsPrivate")
	Expect(actual.ScopeType).To(Equal(requested.ScopeType), "ScopeType")
	Expect(actual.CreatedBy).To(Equal(TestUser), "CreatedBy")
	Expect(actual.UpdatedBy).To(Equal(TestUser), "UpdatedBy")
}
