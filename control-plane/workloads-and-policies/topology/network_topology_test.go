/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology_test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Network Topology", func() {
	var (
		ctx                context.Context
		randoms            *testutils.RandomValues
		mockClusterService *clusterservice.SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = context.Background()
		mockClusterService = clusterservice.NewSimpleMock()
	})

	Describe("CheckNodePoolsTopology", func() {
		Context("Sanity", func() {
			It("client is nil", func() {
				requestedNameSet := []string{"pool-1", "pool-2"}

				_, err := clusterservice.GetNodePoolsByNames(ctx, nil, randoms.ClusterId, requestedNameSet)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no client provided"))
			})
		})

		Context("When there are no topologies", func() {
			It("should return error when no node pool names provided", func() {
				requestedNameSet := []string{}

				mockClusterService.WithNodePools(randoms.ClusterId)

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err).To(Equal(topology.ErrNoNodePoolsNames))
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})

			It("should return error when node pools cannot be located", func() {
				// Mock empty response
				mockClusterService.WithNodePools(randoms.ClusterId)

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err.Error()).To(ContainSubstring("do not exist: pool-1, pool-2"))
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})

			It("should return error when node pools partially be located", func() {
				mockClusterService.WithNodePools(randoms.ClusterId, "pool-1")

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err.Error()).To(ContainSubstring("do not exist: pool-2"))
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})

			It("should topology name be empty and found is false", func() {
				mockClusterService.WithNodePools(randoms.ClusterId, "pool-1", "pool-2")

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})
		})

		Context("When there are topologies", func() {
			It("should return error when there are node pool with topology and another without", func() {
				mockClusterService.
					WithNodePoolsWithTopology(randoms.ClusterId, "only-here", "pool-1").
					WithNodePools(randoms.ClusterId, "pool-2")

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err).To(Equal(topology.ErrTopologyNameDoesNotMatch))
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})

			It("should return error when there are node pools are not homogeneous with topology", func() {
				mockClusterService.
					WithNodePoolsWithTopology(randoms.ClusterId, "one-here", "pool-1").
					WithNodePoolsWithTopology(randoms.ClusterId, "another-there", "pool-2")

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err).To(Equal(topology.ErrTopologyNameDoesNotMatch))
				Expect(found).To(BeFalse())
				Expect(topologyName).To(Equal(""))
			})

			It("node pools has the same topology, should return found true with topology name", func() {
				mockClusterService.
					WithNodePoolsWithTopology(randoms.ClusterId, "same-topology", "pool-1").
					WithNodePoolsWithTopology(randoms.ClusterId, "same-topology", "pool-2")

				requestedNameSet := []string{"pool-1", "pool-2"}

				actualNodePools, err := clusterservice.GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, requestedNameSet)
				Expect(err).To(Not(HaveOccurred()))

				found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, requestedNameSet)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(topologyName).To(Equal("same-topology"))
			})
		})
	})

	Describe("GetByName", func() {
		Context("When there are topologies", func() {
			It("should return error when there are topology without", func() {
				_, err := topology.GetByName(ctx, mockClusterService, "pool-1")
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(topology.ErrTopologyNotFound))
			})
		})
	})
})
