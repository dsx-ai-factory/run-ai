/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	settingPkg "github.com/run-ai/runai/runai-common-packages/setting"
	settingPkgManager "github.com/run-ai/runai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/run-ai/runai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("SettingsService", func() {
	const settingsTestTenantId = "test-tenant-id"

	var (
		settingPkgMock *settingPkgMocks.SettingsManager
		testSubject    *SettingsService
		ctx            context.Context
	)

	BeforeEach(func() {
		settingPkgMock = settingPkgMocks.NewSettingsManager(GinkgoT())
		settingPkg.SetMock(settingPkgMock)

		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, settingsTestTenantId)

		testSubject = &SettingsService{
			settingsManager: *settings.NewSettingsManager(),
		}
	})

	Describe("GetCatalogSettings", func() {
		It("should return all settings successfully", func() {
			mockSettingsList := []settingPkgManager.Settinger{
				&settings.EnableNgcIntegration{Enabled: true},
			}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.CatalogCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetCatalogSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody, ok := response.Body.(*api.CatalogSettingsGetResponse)
			Expect(ok).To(BeTrue())
			Expect(responseBody.EnableNgcIntegration).NotTo(BeNil())
			Expect(*responseBody.EnableNgcIntegration.Value.Enabled).To(BeTrue())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetCatalogSettings(emptyCtx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.CatalogCategory).
				Return(([]settingPkgManager.Settinger)(nil), fmt.Errorf("db error"))

			response, err := testSubject.GetCatalogSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		Context("with explicit tenant id", func() {
			It("should allow management tenant to query another tenant", func() {
				otherTenant := "other-tenant"
				mgmtCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

				mockSettingsList := []settingPkgManager.Settinger{
					&settings.EnableNgcIntegration{Enabled: false},
				}
				settingPkgMock.On("GetAll", mock.Anything, otherTenant, settings.CatalogCategory).Return(mockSettingsList, nil)

				response, err := testSubject.GetCatalogSettings(mgmtCtx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				settingPkgMock.AssertExpectations(GinkgoT())
			})

			It("should return 401 when non-management tenant queries another tenant", func() {
				otherTenant := "other-tenant"

				response, err := testSubject.GetCatalogSettings(ctx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetCatalogSettingByKey", func() {
		It("should return setting by key successfully", func() {
			mockSetting := &settings.EnableNgcIntegration{Enabled: true}
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION)).
				Return(mockSetting, nil)

			response, err := testSubject.GetCatalogSettingByKey(ctx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetCatalogSettingByKey(emptyCtx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			invalidKey := api.CatalogSettingsKeyEnum("invalid-key")

			response, err := testSubject.GetCatalogSettingByKey(ctx, invalidKey, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION)).
				Return(nil, fmt.Errorf("db connection error"))

			response, err := testSubject.GetCatalogSettingByKey(ctx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})
	})

	Describe("UpdateCatalogSettingByKey", func() {
		It("should update setting successfully", func() {
			updateRequest := api.UpdateCatalogSettingByKeyRequest{Enabled: ptr.To(true)}
			expectedValue := settings.EnableNgcIntegration{Enabled: true}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION), true, expectedValue).Return(nil)

			response, err := testSubject.UpdateCatalogSettingByKey(ctx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()
			updateRequest := api.UpdateCatalogSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateCatalogSettingByKey(emptyCtx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails to update", func() {
			updateRequest := api.UpdateCatalogSettingByKeyRequest{Enabled: ptr.To(false)}
			expectedValue := settings.EnableNgcIntegration{Enabled: false}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION), true, expectedValue).
				Return(fmt.Errorf("db write error"))

			response, err := testSubject.UpdateCatalogSettingByKey(ctx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant updates another tenant", func() {
			otherTenant := "other-tenant"
			updateRequest := api.UpdateCatalogSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateCatalogSettingByKey(ctx, api.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, updateRequest, otherTenant)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
