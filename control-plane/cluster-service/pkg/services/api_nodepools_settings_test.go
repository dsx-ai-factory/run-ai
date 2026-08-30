package services_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/settings"
	org_unit_service_client_mock "github.com/run-ai/runai/runai-clients/org-unit-service-client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	settingPkg "github.com/run-ai/runai/runai-common-packages/setting"
	settingPkgMocks "github.com/run-ai/runai/runai-common-packages/setting/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("ApiNodepoolsSettings", func() {
	var ()
})

var _ = Describe("ApiNodepoolsSettings - node pool settings from package", Ordered, func() {
	tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	orgClient := &org_unit_service_client_mock.IOrgUnitClient{}

	ctx := createContext()
	var nodepoolService api.NodePoolsAPIServicer
	var createdNodepools []api.Nodepool
	var clusterId string
	const testSettingsSuffix = "-settings"

	// second tenant
	var otherCtx context.Context
	var otherClusterId string
	var otherCreatedNodepools []api.Nodepool

	BeforeAll(func() {
		tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
		orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		nodepoolService = services.NewNodePoolsApiService(tenantSettingsClient, orgClient)

		cluster, err := dao.CreateCluster(ctx, &models.Cluster{
			Name:     "cluster-for-np-setting",
			TenantId: 1,
			Domain:   "xyz",
			Version:  "v2.23",
		})
		Expect(err).To(BeNil())
		clusterId = cluster.ID.String()

		// create a couple of nodepools in this tenant/cluster
		for i := 0; i < 2; i++ {
			req := nodepoolCreateFields(clusterId, fmt.Sprintf("setting-%d", i), nil)
			response, apiErr := nodepoolService.CreateNodepool(ctx, *req)
			Expect(apiErr).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			np := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
		}

		// setup other tenant context and data
		otherCtx = createContextForTenant("2", testUser)
		otherCluster, err2 := dao.CreateCluster(otherCtx, &models.Cluster{
			Name:     "cluster-for-np-setting-2",
			TenantId: 2,
			Domain:   "xyz",
			Version:  "v2.23",
		})
		Expect(err2).To(BeNil())
		otherClusterId = otherCluster.ID.String()
		for i := 0; i < 2; i++ {
			req := nodepoolCreateFields(otherClusterId, fmt.Sprintf("setting2-%d", i), nil)
			response, apiErr := nodepoolService.CreateNodepool(otherCtx, *req)
			Expect(apiErr).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			np := response.Body.(api.Nodepool)
			otherCreatedNodepools = append(otherCreatedNodepools, np)
		}
	})

	AfterAll(func() {
		for _, np := range createdNodepools {
			_ = dao.DeleteNodepool(ctx, 1, testUser, np.Id)
		}
		createdNodepools = nil
		_ = dao.DeleteCluster(ctx, clusterId, 1)

		for _, np := range otherCreatedNodepools {
			_ = dao.DeleteNodepool(otherCtx, 2, testUser, np.Id)
		}
		otherCreatedNodepools = nil
		_ = dao.DeleteCluster(otherCtx, otherClusterId, 2)
	})

	Context("test cpu/weights is taken from settings on nodepool create/update", func() {
		var createdNodepoolIds []string

		AfterAll(func() {
			for _, npId := range createdNodepoolIds {
				_ = dao.DeleteNodepool(ctx, 1, testUser, npId)
			}
		})

		It("validate the default nodepool was created with correct config", func() {
			// cpu_quota / adjustable_weights defaults are now read from the local settings package
			settingPkgMock := settingPkgMocks.NewSettingsManager(GinkgoT())
			settingPkgMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
				Return(&settings.CPUResourcesQuota{Enabled: true}, nil)
			settingPkgMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)).
				Return(&settings.ProjectManualOverQuota{Enabled: false}, nil)
			settingPkg.SetMock(settingPkgMock)

			res, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
				Operation: api.NOTIFICATIONOPERATION_CREATE,
				Id:        clusterId,
				Name:      "cluster-for-np-setting",
			})
			Expect(err).To(BeNil(), "Expected cluster creation notification to succeed")
			Expect(res.Code).To(Equal(http.StatusNoContent), "Expected cluster creation notification to succeed")

			nodePools, dbErr := dao.GetNodepools(ctx, 1, *mustPermittedScopes(clusterId), 0, 0, "", "id", nil)
			Expect(dbErr).To(BeNil())
			Expect(nodePools).ToNot(BeEmpty())
			defaultFound := false
			for _, np := range nodePools {
				if np.Name == "default" {
					defaultFound = true
					Expect(np.AllowCpuQuota).To(Equal(true))
					Expect(np.AllowAdjustableWeights).To(Equal(false))
					Expect(np.AllowQuota).To(Equal(true))
					Expect(np.AllowRanks).To(Equal(true))
					createdNodepoolIds = append(createdNodepoolIds, fmt.Sprintf("%d", np.ID))
				}
			}
			Expect(defaultFound).To(BeTrue())
		})

		It("nodepool create/update", func() {
			// on create: cpu_quota enabled, adjustable_weights disabled (read from settings package)
			createMock := settingPkgMocks.NewSettingsManager(GinkgoT())
			createMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
				Return(&settings.CPUResourcesQuota{Enabled: true}, nil)
			createMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)).
				Return(&settings.ProjectManualOverQuota{Enabled: false}, nil)
			settingPkg.SetMock(createMock)

			req := nodepoolCreateFields(clusterId, "np-weights-cpu-from-settings", nil)
			response, apiErr := nodepoolService.CreateNodepool(ctx, *req)
			Expect(apiErr).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			np := response.Body.(api.Nodepool)
			createdNodepoolIds = append(createdNodepoolIds, np.Id)
			Expect(np.SchedulingCapabilities.CpuQuota).To(Equal(true))
			Expect(np.SchedulingCapabilities.AdjustableWeights).To(Equal(false))

			// on update: flip the settings; UpdateNodepool re-reads them from the settings package
			updateMock := settingPkgMocks.NewSettingsManager(GinkgoT())
			updateMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)).
				Return(&settings.CPUResourcesQuota{Enabled: false}, nil)
			updateMock.On("Get", mock.Anything, tenantIdStr, string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)).
				Return(&settings.ProjectManualOverQuota{Enabled: true}, nil)
			settingPkg.SetMock(updateMock)

			updateReq := nodepoolUpdateFields("-ts", nil)
			updateResp, updateApiErr := nodepoolService.UpdateNodepool(ctx, np.Id, *updateReq)
			Expect(updateApiErr).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updatedNp := updateResp.Body.(api.Nodepool)
			Expect(updatedNp.SchedulingCapabilities.CpuQuota).To(Equal(false))
			Expect(updatedNp.SchedulingCapabilities.AdjustableWeights).To(Equal(true))
		})
	})

	Context("UpdateNodePoolSettingByKey propagates the toggle to all tenant nodepools", func() {
		// disabling cpuResourcesQuota also calls org-unit-service to wipe project/department quota; stub it.
		orgClientForSettings := &org_unit_service_client_mock.IOrgUnitClient{}
		orgClientForSettings.On("UpdateOrgUnitSettings", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		settingsService := services.NewSettingsService(orgClientForSettings)

		setNodePoolSettingMock := func(key api.NodePoolSettingsKeyEnum, expectedValue interface{}) {
			settingMock := settingPkgMocks.NewSettingsManager(GinkgoT())
			settingMock.On("Update", mock.Anything, tenantIdStr, string(key), true, expectedValue).Return(nil)
			settingPkg.SetMock(settingMock)
		}

		It("enabling cpuResourcesQuota sets AllowCpuQuota on tenant nodepools where quota is allowed", func() {
			setNodePoolSettingMock(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, settings.CPUResourcesQuota{Enabled: true})

			resp, err := settingsService.UpdateNodePoolSettingByKey(ctx,
				api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			nps, dbErr := dao.GetNodepools(ctx, 1, *mustPermittedScopes(clusterId), 0, 0, "", "", nil)
			Expect(dbErr).To(BeNil())
			Expect(nps).ToNot(BeEmpty())
			for _, m := range nps {
				// cpu quota is enabled only on nodepools where quota itself is allowed
				Expect(m.AllowCpuQuota).To(Equal(m.AllowQuota))
			}
		})

		It("enabling projectManualOverQuota sets AllowAdjustableWeights on all tenant nodepools", func() {
			setNodePoolSettingMock(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA, settings.ProjectManualOverQuota{Enabled: true})

			resp, err := settingsService.UpdateNodePoolSettingByKey(ctx,
				api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA, api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(true)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			nps, dbErr := dao.GetNodepools(ctx, 1, *mustPermittedScopes(clusterId), 0, 0, "", "", nil)
			Expect(dbErr).To(BeNil())
			Expect(nps).ToNot(BeEmpty())
			for _, m := range nps {
				Expect(m.AllowAdjustableWeights).To(BeTrue())
			}
		})

		It("disabling cpuResourcesQuota clears AllowCpuQuota on all tenant nodepools", func() {
			setNodePoolSettingMock(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, settings.CPUResourcesQuota{Enabled: false})

			resp, err := settingsService.UpdateNodePoolSettingByKey(ctx,
				api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA, api.UpdateNodePoolSettingByKeyRequest{Enabled: ptr.To(false)}, "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			nps, dbErr := dao.GetNodepools(ctx, 1, *mustPermittedScopes(clusterId), 0, 0, "", "", nil)
			Expect(dbErr).To(BeNil())
			Expect(nps).ToNot(BeEmpty())
			for _, m := range nps {
				Expect(m.AllowCpuQuota).To(BeFalse())
			}
		})
	})
})

func mustPermittedScopes(clusterId string) *authorization_api.PermittedScopes {
	// helper lives here to avoid importing test helpers from other files
	ps := authorization_api.PermittedScopes{Clusters: []string{clusterId}}
	return &ps
}

// createContextForTenant duplicates createContext with different tenant id
func createContextForTenant(tenantIdStr string, subject string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subject)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{tenantIdStr},
	})
	ctx = addAuditToCtx(ctx)
	return ctx
}
