/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	settingPkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgManager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Describe("GetAuthorizationSettings", func() {
		It("should return all settings successfully", func() {
			mockSettingsList := []settingPkgManager.Settinger{
				&settings.EnableScopedAuthorization{Enabled: true},
			}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.AuthorizationCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetAuthorizationSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody, ok := response.Body.(*api.AuthorizationSettingsGetResponse)
			Expect(ok).To(BeTrue())
			Expect(responseBody.EnableScopedAuthorization.Enabled).NotTo(BeNil())
			Expect(*responseBody.EnableScopedAuthorization.Enabled).To(BeTrue())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetAuthorizationSettings(emptyCtx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.AuthorizationCategory).
				Return(([]settingPkgManager.Settinger)(nil), fmt.Errorf("db error"))

			response, err := testSubject.GetAuthorizationSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		Context("with explicit tenant id", func() {
			It("should allow management tenant to query another tenant", func() {
				otherTenant := "other-tenant"
				mgmtCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

				mockSettingsList := []settingPkgManager.Settinger{
					&settings.EnableScopedAuthorization{Enabled: false},
				}
				settingPkgMock.On("GetAll", mock.Anything, otherTenant, settings.AuthorizationCategory).Return(mockSettingsList, nil)

				response, err := testSubject.GetAuthorizationSettings(mgmtCtx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				settingPkgMock.AssertExpectations(GinkgoT())
			})

			It("should return 401 when non-management tenant queries another tenant", func() {
				otherTenant := "other-tenant"

				response, err := testSubject.GetAuthorizationSettings(ctx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetAuthorizationSettingByKey", func() {
		It("should return EnableScopedAuthorization successfully", func() {
			mockSetting := &settings.EnableScopedAuthorization{Enabled: true}
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION)).
				Return(mockSetting, nil)

			response, err := testSubject.GetAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetAuthorizationSettingByKey(emptyCtx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			invalidKey := api.AuthorizationSettingsKeyEnum("invalid-key")

			response, err := testSubject.GetAuthorizationSettingByKey(ctx, invalidKey, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION)).
				Return(nil, fmt.Errorf("db connection error"))

			response, err := testSubject.GetAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})
	})

	Describe("UpdateAuthorizationSettingByKey", func() {
		It("should update EnableScopedAuthorization successfully", func() {
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{Enabled: ptr.To(true)}
			expectedValue := settings.EnableScopedAuthorization{Enabled: true}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION), true, expectedValue).Return(nil)

			response, err := testSubject.UpdateAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when enabled is missing from request", func() {
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{}

			response, err := testSubject.UpdateAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateAuthorizationSettingByKey(emptyCtx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{Enabled: ptr.To(true)}
			invalidKey := api.AuthorizationSettingsKeyEnum("invalid-key")

			response, err := testSubject.UpdateAuthorizationSettingByKey(ctx, invalidKey, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails to update", func() {
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{Enabled: ptr.To(false)}
			expectedValue := settings.EnableScopedAuthorization{Enabled: false}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION), true, expectedValue).
				Return(fmt.Errorf("db write error"))

			response, err := testSubject.UpdateAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant updates another tenant", func() {
			otherTenant := "other-tenant"
			updateRequest := api.UpdateAuthorizationSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateAuthorizationSettingByKey(ctx, api.AUTHORIZATIONSETTINGSKEYENUM_ENABLE_SCOPED_AUTHORIZATION, updateRequest, otherTenant)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
