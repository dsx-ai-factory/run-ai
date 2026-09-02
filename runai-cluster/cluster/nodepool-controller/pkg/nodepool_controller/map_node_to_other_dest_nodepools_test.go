/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"fmt"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("mapNodeToOtherDestNodePools", Ordered, func() {
	var (
		scheme     *runtime.Scheme
		npc        *NodePoolController
		ctx        context.Context
		fakeClient client.Client

		nodeObjs     []*corev1.Node
		nodePoolObjs []*v1alpha1.NodePool
	)

	BeforeAll(func() {
		scheme = runtime.NewScheme()
		_ = v1alpha1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		ctx = context.Background()

		// Seed the config singleton with runai-vocabulary values so the
		// production code under test sees the same label key/default name
		// the test assertions are written against.
		config.SetCurrent(&config.NodePoolControllerConfig{
			NodePoolNameLabel:     kaiconstants.DefaultNodePoolLabelKey,
			DefaultNodepoolName:   kaiconstants.DefaultNodePoolName,
			SchedulerName:         "runai-scheduler",
			SchedulerNamespace:    "runai",
			MetricsNamespace:      "runai",
			FinalizerDomain:       "run.ai",
			UnschedulableLabelKey: "kai.scheduler/unschedulable",
		})

		nodeObjs = make([]*corev1.Node, 5)
		nodePoolObjs = make([]*v1alpha1.NodePool, 5)
		objs := make([]client.Object, 0, 10)

		for i := 0; i < 5; i++ {
			nodeObjs[i] = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("node%d", i+1),
					Labels: map[string]string{},
				},
				Spec: corev1.NodeSpec{Unschedulable: false},
			}
			nodePoolObjs[i] = &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("np%d", i+1)},
				Spec: v1alpha1.NodePoolSpec{
					LabelKey:   fmt.Sprintf("key%d", i+1),
					LabelValue: fmt.Sprintf("val%d", i+1),
				},
				Status: v1alpha1.NodePoolStatus{},
			}

			objs = append(objs, nodeObjs[i], nodePoolObjs[i])
		}

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&v1alpha1.NodePool{}).
			WithObjects(objs...).Build()
		npc = &NodePoolController{Client: fakeClient, Scheme: scheme}
	})

	BeforeEach(func() {
		// Reset all node labels and unschedulable before each test
		for _, n := range nodeObjs {
			n.Labels = map[string]string{}
			n.Spec.Unschedulable = false
			_ = fakeClient.Update(ctx, n)
			// the fake client normalizes an empty Labels map to nil on Update;
			// restore it so spec bodies can assign labels without panicking
			n.Labels = map[string]string{}
		}
		// Reset all nodepool status messages
		for _, np := range nodePoolObjs {
			np.Status.Message = ""
			_ = fakeClient.Status().Update(ctx, np)
		}
	})

	It("should return matching nodepool when node is unschedulable by us and matches label", func() {
		n := nodeObjs[0]
		n.Labels[config.Get().UnschedulableLabelKey] = "true"
		n.Labels["key1"] = "val1"
		n.Spec.Unschedulable = true
		_ = fakeClient.Update(ctx, n)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "other")
		Expect(requests).To(HaveLen(1))
		Expect(requests[0].NamespacedName.Name).To(Equal("np1"))
	})

	It("should return default nodepool when node is unschedulable by us and matches no nodepool", func() {
		n := nodeObjs[1]
		n.Labels[config.Get().UnschedulableLabelKey] = "true"
		n.Spec.Unschedulable = true
		_ = fakeClient.Update(ctx, n)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "np2")
		Expect(requests).To(HaveLen(1))
		Expect(requests[0].NamespacedName.Name).To(Equal(config.Get().DefaultNodepoolName))
	})

	It("should return nodepool if node is not unschedulable but appears in status message", func() {
		n := nodeObjs[2]

		np := nodePoolObjs[2]
		np.Status.Message = "node3, something else"
		_ = fakeClient.Status().Update(ctx, np)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "other")
		Expect(requests).To(HaveLen(1))
		Expect(requests[0].NamespacedName.Name).To(Equal("np3"))
	})

	It("should return nothing if node is not unschedulable and not in any status message", func() {
		n := nodeObjs[3]

		np := nodePoolObjs[3]
		np.Status.Message = "something else"
		_ = fakeClient.Status().Update(ctx, np)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "other")
		Expect(requests).To(BeEmpty())
	})

	It("should not match if node name appears in message but without allowed suffix", func() {
		n := nodeObjs[4]
		n.Spec.Unschedulable = false
		_ = fakeClient.Update(ctx, n)

		np := nodePoolObjs[4]
		np.Status.Message = "node5X something else" // 'X' is not an allowed suffix
		_ = fakeClient.Status().Update(ctx, np)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "other")
		Expect(requests).To(BeEmpty())
	})

	It("should not treat node as unschedulable by us if Unschedulable=true but missing by-us label", func() {
		n := nodeObjs[0]
		n.Labels = map[string]string{} // Remove the by-us label
		n.Spec.Unschedulable = true
		_ = fakeClient.Update(ctx, n)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "np1")
		Expect(requests).To(BeEmpty())
	})

	It("should fall back to default when the only matching nodepool is already in requests", func() {
		n := nodeObjs[1]
		n.Labels[config.Get().UnschedulableLabelKey] = "true"
		n.Labels["key2"] = "val2"
		n.Spec.Unschedulable = true
		_ = fakeClient.Update(ctx, n)

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "np2")
		Expect(requests).To(HaveLen(1))
		Expect(requests[0].NamespacedName.Name).To(Equal(config.Get().DefaultNodepoolName))
	})

	It("should skip nodepools with DeletionTimestamp set", func() {
		n := nodeObjs[2]
		n.Labels[config.Get().UnschedulableLabelKey] = "true"
		n.Labels["key3"] = "val3"
		n.Spec.Unschedulable = true
		_ = fakeClient.Update(ctx, n)

		// The fake client only sets DeletionTimestamp via Delete while a finalizer holds the
		// object, so add a finalizer and delete it to make the nodepool appear as deleting.
		np := nodePoolObjs[2]
		np.Finalizers = []string{"test/finalizer"}
		Expect(fakeClient.Update(ctx, np)).To(Succeed())
		Expect(fakeClient.Delete(ctx, np)).To(Succeed())

		requests := npc.mapNodeToOtherDestNodePools(ctx, n, "other")
		Expect(requests).To(HaveLen(1))
		Expect(requests[0].NamespacedName.Name).To(Equal(config.Get().DefaultNodepoolName))
	})
})
