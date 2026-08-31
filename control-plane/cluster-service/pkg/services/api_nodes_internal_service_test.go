/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"net/http"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var _ = Describe("NodeService GetNodeInternal tests", Ordered, func() {
	var (
		nodeService           *services.NodeService
		ctxInternal           context.Context
		testInternalCluster   *models.Cluster
		testInternalClusterId string
		testNode              *models.Node
	)

	BeforeAll(func() {
		tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
		nodeService = services.NewNodeService(tenantSettingsClient, nil)

		// Create a simple context with tenant ID
		ctxInternal = context.WithValue(context.Background(), runaictx.CfTenantId, "1")

		// Create test cluster
		var err *api.Error
		testInternalCluster, err = dao.CreateCluster(ctxInternal, &models.Cluster{
			Name:     "test-cluster-internal",
			TenantId: 1,
			Domain:   "test.run.ai",
			Version:  "v2.24",
		})
		Expect(err).To(BeNil())
		testInternalClusterId = testInternalCluster.ID.String()

		// Create test node with topology labels
		networkTopologyLabelValues := pq.StringArray{"zone/rack-1", "rack/shelf-2", "shelf/node-3"}
		testNode = &models.Node{
			Name:                       "test-node-1",
			ClusterUuid:                testInternalClusterId,
			TenantId:                   1,
			Status:                     "Ready",
			NodePool:                   "default",
			GpuType:                    "A100",
			GpuCount:                   8,
			NetworkTopologyLabelValues: &networkTopologyLabelValues,
		}

		_, err = dao.UpsertNode(ctxInternal, testNode)
		Expect(err).To(BeNil())
	})

	AfterAll(func() {
		// Clean up
		if testNode != nil {
			dao.DeleteNode(ctxInternal, testNode.TenantId, testInternalClusterId, testNode.Name)
		}
		if testInternalCluster != nil {
			dao.DeleteCluster(ctxInternal, testInternalClusterId, testInternalCluster.TenantId)
		}
	})

	Context("GetNodeInternal", func() {
		It("should successfully retrieve node with topology labels", func() {
			response, err := nodeService.GetNodeInternal(ctxInternal, testInternalClusterId, "test-node-1")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodeInternal, ok := response.Body.(api.NodeInternal)
			Expect(ok).To(BeTrue(), "Response body should be of type NodeInternal")
			Expect(nodeInternal.Name).To(Equal("test-node-1"))
			Expect(nodeInternal.ClusterUuid).To(Equal(testInternalClusterId))
			Expect(nodeInternal.NetworkTopologyLabelValues).NotTo(BeNil())
			Expect(*nodeInternal.NetworkTopologyLabelValues).To(HaveLen(3))
			Expect(*nodeInternal.NetworkTopologyLabelValues).To(ContainElements("zone/rack-1", "rack/shelf-2", "shelf/node-3"))
		})

		It("should successfully retrieve node without topology labels", func() {
			// Create node without topology labels
			nodeWithoutLabels := &models.Node{
				Name:                       "test-node-2",
				ClusterUuid:                testInternalClusterId,
				TenantId:                   1,
				Status:                     "Ready",
				NodePool:                   "default",
				GpuType:                    "A100",
				GpuCount:                   4,
				NetworkTopologyLabelValues: nil,
			}

			_, daoErr := dao.UpsertNode(ctxInternal, nodeWithoutLabels)
			Expect(daoErr).To(BeNil())
			defer dao.DeleteNode(ctxInternal, nodeWithoutLabels.TenantId, testInternalClusterId, nodeWithoutLabels.Name)

			response, err := nodeService.GetNodeInternal(ctxInternal, testInternalClusterId, "test-node-2")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodeInternal, ok := response.Body.(api.NodeInternal)
			Expect(ok).To(BeTrue(), "Response body should be of type NodeInternal")
			Expect(nodeInternal.Name).To(Equal("test-node-2"))
			Expect(nodeInternal.ClusterUuid).To(Equal(testInternalClusterId))
			Expect(nodeInternal.NetworkTopologyLabelValues).To(BeNil())
		})

		It("should return error for non-existent node", func() {
			response, err := nodeService.GetNodeInternal(ctxInternal, testInternalClusterId, "non-existent-node")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))

			apiError, ok := response.Body.(api.Error)
			Expect(ok).To(BeTrue(), "Response body should be of type Error")
			Expect(apiError.Message).To(ContainSubstring("Failed to get node"))
		})

		It("should return error for non-existent cluster", func() {
			response, err := nodeService.GetNodeInternal(ctxInternal, "non-existent-cluster-uuid", "test-node-1")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))

			apiError, ok := response.Body.(api.Error)
			Expect(ok).To(BeTrue(), "Response body should be of type Error")
			Expect(apiError.Message).To(ContainSubstring("Failed to get node"))
		})

		It("should retrieve node with empty topology labels array", func() {
			// Create node with empty topology labels array
			emptyNetworkTopologyLabelValues := pq.StringArray{}
			nodeWithEmptyLabels := &models.Node{
				Name:                       "test-node-3",
				ClusterUuid:                testInternalClusterId,
				TenantId:                   1,
				Status:                     "NotReady",
				NodePool:                   "gpu-pool",
				GpuType:                    "V100",
				GpuCount:                   2,
				NetworkTopologyLabelValues: &emptyNetworkTopologyLabelValues,
			}

			_, daoErr := dao.UpsertNode(ctxInternal, nodeWithEmptyLabels)
			Expect(daoErr).To(BeNil())
			defer dao.DeleteNode(ctxInternal, nodeWithEmptyLabels.TenantId, testInternalClusterId, nodeWithEmptyLabels.Name)

			response, err := nodeService.GetNodeInternal(ctxInternal, testInternalClusterId, "test-node-3")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodeInternal, ok := response.Body.(api.NodeInternal)
			Expect(ok).To(BeTrue(), "Response body should be of type NodeInternal")
			Expect(nodeInternal.Name).To(Equal("test-node-3"))
			Expect(nodeInternal.ClusterUuid).To(Equal(testInternalClusterId))
			Expect(nodeInternal.NetworkTopologyLabelValues).NotTo(BeNil())
			Expect(*nodeInternal.NetworkTopologyLabelValues).To(HaveLen(0))
		})

		It("should only return internal fields (name, cluster_uuid, network_topology_label_values)", func() {
			response, err := nodeService.GetNodeInternal(ctxInternal, testInternalClusterId, "test-node-1")

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodeInternal, ok := response.Body.(api.NodeInternal)
			Expect(ok).To(BeTrue(), "Response body should be of type NodeInternal")

			Expect(nodeInternal.Name).NotTo(BeEmpty())
			Expect(nodeInternal.ClusterUuid).NotTo(BeEmpty())
			Expect(*nodeInternal.NetworkTopologyLabelValues).To(ContainElements("zone/rack-1", "rack/shelf-2", "shelf/node-3"))
		})
	})
})
