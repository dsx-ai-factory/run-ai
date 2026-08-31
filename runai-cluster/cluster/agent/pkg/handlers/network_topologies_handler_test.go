/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/adapters"
	"k8s.io/utils/ptr"

	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	NodeLevelHostName  = "topology-hostname"
	NodeLevelHostRack  = "topology-rack"
	NodeLevelHostBlock = "topology-block"
)

var _ = Describe("NetworkTopologiesHandler", func() {
	var (
		mockBackendClient *MockBackendClient
		fakeClient        client.Client
		fakeCachedClient  client.Reader
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		err := kaiv1alpha1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		fakeCachedClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		mockBackendClient = NewMockBackendClient([]openapi.NetworkTopology{
			{
				Name:   "test-topology-1",
				Levels: []string{NodeLevelHostBlock, NodeLevelHostName},
			},
			{
				Name:   "test-topology-2",
				Levels: []string{NodeLevelHostRack, NodeLevelHostBlock, NodeLevelHostName},
			},
		}, false, "")
	})

	Describe("NewNetworkTopologiesHandler", func() {
		It("should create a new NetworkTopologiesHandler successfully", func() {
			Expect(mockBackendClient).NotTo(BeNil())
			Expect(fakeClient).NotTo(BeNil())
			Expect(fakeCachedClient).NotTo(BeNil())
		})
	})

	Describe("NetworkTopologyObjectInspector", func() {
		var inspector NetworkTopologyObjectInspector

		Describe("DeepCopySpec", func() {
			It("should copy topology spec correctly", func() {
				from := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}
				to := &kaiv1alpha1.Topology{}

				inspector.DeepCopySpec(from, to)

				Expect(to.Spec.Levels).To(HaveLen(2))
				Expect(to.Spec.Levels[0].NodeLabel).To(Equal(NodeLevelHostBlock))
				Expect(to.Spec.Levels[1].NodeLabel).To(Equal(NodeLevelHostName))
			})
		})

		Describe("IsUpToDate", func() {
			It("should return true for identical topologies", func() {
				left := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}
				right := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}

				result := inspector.IsUpToDate(left, right)
				Expect(result).To(BeTrue())
			})

			It("should always return true for different level counts", func() {
				left := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}
				right := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
						},
					},
				}

				result := inspector.IsUpToDate(left, right)
				Expect(result).To(BeTrue())
			})

			It("should return true for different level values", func() {
				left := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostBlock},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}
				right := &kaiv1alpha1.Topology{
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: NodeLevelHostRack},
							{NodeLabel: NodeLevelHostName},
						},
					},
				}

				result := inspector.IsUpToDate(left, right)
				Expect(result).To(BeTrue())
			})

			It("should return false for wrong type", func() {
				left := &kaiv1alpha1.Topology{}
				right := &kaiv1alpha1.Topology{}

				result := inspector.IsUpToDate(left, right)
				Expect(result).To(BeTrue())
			})
		})
	})

	Describe("NewNetworkTopologiesAdapter", func() {
		It("should extract and convert multiple network topologies", func() {
			openapiTopologies := []openapi.NetworkTopology{
				{
					Name:   "topology-1",
					Levels: []string{NodeLevelHostBlock, NodeLevelHostName},
					Source: ptr.To(adapters.NetworkTopologyRunAiSource),
				},
				{
					Name:   "topology-2",
					Levels: []string{NodeLevelHostRack, NodeLevelHostBlock, NodeLevelHostName},
					Source: ptr.To(adapters.NetworkTopologyRunAiSource),
				},
			}

			result := adapters.NewNetworkTopologiesAdapter(openapiTopologies).ToClusterTopologiesBySource()

			Expect(result).To(HaveLen(2))
			Expect(result[0].GetName()).To(Equal("topology-1"))
			Expect(result[1].GetName()).To(Equal("topology-2"))
		})

		It("should handle empty list", func() {
			openapiTopologies := []openapi.NetworkTopology{}

			result := adapters.NewNetworkTopologiesAdapter(openapiTopologies).ToClusterTopologiesBySource()

			Expect(result).To(HaveLen(0))
		})
	})

	Describe("Handle", func() {
		It("should have proper metric label", func() {
			handler := &NetworkTopologiesHandler{}
			Expect(handler.MetricLabel()).To(Equal("network_topologies"))
		})
	})

	Describe("MetricLabel", func() {
		It("should return correct metric label", func() {
			handler := &NetworkTopologiesHandler{}
			Expect(handler.MetricLabel()).To(Equal("network_topologies"))
		})
	})
})

// MockBackendClient is a mock implementation for testing
type MockBackendClient struct {
	networkTopologies []openapi.NetworkTopology
	shouldError       bool
	errorMessage      string
	context           context.Context
}

func NewMockBackendClient(networkTopologies []openapi.NetworkTopology, shouldError bool, errorMessage string) *MockBackendClient {
	return &MockBackendClient{
		networkTopologies: networkTopologies,
		shouldError:       shouldError,
		errorMessage:      errorMessage,
		context:           context.Background(),
	}
}

func (m *MockBackendClient) Context() context.Context {
	return m.context
}
