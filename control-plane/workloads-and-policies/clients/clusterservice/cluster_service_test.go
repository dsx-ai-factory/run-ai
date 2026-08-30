package clusterservice

import (
	"context"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Network Topology", func() {
	var (
		ctx                context.Context
		randoms            *testutils.RandomValues
		mockClusterService *SimpleMock
	)

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		ctx = context.Background()
		mockClusterService = NewSimpleMock()
	})

	Describe("GetNodePoolsByNames", func() {
		It("should return node pools by names", func() {
			// Mock node pools
			mockClusterService.WithNodePools(randoms.ClusterId, "pool-1", "pool-2", "pool-3")

			result, err := GetNodePoolsByNames(ctx, mockClusterService, randoms.ClusterId, []string{"pool-1", "pool-3"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(2))
			Expect(result[0].GetName()).To(Equal("pool-1"))
			Expect(result[1].GetName()).To(Equal("pool-3"))
		})
	})

	Describe("GetTopologyByName", func() {
		It("should return topology by name", func() {
			// Mock topology
			topology := &openapi.NetworkTopology{
				Name:   "test-topology",
				Levels: []string{"level1", "level2"},
			}
			mockClusterService.WithTopology("test-topology", topology)

			result, err := GetTopologyByName(ctx, mockClusterService, "test-topology")
			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.GetName()).To(Equal("test-topology"))
			Expect(result.GetLevels()).To(HaveLen(2))
		})

		It("should return error when topology not found", func() {
			result, err := GetTopologyByName(ctx, mockClusterService, "non-existent-topology")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("topology not found"))
			Expect(result).To(BeNil())
		})
	})
})
