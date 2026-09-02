package services_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Nodepool gpu types", Ordered, func() {
	const (
		tenantId    = int32(40873)
		tenantIdStr = "40873"

		v100        = "Tesla-V100"
		h100        = "NVIDIA-H100-80GB-HBM3"
		v100Memory  = int64(34359738368)
		h100Memory  = int64(85899345920)
		halfH100Mem = int64(42949672960)

		homogeneousPool   = "gpu-homogeneous"
		heterogeneousPool = "gpu-heterogeneous"
		mixedMemoryPool   = "gpu-mixed-memory"
		unknownMemoryPool = "gpu-unknown-memory"
		noMemoryPool      = "gpu-no-memory"
		cpuOnlyPool       = "cpu-only"
		sharedNamePool    = "shared-name"
	)

	var (
		testCtx         context.Context
		nodepoolService api.NodePoolsAPIServicer
		nodeService     *services.NodeService
		cluster         *models.Cluster
		otherCluster    *models.Cluster
		createdIds      []string
	)

	createNodepool := func(clusterId, name string) string {
		response, err := nodepoolService.CreateNodepool(testCtx, api.NodepoolCreateFields{
			Name:       name,
			LabelKey:   "pool",
			LabelValue: name,
			ClusterId:  clusterId,
		})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated), "create nodepool %s failed: %v", name, response.Body)

		id := response.Body.(api.Nodepool).Id
		createdIds = append(createdIds, id)
		return id
	}

	addNode := func(clusterId, nodeName, nodepoolName string, gpuInfo *api.GpuInfo) {
		response, err := nodeService.UpsertNode(testCtx, clusterId, nodeName, api.NodeInfo{
			Status:    "Ready",
			NodePool:  nodepoolName,
			CreatedAt: time.Now(),
			GpuInfo:   gpuInfo,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	}

	gpuTypesOf := func(nodepools []api.NodepoolMinimal, name string) []api.NodepoolGpuTypeDetails {
		for _, nodepool := range nodepools {
			if nodepool.Name == name {
				return nodepool.GpuTypes
			}
		}
		Fail(fmt.Sprintf("nodepool %s missing from response", name))
		return nil
	}

	listMinimal := func(clusterId string) []api.NodepoolMinimal {
		response, err := nodepoolService.GetNodepoolsMinimal(
			testCtx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "asc", 0, 0)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusOK))
		return response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
	}

	BeforeAll(func() {
		testCtx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		testCtx = context.WithValue(testCtx, runaictx.CfSubjectID, "gpu-types-test-user")
		testCtx = context.WithValue(testCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})
		testCtx = addAuditToCtx(testCtx)

		tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
		tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
		orgClient := &org_unit_service_client_mock.IOrgUnitClient{}
		orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		nodepoolService = services.NewNodePoolsApiService(tenantSettingsClient, orgClient)
		nodeService = services.NewNodeService(tenantSettingsClient, nil)

		var dbErr *api.Error
		cluster, dbErr = dao.CreateCluster(testCtx, &models.Cluster{
			Name:     clusterNameGenerator.Generate("nodepool-gpu-types"),
			TenantId: tenantId,
			Domain:   "gpu-types.test",
			Version:  "3.0.0",
		})
		Expect(dbErr).To(BeNil())

		otherCluster, dbErr = dao.CreateCluster(testCtx, &models.Cluster{
			Name:     clusterNameGenerator.Generate("nodepool-gpu-types"),
			TenantId: tenantId,
			Domain:   "gpu-types.test",
			Version:  "3.0.0",
		})
		Expect(dbErr).To(BeNil())

		clusterId := cluster.ID.String()
		for _, name := range []string{
			homogeneousPool, heterogeneousPool, mixedMemoryPool, unknownMemoryPool, noMemoryPool,
			cpuOnlyPool, sharedNamePool,
		} {
			createNodepool(clusterId, name)
		}
		createNodepool(otherCluster.ID.String(), sharedNamePool)

		addNode(clusterId, "homogeneous-a", homogeneousPool, &api.GpuInfo{GpuType: h100, GpuCount: 8, GpuMemoryBytes: h100Memory})
		addNode(clusterId, "homogeneous-b", homogeneousPool, &api.GpuInfo{GpuType: h100, GpuCount: 4, GpuMemoryBytes: h100Memory})

		addNode(clusterId, "heterogeneous-a", heterogeneousPool, &api.GpuInfo{GpuType: h100, GpuCount: 8, GpuMemoryBytes: h100Memory})
		addNode(clusterId, "heterogeneous-b", heterogeneousPool, &api.GpuInfo{GpuType: v100, GpuCount: 2, GpuMemoryBytes: v100Memory})

		addNode(clusterId, "mixed-memory-a", mixedMemoryPool, &api.GpuInfo{GpuType: h100, GpuCount: 8, GpuMemoryBytes: h100Memory})
		addNode(clusterId, "mixed-memory-b", mixedMemoryPool, &api.GpuInfo{GpuType: h100, GpuCount: 1, GpuMemoryBytes: halfH100Mem})

		addNode(clusterId, "unknown-memory-a", unknownMemoryPool, &api.GpuInfo{GpuType: h100, GpuCount: 8, GpuMemoryBytes: h100Memory})
		addNode(clusterId, "unknown-memory-b", unknownMemoryPool, &api.GpuInfo{GpuType: h100, GpuCount: 4})

		// An old cluster reports its GPU model and count but never per-device memory.
		addNode(clusterId, "no-memory-a", noMemoryPool, &api.GpuInfo{GpuType: v100, GpuCount: 4})
		addNode(clusterId, "no-memory-b", noMemoryPool, &api.GpuInfo{GpuType: v100, GpuCount: 2})

		addNode(clusterId, "cpu-only-a", cpuOnlyPool, nil)

		addNode(otherCluster.ID.String(), "other-cluster-a", sharedNamePool,
			&api.GpuInfo{GpuType: v100, GpuCount: 3, GpuMemoryBytes: v100Memory})
	})

	AfterAll(func() {
		for _, id := range createdIds {
			Expect(dao.DeleteNodepool(testCtx, tenantId, "", id)).To(BeNil())
		}
		for _, c := range []*models.Cluster{cluster, otherCluster} {
			if c == nil {
				continue
			}
			Expect(dao.DeleteClusterNodes(testCtx, tenantId, c.ID.String())).To(Succeed())
			Expect(dao.DeleteCluster(testCtx, c.ID.String(), tenantId)).To(BeNil())
		}
	})

	Context("metadata (minimal) endpoint", func() {
		It("sums devices of a single gpu type across the pool's nodes", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), homogeneousPool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 12},
			))
		})

		It("reports an entry per gpu type in a heterogeneous pool", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), heterogeneousPool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 8},
				api.NodepoolGpuTypeDetails{Type: v100, GpuMemoryBytes: v100Memory, TotalDevices: 2},
			))
		})

		It("splits one gpu type into an entry per memory size", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), mixedMemoryPool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 8},
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: halfH100Mem, TotalDevices: 1},
			))
		})

		// Devices whose nodes never reported memory keep their own entry: folding them into the
		// known-memory one would hand the submission form GB figures for devices nobody measured.
		It("keeps devices with unknown memory in a separate entry from the same model", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), unknownMemoryPool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 8},
				api.NodepoolGpuTypeDetails{Type: h100, TotalDevices: 4},
			))
		})

		// A cluster that predates per-device memory sync reports no memory at all. Its GPU model
		// and device count still have to come through.
		It("reports the model and device count for a cluster that reports no memory", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), noMemoryPool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: v100, TotalDevices: 6},
			))
		})

		It("returns nothing for a pool without gpu nodes", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), cpuOnlyPool)).To(BeEmpty())
		})

		It("does not leak gpu types between same-named pools in different clusters", func() {
			Expect(gpuTypesOf(listMinimal(cluster.ID.String()), sharedNamePool)).To(BeEmpty())
			Expect(gpuTypesOf(listMinimal(otherCluster.ID.String()), sharedNamePool)).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: v100, GpuMemoryBytes: v100Memory, TotalDevices: 3},
			))
		})

		It("returns gpu types when fetching a single nodepool", func() {
			var nodepoolId string
			for _, nodepool := range listMinimal(cluster.ID.String()) {
				if nodepool.Name == homogeneousPool {
					nodepoolId = nodepool.Id
				}
			}
			Expect(nodepoolId).ToNot(BeEmpty())

			response, err := nodepoolService.GetNodepoolMinimal(testCtx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.NodepoolMinimal).GpuTypes).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 12},
			))
		})
	})

	Context("full endpoint", func() {
		It("returns the same gpu types as the metadata endpoint", func() {
			response, err := nodepoolService.GetNodepools(
				testCtx, []string{fmt.Sprintf("clusterId==%s", cluster.ID.String())}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			byName := map[string][]api.NodepoolGpuTypeDetails{}
			for _, nodepool := range response.Body.(api.GetNodepools200Response).Nodepools {
				byName[nodepool.Name] = nodepool.GpuTypes
			}

			Expect(byName[homogeneousPool]).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 12},
			))
			Expect(byName[heterogeneousPool]).To(ConsistOf(
				api.NodepoolGpuTypeDetails{Type: h100, GpuMemoryBytes: h100Memory, TotalDevices: 8},
				api.NodepoolGpuTypeDetails{Type: v100, GpuMemoryBytes: v100Memory, TotalDevices: 2},
			))
			Expect(byName[cpuOnlyPool]).To(BeEmpty())
		})
	})
})
