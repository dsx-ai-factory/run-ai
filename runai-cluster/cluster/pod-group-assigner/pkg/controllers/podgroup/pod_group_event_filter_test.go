/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	nodepoolutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils/utils"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testNodePoolLabelKey    = "test/node-pool"
	testDefaultNodePoolName = "default"
	testUnexistingNodePool  = "test-unexisting-node-pool"
)

func TestPodGroupEventFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PodGroupEventFilter Suite")
}

var _ = BeforeSuite(func() {
	// Configure the package-level config with the runai vocabulary so that
	// code under test (which now reads via config.Config()) finds the same
	// label keys / defaults the test data uses.
	DeferCleanup(config.SetForTest(config.PodGroupAssignerConfig{
		NodePoolLabelKey:           testNodePoolLabelKey,
		DefaultNodepoolName:        testDefaultNodePoolName,
		UnexistingNodepoolSentinel: testUnexistingNodePool,
	}))
})

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(kaiv2alpha2.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func newReconcilerWithObjects(objs ...client.Object) *PodGroupReconciler {
	fc := fakeclient.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(objs...).
		Build()
	return &PodGroupReconciler{cachedClient: fc}
}

func makePodGroup(name, namespace, nodePoolName string) *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{testNodePoolLabelKey: nodePoolName},
		},
	}
}

func makePodGroupNoLabel(name, namespace string) *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}

var _ = Describe("filterNodePoolsObjectsForController", func() {
	DescribeTable("returns true only for Deleting and Unschedulable phases",
		func(phase v1alpha1.NodePoolPhase, want bool) {
			Expect(isNodePoolPhaseRelevantForController(string(phase))).To(Equal(want))
		},
		Entry("Deleting → true", v1alpha1.NodePoolDeleting, true),
		Entry("Unschedulable → true", v1alpha1.NodePoolUnschedulable, true),
		Entry("Ready → false", v1alpha1.NodePoolReady, false),
		Entry("Empty → false", v1alpha1.NodePoolEmpty, false),
		Entry("empty phase → false", v1alpha1.NodePoolPhase(""), false),
	)
})

var _ = Describe("FilterPodGroupControllerEvents", func() {
	var r *PodGroupReconciler

	BeforeEach(func() {
		r = &PodGroupReconciler{}
	})

	Context("when given a PodGroup", func() {
		It("always passes the filter", func() {
			Expect(r.FilterPodGroupControllerEvents(&kaiv2alpha2.PodGroup{})).To(BeTrue())
		})
	})

	Context("when given a NodePool", func() {
		DescribeTable("passes only for Deleting or Unschedulable phases",
			func(phase v1alpha1.NodePoolPhase, want bool) {
				np := &v1alpha1.NodePool{Status: v1alpha1.NodePoolStatus{Phase: phase}}
				Expect(r.FilterPodGroupControllerEvents(np)).To(Equal(want))
			},
			Entry("Deleting → true", v1alpha1.NodePoolDeleting, true),
			Entry("Unschedulable → true", v1alpha1.NodePoolUnschedulable, true),
			Entry("Ready → false", v1alpha1.NodePoolReady, false),
			Entry("Empty → false", v1alpha1.NodePoolEmpty, false),
			Entry("empty phase → false", v1alpha1.NodePoolPhase(""), false),
		)
	})

	Context("when given an unrelated kind", func() {
		It("filters out Pods", func() {
			Expect(r.FilterPodGroupControllerEvents(&corev1.Pod{})).To(BeFalse())
		})

		It("filters out Nodes", func() {
			Expect(r.FilterPodGroupControllerEvents(&corev1.Node{})).To(BeFalse())
		})
	})
})

var _ = Describe("listPodGroupsWithSelector", func() {
	const (
		npA = "np-a"
		npB = "np-b"
	)

	var r *PodGroupReconciler

	BeforeEach(func() {
		r = newReconcilerWithObjects(
			makePodGroup("pg-a-1", "ns1", npA),
			makePodGroup("pg-a-2", "ns2", npA),
			makePodGroup("pg-b", "ns1", npB),
			makePodGroupNoLabel("pg-default", "ns1"),
		)
	})

	It("returns only pod groups matching the label requirement", func() {
		matchA, err := nodepoolutils.RequirementNodePoolNameMatching(npA, config.Config().NodePoolLabelKey)
		Expect(err).NotTo(HaveOccurred())

		pgs, err := r.listPodGroupsWithSelector(*matchA)
		Expect(err).NotTo(HaveOccurred())

		gotKeys := []string{}
		for _, pg := range pgs.Items {
			gotKeys = append(gotKeys, pg.Namespace+"/"+pg.Name)
		}
		Expect(gotKeys).To(ConsistOf("ns1/pg-a-1", "ns2/pg-a-2"))
	})
})

var _ = Describe("MapNodePoolToPodGroupEvent", func() {
	Context("with multiple labeled pod groups in different namespaces", func() {
		const (
			npA = "np-a"
			npB = "np-b"
		)

		var r *PodGroupReconciler

		BeforeEach(func() {
			r = newReconcilerWithObjects(
				makePodGroup("pg-a-1", "ns1", npA),
				makePodGroup("pg-a-2", "ns2", npA),
				makePodGroup("pg-b", "ns1", npB),
			)
		})

		It("returns reconcile requests only for pod groups labeled with the input node pool", func() {
			requests := r.MapNodePoolToPodGroupEvent(context.Background(), &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: npA},
				Status:     v1alpha1.NodePoolStatus{Phase: v1alpha1.NodePoolDeleting},
			})

			Expect(requests).To(ConsistOf(
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns1", Name: "pg-a-1"}},
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns2", Name: "pg-a-2"}},
			))
		})
	})

	Context("when the input node pool is the synthetic default", func() {
		It("matches pod groups that have no node-pool label (DoesNotExist requirement)", func() {
			r := newReconcilerWithObjects(
				makePodGroupNoLabel("pg-default", "ns1"),
				makePodGroup("pg-labeled", "ns1", "np-x"),
			)

			requests := r.MapNodePoolToPodGroupEvent(context.Background(), &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: testDefaultNodePoolName},
			})

			Expect(requests).To(ConsistOf(
				reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns1", Name: "pg-default"}},
			))
		})
	})

	Context("when no pod groups match the node pool", func() {
		It("returns an empty request list", func() {
			r := newReconcilerWithObjects(makePodGroup("pg-a", "ns1", "np-a"))

			requests := r.MapNodePoolToPodGroupEvent(context.Background(), &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "np-nonexistent"},
			})
			Expect(requests).To(BeEmpty())
		})
	})

	Context("when given a non-NodePool object", func() {
		It("panics", func() {
			r := &PodGroupReconciler{}
			Expect(func() {
				r.MapNodePoolToPodGroupEvent(context.Background(), &corev1.Pod{})
			}).To(Panic())
		})
	})
})
