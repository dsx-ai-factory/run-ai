/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	cluster_config "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/mocks"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	settingPkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
)

var _ = Describe("Nodepool scheduling capability transitions", Ordered, func() {
	ctx := createContext()
	var clusterId string
	var createdNodepools []api.Nodepool
	var orgClient *org_unit_service_client_mock.IOrgUnitClient
	var nodepoolService api.NodePoolsAPIServicer

	BeforeAll(func() {
		cluster, err := dao.CreateCluster(ctx, &models.Cluster{Name: "cluster-for-np-capabilities", TenantId: tenantId, Domain: "xyz", Version: "v2.24"})
		Expect(err).To(BeNil())
		clusterId = cluster.ID.String()
		Expect(runai_config.Set(&cluster_config.EditableNodepoolCapabilities{Enabled: true})).To(Succeed())
	})

	AfterAll(func() {
		Expect(runai_config.Set(&cluster_config.EditableNodepoolCapabilities{Enabled: false})).To(Succeed())
		deleteNodepools(createdNodepools)
		_ = dao.DeleteCluster(ctx, clusterId, tenantId)
	})

	useTenantSettings := func(cpuQuota, adjustableWeights bool) {
		settingsMock := settingPkgMocks.NewSettingsManager(GinkgoT())
		settingsMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
			Return(&settings.CPUResourcesQuota{Enabled: cpuQuota}, nil).Maybe()
		settingsMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)).
			Return(&settings.ProjectManualOverQuota{Enabled: adjustableWeights}, nil).Maybe()
		settingPkg.SetMock(settingsMock)
	}

	BeforeEach(func() {
		useTenantSettings(false, false)
		tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
		tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
		orgClient = org_unit_service_client_mock.NewIOrgUnitClient(GinkgoT())
		orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		nodepoolService = services.NewNodePoolsApiService(tenantSettingsClient, orgClient)
	})

	createNodepool := func(onCluster string) api.Nodepool {
		suffix := fmt.Sprintf("-capabilities-%d", len(createdNodepools))
		response, err := nodepoolService.CreateNodepool(ctx, *nodepoolCreateFields(onCluster, suffix, nil))
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated))
		np := response.Body.(api.Nodepool)
		createdNodepools = append(createdNodepools, np)
		return np
	}

	patchCapabilities := func(nodepoolId string, capabilities api.SchedulingCapabilitiesNullable) api.ImplResponse {
		response, err := nodepoolService.PatchNodepool(ctx, nodepoolId, api.NodepoolUpdateFields{SchedulingCapabilities: capabilities})
		Expect(err).To(BeNil())
		return response
	}

	storedCapabilities := func(nodepoolId string) api.SchedulingCapabilities {
		response, err := nodepoolService.GetNodepool(ctx, nodepoolId)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK))
		return response.Body.(api.Nodepool).SchedulingCapabilities
	}

	expectSummary := func(nodepoolId string, summaryType openapi.OrgUnitsNodepoolSummaryType, result []openapi.OrgUnitsNodepoolSummaryInner, err error) {
		orgClient.On("NodePoolSummary", mock.Anything, nodepoolId, summaryType).Return(result, err).Once()
	}

	expectReset := func(nodepoolId string, action openapi.NodepoolResourcesResetAction, err error) {
		orgClient.On("NodePoolResetResources", mock.Anything, clusterId, nodepoolId, action).Return(err).Once()
	}

	turnOnCpuQuotaAndWeights := func(nodepoolId string) {
		response := patchCapabilities(nodepoolId, api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(true), AdjustableWeights: ptr.To(true)})
		Expect(response.Code).To(Equal(http.StatusOK))
	}

	It("turns flags on without any org-unit call", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)

		stored := storedCapabilities(np.Id)
		Expect(stored.CpuQuota).To(BeTrue())
		Expect(stored.AdjustableWeights).To(BeTrue())
	})

	It("falls back to the defaults on a PUT without schedulingCapabilities", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectReset(np.Id, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT, nil)

		response, err := nodepoolService.UpdateNodepool(ctx, np.Id, *nodepoolUpdateFields("-capabilities-put", nil))
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK))

		stored := storedCapabilities(np.Id)
		Expect(stored.Quota).To(BeTrue())
		Expect(stored.Ranks).To(BeTrue())
		Expect(stored.CpuQuota).To(BeFalse())
		Expect(stored.AdjustableWeights).To(BeFalse())
	})

	It("resets the org-unit CPU quotas when cpuQuota turns off, and keeps weights when weights turn off", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectReset(np.Id, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT, nil)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(false), AdjustableWeights: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusOK))

		stored := storedCapabilities(np.Id)
		Expect(stored.CpuQuota).To(BeFalse())
		Expect(stored.AdjustableWeights).To(BeFalse())
	})

	It("does not save the nodepool when a reset fails", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectReset(np.Id, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT, errors.New("org-unit-service is down"))

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(storedCapabilities(np.Id).CpuQuota).To(BeTrue())
	})

	It("turns weights off on a PUT without any org-unit call", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)

		fields := nodepoolUpdateFields(fmt.Sprintf("-capabilities-put-off-%d", len(createdNodepools)), nil)
		fields.SchedulingCapabilities = api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(true), AdjustableWeights: ptr.To(false)}
		response, err := nodepoolService.UpdateNodepool(ctx, np.Id, *fields)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK))

		orgClient.AssertNotCalled(GinkgoT(), "NodePoolResetResources", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		stored := storedCapabilities(np.Id)
		Expect(stored.AdjustableWeights).To(BeFalse())
		Expect(stored.CpuQuota).To(BeTrue())
	})

	It("names the resets already done when the nodepool save fails", func() {
		taken := createNodepool(clusterId)
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectReset(np.Id, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT, nil)

		fields := api.NodepoolUpdateFields{
			LabelKey:               ptr.To(taken.LabelKey),
			LabelValue:             ptr.To(taken.LabelValue),
			SchedulingCapabilities: api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(false)},
		}
		response, err := nodepoolService.UpdateNodepool(ctx, np.Id, fields)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusConflict))
		Expect(response.Body.(api.Error).Message).To(ContainSubstring("Already reset: CPU_QUOTAS_TO_DEFAULT"))
		Expect(storedCapabilities(np.Id).CpuQuota).To(BeTrue())
	})

	It("refuses to turn quota off while cpuQuota stays on", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Quota: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(storedCapabilities(np.Id).Quota).To(BeTrue())
	})

	It("refuses to turn quota off while org units have quota, and lists them", func() {
		np := createNodepool(clusterId)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, []openapi.OrgUnitsNodepoolSummaryInner{
			{Type: ptr.To("project"), Name: ptr.To("team-a")},
			{Type: ptr.To("department"), Name: ptr.To("research")},
		}, nil)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Quota: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		message := response.Body.(api.Error).Message
		Expect(message).To(ContainSubstring("project team-a, department research"))
		Expect(message).To(ContainSubstring("QUOTAS_TO_DEFAULT"))
		Expect(storedCapabilities(np.Id).Quota).To(BeTrue())
	})

	It("checks every flag before it resets any org unit", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, []openapi.OrgUnitsNodepoolSummaryInner{
			{Type: ptr.To("project"), Name: ptr.To("team-a")},
		}, nil)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Quota: ptr.To(false), CpuQuota: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		orgClient.AssertNotCalled(GinkgoT(), "NodePoolResetResources", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		Expect(storedCapabilities(np.Id).CpuQuota).To(BeTrue())
	})

	It("turns quota and cpuQuota off together when no org unit has quota", func() {
		np := createNodepool(clusterId)
		turnOnCpuQuotaAndWeights(np.Id)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, nil, nil)
		expectReset(np.Id, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT, nil)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Quota: ptr.To(false), CpuQuota: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusOK))

		stored := storedCapabilities(np.Id)
		Expect(stored.Quota).To(BeFalse())
		Expect(stored.CpuQuota).To(BeFalse())
	})

	It("refuses to turn ranks off while org units have a non-default rank", func() {
		np := createNodepool(clusterId)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK, []openapi.OrgUnitsNodepoolSummaryInner{
			{Type: ptr.To("project"), Name: ptr.To("team-b")},
		}, nil)

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Ranks: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.(api.Error).Message).To(ContainSubstring("RANKS_TO_DEFAULT"))
		Expect(storedCapabilities(np.Id).Ranks).To(BeTrue())
	})

	It("fails with 503 when the summary API fails", func() {
		np := createNodepool(clusterId)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK, nil, errors.New("org-unit-service is down"))

		response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Ranks: ptr.To(false)})
		Expect(response.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(storedCapabilities(np.Id).Ranks).To(BeTrue())
	})

	It("turns ranks off and on again", func() {
		np := createNodepool(clusterId)
		expectSummary(np.Id, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK, nil, nil)

		Expect(patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Ranks: ptr.To(false)}).Code).To(Equal(http.StatusOK))
		Expect(patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Ranks: ptr.To(true)}).Code).To(Equal(http.StatusOK))
		Expect(storedCapabilities(np.Id).Ranks).To(BeTrue())
	})

	Context("while editing capabilities is turned off", func() {
		BeforeEach(func() {
			Expect(runai_config.Set(&cluster_config.EditableNodepoolCapabilities{Enabled: false})).To(Succeed())
		})

		AfterEach(func() {
			Expect(runai_config.Set(&cluster_config.EditableNodepoolCapabilities{Enabled: true})).To(Succeed())
		})

		It("ignores the flags in a PATCH", func() {
			np := createNodepool(clusterId)

			response := patchCapabilities(np.Id, api.SchedulingCapabilitiesNullable{Ranks: ptr.To(false)})
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(storedCapabilities(np.Id).Ranks).To(BeTrue())
		})

		It("ignores the flags in a PUT and takes cpuQuota and weights from the tenant settings", func() {
			np := createNodepool(clusterId)
			useTenantSettings(true, true)

			fields := nodepoolUpdateFields(fmt.Sprintf("-capabilities-gated-%d", len(createdNodepools)), nil)
			fields.SchedulingCapabilities = api.SchedulingCapabilitiesNullable{CpuQuota: ptr.To(false), AdjustableWeights: ptr.To(false)}
			response, err := nodepoolService.UpdateNodepool(ctx, np.Id, *fields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			stored := storedCapabilities(np.Id)
			Expect(stored.CpuQuota).To(BeTrue())
			Expect(stored.AdjustableWeights).To(BeTrue())
		})
	})
})
