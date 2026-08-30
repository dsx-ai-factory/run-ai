package test

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/services"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var _ = Describe("Registries API tests", Ordered, func() {
	registryIds := map[string]string{} // registryId -> tenantId

	var registriesApiService api.RegistriesAPIServicer
	var ctx context.Context

	BeforeEach(func() {
		ctx = SetContext(TestTenant, TestUser)
		registriesApiService = services.NewRegistriesAPIService()
		// Reset to default mock that allows access
		authorization.SetMock(&SimpleMockFactory{})
	})

	AfterEach(func() {
		// Clean up created registries
		for registryId := range registryIds {
			cleanupCtx := SetContext(TestTenant, TestUser)
			_ = dao.DeleteRegistry(cleanupCtx, registryId)
		}
		registryIds = map[string]string{}
	})

	Context("create registry", func() {
		var (
			request api.RegistryIntegrationCreationRequest
		)

		BeforeEach(func() {
			request = api.RegistryIntegrationCreationRequest{
				Name:      "test-registry",
				Url:       "https://registry.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}
		})

		It("create sanity", func() {
			now := time.Now()
			time.Sleep(1 * time.Millisecond)

			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))

			created := res.Body.(api.RegistryIntegration)
			registryIds[created.Id] = TestTenant

			validateRegistry(created, request)
			Expect(created.CreatedAt.After(now)).To(BeTrue())
			Expect(created.UpdatedAt.After(now)).To(BeTrue())
		})

		It("create with different name", func() {
			request.Name = "another-registry"
			request.Url = "https://another.registry.com"

			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))

			created := res.Body.(api.RegistryIntegration)
			registryIds[created.Id] = TestTenant

			Expect(created.Name).To(Equal(request.Name))
			Expect(created.Url).To(Equal(request.Url))
		})

		It("create with missing subjectId should fail", func() {
			// Empty context - missing both tenant and subject
			ctx = context.Background()
			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))

			// Only tenantId set - missing subjectId
			ctx = context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)
			res, err = registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("create with conflict: multiple registries with same name", func() {
			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			created := res.Body.(api.RegistryIntegration)
			registryIds[created.Id] = TestTenant

			// Try to create another registry with the same name
			request.Url = "https://different.registry.com"
			res, err = registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry err should be nil")
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("already exists"))
		})

		It("create with project scope type should fail", func() {
			request.ScopeType = api.SCOPETYPE_PROJECT
			request.ScopeId = "1"
			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported for registry integrations"))
		})

		It("create with system scope should fail for non-migrating", func() {
			request.ScopeType = api.SCOPETYPE_SYSTEM
			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("system scope is not supported"))
		})

		It("create with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("get registry", func() {
		var (
			testRegistryId string
			request        api.RegistryIntegrationCreationRequest
		)

		BeforeEach(func() {
			request = api.RegistryIntegrationCreationRequest{
				Name:      "test-registry-get",
				Url:       "https://registry.get.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testRegistryId = res.Body.(api.RegistryIntegration).Id
			registryIds[testRegistryId] = TestTenant
		})

		It("get sanity", func() {
			res, err := registriesApiService.GetRegistryIntegrationById(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			registry := res.Body.(api.RegistryIntegration)
			validateRegistry(registry, request)
		})

		It("get with invalid registry id", func() {
			res, err := registriesApiService.GetRegistryIntegrationById(ctx, "invalid-uuid")
			Expect(err).To(BeNil())
			// Invalid UUID returns a validation error (wrapped as internal server error in DAO)
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("get with non-existent registry id", func() {
			nonExistentId := uuid.NewString()
			res, err := registriesApiService.GetRegistryIntegrationById(ctx, nonExistentId)
			Expect(err).To(BeNil())
			// Note: GetRegistryById returns 404 when registry not found
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("get with different tenant - should not find", func() {
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err := registriesApiService.GetRegistryIntegrationById(ctx2, testRegistryId)
			Expect(err).To(BeNil())
			// Registry should not be found for different tenant
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("list registries", func() {
		var (
			testRegistryIds []string
		)

		BeforeEach(func() {
			testRegistryIds = []string{}

			// Create multiple registries
			for i := 0; i < 3; i++ {
				request := api.RegistryIntegrationCreationRequest{
					Name:      "list-registry-" + uuid.NewString()[:8],
					Url:       "https://registry" + uuid.NewString()[:8] + ".example.com",
					ScopeType: api.SCOPETYPE_TENANT,
					ScopeId:   TestTenant,
				}

				res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
				Expect(err).To(BeNil(), "Create registry should succeed")
				Expect(res.Code).To(Equal(http.StatusOK))
				created := res.Body.(api.RegistryIntegration)
				testRegistryIds = append(testRegistryIds, created.Id)
				registryIds[created.Id] = TestTenant
			}
		})

		It("list sanity", func() {
			res, err := registriesApiService.ListRegistryIntegrations(ctx)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			response := res.Body.(api.RegistryIntegrationListResponse)
			Expect(len(response.Registries)).To(BeNumerically(">=", 3))

			// Verify our created registries are in the list
			foundIds := make(map[string]bool)
			for _, registry := range response.Registries {
				foundIds[registry.Id] = true
			}

			for _, id := range testRegistryIds {
				Expect(foundIds[id]).To(BeTrue(), "Registry %s should be in the list", id)
			}
		})

		It("list with different tenant - should not see other tenant registries", func() {
			ctx2 := SetContext(TestTenant2, TestUser)

			res, err := registriesApiService.ListRegistryIntegrations(ctx2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			response := res.Body.(api.RegistryIntegrationListResponse)

			// Should not find any of our created registries
			foundIds := make(map[string]bool)
			for _, registry := range response.Registries {
				foundIds[registry.Id] = true
			}

			for _, id := range testRegistryIds {
				Expect(foundIds[id]).To(BeFalse(), "Registry %s should NOT be in the list for different tenant", id)
			}
		})
	})

	Context("update registry", func() {
		var (
			testRegistryId string
			createRequest  api.RegistryIntegrationCreationRequest
		)

		BeforeEach(func() {
			createRequest = api.RegistryIntegrationCreationRequest{
				Name:      "test-registry-update",
				Url:       "https://registry.update.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := registriesApiService.CreateRegistryIntegration(ctx, createRequest)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testRegistryId = res.Body.(api.RegistryIntegration).Id
			registryIds[testRegistryId] = TestTenant
		})

		It("update name sanity", func() {
			time.Sleep(1 * time.Millisecond)
			ctx = SetContext(TestTenant, TestUser2)

			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "updated-registry-name",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.RegistryIntegration)
			Expect(updated.Name).To(Equal(updateRequest.Name))
			Expect(updated.Url).To(Equal(createRequest.Url)) // URL should remain unchanged
			Expect(updated.UpdatedBy).To(Equal(TestUser2))
			Expect(updated.UpdatedAt.After(updated.CreatedAt)).To(BeTrue())
		})

		It("update url sanity", func() {
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Url: "https://updated.registry.com",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.RegistryIntegration)
			Expect(updated.Url).To(Equal(updateRequest.Url))
			Expect(updated.Name).To(Equal(createRequest.Name)) // Name should remain unchanged
		})

		It("update both name and url", func() {
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "fully-updated-registry",
				Url:  "https://fully.updated.registry.com",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			updated := res.Body.(api.RegistryIntegration)
			Expect(updated.Name).To(Equal(updateRequest.Name))
			Expect(updated.Url).To(Equal(updateRequest.Url))
		})

		It("update with invalid registry id", func() {
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "new-name",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, "invalid-uuid", updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("update with non-existent registry id", func() {
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "new-name",
			}

			nonExistentId := uuid.NewString()
			res, err := registriesApiService.UpdateRegistryIntegration(ctx, nonExistentId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("update with conflicting name", func() {
			// First, create another registry
			anotherRequest := api.RegistryIntegrationCreationRequest{
				Name:      "another-registry-for-conflict",
				Url:       "https://another.registry.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := registriesApiService.CreateRegistryIntegration(ctx, anotherRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
			another := res.Body.(api.RegistryIntegration)
			registryIds[another.Id] = TestTenant

			// Try to update the first registry with the name of the second
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: anotherRequest.Name,
			}

			res, err = registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusConflict))
			apiErr := res.Body.(*api.Error)
			Expect(apiErr.Message).To(ContainSubstring("already exists"))
		})

		It("update with same name (no change) should succeed", func() {
			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: createRequest.Name,
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("update with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "new-name",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("update with missing subject id", func() {
			ctx = context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)

			updateRequest := api.RegistryIntegrationUpdateRequest{
				Name: "new-name",
			}

			res, err := registriesApiService.UpdateRegistryIntegration(ctx, testRegistryId, updateRequest)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("delete registry", func() {
		var (
			testRegistryId  string
			testRegistryId2 string
		)

		BeforeEach(func() {
			request := api.RegistryIntegrationCreationRequest{
				Name:      "test-registry-delete-" + uuid.NewString()[:8],
				Url:       "https://registry.delete1-" + uuid.NewString()[:8] + ".example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   TestTenant,
			}

			res, err := registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testRegistryId = res.Body.(api.RegistryIntegration).Id
			registryIds[testRegistryId] = TestTenant

			// Create another one for tenant isolation tests
			request.Name = "test-registry-delete-2-" + uuid.NewString()[:8]
			request.Url = "https://registry.delete2-" + uuid.NewString()[:8] + ".example.com"
			res, err = registriesApiService.CreateRegistryIntegration(ctx, request)
			Expect(err).To(BeNil(), "Create registry should succeed")
			Expect(res.Code).To(Equal(http.StatusOK))
			testRegistryId2 = res.Body.(api.RegistryIntegration).Id
			registryIds[testRegistryId2] = TestTenant
		})

		It("delete sanity", func() {
			res, err := registriesApiService.DeleteRegistryIntegration(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))
			delete(registryIds, testRegistryId)

			// Verify it's deleted
			res, err = registriesApiService.GetRegistryIntegrationById(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNotFound))

			// Verify the other registry still exists
			res, err = registriesApiService.GetRegistryIntegrationById(ctx, testRegistryId2)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete with invalid registry id", func() {
			res, err := registriesApiService.DeleteRegistryIntegration(ctx, "invalid-uuid")
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

		It("delete with non-existent registry id", func() {
			nonExistentId := uuid.NewString()
			res, err := registriesApiService.DeleteRegistryIntegration(ctx, nonExistentId)
			Expect(err).To(BeNil())
			// Non-existent registry returns 404
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})

		It("delete with different tenant - should not find", func() {
			ctx2 := SetContext(TestTenant2, TestUser)
			res, err := registriesApiService.DeleteRegistryIntegration(ctx2, testRegistryId)
			Expect(err).To(BeNil())
			// Registry should not be found for different tenant
			Expect(res.Code).To(Equal(http.StatusNotFound))

			// Verify it still exists in the original tenant
			res, err = registriesApiService.GetRegistryIntegrationById(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete with insufficient permissions", func() {
			authorization.SetMock(&NoPermissionsMockFactory{})

			res, err := registriesApiService.DeleteRegistryIntegration(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusForbidden))

			// Verify it still exists
			authorization.SetMock(&SimpleMockFactory{})
			res, err = registriesApiService.GetRegistryIntegrationById(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("delete twice should fail on second attempt", func() {
			res, err := registriesApiService.DeleteRegistryIntegration(ctx, testRegistryId)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusAccepted))
			delete(registryIds, testRegistryId)

			res, err = registriesApiService.DeleteRegistryIntegration(ctx, testRegistryId)
			Expect(err).To(BeNil())
			// Second delete returns 404 since registry no longer exists
			Expect(res.Code).To(Equal(http.StatusNotFound))
		})
	})
})

func validateRegistry(actual api.RegistryIntegration, requested api.RegistryIntegrationCreationRequest) {
	_, err := uuid.Parse(actual.Id)
	Expect(err).To(BeNil(), "Registry ID should be a valid UUID")
	Expect(actual.CreatedAt).ToNot(BeZero(), "CreatedAt should not be zero")
	Expect(actual.UpdatedAt).ToNot(BeZero(), "UpdatedAt should not be zero")
	Expect(actual.Name).To(Equal(requested.Name), "Name")
	Expect(actual.Url).To(Equal(requested.Url), "Url")
	Expect(actual.ScopeType).To(Equal(requested.ScopeType), "ScopeType")
	Expect(actual.CreatedBy).To(Equal(TestUser), "CreatedBy")
	Expect(actual.UpdatedBy).To(Equal(TestUser), "UpdatedBy")
}
