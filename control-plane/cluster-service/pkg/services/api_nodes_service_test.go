/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
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
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

var _ = Describe("NodeService excluded-node filtering", Ordered, func() {
	const tenantId = int32(1)

	var (
		nodeService *services.NodeService
		ctx         context.Context
		testCluster *models.Cluster
		clusterId   string
	)

	// nodeForSync builds a sync payload entry with the given node-pool.
	nodeForSync := func(name, nodePool string) api.NodeForSync {
		return api.NodeForSync{
			Name: name,
			NodeInfo: api.NodeInfo{
				Status:   "Ready",
				NodePool: nodePool,
			},
		}
	}

	// clusterNodeNames returns the names of all nodes currently stored for the cluster.
	clusterNodeNames := func() []string {
		nodes, err := dao.GetNodesByClusterUuid(ctx, tenantId, clusterId,
			authorization_api.PermittedScopes{System: ptr.To(true)}, nil, "", "", 0, 0, "")
		Expect(err).To(BeNil())
		names := make([]string, 0, len(*nodes))
		for _, node := range *nodes {
			names = append(names, node.Name)
		}
		return names
	}

	BeforeAll(func() {
		tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
		nodeService = services.NewNodeService(tenantSettingsClient, nil)

		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "1")
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{"1"},
		})

		var apiErr *api.Error
		testCluster, apiErr = dao.CreateCluster(ctx, &models.Cluster{
			Name:     "test-cluster-excluded-nodes",
			TenantId: tenantId,
			Domain:   "test.run.ai",
			Version:  "v2.24",
		})
		Expect(apiErr).To(BeNil())
		clusterId = testCluster.ID.String()
	})

	AfterAll(func() {
		if testCluster != nil {
			_ = dao.DeleteClusterNodes(ctx, tenantId, clusterId)
			dao.DeleteCluster(ctx, clusterId, tenantId)
		}
	})

	Context("SyncNodes", func() {
		It("persists managed nodes and skips excluded ones", func() {
			req := api.NodesSyncRequest{
				NodesForSync: []api.NodeForSync{
					nodeForSync("managed-a", "default"),
					nodeForSync("excluded-x", services.ExcludedNodepoolName),
					nodeForSync("managed-b", "gpu-pool"),
				},
			}

			response, err := nodeService.SyncNodes(ctx, clusterId, req)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			Expect(clusterNodeNames()).To(ConsistOf("managed-a", "managed-b"))
		})

		It("cleans up a stale excluded node already in the DB (self-heal)", func() {
			// Simulate a stale excluded row that an older agent persisted directly.
			_, daoErr := dao.UpsertNode(ctx, &models.Node{
				Name:        "stale-excluded",
				ClusterUuid: clusterId,
				TenantId:    tenantId,
				Status:      "Ready",
				NodePool:    services.ExcludedNodepoolName,
			})
			Expect(daoErr).To(BeNil())
			Expect(clusterNodeNames()).To(ContainElement("stale-excluded"))

			// A full resync that reports only the managed node should reconcile the
			// whole cluster: excluded + unreferenced nodes get pruned.
			req := api.NodesSyncRequest{
				NodesForSync: []api.NodeForSync{
					nodeForSync("managed-a", "default"),
				},
			}

			response, err := nodeService.SyncNodes(ctx, clusterId, req)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			Expect(clusterNodeNames()).To(ConsistOf("managed-a"))
		})
	})

	Context("UpsertNode", func() {
		It("persists a managed node", func() {
			response, err := nodeService.UpsertNode(ctx, clusterId, "managed-single",
				api.NodeInfo{Status: "Ready", NodePool: "default"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			Expect(clusterNodeNames()).To(ContainElement("managed-single"))
		})
	})
})
