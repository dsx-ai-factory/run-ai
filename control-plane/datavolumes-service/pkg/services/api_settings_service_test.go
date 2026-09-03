/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/settings"
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

	Describe("GetDatavolumesSettings", func() {
		It("should return all settings successfully", func() {
			mockSettingsList := []settingPkgManager.Settinger{
				&settings.EnableDataVolumes{Enabled: true},
			}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.DatavolumesCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetDatavolumesSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody, ok := response.Body.(*api.DatavolumesSettingsGetResponse)
			Expect(ok).To(BeTrue())
			Expect(responseBody.EnableDataVolumes).NotTo(BeNil())
			Expect(*responseBody.EnableDataVolumes.Value.Enabled).To(BeTrue())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetDatavolumesSettings(emptyCtx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.DatavolumesCategory).
				Return(([]settingPkgManager.Settinger)(nil), fmt.Errorf("db error"))

			response, err := testSubject.GetDatavolumesSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		Context("with explicit tenant id", func() {
			It("should allow management tenant to query another tenant", func() {
				otherTenant := "other-tenant"
				mgmtCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

				mockSettingsList := []settingPkgManager.Settinger{
					&settings.EnableDataVolumes{Enabled: false},
				}
				settingPkgMock.On("GetAll", mock.Anything, otherTenant, settings.DatavolumesCategory).Return(mockSettingsList, nil)

				response, err := testSubject.GetDatavolumesSettings(mgmtCtx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				settingPkgMock.AssertExpectations(GinkgoT())
			})

			It("should return 401 when non-management tenant queries another tenant", func() {
				otherTenant := "other-tenant"

				response, err := testSubject.GetDatavolumesSettings(ctx, otherTenant)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetDatavolumesSettingByKey", func() {
		It("should return setting by key successfully", func() {
			mockSetting := &settings.EnableDataVolumes{Enabled: true}
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES)).
				Return(mockSetting, nil)

			response, err := testSubject.GetDatavolumesSettingByKey(ctx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetDatavolumesSettingByKey(emptyCtx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			invalidKey := api.DatavolumesSettingsKeyEnum("invalid-key")

			response, err := testSubject.GetDatavolumesSettingByKey(ctx, invalidKey, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES)).
				Return(nil, fmt.Errorf("db connection error"))

			response, err := testSubject.GetDatavolumesSettingByKey(ctx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})
	})

	Describe("UpdateDatavolumesSettingByKey", func() {
		It("should update setting successfully", func() {
			updateRequest := api.UpdateDatavolumesSettingByKeyRequest{Enabled: ptr.To(true)}
			expectedValue := settings.EnableDataVolumes{Enabled: true}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES), true, expectedValue).Return(nil)

			response, err := testSubject.UpdateDatavolumesSettingByKey(ctx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()
			updateRequest := api.UpdateDatavolumesSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateDatavolumesSettingByKey(emptyCtx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails to update", func() {
			updateRequest := api.UpdateDatavolumesSettingByKeyRequest{Enabled: ptr.To(false)}
			expectedValue := settings.EnableDataVolumes{Enabled: false}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES), true, expectedValue).
				Return(fmt.Errorf("db write error"))

			response, err := testSubject.UpdateDatavolumesSettingByKey(ctx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusInternalServerError))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant updates another tenant", func() {
			otherTenant := "other-tenant"
			updateRequest := api.UpdateDatavolumesSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateDatavolumesSettingByKey(ctx, api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, updateRequest, otherTenant)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})
