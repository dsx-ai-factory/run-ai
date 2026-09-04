/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	managed_nodes_config "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/managed-nodes-config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func nrtForReconcileTest(nodeName, zoneType, policy string) *nrtv1alpha2.NodeResourceTopology {
	nrt := &nrtv1alpha2.NodeResourceTopology{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	if zoneType != "" {
		nrt.Zones = nrtv1alpha2.ZoneList{{Name: "zone-0", Type: zoneType}}
	}
	if policy != "" {
		nrt.Attributes = nrtv1alpha2.AttributeList{{
			Name:  nodepool_controller.NrtAttrTopologyManagerPolicy,
			Value: policy,
		}}
	}
	return nrt
}

func nodeConditionForReconcileTest(node *corev1.Node, conditionType corev1.NodeConditionType) *corev1.NodeCondition {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			return &node.Status.Conditions[i]
		}
	}
	return nil
}

func nrtNodeMessageForReconcileTest(missingItem string) string {
	return fmt.Sprintf(nodepool_controller.NrtPrereqMessageTemplate, missingItem)
}

func nrtNodePoolMessageForReconcileTest(missingItem string) string {
	return fmt.Sprintf(nodepool_controller.NrtNodePoolPrereqMessageTemplate, missingItem)
}

var _ = Describe("NRT health through full NodePool Reconcile", func() {
	const (
		nodePoolName  = "numa-node-pool"
		nodeAName     = "node-a"
		nodeBName     = "node-b"
		nodePoolKey   = "nodePoolNrtKey"
		nodePoolValue = "nodePoolNrtValue"
	)

	DescribeTable("cascades node NRT health into nodepool status",
		func(nodeANrt *nrtv1alpha2.NodeResourceTopology, expectedReason, expectedMissingItem,
			expectedNodeAPolicy string, expectedPhase v1alpha1.NodePoolPhase) {
			ctx, cancel := context.WithCancel(context.Background())
			stopper := make(chan struct{}, 1)
			DeferCleanup(func() {
				close(stopper)
				cancel()
			})

			testCase := &TestCase{
				Name: "NRT health full reconcile",
				NodePools: []TestNodePool{{
					Name:       nodePoolName,
					LabelKey:   nodePoolKey,
					LabelValue: nodePoolValue,
				}},
				Nodes: map[string]TestNode{
					nodeAName: {
						Name:   nodeAName,
						Labels: map[string]string{nodePoolKey: nodePoolValue},
					},
					nodeBName: {
						Name:   nodeBName,
						Labels: map[string]string{nodePoolKey: nodePoolValue},
					},
				},
				ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
					nodePoolName: {},
				},
			}

			k8sClient := preTestSetup(ctx, stopper, testCase, scheme)
			updateNodePoolSchedulingShardConfig(nodePoolName, &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{
					nodepool_controller.NumaPluginName: {Enabled: ptr.To(true)},
				},
			}, k8sClient)

			if nodeANrt != nil {
				Expect(k8sClient.Create(ctx, nodeANrt.DeepCopy())).To(Succeed())
			}
			Expect(k8sClient.Create(ctx,
				nrtForReconcileTest(nodeBName, nodepool_controller.NrtZoneTypeNode, "Restricted"))).To(Succeed())

			npc := nodepool_controller.NewNodePoolController(
				k8sClient,
				scheme,
				&common.NodePoolControllerParams{},
			)
			npc.SetServiceMonitorEnabled(true)
			npc.SetNrtEnabled(true)
			mncc := managed_nodes_config.NewManagedNodesConfigController(k8sClient, scheme, npc)

			reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)

			Eventually(func(g Gomega) {
				nodePool := getNodePoolFromClient(nodePoolName, k8sClient)
				g.Expect(nodePool).NotTo(BeNil())
				g.Expect(nodePool.Status.Phase).To(Equal(expectedPhase))

				nodeA := getNodeFromClient(nodeAName, k8sClient)
				nodeB := getNodeFromClient(nodeBName, k8sClient)
				g.Expect(nodeA).NotTo(BeNil())
				g.Expect(nodeB).NotTo(BeNil())

				nodeACondition := nodeConditionForReconcileTest(
					nodeA,
					nodepool_controller.NrtHealthyConditionType,
				)
				nodeBCondition := nodeConditionForReconcileTest(
					nodeB,
					nodepool_controller.NrtHealthyConditionType,
				)
				g.Expect(nodeBCondition).To(BeNil())
				g.Expect(nodeB.Annotations).To(HaveKeyWithValue(
					v1alpha1.AnnotationTopologyManagerPolicy,
					"restricted",
				))

				if expectedReason == "" {
					g.Expect(nodeACondition).To(BeNil())
					g.Expect(nodePool.Status.Message).To(BeEmpty())
					g.Expect(nodePool.Status.Conditions).To(BeEmpty())
					g.Expect(nodePool.Status.Nodes).To(ConsistOf(
						v1alpha1.NodeInNodePool{Name: nodeAName, Status: v1alpha1.NodeReady},
						v1alpha1.NodeInNodePool{Name: nodeBName, Status: v1alpha1.NodeReady},
					))
				} else {
					expectedNodeMessage := nrtNodeMessageForReconcileTest(expectedMissingItem)
					expectedNodePoolMessage := nrtNodePoolMessageForReconcileTest(expectedMissingItem)

					g.Expect(nodeACondition).NotTo(BeNil())
					g.Expect(nodeACondition.Status).To(Equal(corev1.ConditionTrue))
					g.Expect(nodeACondition.Reason).To(Equal(expectedReason))
					g.Expect(nodeACondition.Message).To(Equal(expectedNodeMessage))

					g.Expect(nodePool.Status.Message).To(Equal(expectedNodePoolMessage))
					g.Expect(nodePool.Status.Conditions).To(HaveLen(1))
					nodePoolCondition := nodePool.Status.Conditions[0]
					g.Expect(nodePoolCondition.Type).To(Equal(
						v1alpha1.NodePoolMissingNrtHealthyPrerequisite,
					))
					g.Expect(nodePoolCondition.Status).To(Equal(corev1.ConditionTrue))
					g.Expect(nodePoolCondition.Reason).To(Equal(
						v1alpha1.NodePoolMissingNrtHealthyPrerequisiteReason,
					))
					g.Expect(nodePoolCondition.Message).To(Equal(expectedNodePoolMessage))
					g.Expect(nodePool.Status.Nodes).To(ConsistOf(
						v1alpha1.NodeInNodePool{
							Name:   nodeAName,
							Status: v1alpha1.NodeMissingNrtHealthyPrerequisite,
						},
						v1alpha1.NodeInNodePool{Name: nodeBName, Status: v1alpha1.NodeReady},
					))
				}

				if expectedNodeAPolicy == "" {
					g.Expect(nodeA.Annotations).NotTo(HaveKey(
						v1alpha1.AnnotationTopologyManagerPolicy,
					))
				} else {
					g.Expect(nodeA.Annotations).To(HaveKeyWithValue(
						v1alpha1.AnnotationTopologyManagerPolicy,
						expectedNodeAPolicy,
					))
				}
			}, validateTestTimeout, validateTestInterval).Should(Succeed())
		},
		Entry(
			"policy none",
			nrtForReconcileTest(nodeAName, nodepool_controller.NrtZoneTypeNode, "None"),
			nodepool_controller.NrtHealthReasonPolicyNotEnforcing,
			nodepool_controller.NrtMissingItemPolicy,
			nodepool_controller.NrtPolicyNone,
			v1alpha1.NodePoolMissingPrerequisites,
		),
		Entry(
			"missing NRT",
			nil,
			nodepool_controller.NrtHealthReasonMissing,
			nodepool_controller.NrtMissingItemNoObject,
			"",
			v1alpha1.NodePoolMissingPrerequisites,
		),
		Entry(
			"NRT without NUMA zones",
			nrtForReconcileTest(nodeAName, "Socket", "Restricted"),
			nodepool_controller.NrtHealthReasonNoNumaZones,
			nodepool_controller.NrtMissingItemNoNumaZones,
			"restricted",
			v1alpha1.NodePoolMissingPrerequisites,
		),
		Entry(
			"healthy enforcing NRT",
			nrtForReconcileTest(nodeAName, nodepool_controller.NrtZoneTypeNode, "Single-NUMA-Node"),
			"",
			"",
			"single-numa-node",
			v1alpha1.NodePoolReady,
		),
	)
})
