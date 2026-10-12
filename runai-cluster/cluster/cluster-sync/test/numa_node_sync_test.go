/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	runaiv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("NUMA node sync", func() {
	var (
		server       *ghttp.Server
		syncRequests chan openapi.NumaNodesSyncRequest
		syncPath     = fmt.Sprintf("/api/v1/clusters/%s/nodes/numa/sync", ClusterIdForTests)
	)

	BeforeEach(func() {
		server = CreateBasicRunaiBackend()
		server.AllowUnhandledRequests = true
		DeferCleanup(server.Close)
		syncRequests = make(chan openapi.NumaNodesSyncRequest, 10)
		DeferCleanup(func() { close(syncRequests) })

		server.RouteToHandler("POST", syncPath, func(writer http.ResponseWriter, request *http.Request) {
			var syncRequest openapi.NumaNodesSyncRequest
			Expect(json.NewDecoder(request.Body).Decode(&syncRequest)).To(Succeed())
			syncRequests <- syncRequest
			ghttp.RespondWith(http.StatusNoContent, nil)(writer, request)
		})
	})

	Context("when syncing a single physical node", func() {
		It("sends an add without deleteMissing", func() {
			nrt := buildNrt("node-a", "node-0")
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(nrt),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)

			Expect(handler.HandleAdd(nrt)).To(Succeed())
			request := <-syncRequests
			Expect(request.DeleteMissing).To(BeNil())
			Expect(request.From).To(BeNil())
			Expect(request.To).To(BeNil())
			Expect(request.Nodes).To(HaveLen(1))
			Expect(request.Nodes[0].Name).To(Equal("node-a"))
			Expect(request.Nodes[0].NumaNodes).To(HaveLen(1))
			Expect(request.Nodes[0].NumaNodes[0].Name).To(Equal("node-0"))
			Expect(request.Nodes[0].NumaNodes[0].TotalCpuCores).To(Equal(7.5))
			Expect(request.Nodes[0].NumaNodes[0].AllocatedCpuCores).To(Equal(1.5))
		})

		It("clears the physical node when its NRT is deleted", func() {
			nrt := buildNrt("node-a", "node-0")
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)

			Expect(handler.HandleDelete(nrt)).To(Succeed())
			request := <-syncRequests
			Expect(request.DeleteMissing).To(BeNil())
			Expect(request.Nodes).To(Equal([]openapi.NodeNumaForSync{{
				Name:      "node-a",
				NumaNodes: []openapi.NumaNodeDetails{},
			}}))
		})
	})

	Context("when syncing the full cluster", func() {
		It("includes only nodes whose nodepool has NUMA-aware scheduling enabled", func() {
			enabledNrt := buildNrt("node-b", "node-0")
			disabledNrt := buildNrt("node-a", "node-1")
			defaultNrt := buildNrt("node-c", "node-0")
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					enabledNrt,
					disabledNrt,
					defaultNrt,
					buildNodeForNumaSync("node-b", "numa-enabled"),
					buildNodeForNumaSync("node-a", "numa-disabled"),
					buildNodeForNumaSync("node-c", ""),
					buildNodePoolForNumaSync("numa-enabled", true),
					buildNodePoolForNumaSync("numa-disabled", false),
					buildNodePoolForNumaSync(runaiv1alpha1.DefaultNodePoolName, true),
				),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)

			handler.SyncAll()
			request := <-syncRequests
			Expect(request.DeleteMissing).NotTo(BeNil())
			Expect(*request.DeleteMissing).To(BeTrue())
			Expect(request.From).To(BeNil())
			Expect(request.To).To(BeNil())
			Expect(request.Nodes).To(HaveLen(2))
			Expect(request.Nodes[0].Name).To(Equal("node-b"))
			Expect(request.Nodes[1].Name).To(Equal("node-c"))
		})

		It("does not add an NRT already included in a successful full sync", func() {
			nrt := buildNrt("node-a", "node-0")
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					nrt,
					buildNodeForNumaSync("node-a", "numa-enabled"),
					buildNodePoolForNumaSync("numa-enabled", true),
				),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)

			handler.SyncAll()
			Eventually(syncRequests).Should(Receive())
			Expect(handler.HandleAdd(nrt)).To(Succeed())
			Consistently(syncRequests).ShouldNot(Receive())
		})
	})

	Context("when a node pool's NUMA-aware scheduling is toggled", func() {
		It("sends the pool's NUMA nodes once the setting is turned on", func() {
			nodePool := buildNodePoolForNumaSync("numa-pool", false)
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					buildNrt("node-a", "node-0"),
					buildNodeForNumaSync("node-a", "numa-pool"),
					nodePool,
				),
			)
			numaNodeHandler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			nodeHandler := handlers.NewNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			nodeHandler.SetNumaNodeHandler(numaNodeHandler)

			numaNodeHandler.SyncAll()
			Expect((<-syncRequests).Nodes).To(BeEmpty())

			nodePool.Spec.SchedulingShardConfig.Plugins["numa"] = kaiv1.PluginConfig{
				Enabled: ptr.To(true),
			}
			Expect(fakeClient.Update(context.Background(), nodePool)).To(Succeed())

			nodeHandler.SyncNumaNodesForNodepool("numa-pool")
			request := <-syncRequests
			Expect(request.Nodes).To(HaveLen(1))
			Expect(request.Nodes[0].Name).To(Equal("node-a"))
			Expect(request.Nodes[0].NumaNodes).To(HaveLen(1))
			Expect(request.Nodes[0].NumaNodes[0].Name).To(Equal("node-0"))
		})

		It("clears the pool's NUMA nodes once the setting is turned off", func() {
			nodePool := buildNodePoolForNumaSync("numa-pool", true)
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					buildNrt("node-a", "node-0"),
					buildNodeForNumaSync("node-a", "numa-pool"),
					nodePool,
				),
			)
			numaNodeHandler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			nodeHandler := handlers.NewNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			nodeHandler.SetNumaNodeHandler(numaNodeHandler)

			numaNodeHandler.SyncAll()
			Expect((<-syncRequests).Nodes).To(HaveLen(1))

			nodePool.Spec.SchedulingShardConfig.Plugins["numa"] = kaiv1.PluginConfig{
				Enabled: ptr.To(false),
			}
			Expect(fakeClient.Update(context.Background(), nodePool)).To(Succeed())

			nodeHandler.SyncNumaNodesForNodepool("numa-pool")
			request := <-syncRequests
			// deleteMissing lets the control plane drop the NUMA rows of the
			// nodes that fell out of the snapshot.
			Expect(request.DeleteMissing).NotTo(BeNil())
			Expect(*request.DeleteMissing).To(BeTrue())
			Expect(request.Nodes).To(BeEmpty())
		})

		It("does nothing when NUMA sync is disabled", func() {
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()),
			)
			nodeHandler := handlers.NewNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)

			nodeHandler.SyncNumaNodesForNodepool("numa-pool")
			Consistently(syncRequests).ShouldNot(Receive())
		})
	})

	Context("when the NRT informer receives an add event", func() {
		It("uses HandleAdd for an unlabeled node in the NUMA-enabled default nodepool", func() {
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					buildNodeForNumaSync("node-a", ""),
					buildNodePoolForNumaSync(runaiv1alpha1.DefaultNodePoolName, true),
				),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			controllers.AddNumaNodeHandlerToCache(fakeClient, handler)

			ctx, cancel := context.WithCancel(context.Background())
			stopper := make(chan struct{})
			DeferCleanup(func() {
				cancel()
				close(stopper)
				fakeClient.Shutdown()
			})
			fakeClient.Start(ctx, stopper)

			Expect(fakeClient.Create(ctx, buildNrt("node-a", "node-0"))).To(Succeed())
			Eventually(syncRequests).Should(Receive(WithTransform(
				func(request openapi.NumaNodesSyncRequest) string {
					return request.Nodes[0].Name
				},
				Equal("node-a"),
			)))
		})

		It("filters out a node whose nodepool does not have NUMA-aware scheduling enabled", func() {
			fakeClient := fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					buildNodeForNumaSync("node-a", "numa-disabled"),
					buildNodePoolForNumaSync("numa-disabled", false),
				),
			)
			handler := handlers.NewNumaNodeHandler(
				backendclient.NewBackendClient(server.URL(), nil),
				fakeClient,
			)
			controllers.AddNumaNodeHandlerToCache(fakeClient, handler)

			ctx, cancel := context.WithCancel(context.Background())
			stopper := make(chan struct{})
			DeferCleanup(func() {
				cancel()
				close(stopper)
				fakeClient.Shutdown()
			})
			fakeClient.Start(ctx, stopper)

			Expect(fakeClient.Create(ctx, buildNrt("node-a", "node-0"))).To(Succeed())
			Consistently(syncRequests).ShouldNot(Receive())
		})
	})
})

func buildNrt(nodeName string, numaNodeName string) *nrtv1alpha2.NodeResourceTopology {
	return &nrtv1alpha2.NodeResourceTopology{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		Zones: nrtv1alpha2.ZoneList{{
			Name: numaNodeName,
			Type: "Node",
			Resources: nrtv1alpha2.ResourceInfoList{
				{
					Name:        "cpu",
					Capacity:    resource.MustParse("8"),
					Allocatable: resource.MustParse("7500m"),
					Available:   resource.MustParse("6"),
				},
			},
		}},
	}
}

func buildNodeForNumaSync(name string, nodePoolName string) *corev1.Node {
	labels := map[string]string{}
	if nodePoolName != "" {
		labels[runaiv1alpha1.NodePoolNameLabel] = nodePoolName
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func buildNodePoolForNumaSync(name string, numaAwareScheduling bool) *kainodepool.NodePool {
	return &kainodepool.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kainodepool.NodePoolSpec{
			SchedulingShardConfig: &kainodepool.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{
					"numa": {Enabled: &numaAwareScheduling},
				},
			},
		},
	}
}
