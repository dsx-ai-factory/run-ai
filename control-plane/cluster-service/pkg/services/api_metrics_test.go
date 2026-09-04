/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	cd "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	clusterinstall "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-installation/mocks"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	smocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services"
	csc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/services/cluster-state-calculator"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	cluster_settings "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/settings"
	idm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/mocks"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	org_unit_client "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	settingPkgMocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("ApiClustersService", Ordered, func() {
	type Mocks struct {
		clusterStateCalculator      *csc_mocks.ClusterStateCalculator
		tenantSettingsClient        *bc_mocks.TenantSettingsClient
		idmClient                   *idm_client.IIdm
		orgUnitClient               *org_unit_client.IOrgUnitClient
		settingPkg                  *settingPkgMocks.SettingsManager
		clusterAccessControlCreator *smocks.ClusterAccessControlManager
	}

	const (
		tenantId            = 1
		fullUrl             = "tenant.test.run.ai"
		testSuite           = "cluster-service"
		baseTimestampUnix   = 1708958503653
		numOfSamples        = 10
		numOfMultipleResult = 8
	)

	var (
		mocks              *Mocks
		clustersApiService api.ClustersAPIServicer
		nodePoolApiService api.NodePoolsAPIServicer
		ctx                context.Context
		categories         = []string{"Build", "Train"}
	)

	BeforeEach(func() {
		// Setup empty mocks
		mocks = &Mocks{
			clusterStateCalculator:      &csc_mocks.ClusterStateCalculator{},
			tenantSettingsClient:        &bc_mocks.TenantSettingsClient{},
			idmClient:                   &idm_client.IIdm{},
			orgUnitClient:               &org_unit_client.IOrgUnitClient{},
			settingPkg:                  settingPkgMocks.NewSettingsManager(GinkgoT()),
			clusterAccessControlCreator: &smocks.ClusterAccessControlManager{},
		}
		mocks.clusterStateCalculator.On("CalculateState", mock.Anything).Return("")
		mocks.clusterAccessControlCreator.On("CreateAccessControl", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		setting_pkg.SetMock(mocks.settingPkg)

		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})

		clusterDeleter := cd.NewClusterDeleter(authorization_client.NewAuthorizationClient(), mocks.orgUnitClient, nil)
		clusterInstall := clusterinstall.NewClusterInstallService(GinkgoT())
		nodePoolApiService = services.NewNodePoolsApiService(mocks.tenantSettingsClient, nil)
		clustersApiService = services.NewClustersApiService(mocks.clusterStateCalculator, mocks.idmClient, clusterDeleter, clusterInstall, nil, nodePoolApiService, mocks.orgUnitClient,
			mocks.clusterAccessControlCreator)
	})

	BeforeAll(func() {
		// set all metrics mocks once before all tests
		metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
			runaimetrics.RunaiGpuUtilizationPerClusterMetric{}.GetName():              getDefaultMetricMockResult(1, baseTimestampUnix),
			runaimetrics.RunaiGpuMemoryUtilizationPerClusterMetric{}.GetName():        getDefaultMetricMockResult(2, baseTimestampUnix),
			runaimetrics.RunaiCpuUtilizationPerClusterMetric{}.GetName():              getDefaultMetricMockResult(3, baseTimestampUnix),
			runaimetrics.RunaiMemoryUtilizationPerClusterMetric{}.GetName():           getDefaultMetricMockResult(4, baseTimestampUnix),
			runaimetrics.RunaiAllocatedGpuCountPerClusterMetric{}.GetName():           getDefaultMetricMockResult(5, baseTimestampUnix),
			runaimetrics.RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric{}.GetName():  getDefaultMetricMockResult(6, baseTimestampUnix),
			runaimetrics.RunaiGpuNodesCountPerNodepoolMetric{}.GetName():              getDefaultMetricMockResult(1, baseTimestampUnix),
			runaimetrics.RunaiUnallocatedGpuCountPerNodepoolMetric{}.GetName():        getDefaultMetricMockResult(2, baseTimestampUnix),
			runaimetrics.RunaiGpuUtilizationDistributionPerNodepoolMetric{}.GetName(): getMetricListMockResult([]int{0, 2, 1, 0}, baseTimestampUnix),
			runaimetrics.RunaiGpuUtilizationPerGpuMetric{}.GetName():                  getDefaultMetricMockMultipleResults(50, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiGpuUtilizationPerNodeMetric{}.GetName():                 getDefaultMetricMockResult(60, baseTimestampUnix),
			runaimetrics.RunaiGpuMemoryUtilizationPerGpuMetric{}.GetName():            getDefaultMetricMockMultipleResults(70, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiGpuMemoryUtilizationPerNodeMetric{}.GetName():           getDefaultMetricMockResult(60, baseTimestampUnix),
			runaimetrics.RunaiGpuMemoryUsedBytesPerGpuMetric{}.GetName():              getDefaultMetricMockMultipleResults(70, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiAllocatedGpuCountPerGpuMetric{}.GetName():               getDefaultMetricMockMultipleResults(80, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiAllocatedGpuMemoryBytesPerGpuMetric{}.GetName():         getDefaultMetricMockMultipleResults(90, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiIdleSecondsPerGpuMetric{}.GetName():                     getDefaultMetricMockMultipleResults(100, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiUsedGpuMemoryBytesPerNodeMetric{}.GetName():             getDefaultMetricMockResult(60, baseTimestampUnix),
			runaimetrics.RunaiUsedCpuCoresPerNodeMetric{}.GetName():                   getDefaultMetricMockResult(10, baseTimestampUnix),
			runaimetrics.RunaiCpuUtilizationPerNodeMetric{}.GetName():                 getDefaultMetricMockResult(20, baseTimestampUnix),
			runaimetrics.RunaiUsedMemoryBytesPerNodeMetric{}.GetName():                getDefaultMetricMockResult(30, baseTimestampUnix),
			runaimetrics.RunaiMemoryUtilizationPerNodeMetric{}.GetName():              getDefaultMetricMockResult(40, baseTimestampUnix),
			runaimetrics.RunaiGpuPipeFp64ActivePerGpuMetric{}.GetName():               getDefaultMetricMockMultipleResults(30, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiGpuNvlinkRxBytesPerGpuMetric{}.GetName():                getDefaultMetricMockMultipleResults(40, baseTimestampUnix, numOfSamples, numOfMultipleResult),
			runaimetrics.RunaiWorkloadsCountMetric{
				GroupBy: metrics_client.GroupByCategory,
			}.GetName(): getDefaultMetricMockMultipleResultsWithGroupBy(40, baseTimestampUnix, numOfSamples, "workload_category", categories),
			// Dont have a mock for metrics.RunaiGpuCountPerClusterMetric{}.GetName() to test a case where no result returned
		})
	})

	Describe("Cluster, Nodepool and Node Metrics", Ordered, func() {
		var metricsCluster1Id string
		BeforeAll(func() {
			metricsCluster1, _ := dao.CreateCluster(ctx, &models.Cluster{
				Common:   models.Common{},
				Name:     "cluster-metrics",
				Version:  "2.22",
				TenantId: tenantId,
			})
			metricsCluster1Id = metricsCluster1.ID.String()
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Tenants: []string{tenantIdStr}, Clusters: []string{metricsCluster1Id},
			})
		})

		AfterAll(func() {
			err := dao.DeleteCluster(ctx, metricsCluster1Id, tenantId)
			Expect(err).To(BeNil())
		})

		It("Get workload count metric - without groupBy", func() {
			numberOfSamples := int32(10)
			nodepoolName := "nodepool-1"
			resp, err := clustersApiService.GetClusterMetrics(ctx, metricsCluster1Id, time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_WORKLOADS_COUNT}, numberOfSamples, string(metrics_client.GroupByCategory), nodepoolName)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			metrics, ok := resp.Body.(api.GroupedMetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(len(categories)))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.METRICSTYPE_WORKLOADS_COUNT))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(int(numberOfSamples)))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
			Expect(len(metrics.Measurements[0].Groups)).To(Equal(1))
			Expect(metrics.Measurements[0].Groups[0].Key).To(Equal(string(metrics_client.GroupByCategory)))
			Expect(metrics.Measurements[0].Groups[0].Value).To(Equal(categories[0]))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("40"))
			Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("41"))
			Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("42"))

			Expect(len(metrics.Measurements[1].Groups)).To(Equal(1))
			Expect(metrics.Measurements[1].Groups[0].Key).To(Equal(string(metrics_client.GroupByCategory)))
			Expect(metrics.Measurements[1].Groups[0].Value).To(Equal(categories[1]))
		})

		It("Get multiple metric - cluster", func() {
			resp, err := clustersApiService.GetClusterMetrics(ctx, metricsCluster1Id, time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_GPU_UTILIZATION, api.METRICSTYPE_ALLOCATED_GPU}, 10, "", "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metrics, ok := resp.Body.(api.GroupedMetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(2))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.METRICSTYPE_GPU_UTILIZATION))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.METRICSTYPE_ALLOCATED_GPU))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(3))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("1"))
			Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("2"))
			Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("3"))
		})

		It("Get multiple metric with single failure metric - cluster", func() {
			resp, err := clustersApiService.GetClusterMetrics(ctx, metricsCluster1Id, time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_GPU_UTILIZATION, api.METRICSTYPE_ALLOCATED_GPU, api.METRICSTYPE_TOTAL_GPU}, 10, "", "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))

			metrics, ok := resp.Body.(api.GroupedMetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(3))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.METRICSTYPE_GPU_UTILIZATION))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.METRICSTYPE_ALLOCATED_GPU))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.METRICSTYPE_TOTAL_GPU))
			Expect(len(*(metrics.Measurements)[0].Values)).To(Equal(3))
			Expect(len(*(metrics.Measurements)[1].Values)).To(Equal(3))
			Expect(metrics.Measurements[2].Values).To(BeNil())
		})

		It("Get multiple metric - nodePool", func() {
			Skip("Flaky test")
			resp, err := nodePoolApiService.GetNodepoolMetrics(ctx, metricsCluster1Id, "default", time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_TOTAL_GPU_NODES, api.METRICSTYPE_UNALLOCATED_GPU, api.METRICSTYPE_GPU_UTILIZATION_DISTRIBUTION}, 10)

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(200))

			metrics, ok := resp.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(3))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.METRICSTYPE_TOTAL_GPU_NODES))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.METRICSTYPE_UNALLOCATED_GPU))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.METRICSTYPE_GPU_UTILIZATION_DISTRIBUTION))
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(3))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
			Expect((*metrics.Measurements[2].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[2].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
			Expect((*metrics.Measurements[2].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("1"))
			Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("2"))
			Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("3"))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("2"))
			Expect((*metrics.Measurements[1].Values)[1].Value).To(Equal("3"))
			Expect((*metrics.Measurements[1].Values)[2].Value).To(Equal("4"))
			Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("[{\"0%-25%\":\"0\"},{\"25%-50%\":\"2\"},{\"50%-75%\":\"1\"},{\"75%-100%\":\"0\"}]"))
			Expect((*metrics.Measurements[2].Values)[1].Value).To(Equal("[{\"0%-25%\":\"1\"},{\"25%-50%\":\"3\"},{\"50%-75%\":\"2\"},{\"75%-100%\":\"1\"}]"))
			Expect((*metrics.Measurements[2].Values)[2].Value).To(Equal("[{\"0%-25%\":\"2\"},{\"25%-50%\":\"4\"},{\"50%-75%\":\"3\"},{\"75%-100%\":\"2\"}]"))
		})

		It("groupBy category not supported by old clusters", func() {
			olderCluster, _ := dao.CreateCluster(ctx, &models.Cluster{
				Common:   models.Common{},
				Name:     "old-cluster",
				Version:  "2.21",
				TenantId: tenantId,
			})
			olderClusterId := olderCluster.ID.String()
			olderClusterCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Tenants: []string{tenantIdStr}, Clusters: []string{metricsCluster1Id},
			})
			numberOfSamples := int32(10)
			resp, err := clustersApiService.GetClusterMetrics(olderClusterCtx, olderClusterId, time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_WORKLOADS_COUNT}, numberOfSamples, string(metrics_client.GroupByCategory), "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			deleteErr := dao.DeleteCluster(ctx, olderClusterId, tenantId)
			Expect(deleteErr).To(BeNil())
		})

		Context("Get node metrics", func() {
			var (
				nodeService *services.NodeService
				nodeName    = "node1"
				nodeId      string
				cluster1Id  string
			)

			BeforeAll(func() {
				nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)
				cluster1, _ := dao.CreateCluster(ctx, &models.Cluster{
					Common:   models.Common{},
					Name:     "cluster-1",
					Version:  "2.19",
					TenantId: tenantId,
				})
				cluster1Id = cluster1.ID.String()

				nodeInfo := api.NodeInfo{
					NodePool: "default",
					Status:   "Ready",
					GpuInfo: &api.GpuInfo{
						GpuType:  "type",
						GpuCount: 16,
					},
				}
				result, err := nodeService.UpsertNode(ctx, cluster1Id, nodeName, nodeInfo)
				Expect(err).To(BeNil())
				Expect(result.Code).To(Equal(http.StatusNoContent))

				res, err := nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("clusterUuid==%s", cluster1Id)}, "", "", 0, 0, "")
				Expect(err).To(BeNil())
				apiNodes := res.Body.(api.Nodes).Nodes
				Expect(len(apiNodes)).To(Equal(1))
				nodeId = apiNodes[0].Id
			})

			AfterAll(func() {
				err := dao.DeleteCluster(ctx, cluster1Id, tenantId)
				Expect(err).To(BeNil())
			})

			It("Get node metrics successfully and validate", func() {
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_UTILIZATION, api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))
				metrics, ok := res.Body.(api.MetricsResponse)
				Expect(ok).To(BeTrue())
				Expect(len(metrics.Measurements)).To(Equal(2 + 8*1))
				Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION))
				Expect(metrics.Measurements[0].Labels).To(BeNil())
				Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("60"))
				Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("61"))
				Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("62"))

				Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_CPU_UTILIZATION))
				Expect(metrics.Measurements[1].Labels).To(BeNil())
				Expect((*metrics.Measurements[1].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[1].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[1].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("20"))
				Expect((*metrics.Measurements[1].Values)[1].Value).To(Equal("21"))
				Expect((*metrics.Measurements[1].Values)[2].Value).To(Equal("22"))

				Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU))
				Expect(metrics.Measurements[2].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[2].Labels)["gpu"]).To(Equal("0"))
				Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("50"))
				Expect((*metrics.Measurements[2].Values)[1].Value).To(Equal("51"))
				Expect((*metrics.Measurements[2].Values)[2].Value).To(Equal("52"))

				Expect(metrics.Measurements[6].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU))
				Expect(metrics.Measurements[6].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[6].Labels)["gpu"]).To(Equal("4"))
				Expect((*metrics.Measurements[6].Values)[0].Value).To(Equal("50"))
				Expect((*metrics.Measurements[6].Values)[1].Value).To(Equal("51"))
				Expect((*metrics.Measurements[6].Values)[2].Value).To(Equal("52"))
			})

			It("Get node metrics - request too many", func() {
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_UTILIZATION, api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU,
					api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_UTILIZATION, api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU,
					api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_UTILIZATION, api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU,
					api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_UTILIZATION, api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				// bad request
				Expect(res.Code).To(Equal(400))
			})

			It("Get node metrics - cluster scoped permission without tenant scope succeeds", func() {
				// simulate a subject that has only cluster scope on the node's cluster (no tenant scope)
				clusterScopedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Clusters: []string{cluster1Id},
				})
				res, err := nodeService.GetNodeMetrics(clusterScopedCtx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_UTILIZATION, api.NODEMETRICTYPE_CPU_USAGE_CORES,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))
			})

			It("Get node metrics - cluster scoped to another cluster is denied", func() {
				// simulate scope on a different cluster than the node's cluster
				otherClusterId := uuid.NewString()
				mismatchedScopeCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Clusters: []string{otherClusterId},
				})
				res, err := nodeService.GetNodeMetrics(mismatchedScopeCtx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_UTILIZATION, api.NODEMETRICTYPE_CPU_USAGE_CORES,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				// node exists but not visible under this cluster scope → treated as not found
				Expect(res.Code).To(Or(Equal(404), Equal(403)))
			})

			It("Get node metrics - department scoped permission is denied", func() {
				departmentScopedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Departments: []string{uuid.NewString()},
				})
				res, err := nodeService.GetNodeMetrics(departmentScopedCtx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_UTILIZATION, api.NODEMETRICTYPE_CPU_USAGE_CORES,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Or(Equal(404), Equal(403)))
			})

			It("Get node metrics - unexisting node", func() {
				res, err := nodeService.GetNodeMetrics(ctx, uuid.NewString(), []api.NodeMetricType{
					api.NODEMETRICTYPE_CPU_MEMORY_UTILIZATION, api.NODEMETRICTYPE_CPU_MEMORY_USAGE_BYTES, api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				// not found
				Expect(res.Code).To(Equal(404))
			})

			It("Get node metrics - existing node but no permissions", func() {
				ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Tenants: []string{"2"},
				})
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_UTILIZATION, api.NODEMETRICTYPE_CPU_USAGE_CORES, api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				// forbidden
				Expect(res.Code).To(Or(Equal(404), Equal(403)))
			})

			It("Get node metrics as csv", func() {
				ctx = context.WithValue(ctx, runaictx.CfAccept, "text/csv")
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_UTILIZATION, api.NODEMETRICTYPE_CPU_USAGE_CORES, api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))

				response, ok := res.Body.(runaihttp.CsvFile)
				Expect(ok).To(BeTrue())
				metrics := response.Data
				Expect(metrics[0][0]).To(BeEquivalentTo("Timestamp"))
				Expect(metrics[0][1]).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_UTILIZATION))
				Expect(metrics[0][2]).To(BeEquivalentTo(api.NODEMETRICTYPE_CPU_USAGE_CORES))
				Expect(metrics[0][3]).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU))
				Expect(metrics[1][0]).To(Equal(stringTimestamp(baseTimestampUnix)))
				Expect(metrics[2][0]).To(Equal(stringTimestamp(baseTimestampUnix + 1000)))
				Expect(metrics[3][0]).To(Equal(stringTimestamp(baseTimestampUnix + 2000)))
				Expect(metrics[1][1]).To(Equal("60"))
				Expect(metrics[1][2]).To(Equal("10"))
				Expect(metrics[1][3]).To(Equal("70"))
				Expect(metrics[2][1]).To(Equal("61"))
				Expect(metrics[2][2]).To(Equal("11"))
				Expect(metrics[2][3]).To(Equal("71"))
				Expect(metrics[3][1]).To(Equal("62"))
				Expect(metrics[3][2]).To(Equal("12"))
				Expect(metrics[3][3]).To(Equal("72"))
				// TODO: fix after fixing the csv response to include labels
			})

			It("Get node metrics - more metrics to test", func() {
				err := cache.ClearCache()
				Expect(err).To(BeNil())

				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").Unset()
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").
					Return(&cluster_settings.EnableAdvancedGPUMetrics{Enabled: false}, nil)
				nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU, api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(400))

				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").Unset()
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").
					Return(&cluster_settings.EnableAdvancedGPUMetrics{Enabled: true}, nil)

				err = cache.ClearCache()
				Expect(err).To(BeNil())
				nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)

				res, err = nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU, api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))
				metrics, ok := res.Body.(api.MetricsResponse)
				Expect(ok).To(BeTrue())
				Expect(len(metrics.Measurements)).To(Equal(2 * numOfMultipleResult))
				Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU))
				Expect(metrics.Measurements[0].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[0].Labels)["gpu"]).To(Equal("0"))
				Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("30"))
				Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("31"))
				Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("32"))

				Expect(metrics.Measurements[4].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU))
				Expect(metrics.Measurements[4].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[4].Labels)["gpu"]).To(Equal("4"))
				Expect((*metrics.Measurements[4].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[4].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[4].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[4].Values)[0].Value).To(Equal("30"))
				Expect((*metrics.Measurements[4].Values)[1].Value).To(Equal("31"))
				Expect((*metrics.Measurements[4].Values)[2].Value).To(Equal("32"))

				Expect(metrics.Measurements[8].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU))
				Expect(metrics.Measurements[8].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[8].Labels)["gpu"]).To(Equal("0"))
				Expect((*metrics.Measurements[8].Values)[0].Value).To(Equal("40"))
				Expect((*metrics.Measurements[8].Values)[1].Value).To(Equal("41"))
				Expect((*metrics.Measurements[8].Values)[2].Value).To(Equal("42"))
			})

			It("Get node metrics per gpu - more metrics to test", func() {
				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_ALLOCATION_PER_GPU, api.NODEMETRICTYPE_GPU_MEMORY_ALLOCATION_BYTES_PER_GPU, api.NODEMETRICTYPE_GPU_IDLE_TIME_SECONDS_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))
				metrics, ok := res.Body.(api.MetricsResponse)
				Expect(ok).To(BeTrue())
				Expect(len(metrics.Measurements)).To(Equal(3 * numOfMultipleResult))
				Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_ALLOCATION_PER_GPU))
				Expect(metrics.Measurements[0].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[0].Labels)["gpu"]).To(Equal("0"))
				Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[0].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[0].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("80"))
				Expect((*metrics.Measurements[0].Values)[1].Value).To(Equal("81"))
				Expect((*metrics.Measurements[0].Values)[2].Value).To(Equal("82"))
				for i := 0; i < numOfMultipleResult; i++ {
					Expect(metrics.Measurements[i].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_ALLOCATION_PER_GPU))
					Expect(metrics.Measurements[i].Labels).ToNot(BeNil())
					Expect((*metrics.Measurements[i].Labels)["gpu"]).To(Equal(strconv.Itoa(i)))
				}

				Expect(metrics.Measurements[numOfMultipleResult].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_MEMORY_ALLOCATION_BYTES_PER_GPU))
				Expect(metrics.Measurements[numOfMultipleResult].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[0].Value).To(Equal("90"))
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[1].Value).To(Equal("91"))
				Expect((*metrics.Measurements[numOfMultipleResult].Values)[2].Value).To(Equal("92"))
				for i := numOfMultipleResult; i < 2*numOfMultipleResult; i++ {
					Expect(metrics.Measurements[i].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_MEMORY_ALLOCATION_BYTES_PER_GPU))
					Expect(metrics.Measurements[i].Labels).ToNot(BeNil())
					Expect((*metrics.Measurements[i].Labels)["gpu"]).To(Equal(strconv.Itoa(i - numOfMultipleResult)))
				}

				Expect(metrics.Measurements[2*numOfMultipleResult].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_IDLE_TIME_SECONDS_PER_GPU))
				Expect(metrics.Measurements[2*numOfMultipleResult].Labels).ToNot(BeNil())
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[1].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 1000)))
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[2].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix + 2000)))
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[0].Value).To(Equal("100"))
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[1].Value).To(Equal("101"))
				Expect((*metrics.Measurements[2*numOfMultipleResult].Values)[2].Value).To(Equal("102"))
				for i := 2 * numOfMultipleResult; i < 3*numOfMultipleResult; i++ {
					Expect(metrics.Measurements[i].Type).To(BeEquivalentTo(api.NODEMETRICTYPE_GPU_IDLE_TIME_SECONDS_PER_GPU))
					Expect(metrics.Measurements[i].Labels).ToNot(BeNil())
					Expect((*metrics.Measurements[i].Labels)["gpu"]).To(Equal(strconv.Itoa(i - 2*numOfMultipleResult)))
				}
			})

			It("Toggling EnableAdvancedGPUMetrics in settings table gates advanced metrics", func() {
				err := cache.ClearCache()
				Expect(err).To(BeNil())

				// set FF to false - should return 400
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").Unset()
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").
					Return(&cluster_settings.EnableAdvancedGPUMetrics{Enabled: false}, nil)
				nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)

				res, err := nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(400))

				// flip the setting to true - should now return 200
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").Unset()
				mocks.settingPkg.On("Get", mock.Anything, tenantIdStr, "enableAdvancedGpuMetrics").
					Return(&cluster_settings.EnableAdvancedGPUMetrics{Enabled: true}, nil)

				nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)
				res, err = nodeService.GetNodeMetrics(ctx, nodeId, []api.NodeMetricType{
					api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU,
				},
					time.Now(), time.Now().Add(-1*time.Minute), 10)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(200))
			})
		})
	})

	Describe("ClusterMetrics csv", Ordered, func() {
		const baseTimestampUnix = 1708958503653

		BeforeEach(func() {
			err := cache.ClearCache()
			Expect(err).To(BeNil())
			ctx = context.WithValue(ctx, runaictx.CfAccept, "text/csv")
		})

		It("Get multiple metric", func() {
			resp, err := clustersApiService.GetClusterMetrics(ctx, uuid.NewString(), time.Now(), time.Now().Add(-1*time.Minute),
				[]api.MetricsType{api.METRICSTYPE_GPU_UTILIZATION, api.METRICSTYPE_ALLOCATED_GPU, api.METRICSTYPE_TOTAL_GPU}, 10, "", "")
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(207))

			response, ok := resp.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			metrics := response.Data
			Expect(metrics[0][0]).To(BeEquivalentTo("Timestamp"))
			Expect(metrics[0][1]).To(BeEquivalentTo(api.METRICSTYPE_GPU_UTILIZATION))
			Expect(metrics[0][2]).To(BeEquivalentTo(api.METRICSTYPE_ALLOCATED_GPU))
			Expect(metrics[0][3]).To(BeEquivalentTo(api.METRICSTYPE_TOTAL_GPU))
			Expect(metrics[1][1]).To(Equal("1"))
			Expect(metrics[2][1]).To(Equal("2"))
			Expect(metrics[3][1]).To(Equal("3"))
			Expect(metrics[1][0]).To(Equal(stringTimestamp(baseTimestampUnix)))
			Expect(metrics[2][0]).To(Equal(stringTimestamp(baseTimestampUnix + 1000)))
			Expect(metrics[3][0]).To(Equal(stringTimestamp(baseTimestampUnix + 2000)))
		})
	})
})

func getDefaultMetricMockResult(baseValue int, baseTimestampUnix float64) openapi.GetMetrics200Response {
	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data: []openapi.Result{
			{
				Measurements: []openapi.Measurement{
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix)),
						Value:     fmt.Sprintf("%d", baseValue),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 1000),
						Value:     fmt.Sprintf("%d", baseValue+1),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 2000),
						Value:     fmt.Sprintf("%d", baseValue+2),
					},
				},
			},
		},
	}
}

func getMetricListMockResult(baseValue []int, baseTimestampUnix float64) openapi.GetMetrics200Response {
	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data: []openapi.Result{
			{
				MetricLabels: map[string]string{
					"range": "0",
				},
				Measurements: []openapi.Measurement{
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix)),
						Value:     fmt.Sprintf("%d", baseValue[0]),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 1000),
						Value:     fmt.Sprintf("%d", baseValue[0]+1),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 2000),
						Value:     fmt.Sprintf("%d", baseValue[0]+2),
					},
				},
			},
			{
				MetricLabels: map[string]string{
					"range": "1",
				},
				Measurements: []openapi.Measurement{
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix)),
						Value:     fmt.Sprintf("%d", baseValue[1]),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 1000),
						Value:     fmt.Sprintf("%d", baseValue[1]+1),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 2000),
						Value:     fmt.Sprintf("%d", baseValue[1]+2),
					},
				},
			},
			{
				MetricLabels: map[string]string{
					"range": "2",
				},
				Measurements: []openapi.Measurement{
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix)),
						Value:     fmt.Sprintf("%d", baseValue[2]),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 1000),
						Value:     fmt.Sprintf("%d", baseValue[2]+1),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 2000),
						Value:     fmt.Sprintf("%d", baseValue[2]+2),
					},
				},
			},
			{
				MetricLabels: map[string]string{
					"range": "3",
				},
				Measurements: []openapi.Measurement{
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix)),
						Value:     fmt.Sprintf("%d", baseValue[3]),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 1000),
						Value:     fmt.Sprintf("%d", baseValue[3]+1),
					},
					{
						Timestamp: time.UnixMilli(int64(baseTimestampUnix) + 2000),
						Value:     fmt.Sprintf("%d", baseValue[3]+2),
					},
				},
			},
		},
	}
}

func getDefaultMetricMockMultipleResults(baseValue int, baseTimestampUnix float64, numOfSamples, numberOfResults int) openapi.GetMetrics200Response {
	return getMetricMockResultInner(baseValue, baseTimestampUnix, false, numOfSamples, numberOfResults)
}

func getDefaultMetricMockMultipleResultsWithGroupBy(baseValue int, baseTimestampUnix float64, numOfSamples int, groupByLabel string, groupByValues []string) openapi.GetMetrics200Response {
	data := []openapi.Result{}
	for _, value := range groupByValues {
		data = append(data, openapi.Result{
			Measurements: generateMeasurements(baseValue, numOfSamples, baseTimestampUnix, false),
			MetricLabels: map[string]string{groupByLabel: value},
		})
	}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data:    data,
	}
}

func generateMeasurements(baseValue, numOfSamples int, baseTimestampUnix float64, constValues bool) []openapi.Measurement {
	baseValueStr := fmt.Sprintf("%d", baseValue)
	measurements := []openapi.Measurement{}
	for i := 0; i < numOfSamples; i++ {
		value := baseValueStr
		if !constValues {
			value = fmt.Sprintf("%d", baseValue+i)
		}
		measurements = append(measurements, openapi.Measurement{
			Timestamp: time.UnixMilli(int64(baseTimestampUnix) + int64(i*1000)),
			Value:     value,
		})
	}
	return measurements
}

func getMetricMockResultInner(baseValue int, baseTimestampUnix float64, constValues bool, numOfSamples int, numberOfResults int) openapi.GetMetrics200Response {
	data := []openapi.Result{}
	for i := 0; i < numberOfResults; i++ {
		data = append(data, openapi.Result{
			Measurements: generateMeasurements(baseValue, numOfSamples, baseTimestampUnix, constValues),
			MetricLabels: map[string]string{"gpu": fmt.Sprintf("%d", i)},
		})
	}

	return openapi.GetMetrics200Response{
		Status:  "success",
		Query:   "",
		IsRange: true,
		Data:    data,
	}
}

func stringTimestamp(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).String()
}
