/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"net/http"
	"time"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node NUMA service", Ordered, func() {
	const tenantId = int32(40872)

	var (
		testCtx     context.Context
		nodeService *services.NodeService
		cluster     *models.Cluster
		nodes       map[string]*models.Node
	)

	numaDetails := func(name string, cpu float64, memory int64, gpus int32) api.NumaNodeDetails {
		return api.NumaNodeDetails{
			Name:                    name,
			TotalCpuCores:           cpu,
			AllocatedCpuCores:       1,
			TotalCpuMemoryBytes:     memory,
			AllocatedCpuMemoryBytes: 1024,
			TotalGpus:               gpus,
			AllocatedGpus:           0,
		}
	}

	syncNuma := func(request api.NumaNodesSyncRequest) api.ImplResponse {
		response, err := nodeService.SyncNumaNodes(testCtx, cluster.ID.String(), request)
		Expect(err).NotTo(HaveOccurred())
		return response
	}

	BeforeAll(func() {
		testCtx = context.WithValue(context.Background(), runaictx.CfTenantId, "40872")
		testCtx = context.WithValue(testCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{"40872"},
		})
		nodeService = services.NewNodeService(&bc_mocks.TenantSettingsClient{}, nil)

		var dbErr *api.Error
		cluster, dbErr = dao.CreateCluster(testCtx, &models.Cluster{
			Name:     clusterNameGenerator.Generate("node-numa"),
			TenantId: tenantId,
			Domain:   "numa.test",
			Version:  "3.0.0",
		})
		Expect(dbErr).To(BeNil())

		nodes = map[string]*models.Node{}
		policy := api.TOPOLOGYMANAGERPOLICY_RESTRICTED
		for _, nodeName := range []string{"node-a", "node-b", "node-c"} {
			response, err := nodeService.UpsertNode(testCtx, cluster.ID.String(), nodeName, api.NodeInfo{
				Status:    "Ready",
				NodePool:  "default",
				CreatedAt: time.Now(),
				Numa: &api.NumaInfo{
					TopologyManagerPolicy: &policy,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			node, getErr := dao.GetNodeInternal(
				testCtx, tenantId, cluster.ID.String(), nodeName, []string{"id", "name", "tenant_id", "cluster_uuid"},
			)
			Expect(getErr).To(BeNil())
			nodes[nodeName] = node
		}
	})

	AfterAll(func() {
		if cluster == nil {
			return
		}
		Expect(dao.DeleteClusterNodes(testCtx, tenantId, cluster.ID.String())).To(Succeed())
		Expect(dao.DeleteCluster(testCtx, cluster.ID.String(), tenantId)).To(BeNil())
	})

	It("syncs snapshots and returns extended node details", func() {
		response := syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name: "node-a",
				NumaNodes: []api.NumaNodeDetails{
					numaDetails("node-0", 24.5, 64*1024*1024*1024, 4),
					numaDetails("node-1", 24.5, 64*1024*1024*1024, 4),
				},
			}},
		})
		Expect(response.Code).To(Equal(http.StatusNoContent))

		getResponse, err := nodeService.GetNode(testCtx, nodes["node-a"].ID.String())
		Expect(err).NotTo(HaveOccurred())
		Expect(getResponse.Code).To(Equal(http.StatusOK))
		extended, ok := getResponse.Body.(api.NodeExtended)
		Expect(ok).To(BeTrue())
		Expect(extended.Name).To(Equal("node-a"))
		Expect(extended.Numa.TopologyManagerPolicy).NotTo(BeNil())
		Expect(*extended.Numa.TopologyManagerPolicy).To(Equal(api.TOPOLOGYMANAGERPOLICY_RESTRICTED))
		Expect(extended.Numa.Nodes).To(HaveLen(2))
		Expect(extended.Numa.Nodes[0].Name).To(Equal("node-0"))
		Expect(extended.Numa.Nodes[0].TotalCpuCores).To(Equal(24.5))
		Expect(extended.Numa.Nodes[1].TotalGpus).To(Equal(int32(4)))
	})

	It("preserves successful node snapshots when another node fails", func() {
		response := syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{
				{
					Name:      "node-b",
					NumaNodes: []api.NumaNodeDetails{numaDetails("node-0", 8, 16*1024*1024*1024, 1)},
				},
				{
					Name:      "missing-node",
					NumaNodes: []api.NumaNodeDetails{},
				},
			},
		})
		Expect(response.Code).To(Equal(http.StatusInternalServerError))

		numaNodes, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-b"].ID)
		Expect(dbErr).To(BeNil())
		Expect(numaNodes).To(HaveLen(1))
		Expect(numaNodes[0].Name).To(Equal("node-0"))
	})

	It("upserts a node snapshot, deletes stale NUMA nodes, and clears an empty snapshot", func() {
		Expect(syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name: "node-a",
				NumaNodes: []api.NumaNodeDetails{
					numaDetails("kept", 8, 16*1024*1024*1024, 1),
					numaDetails("stale", 4, 8*1024*1024*1024, 0),
				},
			}},
		}).Code).To(Equal(http.StatusNoContent))

		numaNodes, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-a"].ID)
		Expect(dbErr).To(BeNil())
		Expect(numaNodes).To(HaveLen(2))
		keptId := numaNodes[0].ID

		Expect(syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name: "node-a",
				NumaNodes: []api.NumaNodeDetails{
					numaDetails("kept", 12, 32*1024*1024*1024, 2),
					numaDetails("new", 6, 8*1024*1024*1024, 0),
				},
			}},
		}).Code).To(Equal(http.StatusNoContent))

		numaNodes, dbErr = dao.GetNumaNodes(testCtx, tenantId, nodes["node-a"].ID)
		Expect(dbErr).To(BeNil())
		Expect(numaNodes).To(HaveLen(2))
		Expect(numaNodes[0].Name).To(Equal("kept"))
		Expect(numaNodes[0].ID).To(Equal(keptId))
		Expect(numaNodes[0].TotalCpuCores).To(Equal(12.0))
		Expect(numaNodes[1].Name).To(Equal("new"))

		Expect(syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name:      "node-a",
				NumaNodes: []api.NumaNodeDetails{},
			}},
		}).Code).To(Equal(http.StatusNoContent))
		numaNodes, dbErr = dao.GetNumaNodes(testCtx, tenantId, nodes["node-a"].ID)
		Expect(dbErr).To(BeNil())
		Expect(numaNodes).To(BeEmpty())
	})

	It("rejects an inverted full-sync range before persisting snapshots", func() {
		response := syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name:      "node-c",
				NumaNodes: []api.NumaNodeDetails{numaDetails("not-synced", 8, 16*1024*1024*1024, 1)},
			}},
			DeleteMissing: true,
			From:          "node-z",
			To:            "node-a",
		})
		Expect(response.Code).To(Equal(http.StatusBadRequest))

		numaNodes, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-c"].ID)
		Expect(dbErr).To(BeNil())
		Expect(numaNodes).To(BeEmpty())
	})

	It("deletes only missing node snapshots inside the requested full-sync range", func() {
		for _, nodeName := range []string{"node-a", "node-b", "node-c"} {
			Expect(syncNuma(api.NumaNodesSyncRequest{
				Nodes: []api.NodeNumaForSync{{
					Name:      nodeName,
					NumaNodes: []api.NumaNodeDetails{numaDetails("node-0", 4, 8*1024*1024*1024, 0)},
				}},
			}).Code).To(Equal(http.StatusNoContent))
		}

		response := syncNuma(api.NumaNodesSyncRequest{
			Nodes: []api.NodeNumaForSync{{
				Name:      "node-b",
				NumaNodes: []api.NumaNodeDetails{numaDetails("new-node-0", 6, 8*1024*1024*1024, 0)},
			}},
			DeleteMissing: true,
			From:          "node-a",
			To:            "node-b",
		})
		Expect(response.Code).To(Equal(http.StatusNoContent))

		nodeANuma, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-a"].ID)
		Expect(dbErr).To(BeNil())
		Expect(nodeANuma).To(BeEmpty())
		nodeBNuma, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-b"].ID)
		Expect(dbErr).To(BeNil())
		Expect(nodeBNuma).To(HaveLen(1))
		Expect(nodeBNuma[0].Name).To(Equal("new-node-0"))
		nodeCNuma, dbErr := dao.GetNumaNodes(testCtx, tenantId, nodes["node-c"].ID)
		Expect(dbErr).To(BeNil())
		Expect(nodeCNuma).To(HaveLen(1))
	})
})
