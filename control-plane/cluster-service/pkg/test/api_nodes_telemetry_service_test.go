/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/metrics"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	TestTenantInt = 1
	TestTenant    = "1"
	ReadyStatus   = "Ready"
)

var (
	groupByKeyToMetricLabel = map[string]string{}
	baseTimestampUnix       = time.Now()
)

func setContext() context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{TestTenant},
	})

	return ctx
}

func setMgmtContext() context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "2")
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		System: ptr.To(true),
	})

	return ctx
}

var _ = Describe("Nodes API tests", Ordered, func() {
	type Mocks struct {
		tenantSettingsClient *bc_mocks.TenantSettingsClient
	}

	ctx := setContext()
	mgmtCtx := setMgmtContext()
	var mocks *Mocks

	testCreatedNodes := []models.Node{}
	var nodeService *services.NodeService
	var cluster1Id string
	var cluster2Id string
	var cluster3Id string
	var cluster4Id string
	nodepool1 := "default"
	nodepool2 := "pool2"
	nodepool3 := "pool3"
	nodepool4 := "pool4"

	BeforeAll(func() {
		mocks = &Mocks{
			tenantSettingsClient: &bc_mocks.TenantSettingsClient{},
		}

		nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)
		groupByKeyToMetricLabel = nodeService.ModelFieldToMetricLabelMap
		cluster1, _ := dao.CreateCluster(ctx, &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-1",
			Version:  "2.16",
			TenantId: 1,
		})
		cluster2, _ := dao.CreateCluster(ctx, &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-2",
			Version:  "0.0.0-1001030.master-982a2c2",
			TenantId: 1,
		})
		cluster3, _ := dao.CreateCluster(ctx, &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-3",
			Version:  "2.15.1",
			TenantId: 1,
		})
		cluster4, _ := dao.CreateCluster(ctx, &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-4",
			Version:  "2.15.1",
			TenantId: 2,
		})
		cluster1Id = cluster1.ID.String()
		cluster2Id = cluster2.ID.String()
		cluster3Id = cluster3.ID.String()
		cluster4Id = cluster4.ID.String()
		node := addNode(ctx, cluster1Id, "node1", nodepool1, 0, ReadyStatus, *nodeService)
		node1 := addNode(ctx, cluster1Id, "node2", nodepool1, 1, ReadyStatus, *nodeService)
		node2 := addNode(ctx, cluster1Id, "node3", nodepool1, 1, "NotReady", *nodeService)
		node3 := addNode(ctx, cluster1Id, "node4", nodepool2, 1, ReadyStatus, *nodeService)
		node4 := addNode(ctx, cluster2Id, "node5", nodepool3, 3, ReadyStatus, *nodeService)
		node5 := addNode(mgmtCtx, cluster4Id, "node6", nodepool1, 3, ReadyStatus, *nodeService)

		testCreatedNodes = append(testCreatedNodes, *node, *node1, *node2, *node3, *node4, *node5)
	})

	BeforeAll(func() {
		metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
			runaimetrics.RunaiFreeGpuCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"24"}}}, nil, ""),
			runaimetrics.RunaiFreeGpuCountPerClusterMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"24"}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.ClusterId, cluster2Id}},
				{{"16"}, {metrics.ClusterId, cluster3Id}},
			}, nil, ""),
			runaimetrics.RunaiFreeGpuCountPerNodepoolMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"14"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"10"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
			}, []string{cluster1Id}, ""),
			runaimetrics.RunaiFreeGpuCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"12"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"16"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}, []string{cluster1Id}, ""),
			runaimetrics.RunaiGpuNodesCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"40"}}}, nil, ""),

			runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"51"}}}, nil, ""),
			runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"24"}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.ClusterId, cluster2Id}},
				{{"16"}, {metrics.ClusterId, cluster3Id}},
			}, nil, ""),

			runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"15"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"6"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}, []string{cluster1Id}, ""),
			runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
				{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"12"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"16"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}, nil, ""),
		})
	})

	AfterAll(func() {
		CleanTables()
	})

	Context("ready gpu nodes", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("nodepool filter without cluster filter", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3")) // one of the nodes has no gpus + one of the nodes is not ready
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("cluster filter nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter cluster, nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool, metrics.ClusterId)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter, node groups", func() {
			groups := []string{metrics.Node}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}},
				{{"1"}, {metrics.Node, "node4"}},
				{{"1"}, {metrics.Node, "node5"}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter, node and nodepool groups", func() {
			groups := []string{metrics.Node, metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster4Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("1"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - no filter, cluster groups", func() {
			groups := []string{metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.ClusterId, cluster2Id}},
				{{"1"}, {metrics.ClusterId, cluster4Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})

		Context("cluster not connected - without version", func() {
			var clsuter5Id string

			BeforeAll(func() {
				cluster5, _ := dao.CreateCluster(ctx, &models.Cluster{
					Common:   models.Common{},
					Name:     "cluster-5",
					TenantId: 1,
				})
				clsuter5Id = cluster5.ID.String()
			})
			AfterAll(func() {
				dao.DeleteCluster(ctx, clsuter5Id, 1)
			})

			It("no filters at all", func() {
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", []string{}, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("3")) // one of the nodes has no gpus + one of the nodes is not ready
				Expect(len(telemetryValues[0].Groups)).To(Equal(0))
			})
		})
	})

	Context("fully free nodes", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("metric filter without cluster filter", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("no filter no groups", func() {
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES, "", "", nil, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("51")) // ?????
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("cluster filter nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"15"}, {metrics.Nodepool, nodepool1}},
				{{"11"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter cluster, nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool, metrics.ClusterId)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"15"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter, node groups", func() {
			groups := []string{metrics.Node}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"0"}, {metrics.Node, "node1"}},
				{{"12"}, {metrics.Node, "node2"}},
				{{"0"}, {metrics.Node, "node3"}},
				{{"11"}, {metrics.Node, "node4"}},
				{{"16"}, {metrics.Node, "node5"}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})

		Context("cluster not connected - without version", func() {
			var clsuter5Id string

			BeforeAll(func() {
				cluster5, _ := dao.CreateCluster(ctx, &models.Cluster{
					Common:   models.Common{},
					Name:     "cluster-5",
					TenantId: 1,
				})
				clsuter5Id = cluster5.ID.String()
			})
			AfterAll(func() {
				dao.DeleteCluster(ctx, clsuter5Id, 1)
			})

			It("no filters at all", func() {
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", []string{}, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("3")) // one of the nodes has no gpus + one of the nodes is not ready
				Expect(len(telemetryValues[0].Groups)).To(Equal(0))
			})
		})
	})

	Context("telemetry - csv", func() {
		BeforeAll(func() {
			ctx = setContext()
			ctx = context.WithValue(ctx, runaictx.CfAccept, "text/csv")
			mgmtCtx = setMgmtContext()
			mgmtCtx = context.WithValue(mgmtCtx, runaictx.CfAccept, "text/csv")
		})
		AfterAll(func() {
			ctx = setContext()
			mgmtCtx = setMgmtContext()
		})
		It("ready gpus", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response, ok := apiResponse.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			telemetry := response.Data

			Expect(len(telemetry)).To(Equal(3))
			Expect(telemetry[0][0]).To(Equal("value"))
			Expect(telemetry[0][1]).To(Equal(metrics.Nodepool + ".value"))
			Expect(telemetry[0][2]).To(Equal(metrics.Nodepool + ".name"))
			Expect(telemetry[1][0]).To(Equal("1"))
			Expect(telemetry[1][1]).To(Equal(nodepool1))
			Expect(telemetry[2][0]).To(Equal("1"))
			Expect(telemetry[2][1]).To(Equal(nodepool2))
		})
		It("mgmt ctx - ready gpus", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response, ok := apiResponse.Body.(runaihttp.CsvFile)
			Expect(ok).To(BeTrue())
			telemetry := response.Data

			Expect(len(telemetry)).To(Equal(3))
			Expect(telemetry[0][0]).To(Equal("value"))
			Expect(telemetry[0][1]).To(Equal(metrics.Nodepool + ".value"))
			Expect(telemetry[0][2]).To(Equal(metrics.Nodepool + ".name"))
			Expect(telemetry[1][0]).To(Equal("1"))
			Expect(telemetry[1][1]).To(Equal(nodepool1))
			Expect(telemetry[2][0]).To(Equal("1"))
			Expect(telemetry[2][1]).To(Equal(nodepool2))
		})
	})

	Context("ready gpus", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("nodepool filter with no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("5"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("cluster filter nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter cluster, nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool, metrics.ClusterId)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter, node and cluster groups", func() {
			groups := []string{metrics.Node, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter, node and nodepool groups", func() {
			groups := []string{metrics.Node, metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, cluster2Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"3"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_READY_GPUS, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - tenant filter, node and cluster groups", func() {
			groups := []string{metrics.Node, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{TestTenant})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
	})

	Context("total gpu nodes", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("nodepool filter without cluster filter", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("4"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("cluster filter nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"2"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter cluster, nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool, metrics.ClusterId)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(3))

			expectedTelemetryValuesRaw := [][][]string{
				{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter, node and cluster groups", func() {
			groups := []string{metrics.Node, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter, node and nodepool groups", func() {
			groups := []string{metrics.Node, metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - no filter, node and cluster groups", func() {
			groups := []string{metrics.Node, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
				{{"1"}, {metrics.Node, "node6"}, {metrics.ClusterId, cluster4Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
	})

	Context("total gpus", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("nodepool filter without cluster filter", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("6"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("cluster filter nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"2"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter cluster, nodepool groups", func() {
			var groups []string
			groups = append(groups, metrics.Nodepool, metrics.ClusterId)
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("no filter, node and cluster groups", func() {
			groups := []string{metrics.Node, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster1Id}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter, node and nodepool groups", func() {
			groups := []string{metrics.Node, metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"1"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}},
				{{"1"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - cluster filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster3Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("0"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("mgmt tenant - no filter, cluster groups", func() {
			groups := []string{metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"3"}, {metrics.ClusterId, cluster1Id}},
				{{"3"}, {metrics.ClusterId, cluster2Id}},
				{{"3"}, {metrics.ClusterId, cluster4Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
	})

	Context("free gpus", func() {
		BeforeAll(func() {
			ctx = setContext()
		})
		It("nodepool filter without cluster filter", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", nodepool1, groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("mgmt ctx - not impl telemetry type", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(mgmtCtx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusNotImplemented))
		})
		It("no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("24"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("no filter cluster group", func() {
			groups := []string{metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"24"}, {metrics.ClusterId, cluster1Id}},
				{{"11"}, {metrics.ClusterId, cluster2Id}},
				{{"16"}, {metrics.ClusterId, cluster3Id}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		It("cluster filter nodepool groups", func() {
			groups := []string{metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"14"}, {metrics.Nodepool, nodepool1}},
				{{"10"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		//TODO: uncomment once we have a way of resetting the mock server
		/*
			It("nodepool and cluster filters, node group", func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"12"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"11"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
						{{"16"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, []string{cluster1Id}, nodepool1),
				})

				groups := []string{services.Node}
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, cluster1Id, nodepool1, groups, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values

				expectedTelemetryValuesRaw := [][][]string{
					{{"0"}, {metrics.Node, "node1"}},
					{{"12"}, {metrics.Node, "node2"}},
					{{"0"}, {metrics.Node, "node3"}},
				}
				validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
			})
				It("no filter cluster, nodepool groups", func() {
					metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
						metrics.RunaiFreeGpuCountPerNodepoolMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
							{{"14"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
							{{"10"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
							{{"24"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
						}, nil, ""),
					})

					groups := []string{metrics.Nodepool, metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"14"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"10"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
						{{"24"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
			It("no filter, node and cluster groups", func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"12"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster1Id}},
						{{"11"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster1Id}},
						{{"16"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, []string{cluster1Id}, ""),
				})

				groups := []string{metrics.Node, metrics.ClusterId}
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values

				expectedTelemetryValuesRaw := [][][]string{
					{{"0"}, {metrics.Node, "node1"}, {metrics.ClusterId, cluster1Id}},
					{{"12"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster1Id}},
					{{"0"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster1Id}},
					{{"11"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster1Id}},
					{{"16"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
				}
				validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
			})
		*/
		It("cluster filter, node and nodepool groups", func() {
			groups := []string{metrics.Node, metrics.Nodepool}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values

			expectedTelemetryValuesRaw := [][][]string{
				{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}},
				{{"12"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}},
				{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}},
				{{"11"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}},
			}
			validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
		})
		//TODO: uncomment once we have a way of resetting the mock server
		/*
			It("cluster filter no groups", func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"16"}}}, []string{cluster3Id}, ""),
				})

				var groups []string
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, cluster3Id, "", groups, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("16"))
				Expect(len(telemetryValues[0].Groups)).To(Equal(0))
			})
		*/
	})

	Context("permitted scopes one cluster", func() {
		BeforeAll(func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Clusters: []string{cluster2Id},
			})
		})
		It("ready gpu nodes - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("1"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("ready gpus - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("total gpu nodes - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("1"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("total gpus - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		//TODO: uncomment once we have a way of resetting the mock server
		/*
			It("free gpus - no filter no groups", func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerClusterMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"24"}, {metrics.ClusterId, cluster1Id}},
						{{"11"}, {metrics.ClusterId, cluster2Id}},
						{{"16"}, {metrics.ClusterId, cluster3Id}},
					}, []string{cluster2Id}, ""),
				})

				var groups []string
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("11"))
				Expect(len(telemetryValues[0].Groups)).To(Equal(0))

			})
		*/
	})

	Context("permitted scopes two clusters", func() {
		BeforeAll(func() {
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Clusters: []string{cluster2Id, cluster1Id},
			})
		})
		It("ready gpu nodes - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3")) // and not 4, since one of the nodes has no GPUs
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("ready gpus - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_READY_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("5"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("total gpu nodes - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("4"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("total gpus - no filter no groups", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("6"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		//TODO: uncomment once we have a way of resetting the mock server
		/*
			It("free gpus - no filter no groups", func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"35"}}}, []string{cluster1Id, cluster2Id}, ""),
				})

				var groups []string
				apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
				Expect(apiResponse.Code).To(Equal(http.StatusOK))
				telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
				Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
				Expect(telemetry.Timestamp).NotTo(BeNil())
				telemetryValues := telemetry.Values
				Expect(len(telemetryValues)).To(Equal(1))
				Expect(telemetryValues[0].Value).To(Equal("35"))
				Expect(len(telemetryValues[0].Groups)).To(Equal(0))

			})
		*/
	})

	Context("api validations", func() {
		It("invalid telemetry type", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, "invalid", "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("invalid group by", func() {
			groups := []string{"invalid"}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))

			groups = append(groups, metrics.ClusterId)
			apiResponse, _ = nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("3 elements in group by", func() {
			groups := []string{metrics.Node, metrics.Nodepool, metrics.ClusterId}
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))

			apiResponse, _ = nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("early version", func() {
		BeforeAll(func() {
			ctx = setContext()
			cluster4, _ := dao.CreateCluster(ctx, &models.Cluster{
				Common:   models.Common{},
				Name:     "cluster-4",
				Version:  "2.14.1",
				TenantId: 1,
			})
			cluster4Id = cluster4.ID.String()
			node5 := addNode(ctx, cluster4Id, "node6", nodepool4, 3, ReadyStatus, *nodeService)
			testCreatedNodes = append(testCreatedNodes, *node5)
		})

		//TODO: uncomment once we have a way of resetting the mock server
		/*
			BeforeAll(func() {
				metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
					metrics.RunaiFreeGpuCountPerTenantMetric{}.GetName():  getDefaultPerClusterMockResult([][][]string{{{"24"}}}, nil, ""),
					metrics.RunaiGpuNodesCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"40"}}}, nil, ""),
					metrics.RunaiGpuNodesCountPerClusterMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"40"}, {metrics.ClusterId, cluster1Id}},
						{{"20"}, {metrics.ClusterId, cluster4Id}},
					}, nil, ""),
					metrics.RunaiGpuNodesCountPerNodepoolMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
						{{"20"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, []string{cluster4Id}, ""),
					metrics.RunaiGpuNodesCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"16"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"10"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, nil, ""),
					metrics.RunaiGpuCountPerTenantMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{{{"6"}}}, nil, ""),
					metrics.RunaiGpuCountPerClusterMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"6"}, {metrics.ClusterId, cluster1Id}},
						{{"3"}, {metrics.ClusterId, cluster4Id}},
					}, nil, ""),
					metrics.RunaiGpuCountPerNodepoolMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
						{{"3"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, []string{cluster4Id}, ""),
					metrics.RunaiGpuCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
						{{"1"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"2"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"3"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"4"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
					}, nil, ""),
				})
			})
		*/

		It("get ready gpu nodes telemetry without cluster filter - should be not implemented", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("get ready gpu nodes telemetry with cluster filter - should be ok", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_READY_GPU_NODES, cluster1Id, "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
		})
		It("get free gpus without cluster filter - should be ok", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_FREE_GPUS, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_FREE_GPUS)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("24"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})
		It("get total gpu nodes telemetry without cluster filter - should be ok", func() {
			var groups []string
			apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
			Expect(apiResponse.Code).To(Equal(http.StatusOK))

			telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
			Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("40"))
			Expect(len(telemetryValues[0].Groups)).To(Equal(0))
		})

		//TODO: uncomment once we have a way of resetting the mock server
		/*
			Context("total gpu nodes with old cluster - from metrics instead of db", func() {
				It("no cluster filter", func() {
					var groups []string
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))

					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values
					Expect(len(telemetryValues)).To(Equal(1))
					Expect(telemetryValues[0].Value).To(Equal("40"))
					Expect(len(telemetryValues[0].Groups)).To(Equal(0))
				})
				It("cluster filter nodepool groups", func() {
					var groups []string
					groups = append(groups, metrics.Nodepool)
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}},
						{{"5"}, {metrics.Nodepool, nodepool2}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("no filter cluster, nodepool groups", func() {
					//TODO: add a new mock without the cluster4Id filter
					groups := []string{metrics.Nodepool, metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("no filter, node and cluster groups", func() {
					groups := []string{metrics.Node, metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"0"}, {metrics.Node, "node1"}, {metrics.ClusterId, cluster4Id}},
						{{"16"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster4Id}},
						{{"0"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster4Id}},
						{{"10"}, {metrics.Node, "node4"}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {metrics.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("cluster filter, node and nodepool groups", func() {
					metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
						metrics.RunaiGpuNodesCountPerNodeMetric{}.GetName(): getDefaultPerClusterMockResult([][][]string{
							{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
							{{"16"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
							{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
							{{"10"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
							{{"5"}, {metrics.Node, "node5"}, {metrics.Nodepool, nodepool3}, {metrics.ClusterId, cluster2Id}},
						}, []string{cluster4Id}, ""),
					})

					groups := []string{metrics.Node, metrics.Nodepool}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"0"}, {metrics.Node, "node1"}, {metrics.Nodepool, nodepool1}},
						{{"16"}, {metrics.Node, "node2"}, {metrics.Nodepool, nodepool1}},
						{{"0"}, {metrics.Node, "node3"}, {metrics.Nodepool, nodepool1}},
						{{"10"}, {metrics.Node, "node4"}, {metrics.Nodepool, nodepool2}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("cluster filter no groups", func() {
					var groups []string
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values
					Expect(len(telemetryValues)).To(Equal(1))
					Expect(telemetryValues[0].Value).To(Equal("0"))
					Expect(len(telemetryValues[0].Groups)).To(Equal(0))
				})
				It("no filter, cluster groups", func() {
					groups := []string{metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPU_NODES, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"40"}, {metrics.ClusterId, cluster1Id}},
						{{"20"}, {metrics.ClusterId, cluster4Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
			})

			Context("total gpus with old cluster - from metrics instead of db", func() {
				It("no filter no groups", func() {
					var groups []string
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))

					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values
					Expect(len(telemetryValues)).To(Equal(1))
					Expect(telemetryValues[0].Value).To(Equal("6"))
					Expect(len(telemetryValues[0].Groups)).To(Equal(0))
				})
				It("cluster filter nodepool groups", func() {
					var groups []string
					groups = append(groups, metrics.Nodepool)
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}},
						{{"1"}, {metrics.Nodepool, nodepool2}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("no filter cluster, nodepool groups", func() {
					//TODO: add a new mock without the cluster4Id filter
					groups := []string{metrics.Nodepool, metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, cluster4Id, "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"2"}, {metrics.Nodepool, nodepool1}, {metrics.ClusterId, cluster4Id}},
						{{"1"}, {metrics.Nodepool, nodepool2}, {metrics.ClusterId, cluster4Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("no filter, node and cluster groups", func() {
					groups := []string{metrics.Node, metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"1"}, {metrics.Node, "node1"}, {metrics.ClusterId, cluster4Id}},
						{{"2"}, {metrics.Node, "node2"}, {metrics.ClusterId, cluster4Id}},
						{{"3"}, {metrics.Node, "node3"}, {metrics.ClusterId, cluster4Id}},
						{{"4"}, {services.Node, "node4"}, {metrics.ClusterId, cluster4Id}},
						{{"5"}, {services.Node, "node5"}, {metrics.ClusterId, cluster2Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				It("no filter, cluster groups", func() {
					groups := []string{metrics.ClusterId}
					apiResponse, _ := nodeService.GetNodeTelemetry(ctx, api.NODETELEMETRYTYPE_TOTAL_GPUS, "", "", groups, []string{})
					Expect(apiResponse.Code).To(Equal(http.StatusOK))
					telemetry := httpResponseBodyToTelemetryResponse(apiResponse)
					Expect(telemetry.Type).To(Equal(string(api.NODETELEMETRYTYPE_TOTAL_GPUS)))
					Expect(telemetry.Timestamp).NotTo(BeNil())
					telemetryValues := telemetry.Values

					expectedTelemetryValuesRaw := [][][]string{
						{{"6"}, {metrics.ClusterId, cluster1Id}},
						{{"3"}, {metrics.ClusterId, cluster4Id}},
					}
					validateTelemetryValues(telemetryValues, expectedTelemetryValuesRaw)
				})
				//TODO: add "cluster filter, node and nodepool groups" and "cluster filter no groups"
			})
		*/

		AfterAll(func() {
			dao.DeleteNode(ctx, TestTenantInt, cluster4Id, "node-6")
			dao.DeleteCluster(ctx, cluster4Id, 1)
		})
	})
})

// expectedTelemetryValuesRaw - a list of telemetry values, each element is a api.TelemetryResponseValuesInner, in each element:
// the first []string{} should be of length 1 and is the value,
// the rest are the groups,
// in each group the []string{} should be of length 2 - the first is the key, the second is the value.
func validateTelemetryValues(actualTelemetryValues []api.TelemetryResponseValuesInner, expectedTelemetryValuesRaw [][][]string) {
	expectedTelemetryValues, errMsg := buildTelemetryGroups(expectedTelemetryValuesRaw)
	ExpectWithOffset(1, errMsg).To(BeEmpty())

	ExpectWithOffset(1, len(actualTelemetryValues)).To(Equal(len(expectedTelemetryValues)),
		"Expected telemetry values to be of length %d, but got %d", len(expectedTelemetryValues), len(actualTelemetryValues))

	for _, expected := range expectedTelemetryValues {
		found := false
		for _, actual := range actualTelemetryValues {
			if actual.Value == expected.Value && validateGroups(actual.Groups, expected.Groups) {
				found = true
				break
			}
		}

		ExpectWithOffset(1, found).To(BeTrue(),
			"Expected telemetry value to be found in actual, expected: %+v, all actual: %+v", expected, actualTelemetryValues)
	}
}

func validateGroups(actualGroups, expectedGroups []api.TelemetryResponseValuesInnerGroupsInner) bool {
	if len(actualGroups) != len(expectedGroups) {
		return false
	}

	for _, expectedGroup := range expectedGroups {
		found := false
		for _, actualGroup := range actualGroups {
			if actualGroup.Key == expectedGroup.Key && actualGroup.Value == expectedGroup.Value {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}

func buildTelemetryGroups(expectedTelemetryValues [][][]string) ([]api.TelemetryResponseValuesInner, string) {
	telemetryGroups := []api.TelemetryResponseValuesInner{}
	for _, expectedTelemetryValueAndGroup := range expectedTelemetryValues {
		if len(expectedTelemetryValueAndGroup) == 0 {
			return nil, "Expected telemetry element to be of length >=1"
		}

		valueArray := expectedTelemetryValueAndGroup[0]
		if len(valueArray) != 1 {
			return nil, "Expected telemetry value to be of length 1"
		}
		value := valueArray[0]

		groups := []api.TelemetryResponseValuesInnerGroupsInner{}
		for _, group := range expectedTelemetryValueAndGroup[1:] {
			groups = append(groups, api.TelemetryResponseValuesInnerGroupsInner{Key: group[0], Value: group[1]})
		}
		telemetryGroups = append(telemetryGroups, api.TelemetryResponseValuesInner{Value: value, Groups: groups})
	}
	return telemetryGroups, ""
}

func addNode(ctx context.Context, clusterId string, nodeName string, nodepool string, gpuCount int32, status string, service services.NodeService) *models.Node {
	nodeInfo := api.NodeInfo{
		NodePool: nodepool,
		Status:   status,
		GpuInfo: &api.GpuInfo{
			GpuType:  "type",
			GpuCount: gpuCount,
		},
	}
	result, _ := service.UpsertNode(ctx, clusterId, nodeName, nodeInfo)
	Expect(result.Code).To(Equal(http.StatusNoContent))
	return &models.Node{
		Name:        nodeName,
		ClusterUuid: clusterId,
	}
}

func httpResponseBodyToTelemetryResponse(httpResponse api.ImplResponse) api.TelemetryResponse {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.TelemetryResponse{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}

func getDefaultPerClusterMockResult(expectedTelemetryValuesRaw [][][]string, clustersFilter []string, nodepoolFilter string) openapi.GetMetrics200Response {
	results := []openapi.Result{}
	for _, expectedTelemetryValueAndGroup := range expectedTelemetryValuesRaw {
		value := expectedTelemetryValueAndGroup[0][0]

		metricLabels := map[string]string{}
		filterOut := false
		for _, group := range expectedTelemetryValueAndGroup[1:] {
			if (group[0] == metrics.ClusterId && clustersFilter != nil && !slices.Contains(clustersFilter, group[1])) ||
				(group[0] == metrics.Nodepool && nodepoolFilter != "" && group[1] != nodepoolFilter) {
				filterOut = true
				break
			}

			metricLabels[groupByKeyToMetricLabel[group[0]]] = group[1]
		}

		if filterOut {
			continue
		}

		results = append(results, openapi.Result{
			Measurements: []openapi.Measurement{{Value: value, Timestamp: baseTimestampUnix}},
			MetricLabels: metricLabels,
		})
	}

	return openapi.GetMetrics200Response{
		Data: results,
	}
}
