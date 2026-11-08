/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node labels", func() {
	const tenantID = int32(19451)
	var ctx context.Context
	var clusterID string
	var nodeService *services.NodeService

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "19451")
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{"19451"}})
		cluster, err := dao.CreateCluster(ctx, &models.Cluster{
			Name: clusterNameGenerator.Generate("node-labels"), TenantId: tenantID,
			Domain: "node-labels.example.com", Version: "3.1",
		})
		Expect(err).To(BeNil())
		clusterID = cluster.ID.String()
		DeferCleanup(func() {
			nodesErr := dao.DeleteClusterNodes(ctx, tenantID, clusterID)
			clusterErr := dao.DeleteCluster(ctx, clusterID, tenantID)
			Expect(nodesErr).To(BeNil())
			Expect(clusterErr).To(BeNil())
		})
		nodeService = services.NewNodeService(nil, nil)
	})

	nodeInfoWithLabels := func(labels map[string]string) api.NodeInfo {
		return api.NodeInfo{Status: "Ready", NodePool: "default", Labels: labels}
	}
	saveNodeLabels := func(labels map[string]string) {
		GinkgoHelper()
		response, err := nodeService.UpsertNode(ctx, clusterID, "node-1", nodeInfoWithLabels(labels))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	}
	syncNodes := func(nodes ...api.NodeForSync) {
		GinkgoHelper()
		response, err := nodeService.SyncNodes(ctx, clusterID, api.NodesSyncRequest{NodesForSync: nodes})
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	}
	expectStoredLabels := func(expectedLabelsByNode map[string]map[string]string) {
		GinkgoHelper()
		permittedScopes := authorization_api.PermittedScopes{Clusters: []string{clusterID}}
		nodes, dbErr := dao.GetNodesByClusterUuid(ctx, tenantID, clusterID, permittedScopes, nil, "", "", 0, 0, "")
		Expect(dbErr).To(BeNil())
		actualLabelsByNode := make(map[string]map[string]string, len(*nodes))
		for _, node := range *nodes {
			actualLabelsByNode[node.Name] = map[string]string(node.Labels)
		}
		Expect(actualLabelsByNode).To(Equal(expectedLabelsByNode))
	}

	It("stores every label received from node sync", func() {
		labels := map[string]string{
			"gpu-type":                       "a100",
			"example.com/gpu_model":          "A100-SXM4.80GB",
			"node-role.kubernetes.io/worker": "",
		}
		saveNodeLabels(labels)
		expectStoredLabels(map[string]map[string]string{"node-1": labels})
	})

	DescribeTable("updating labels for one node",
		func(updatedLabels map[string]string) {
			By("Saving the node's original labels")
			saveNodeLabels(map[string]string{"gpu-type": "a100", "removed-label": "old"})

			By("Sending a node update containing its complete new label set")
			saveNodeLabels(updatedLabels)

			By("Checking that only the new labels are stored, with no old labels left behind")
			expectStoredLabels(map[string]map[string]string{"node-1": updatedLabels})
		},
		Entry("replaces old labels with the new label set", map[string]string{"gpu-type": "h100", "new-label": "value"}),
		Entry("clears labels when the new label set is empty", map[string]string{}),
	)

	It("replaces labels for each node when multiple nodes are sent in one sync request", func() {
		node1Labels := map[string]string{"gpu-type": "a100", "removed-label": "old"}
		node2Labels := map[string]string{"gpu-type": "h100"}

		By("Sending two nodes with different labels in one sync request")
		syncNodes(
			api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(node1Labels)},
			api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(node2Labels)},
		)

		By("Checking that both nodes were saved with their own labels")
		expectStoredLabels(map[string]map[string]string{"node-1": node1Labels, "node-2": node2Labels})

		By("Sending another sync request that changes node-1's labels and clears node-2's labels")
		updatedNode1Labels := map[string]string{"gpu-type": "l40", "new-label": "value"}
		syncNodes(
			api.NodeForSync{Name: "node-1", NodeInfo: nodeInfoWithLabels(updatedNode1Labels)},
			api.NodeForSync{Name: "node-2", NodeInfo: nodeInfoWithLabels(map[string]string{})},
		)

		By("Checking that node-1 has only its new labels and node-2 has no labels")
		expectStoredLabels(map[string]map[string]string{"node-1": updatedNode1Labels, "node-2": {}})
	})

	It("accepts a new node from an older sender that omits labels", func() {
		By("Creating a node without the optional labels field")
		saveNodeLabels(nil)
		expectStoredLabels(map[string]map[string]string{"node-1": {}})
	})

	It("clears stored labels when a node update omits the labels field", func() {
		By("Saving labels before receiving an update without labels")
		saveNodeLabels(map[string]string{"gpu-type": "a100"})
		saveNodeLabels(nil)

		By("Checking that the old labels were replaced with an empty set")
		expectStoredLabels(map[string]map[string]string{"node-1": {}})
	})

	It("clears stored labels when a full sync omits the labels field", func() {
		By("Saving labels before receiving a full sync without labels")
		saveNodeLabels(map[string]string{"gpu-type": "a100"})
		syncNodes(api.NodeForSync{
			Name:     "node-1",
			NodeInfo: api.NodeInfo{Status: "Ready", NodePool: "default"},
		})
		expectStoredLabels(map[string]map[string]string{"node-1": {}})
	})

})
