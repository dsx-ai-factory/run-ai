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
	"regexp"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	sdkkaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"github.com/onsi/gomega/ghttp"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// nodePoolTwin mirrors a legacy node pool fixture onto the kai.resources CRD, so these specs
// resolve the same topology under both values of the compile-time UseKaiNodePools flag.
// On that CRD the topology name lives in spec.preferredNetworkTopologyName.
// TODO: delete along with the legacy fixtures once commonconstants.UseKaiNodePools=true
func nodePoolTwin(nodePool *v1alpha1.NodePool) *sdkkaiv1alpha1.NodePool {
	return &sdkkaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: nodePool.Name},
		Spec: sdkkaiv1alpha1.NodePoolSpec{
			PreferredNetworkTopologyName: nodePool.Spec.NetworkTopologyName,
		},
	}
}

var _ = Describe("NodeHandler", func() {
	var nodesSyncPath = fmt.Sprintf("/api/v1/clusters/%s/nodes/sync", ClusterIdForTests)
	var nodeHandler *handlers.NodeHandler

	nodePath, err := regexp.Compile(fmt.Sprintf("/api/v1/clusters/%s/nodes/.+", ClusterIdForTests))
	Expect(err).To(BeNil())
	Expect(nodePath.MatchString(fmt.Sprintf("/api/v1/clusters/%s/nodes/runai-worker-node2", ClusterIdForTests))).To(BeTrue())

	var ctx context.Context
	var cancel context.CancelFunc
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient

	var stopper chan struct{}
	var numberOfsyncedNodes chan int
	var updatedNodeName chan string
	var deletedNodeName chan string
	var fakeCachedClient cachedclient.CachedClient

	BeforeEach(func() {
		numberOfsyncedNodes = make(chan int, 1)
		DeferCleanup(func() { close(numberOfsyncedNodes) })
		updatedNodeName = make(chan string, 1)
		DeferCleanup(func() { close(updatedNodeName) })
		deletedNodeName = make(chan string, 1)
		DeferCleanup(func() { close(deletedNodeName) })
		GinkgoWriter.Println("Creating server for test...")
		server = CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		server.AllowUnhandledRequests = true
		stopper = make(chan struct{}, 1)
		DeferCleanup(func() { close(stopper) })
	})

	Context("runai backend is available", func() {
		var existingNodeName = "node-2"
		var existingNode *corev1.Node

		var nodeNameToAdd = "another-node"
		var nodeToAdd *corev1.Node
		var excludedNodeToAdd *corev1.Node

		BeforeEach(func() {
			existingNode = BuildNode(existingNodeName)
			nodeToAdd = BuildNode(nodeNameToAdd)
			excludedNodeToAdd = BuildExcludedNode("excluded-node")

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", nodesSyncPath, func(writer http.ResponseWriter, request *http.Request) {
				var nodesSyncRequest swagger.NodesSyncRequest
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&nodesSyncRequest)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				numberOfsyncedNodes <- len(nodesSyncRequest.NodesForSync)
			})
			// Update node info mock
			server.RouteToHandler("PUT", nodePath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				nodeNameStr := strings.TrimPrefix(request.URL.Path, fmt.Sprintf("/api/v1/clusters/%s/nodes/", ClusterIdForTests))
				updatedNodeName <- nodeNameStr
			})
			// Delete node info mock
			server.RouteToHandler("DELETE", nodePath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				nodeNameStr := strings.TrimPrefix(request.URL.Path, fmt.Sprintf("/api/v1/clusters/%s/nodes/", ClusterIdForTests))
				deletedNodeName <- nodeNameStr
			})

			ctx, cancel = context.WithCancel(context.TODO())
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Starting nodes event handling and sync")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(existingNode)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)
			controllers.AddNodeHandlerToCache(backendClient, fakeCachedClient, nodeHandler)

			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
			fakeCachedClient.Start(ctx, stopper)
		})

		JustAfterEach(func() {
			cancel()
		})

		It("Test node add + sync", func() {
			Eventually(numberOfsyncedNodes).Should(Receive(Equal(1)))
			err := nodeHandler.HandleAdd(nodeToAdd)
			Expect(err).To(BeNil())
			Eventually(updatedNodeName).WithTimeout(3 * time.Second).Should(Receive(Equal(nodeNameToAdd)))
		})
		It("Test node update + sync", func() {
			Eventually(numberOfsyncedNodes).Should(Receive(Equal(1)))
			existingNode.Labels = map[string]string{v1alpha1.NodePoolNameLabel: "pool-a"}
			err := nodeHandler.HandleUpdate(existingNode)
			Expect(err).To(BeNil())
			Eventually(updatedNodeName).WithTimeout(3 * time.Second).Should(Receive(Equal(existingNodeName)))
		})
		It("Test node update without change", func() {
			existingNode.Status.Conditions = append(existingNode.Status.Conditions, corev1.NodeCondition{
				Type:    corev1.NodeMemoryPressure,
				Status:  corev1.ConditionTrue,
				Reason:  "",
				Message: "",
			})
			err := nodeHandler.HandleUpdate(existingNode)
			Expect(err).To(BeNil())
			Eventually(updatedNodeName).WithTimeout(3 * time.Second).ShouldNot(Receive(Equal(existingNodeName)))
		})
		It("Test node delete + sync", func() {
			Eventually(numberOfsyncedNodes).Should(Receive(Equal(1)))
			err := nodeHandler.HandleDelete(existingNode)
			Expect(err).To(BeNil())
			Eventually(deletedNodeName).WithTimeout(3 * time.Second).Should(Receive(Equal(existingNodeName)))
		})
		It("Test node add of a non run:ai node does not add node", func() {
			err := fakeCachedClient.Create(ctx, excludedNodeToAdd)
			Expect(err).To(BeNil())
			Eventually(updatedNodeName).WithTimeout(3 * time.Second).ShouldNot(Receive())
		})
	})

	Context("NodeChanged detects network topology label changes", func() {
		var nodeHandler *handlers.NodeHandler

		var nodePoolName = "pool-with-topology"
		var topologyName = "test-topology"

		var nodePool *v1alpha1.NodePool
		var topology *kaiv1alpha1.Topology

		BeforeEach(func() {
			nodePool = &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodePoolName,
				},
				Spec: v1alpha1.NodePoolSpec{
					NetworkTopologyName: topologyName,
				},
			}
			topology = &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: topologyName,
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "topology.kubernetes.io/zone"},
						{NodeLabel: "topology.kubernetes.io/rack"},
					},
				},
			}
		})

		JustBeforeEach(func() {
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), topology)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(nil, fakeCachedClient)
		})

		It("should detect change when topology label value changes", func() {
			oldNode := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel:    nodePoolName,
						"topology.kubernetes.io/zone": "us-west-1",
						"topology.kubernetes.io/rack": "rack-1",
					},
				},
			}
			newNode := oldNode.DeepCopy()
			newNode.Labels["topology.kubernetes.io/rack"] = "rack-2"

			Expect(nodeHandler.NodeChanged(oldNode, newNode)).To(BeTrue())
		})

		It("should not detect change when topology labels are the same", func() {
			oldNode := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel:    nodePoolName,
						"topology.kubernetes.io/zone": "us-west-1",
						"topology.kubernetes.io/rack": "rack-1",
					},
				},
			}
			newNode := oldNode.DeepCopy()

			Expect(nodeHandler.NodeChanged(oldNode, newNode)).To(BeFalse())
		})

		It("should detect change when topology label is added to node", func() {
			oldNode := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel:    nodePoolName,
						"topology.kubernetes.io/zone": "us-west-1",
						"topology.kubernetes.io/rack": "rack-1",
					},
				},
			}
			newNode := oldNode.DeepCopy()
			newNode.Labels["topology.kubernetes.io/rack"] = "rack-2"

			Expect(nodeHandler.NodeChanged(oldNode, newNode)).To(BeTrue())
		})

		It("should detect change when the topology manager policy annotation changes", func() {
			oldNode := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "node-1",
					Labels:      map[string]string{v1alpha1.NodePoolNameLabel: nodePoolName},
					Annotations: map[string]string{sdkkaiv1alpha1.AnnotationTopologyManagerPolicy: "none"},
				},
			}
			newNode := oldNode.DeepCopy()
			newNode.Annotations[sdkkaiv1alpha1.AnnotationTopologyManagerPolicy] = "single-numa-node"

			Expect(nodeHandler.NodeChanged(oldNode, newNode)).To(BeTrue())
		})
	})

	Context("SyncNodesForNodepool on topology change", func() {
		var syncRequest chan swagger.NodesSyncRequest

		gpuPoolName := "gpu-pool"
		zoneLabelKey := "topology.kubernetes.io/zone"
		rackLabelKey := "topology.kubernetes.io/rack"
		regionLabelKey := "topology.kubernetes.io/region"

		BeforeEach(func() {
			syncRequest = make(chan swagger.NodesSyncRequest, 1)

			server = CreateBasicRunaiBackend()
			DeferCleanup(func() { server.Close() })
			server.AllowUnhandledRequests = true

			server.RouteToHandler("POST", nodesSyncPath, func(writer http.ResponseWriter, request *http.Request) {
				var req swagger.NodesSyncRequest
				err := json.NewDecoder(request.Body).Decode(&req)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				syncRequest <- req
			})

			backendClient = backendclient.NewBackendClient(server.URL(), nil)
		})

		It("should sync only labeled nodes for a non-default nodepool with correct topology labels", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: gpuPoolName},
				Spec:       v1alpha1.NodePoolSpec{NetworkTopologyName: "topo-1"},
			}
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{Name: "topo-1"},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: zoneLabelKey},
						{NodeLabel: rackLabelKey},
					},
				},
			}
			node1 := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gpu-node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: gpuPoolName,
						zoneLabelKey:               "us-west-1",
						rackLabelKey:               "rack-a",
					},
				},
			}
			node2 := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gpu-node-2",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: gpuPoolName,
						zoneLabelKey:               "us-east-1",
						rackLabelKey:               "rack-b",
					},
				},
			}
			nodeOtherPool := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "other-pool-node",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: "other-pool",
						zoneLabelKey:               "us-east-1",
					},
				},
			}
			nodeNoLabel := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default-node",
				},
			}

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), topology, node1, node2, nodeOtherPool, nodeNoLabel)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)

			nodeHandler.SyncNodesForNodepool(gpuPoolName)

			var req swagger.NodesSyncRequest
			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(HaveLen(4))

			syncedNames := map[string]bool{}
			for _, n := range req.NodesForSync {
				syncedNames[n.Name] = true
				if n.Name == "gpu-node-1" {
					Expect(n.NodeInfo.NetworkTopologyLabels).To(HaveLen(2))
					labelMap := map[string]string{}
					for _, l := range n.NodeInfo.NetworkTopologyLabels {
						labelMap[l.LabelKey] = l.LabelValue
					}
					Expect(labelMap[zoneLabelKey]).To(Equal("us-west-1"))
					Expect(labelMap[rackLabelKey]).To(Equal("rack-a"))
				} else if n.Name == "gpu-node-2" {
					Expect(n.NodeInfo.NetworkTopologyLabels).To(HaveLen(2))
					labelMap := map[string]string{}
					for _, l := range n.NodeInfo.NetworkTopologyLabels {
						labelMap[l.LabelKey] = l.LabelValue
					}
					Expect(labelMap[zoneLabelKey]).To(Equal("us-east-1"))
					Expect(labelMap[rackLabelKey]).To(Equal("rack-b"))
				}
			}
			Expect(syncedNames).To(HaveKey("gpu-node-1"))
			Expect(syncedNames).To(HaveKey("gpu-node-2"))
			Expect(syncedNames).To(HaveKey("other-pool-node"))
			Expect(syncedNames).To(HaveKey("default-node"))
		})

		It("should sync only unlabeled nodes for the default nodepool", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
				Spec:       v1alpha1.NodePoolSpec{NetworkTopologyName: "default-topo"},
			}
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{Name: "default-topo"},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: zoneLabelKey},
					},
				},
			}
			defaultNode1 := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default-node-1",
					Labels: map[string]string{
						zoneLabelKey: "zone-a",
					},
				},
			}
			defaultNode2 := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default-node-2",
					Labels: map[string]string{
						zoneLabelKey: "zone-b",
					},
				},
			}
			labeledNode := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "labeled-node",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: "other-pool",
						zoneLabelKey:               "zone-c",
					},
				},
			}

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), topology, defaultNode1, defaultNode2, labeledNode)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)

			nodeHandler.SyncNodesForNodepool("default")

			var req swagger.NodesSyncRequest
			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(HaveLen(3))

			syncedNames := map[string]bool{}
			for _, n := range req.NodesForSync {
				syncedNames[n.Name] = true
				if n.Name == "default-node-1" || n.Name == "default-node-2" {
					Expect(n.NodeInfo.NetworkTopologyLabels).To(HaveLen(1))
					Expect(n.NodeInfo.NetworkTopologyLabels[0].LabelKey).To(Equal(zoneLabelKey))
				}
			}
			Expect(syncedNames).To(HaveKey("default-node-1"))
			Expect(syncedNames).To(HaveKey("default-node-2"))
			Expect(syncedNames).To(HaveKey("labeled-node"))
		})

		It("should not send a sync request when no nodes match the nodepool", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-pool"},
				Spec:       v1alpha1.NodePoolSpec{NetworkTopologyName: "some-topo"},
			}
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{Name: "some-topo"},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: zoneLabelKey},
					},
				},
			}

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), topology)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)

			nodeHandler.SyncNodesForNodepool("empty-pool")

			var req swagger.NodesSyncRequest
			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(BeEmpty())
		})

		It("should sync nodes with no topology labels when topology is removed", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: gpuPoolName},
				Spec:       v1alpha1.NodePoolSpec{NetworkTopologyName: ""},
			}
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gpu-node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: gpuPoolName,
						zoneLabelKey:               "us-west-1",
					},
				},
			}

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), node)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)

			nodeHandler.SyncNodesForNodepool(gpuPoolName)

			var req swagger.NodesSyncRequest
			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(HaveLen(1))
			Expect(req.NodesForSync[0].Name).To(Equal("gpu-node-1"))
			Expect(req.NodesForSync[0].NodeInfo.NetworkTopologyLabels).To(BeNil())
		})

		It("should use new topology labels when topology is changed", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: gpuPoolName},
				Spec:       v1alpha1.NodePoolSpec{NetworkTopologyName: "old-topo"},
			}
			oldTopology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{Name: "old-topo"},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: zoneLabelKey},
					},
				},
			}
			newTopology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{Name: "new-topo"},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: regionLabelKey},
						{NodeLabel: zoneLabelKey},
						{NodeLabel: rackLabelKey},
					},
				},
			}
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gpu-node-1",
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel: gpuPoolName,
						regionLabelKey:             "us-west",
						zoneLabelKey:               "us-west-1",
						rackLabelKey:               "rack-a",
					},
				},
			}

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(nodePool, nodePoolTwin(nodePool), oldTopology, newTopology, node)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeHandler = handlers.NewNodeHandler(backendClient, fakeCachedClient)

			// Sync with old topology — expect 1 label (zone only)
			nodeHandler.SyncNodesForNodepool(gpuPoolName)

			var req swagger.NodesSyncRequest
			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(HaveLen(1))
			Expect(req.NodesForSync[0].NodeInfo.NetworkTopologyLabels).To(HaveLen(1))
			Expect(req.NodesForSync[0].NodeInfo.NetworkTopologyLabels[0].LabelKey).To(Equal(zoneLabelKey))
			Expect(req.NodesForSync[0].NodeInfo.NetworkTopologyLabels[0].LabelValue).To(Equal("us-west-1"))

			// Update nodepool to point to new topology
			nodePool.Spec.NetworkTopologyName = "new-topo"
			err := fakeCachedClient.Update(context.Background(), nodePool)
			Expect(err).To(BeNil())

			// Mirror the change onto the kai twin, so this spec still exercises an actual
			// topology change under both values of UseKaiNodePools.
			// TODO: drop once commonconstants.UseKaiNodePools=true
			kaiNodePool := &sdkkaiv1alpha1.NodePool{}
			Expect(fakeCachedClient.Get(context.Background(),
				client.ObjectKeyFromObject(nodePool), kaiNodePool)).To(Succeed())
			kaiNodePool.Spec.PreferredNetworkTopologyName = "new-topo"
			Expect(fakeCachedClient.Update(context.Background(), kaiNodePool)).To(Succeed())

			// Sync again — expect 3 labels from new topology
			nodeHandler.SyncNodesForNodepool(gpuPoolName)

			Eventually(syncRequest).WithTimeout(3 * time.Second).Should(Receive(&req))
			Expect(req.NodesForSync).To(HaveLen(1))
			Expect(req.NodesForSync[0].NodeInfo.NetworkTopologyLabels).To(HaveLen(3))

			labelMap := map[string]string{}
			for _, l := range req.NodesForSync[0].NodeInfo.NetworkTopologyLabels {
				labelMap[l.LabelKey] = l.LabelValue
			}
			Expect(labelMap).To(Equal(map[string]string{
				regionLabelKey: "us-west",
				zoneLabelKey:   "us-west-1",
				rackLabelKey:   "rack-a",
			}))
		})
	})

})
