/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	settingPkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgManager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("ClusterSettingsService", func() {
	const settingsTestTenantId = "test-tenant-id"

	var (
		settingPkgMock *settingPkgMocks.SettingsManager
		testSubject    api.SettingsAPIServicer
		ctx            context.Context
	)

	BeforeEach(func() {
		settingPkgMock = settingPkgMocks.NewSettingsManager(GinkgoT())
		settingPkg.SetMock(settingPkgMock)

		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, settingsTestTenantId)

		testSubject = services.NewSettingsService(org_unit_service_client_mock.NewIOrgUnitClient(GinkgoT()))
	})

	Describe("GetClusterSettings", func() {
		It("should return all settings successfully", func() {
			settingValue := &settings.EnableAdvancedGPUMetrics{Enabled: true}
			mockSettingsList := []settingPkgManager.Settinger{settingValue}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.ClusterCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetClusterSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody, ok := response.Body.(*api.ClusterSettingsGetResponse)
			Expect(ok).To(BeTrue())
			Expect(responseBody.EnableAdvancedGpuMetrics.Enabled).NotTo(BeNil())
			Expect(*responseBody.EnableAdvancedGpuMetrics.Enabled).To(BeTrue())
			Expect(responseBody.EnableAdvancedGpuMetrics.DisplayName).To(Equal(settingValue.DisplayName()))
			Expect(responseBody.EnableAdvancedGpuMetrics.Description).To(Equal(settingValue.Description()))
			Expect(responseBody.EnableAdvancedGpuMetrics.Stage).To(Equal(string(settingValue.Stage())))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when tenant id is missing from context", func() {
			emptyCtx := context.Background()

			response, err := testSubject.GetClusterSettings(emptyCtx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.ClusterCategory).
				Return(([]settingPkgManager.Settinger)(nil), fmt.Errorf("db error"))

			response, err := testSubject.GetClusterSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should allow management tenant to query another tenant", func() {
			otherTenant := "other-tenant"
			mgmtCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

			mockSettingsList := []settingPkgManager.Settinger{
				&settings.EnableAdvancedGPUMetrics{Enabled: false},
			}
			settingPkgMock.On("GetAll", mock.Anything, otherTenant, settings.ClusterCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetClusterSettings(mgmtCtx, otherTenant)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant queries another tenant", func() {
			response, err := testSubject.GetClusterSettings(ctx, "other-tenant")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetClusterSettingByKey", func() {
		It("should return EnableAdvancedGpuMetrics successfully", func() {
			mockSetting := &settings.EnableAdvancedGPUMetrics{Enabled: true}
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS)).
				Return(mockSetting, nil)

			response, err := testSubject.GetClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 for unknown setting key", func() {
			invalidKey := api.ClusterSettingsKeyEnum("invalid-key")

			response, err := testSubject.GetClusterSettingByKey(ctx, invalidKey, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS)).
				Return(nil, fmt.Errorf("db connection error"))

			response, err := testSubject.GetClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})
	})

	Describe("UpdateClusterSettingByKey", func() {
		It("should update EnableAdvancedGpuMetrics successfully", func() {
			updateRequest := api.UpdateClusterSettingByKeyRequest{Enabled: ptr.To(true)}
			expectedValue := settings.EnableAdvancedGPUMetrics{Enabled: true}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS), true, expectedValue).Return(nil)

			response, err := testSubject.UpdateClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when enabled is missing from request", func() {
			updateRequest := api.UpdateClusterSettingByKeyRequest{}

			response, err := testSubject.UpdateClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			updateRequest := api.UpdateClusterSettingByKeyRequest{Enabled: ptr.To(true)}
			invalidKey := api.ClusterSettingsKeyEnum("invalid-key")

			response, err := testSubject.UpdateClusterSettingByKey(ctx, invalidKey, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 500 when setting package fails to update", func() {
			updateRequest := api.UpdateClusterSettingByKeyRequest{Enabled: ptr.To(false)}
			expectedValue := settings.EnableAdvancedGPUMetrics{Enabled: false}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS), true, expectedValue).
				Return(fmt.Errorf("db write error"))

			response, err := testSubject.UpdateClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant updates another tenant", func() {
			updateRequest := api.UpdateClusterSettingByKeyRequest{Enabled: ptr.To(true)}

			response, err := testSubject.UpdateClusterSettingByKey(ctx, api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS, updateRequest, "other-tenant")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("NodePoolSettingsService", func() {
	// numeric so UpdateNodePoolSettingByKey can parse it and propagate to (non-existent) tenant nodepools
	const settingsTestTenantId = "999999"

	var (
		settingPkgMock    *settingPkgMocks.SettingsManager
		orgUnitClientMock *org_unit_service_client_mock.IOrgUnitClient
		testSubject       api.SettingsAPIServicer
		ctx               context.Context
	)

	BeforeEach(func() {
		settingPkgMock = settingPkgMocks.NewSettingsManager(GinkgoT())
		settingPkg.SetMock(settingPkgMock)

		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, settingsTestTenantId)

		orgUnitClientMock = org_unit_service_client_mock.NewIOrgUnitClient(GinkgoT())
		testSubject = services.NewSettingsService(orgUnitClientMock)
	})

	Describe("GetNodePoolSettings", func() {
		It("should return all node pool settings successfully", func() {
			pmoq := &settings.ProjectManualOverQuota{Enabled: true}
			cpu := &settings.CPUResourcesQuota{Enabled: false}
			mockSettingsList := []settingPkgManager.Settinger{pmoq, cpu}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.NodePoolCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetNodePoolSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody, ok := response.Body.(*api.NodePoolSettingsGetResponse)
			Expect(ok).To(BeTrue())
			Expect(responseBody.ProjectManualOverQuota.Enabled).To(Equal(ptr.To(true)))
			Expect(responseBody.ProjectManualOverQuota.DisplayName).To(Equal(pmoq.DisplayName()))
			Expect(responseBody.ProjectManualOverQuota.Description).To(Equal(pmoq.Description()))
			Expect(responseBody.ProjectManualOverQuota.Stage).To(Equal(string(pmoq.Stage())))
			Expect(responseBody.CpuResourcesQuota.Enabled).To(Equal(ptr.To(false)))
			Expect(responseBody.CpuResourcesQuota.DisplayName).To(Equal(cpu.DisplayName()))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should request only the node-pools category (isolation from cluster settings)", func() {
			// Even if the package were to hand back a cluster-category setting, the node-pool
			// response has no field for it, so it cannot leak into the node-pool response.
			mockSettingsList := []settingPkgManager.Settinger{
				&settings.ProjectManualOverQuota{Enabled: true},
				&settings.EnableAdvancedGPUMetrics{Enabled: true},
			}
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.NodePoolCategory).Return(mockSettingsList, nil)

			response, err := testSubject.GetNodePoolSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			responseBody := response.Body.(*api.NodePoolSettingsGetResponse)
			Expect(responseBody.ProjectManualOverQuota.Enabled).To(Equal(ptr.To(true)))
			// cpuResourcesQuota was not returned by GetAll, so it stays zero-valued (no leak of the cluster setting)
			Expect(responseBody.CpuResourcesQuota.Enabled).To(BeNil())

			// GetAll must have been called with the node-pools category, never the clusters category.
			settingPkgMock.AssertCalled(GinkgoT(), "GetAll", mock.Anything, settingsTestTenantId, settings.NodePoolCategory)
			settingPkgMock.AssertNotCalled(GinkgoT(), "GetAll", mock.Anything, settingsTestTenantId, settings.ClusterCategory)
		})

		It("should return 400 when tenant id is missing from context", func() {
			response, err := testSubject.GetNodePoolSettings(context.Background(), "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 503 when setting package fails", func() {
			settingPkgMock.On("GetAll", mock.Anything, settingsTestTenantId, settings.NodePoolCategory).
				Return(([]settingPkgManager.Settinger)(nil), fmt.Errorf("db error"))

			response, err := testSubject.GetNodePoolSettings(ctx, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant queries another tenant", func() {
			response, err := testSubject.GetNodePoolSettings(ctx, "other-tenant")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetNodePoolSettingByKey", func() {
		It("should return projectManualOverQuota with its value", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)).
				Return(&settings.ProjectManualOverQuota{Enabled: true}, nil)

			response, err := testSubject.GetNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			setting, ok := response.Body.(*settings.ProjectManualOverQuota)
			Expect(ok).To(BeTrue())
			Expect(setting.Enabled).To(BeTrue())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return cpuResourcesQuota with its value", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
				Return(&settings.CPUResourcesQuota{Enabled: false}, nil)

			response, err := testSubject.GetNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			setting, ok := response.Body.(*settings.CPUResourcesQuota)
			Expect(ok).To(BeTrue())
			Expect(setting.Enabled).To(BeFalse())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 for unknown setting key", func() {
			response, err := testSubject.GetNodePoolSettingByKey(ctx, api.NodePoolSettingsKeyEnum("invalid-key"), "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 503 when setting package fails", func() {
			settingPkgMock.On("Get", mock.Anything, settingsTestTenantId, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
				Return(nil, fmt.Errorf("db connection error"))

			response, err := testSubject.GetNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})
	})

	Describe("UpdateNodePoolSettingByKey", func() {
		It("should update cpuResourcesQuota and return the new value", func() {
			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA), true, settings.CPUResourcesQuota{Enabled: true}).Return(nil)

			// propagation runs against the real (test) DB; tenant 999999 has no nodepools so it's a no-op
			resp, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(*resp.Body.(api.UpdateNodePoolSettingByKeyRequest).Enabled).To(BeTrue())

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 400 when enabled is missing from request", func() {
			response, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, api.UpdateNodePoolSettingByKeyRequest{}, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 400 for unknown setting key", func() {
			updateRequest := api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}
			response, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NodePoolSettingsKeyEnum("invalid-key"), updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return 503 when setting package fails to update", func() {
			updateRequest := api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}
			expectedValue := settings.ProjectManualOverQuota{Enabled: false}

			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA), true, expectedValue).
				Return(fmt.Errorf("db write error"))

			response, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA, updateRequest, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusServiceUnavailable))

			settingPkgMock.AssertExpectations(GinkgoT())
		})

		It("should return 401 when non-management tenant updates another tenant", func() {
			updateRequest := api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}
			response, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, updateRequest, "other-tenant")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should trigger org-unit DisableCpuQuota when cpuResourcesQuota is disabled", func() {
			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA), true, settings.CPUResourcesQuota{Enabled: false}).Return(nil)
			// org-unit is called with the resolved tenant in context and the cpu_resources flag set to "false"
			orgUnitClientMock.On("UpdateOrgUnitSettings",
				mock.MatchedBy(func(c context.Context) bool {
					tid := runaictx.GetString(c, runaictx.CfTenantId)
					return tid != nil && *tid == settingsTestTenantId
				}),
				"project.quota.cpu_resources", "false").Return(nil)

			resp, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			settingPkgMock.AssertExpectations(GinkgoT())
			orgUnitClientMock.AssertExpectations(GinkgoT())
		})

		It("should NOT trigger org-unit when cpuResourcesQuota is enabled", func() {
			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA), true, settings.CPUResourcesQuota{Enabled: true}).Return(nil)

			resp, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			orgUnitClientMock.AssertNotCalled(GinkgoT(), "UpdateOrgUnitSettings", mock.Anything, mock.Anything, mock.Anything)
		})

		It("should NOT trigger org-unit when disabling projectManualOverQuota", func() {
			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA), true, settings.ProjectManualOverQuota{Enabled: false}).Return(nil)

			resp, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			orgUnitClientMock.AssertNotCalled(GinkgoT(), "UpdateOrgUnitSettings", mock.Anything, mock.Anything, mock.Anything)
		})

		It("should return 503 when org-unit DisableCpuQuota fails", func() {
			settingPkgMock.On("Update", mock.Anything, settingsTestTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA), true, settings.CPUResourcesQuota{Enabled: false}).Return(nil)
			orgUnitClientMock.On("UpdateOrgUnitSettings", mock.Anything, "project.quota.cpu_resources", "false").
				Return(fmt.Errorf("org-unit unavailable"))

			resp, err := testSubject.UpdateNodePoolSettingByKey(ctx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable))

			orgUnitClientMock.AssertExpectations(GinkgoT())
		})

		It("should pass the resolved (target) tenant to org-unit for a management cross-tenant disable", func() {
			const targetTenantId = "888888"
			mgmtCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
			settingPkgMock.On("Update", mock.Anything, targetTenantId,
				string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA), true, settings.CPUResourcesQuota{Enabled: false}).Return(nil)
			// org-unit must wipe the RESOLVED tenant's data, not the management caller's
			orgUnitClientMock.On("UpdateOrgUnitSettings",
				mock.MatchedBy(func(c context.Context) bool {
					tid := runaictx.GetString(c, runaictx.CfTenantId)
					return tid != nil && *tid == targetTenantId
				}),
				"project.quota.cpu_resources", "false").Return(nil)

			resp, err := testSubject.UpdateNodePoolSettingByKey(mgmtCtx, api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA,
				api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}, targetTenantId)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			orgUnitClientMock.AssertExpectations(GinkgoT())
		})
	})
})
