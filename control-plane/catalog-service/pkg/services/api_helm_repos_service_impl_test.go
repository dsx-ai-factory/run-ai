/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testTenantId  = "1"
	testSubjectId = "user@test.ai"
)

func helmRepoTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, testTenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, testSubjectId)
	return ctx
}

type fakeHelmAssetsClient struct {
	key string
}

func (c *fakeHelmAssetsClient) GetNgcApiKeyByIdInternal(_ context.Context, _ string) (*assetsApi.NgcApiKeyInternal, *assetsApi.Error) {
	return &assetsApi.NgcApiKeyInternal{Key: c.key}, nil
}

var _ = Describe("HelmReposServiceImpl", func() {
	var service *helmReposAPIService

	BeforeEach(func() {
		service = NewHelmReposAPIService()
	})

	Describe("createHelmRepo scope validation", func() {
		It("should reject system scope when not migrating", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_SYSTEM,
			}

			_, apiErr := service.createHelmRepo(helmRepoTestContext(), request, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("system scope is not supported"))
		})

		It("should reject project scope", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_PROJECT,
				ScopeId:   "proj-1",
			}

			_, apiErr := service.createHelmRepo(helmRepoTestContext(), request, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported"))
		})

		It("should reject cluster scope", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_CLUSTER,
				ScopeId:   "cluster-1",
			}

			_, apiErr := service.createHelmRepo(helmRepoTestContext(), request, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported"))
		})

		It("should reject department scope", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_DEPARTMENT,
				ScopeId:   "dept-1",
			}

			_, apiErr := service.createHelmRepo(helmRepoTestContext(), request, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported"))
		})

		It("should reject project scope even when migrating", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_PROJECT,
				ScopeId:   "proj-1",
			}

			_, apiErr := service.createHelmRepo(helmRepoTestContext(), request, true)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("non-tenant scope is not supported"))
		})
	})

	Describe("validateHelmRepoURL", func() {
		It("should accept valid https URL", func() {
			Expect(validateHelmRepoURL("https://charts.example.com")).To(BeNil())
		})

		It("should accept valid http URL", func() {
			Expect(validateHelmRepoURL("http://charts.example.com")).To(BeNil())
		})

		It("should accept https URL with path and port", func() {
			Expect(validateHelmRepoURL("https://charts.example.com:8443/repo")).To(BeNil())
		})

		It("should reject ftp scheme", func() {
			err := validateHelmRepoURL("ftp://charts.example.com")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported scheme"))
		})

		It("should reject file scheme", func() {
			err := validateHelmRepoURL("file:///etc/passwd")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported scheme"))
		})

		It("should reject URL without scheme", func() {
			err := validateHelmRepoURL("charts.example.com")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unsupported scheme"))
		})

		It("should reject empty URL", func() {
			err := validateHelmRepoURL("")
			Expect(err).NotTo(BeNil())
		})

		It("should reject loopback IPv4", func() {
			err := validateHelmRepoURL("https://127.0.0.1")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should reject private 10.x.x.x range", func() {
			err := validateHelmRepoURL("https://10.0.0.1/repo")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should reject private 172.16.x.x range", func() {
			err := validateHelmRepoURL("https://172.16.0.1")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should reject private 192.168.x.x range", func() {
			err := validateHelmRepoURL("https://192.168.1.1")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should reject link-local/cloud metadata address", func() {
			err := validateHelmRepoURL("http://169.254.169.254/latest/meta-data")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should reject IPv6 loopback", func() {
			err := validateHelmRepoURL("https://[::1]/repo")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("private/internal address"))
		})

		It("should allow public IP address", func() {
			Expect(validateHelmRepoURL("https://8.8.8.8/repo")).To(BeNil())
		})
	})

	Describe("updateHelmRepo input validation", func() {
		It("should reject invalid UUID", func() {
			request := api.HelmRepoUpdateRequest{Name: "new-name"}

			_, apiErr := service.updateHelmRepo(helmRepoTestContext(), "not-a-uuid", request)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo id"))
		})

		It("should reject empty string as UUID", func() {
			request := api.HelmRepoUpdateRequest{Name: "new-name"}

			_, apiErr := service.updateHelmRepo(helmRepoTestContext(), "", request)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})
	})

	Describe("deleteHelmRepo input validation", func() {
		It("should reject invalid UUID", func() {
			apiErr := service.deleteHelmRepo(helmRepoTestContext(), "not-a-uuid")
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(apiErr.Message).To(ContainSubstring("invalid helm repo id"))
		})

		It("should reject empty string as UUID", func() {
			apiErr := service.deleteHelmRepo(helmRepoTestContext(), "")
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))
		})
	})

	Describe("apiToModelHelmRepo", func() {
		It("should convert a valid tenant-scoped request", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:        "my-repo",
				Url:         "https://charts.example.com",
				IsPrivate:   true,
				ScopeType:   api.SCOPETYPE_TENANT,
				ScopeId:     "original-scope",
				Description: "A test description",
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			Expect(err).To(BeNil())
			Expect(model).NotTo(BeNil())
			Expect(model.Name).To(Equal("my-repo"))
			Expect(model.Url).To(Equal("https://charts.example.com"))
			Expect(model.ScopeType).To(Equal(string(api.SCOPETYPE_TENANT)))
			Expect(model.ScopeId).To(Equal(testTenantId))
			Expect(model.TenantId).NotTo(BeNil())
			Expect(*model.TenantId).To(Equal(testTenantId))
			Expect(model.IsPrivate).To(BeTrue())
			Expect(model.CreatedBy).To(Equal(testSubjectId))
			Expect(model.UpdatedBy).To(Equal(testSubjectId))
			Expect(model.ID).NotTo(Equal(uuid.Nil))
		})

		It("should set description pointer when non-empty", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:        "my-repo",
				Url:         "https://charts.example.com",
				ScopeType:   api.SCOPETYPE_TENANT,
				Description: "some description",
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			Expect(err).To(BeNil())
			Expect(model.Description).NotTo(BeNil())
			Expect(*model.Description).To(Equal("some description"))
		})

		It("should set description to nil when empty", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			Expect(err).To(BeNil())
			Expect(model.Description).To(BeNil())
		})

		It("should fail when subject ID is missing from context", func() {
			ctx := context.Background()
			ctx = context.WithValue(ctx, runaictx.CfTenantId, testTenantId)

			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("subject id is missing"))
			Expect(model).To(BeNil())
		})

		It("should fail when context is empty", func() {
			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
			}

			model, err := apiToModelHelmRepo(context.Background(), request, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("subject id is missing"))
			Expect(model).To(BeNil())
		})

		It("should override ScopeId with tenant ID for tenant scope", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   "should-be-overridden",
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			Expect(err).To(BeNil())
			Expect(model.ScopeId).To(Equal(testTenantId))
		})

		It("should set timestamps close to current time", func() {
			ctx := helmRepoTestContext()
			before := time.Now()

			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
			}

			model, err := apiToModelHelmRepo(ctx, request, false)
			after := time.Now()

			Expect(err).To(BeNil())
			Expect(model.CreatedAt).To(BeTemporally(">=", before))
			Expect(model.CreatedAt).To(BeTemporally("<=", after))
			Expect(model.UpdatedAt).To(BeTemporally(">=", before))
			Expect(model.UpdatedAt).To(BeTemporally("<=", after))
		})

		It("should generate unique UUIDs across calls", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:      "my-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
			}

			model1, err1 := apiToModelHelmRepo(ctx, request, false)
			model2, err2 := apiToModelHelmRepo(ctx, request, false)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(model1.ID).NotTo(Equal(model2.ID))
		})
	})

	Describe("apiToModelHelmRepo migrating mode", func() {
		It("should not require subject ID when migrating", func() {
			ctx := context.Background()
			request := api.HelmRepoCreationRequest{
				Name:      "migrated-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_SYSTEM,
			}

			model, err := apiToModelHelmRepo(ctx, request, true)
			Expect(err).To(BeNil())
			Expect(model).NotTo(BeNil())
			Expect(model.CreatedBy).To(Equal("unknown"))
			Expect(model.UpdatedBy).To(Equal("unknown"))
		})

		It("should set tenant ID to nil when migrating", func() {
			ctx := context.Background()
			request := api.HelmRepoCreationRequest{
				Name:      "migrated-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_SYSTEM,
			}

			model, err := apiToModelHelmRepo(ctx, request, true)
			Expect(err).To(BeNil())
			Expect(model.TenantId).To(BeNil())
		})

		It("should not override ScopeId when migrating with tenant scope and no tenant context", func() {
			ctx := context.Background()
			request := api.HelmRepoCreationRequest{
				Name:      "migrated-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_TENANT,
				ScopeId:   "original-scope-id",
			}

			model, err := apiToModelHelmRepo(ctx, request, true)
			Expect(err).To(BeNil())
			Expect(model.ScopeId).To(Equal("original-scope-id"))
		})

		It("should use subject ID from context when available even in migrating mode", func() {
			ctx := helmRepoTestContext()
			request := api.HelmRepoCreationRequest{
				Name:      "migrated-repo",
				Url:       "https://charts.example.com",
				ScopeType: api.SCOPETYPE_SYSTEM,
			}

			model, err := apiToModelHelmRepo(ctx, request, true)
			Expect(err).To(BeNil())
			Expect(model.CreatedBy).To(Equal(testSubjectId))
			Expect(model.UpdatedBy).To(Equal(testSubjectId))
		})
	})

	Describe("modelToApiHelmRepo", func() {
		It("should convert a model with all fields", func() {
			ctx := helmRepoTestContext()
			tenantIdStr := "1"
			now := time.Now()
			modelId := uuid.New()

			model := models.HelmRepo{
				ID:        modelId,
				Name:      "test-repo",
				Url:       "https://charts.example.com",
				IsPrivate: true,
				ScopeType: string(api.SCOPETYPE_TENANT),
				ScopeId:   "1",
				TenantId:  &tenantIdStr,
				CreatedBy: "creator@test.ai",
				UpdatedBy: "updater@test.ai",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
			}

			result := modelToApiHelmRepo(ctx, model)
			Expect(result.Id).To(Equal(modelId.String()))
			Expect(result.Name).To(Equal("test-repo"))
			Expect(result.Url).To(Equal("https://charts.example.com"))
			Expect(result.IsPrivate).To(BeTrue())
			Expect(result.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
			Expect(result.ScopeId).To(Equal("1"))
			Expect(result.TenantId).To(Equal(int32(1)))
			Expect(result.CreatedBy).To(Equal("creator@test.ai"))
			Expect(result.UpdatedBy).To(Equal("updater@test.ai"))
			Expect(result.CreatedAt).To(Equal(now))
			Expect(result.UpdatedAt).To(Equal(now.Add(time.Hour)))
		})

		It("should use context tenant ID when model TenantId is nil", func() {
			ctx := helmRepoTestContext()
			model := models.HelmRepo{
				ID:        uuid.New(),
				Name:      "system-repo",
				Url:       "https://system.charts.example.com",
				ScopeType: string(api.SCOPETYPE_SYSTEM),
				TenantId:  nil,
				CreatedBy: "system",
				UpdatedBy: "system",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			result := modelToApiHelmRepo(ctx, model)
			Expect(result.TenantId).To(Equal(int32(1)))
		})

		It("should parse numeric tenant ID string correctly", func() {
			ctx := helmRepoTestContext()
			tenantIdStr := "42"
			model := models.HelmRepo{
				ID:        uuid.New(),
				Name:      "repo",
				Url:       "https://charts.example.com",
				ScopeType: string(api.SCOPETYPE_TENANT),
				ScopeId:   "42",
				TenantId:  &tenantIdStr,
				CreatedBy: testSubjectId,
				UpdatedBy: testSubjectId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			result := modelToApiHelmRepo(ctx, model)
			Expect(result.TenantId).To(Equal(int32(42)))
		})

		It("should preserve timestamps exactly", func() {
			ctx := helmRepoTestContext()
			tenantIdStr := "1"
			createdAt := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
			updatedAt := time.Date(2025, 6, 15, 14, 0, 0, 0, time.UTC)

			model := models.HelmRepo{
				ID:        uuid.New(),
				Name:      "repo",
				Url:       "https://charts.example.com",
				ScopeType: string(api.SCOPETYPE_TENANT),
				ScopeId:   "1",
				TenantId:  &tenantIdStr,
				CreatedBy: testSubjectId,
				UpdatedBy: testSubjectId,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			result := modelToApiHelmRepo(ctx, model)
			Expect(result.CreatedAt).To(Equal(createdAt))
			Expect(result.UpdatedAt).To(Equal(updatedAt))
		})

		It("should convert scope type string back to API enum", func() {
			ctx := helmRepoTestContext()
			tenantIdStr := "1"
			model := models.HelmRepo{
				ID:        uuid.New(),
				Name:      "repo",
				Url:       "https://charts.example.com",
				ScopeType: "tenant",
				ScopeId:   "1",
				TenantId:  &tenantIdStr,
				CreatedBy: testSubjectId,
				UpdatedBy: testSubjectId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			result := modelToApiHelmRepo(ctx, model)
			Expect(result.ScopeType).To(Equal(api.SCOPETYPE_TENANT))
		})
	})

	Describe("helm repo browse auth helpers", func() {
		It("should not require an API key when assetId is empty", func() {
			apiKey, apiErr := service.getHelmRepoAPIKey(helmRepoTestContext(), "")
			Expect(apiErr).To(BeNil())
			Expect(apiKey).To(BeEmpty())
		})

		It("should get an API key when assetId is provided", func() {
			service = NewHelmReposAPIService(&fakeHelmAssetsClient{key: "secret-key"})

			apiKey, apiErr := service.getHelmRepoAPIKey(helmRepoTestContext(), "asset-id")
			Expect(apiErr).To(BeNil())
			Expect(apiKey).To(Equal("secret-key"))
		})

		It("should scope authenticated cache keys by assetId hash", func() {
			publicKey := chartsCacheKey("repo-id", "")
			privateKey1 := chartsCacheKey("repo-id", "asset-1")
			privateKey2 := chartsCacheKey("repo-id", "asset-2")

			Expect(publicKey).To(Equal("helm_repo_charts_repo-id"))
			Expect(privateKey1).NotTo(Equal(publicKey))
			Expect(privateKey1).NotTo(Equal(privateKey2))
			Expect(privateKey1).NotTo(ContainSubstring("asset-1"))
		})
	})

	Describe("helm repo presets", func() {
		It("should include public and private NVIDIA NGC repos", func() {
			expected := map[string]bool{
				"NVIDIA":             false,
				"Nvidia Blueprints":  false,
				"Clara":              false,
				"DOCA":               false,
				"Omniverse":          false,
				"TAO":                false,
				"Nemo Microservices": true,
				"Holoscan for Media": true,
				"Isaac":              true,
				"Morpheus":           true,
				"DeepStream":         true,
				"USD Search":         true,
				"Riva":               true,
				"NIM":                true,
				"NVAIE":              true,
			}

			Expect(helmRepoPresets).To(HaveLen(len(expected)))
			for name, isPrivate := range expected {
				preset, exists := helmRepoPresets[name]
				Expect(exists).To(BeTrue(), "missing preset %s", name)
				Expect(preset.Request.IsPrivate).To(Equal(isPrivate), "isPrivate for %s", name)
				Expect(preset.Request.Url).NotTo(BeEmpty(), "url for %s", name)
			}
		})
	})
})
