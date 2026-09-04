/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"encoding/json"

	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func nodeNamed(name string) *corev1.Node {
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

// nrtWith builds an NRT object with a single zone of the given type and an optional
// topologyManagerPolicy attribute (empty string => attribute omitted).
func nrtWith(name, zoneType, policy string) *nrtv1alpha2.NodeResourceTopology {
	nrt := &nrtv1alpha2.NodeResourceTopology{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Zones:      nrtv1alpha2.ZoneList{{Name: "zone-0", Type: zoneType}},
	}
	if policy != "" {
		nrt.Attributes = nrtv1alpha2.AttributeList{{Name: NrtAttrTopologyManagerPolicy, Value: policy}}
	}
	return nrt
}

var _ = Describe("evaluateNrtHealth", func() {
	It("reports Missing with the exact missing-object message when there is no NRT object", func() {
		healthy, reason, message := evaluateNrtHealth(nil)
		Expect(healthy).To(BeFalse())
		Expect(reason).To(Equal(NrtHealthReasonMissing))
		Expect(message).To(Equal("The following prerequisites are not met on the node: " +
			"NodeResourceTopology custom resource missing. As a result, NUMA-aware scheduling may be impacted."))
	})

	It("reports NoNumaZones with the exact message when the NRT exposes no NUMA-node zones", func() {
		nrt := nrtWith("node-a", "Socket", "single-numa-node")
		healthy, reason, message := evaluateNrtHealth(nrt)
		Expect(healthy).To(BeFalse())
		Expect(reason).To(Equal(NrtHealthReasonNoNumaZones))
		Expect(message).To(Equal("The following prerequisites are not met on the node: " +
			"NodeResourceTopology custom resource invalid. As a result, NUMA-aware scheduling may be impacted."))
	})

	DescribeTable("Topology Manager policy handling for a node with a valid NUMA-node zone",
		func(policy string, expectHealthy bool, expectReason string) {
			nrt := nrtWith("node-a", NrtZoneTypeNode, policy)
			healthy, reason, message := evaluateNrtHealth(nrt)
			Expect(healthy).To(Equal(expectHealthy))
			Expect(reason).To(Equal(expectReason))
			if expectHealthy {
				Expect(message).To(BeEmpty())
			} else {
				Expect(message).To(Equal("The following prerequisites are not met on the node: " +
					"Kubelet Topology Manager Policy misconfigured. As a result, NUMA-aware scheduling may be impacted."))
			}
		},
		Entry("none => unhealthy (not enforcing)", "none", false, NrtHealthReasonPolicyNotEnforcing),
		Entry("None (mixed case) => unhealthy", "None", false, NrtHealthReasonPolicyNotEnforcing),
		Entry("best-effort => healthy (recommended default)", "best-effort", true, ""),
		Entry("restricted => healthy", "restricted", true, ""),
		Entry("single-numa-node => healthy", "single-numa-node", true, ""),
		Entry("no policy attribute => unhealthy (cannot confirm enforcement)", "", false, NrtHealthReasonPolicyNotEnforcing),
	)
})

var _ = Describe("NRT condition patch builders", func() {
	It("upsert patch carries only our condition type with the expected fields", func() {
		patchBytes, err := nrtConditionUpsertPatch(corev1.NodeCondition{
			Type:    NrtHealthyConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  NrtHealthReasonMissing,
			Message: "boom",
		})
		Expect(err).NotTo(HaveOccurred())

		var decoded struct {
			Status struct {
				Conditions []map[string]any `json:"conditions"`
			} `json:"status"`
		}
		Expect(json.Unmarshal(patchBytes, &decoded)).To(Succeed())
		Expect(decoded.Status.Conditions).To(HaveLen(1))
		cond := decoded.Status.Conditions[0]
		Expect(cond["type"]).To(Equal(string(NrtHealthyConditionType)))
		Expect(cond["status"]).To(Equal(string(corev1.ConditionTrue)))
		Expect(cond["reason"]).To(Equal(NrtHealthReasonMissing))
		Expect(cond).NotTo(HaveKey("$patch"))
	})

	It("delete patch uses the $patch:delete directive keyed by type", func() {
		patchBytes, err := nrtConditionDeletePatch()
		Expect(err).NotTo(HaveOccurred())

		var decoded struct {
			Status struct {
				Conditions []map[string]any `json:"conditions"`
			} `json:"status"`
		}
		Expect(json.Unmarshal(patchBytes, &decoded)).To(Succeed())
		Expect(decoded.Status.Conditions).To(HaveLen(1))
		cond := decoded.Status.Conditions[0]
		Expect(cond["type"]).To(Equal(string(NrtHealthyConditionType)))
		Expect(cond["$patch"]).To(Equal("delete"))
	})
})

var _ = Describe("ensureNrtCondition / removeNrtCondition", func() {
	var (
		npc  *NodePoolController
		node *corev1.Node
		ctx  context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		// Pre-existing unrelated condition to prove the strategic merge leaves it untouched.
		node = &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-a"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
			},
		}
		cl := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(node).
			WithStatusSubresource(node).
			Build()
		npc = &NodePoolController{Client: cl}
	})

	getNode := func() *corev1.Node {
		fetched := &corev1.Node{}
		Expect(npc.Client.Get(ctx, client.ObjectKey{Name: "node-a"}, fetched)).To(Succeed())
		return fetched
	}

	It("adds the NRT condition while preserving other conditions", func() {
		Expect(npc.ensureNrtCondition(ctx, node, NrtHealthReasonMissing, "boom")).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).NotTo(BeNil(),
			"the in-memory node must reflect the status patch during the same reconcile")

		fetched := getNode()
		Expect(findNodeCondition(fetched, corev1.NodeReady)).NotTo(BeNil())
		ours := findNodeCondition(fetched, NrtHealthyConditionType)
		Expect(ours).NotTo(BeNil())
		Expect(ours.Status).To(Equal(corev1.ConditionTrue))
		Expect(ours.Reason).To(Equal(NrtHealthReasonMissing))
		Expect(ours.Message).To(Equal("boom"))
	})

	It("removes only the NRT condition, keeping other conditions", func() {
		Expect(npc.ensureNrtCondition(ctx, node, NrtHealthReasonMissing, "boom")).To(Succeed())

		// Re-read so the in-memory node reflects the condition we just set.
		node = getNode()
		Expect(npc.removeNrtCondition(ctx, node)).To(Succeed())
		Expect(findNodeCondition(node, corev1.NodeReady)).NotTo(BeNil())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).To(BeNil(),
			"the in-memory node must reflect condition removal during the same reconcile")

		fetched := getNode()
		Expect(findNodeCondition(fetched, NrtHealthyConditionType)).To(BeNil())
		Expect(findNodeCondition(fetched, corev1.NodeReady)).NotTo(BeNil())
	})

	It("removeNrtCondition is a no-op when the condition is absent", func() {
		Expect(npc.removeNrtCondition(ctx, node)).To(Succeed())
		Expect(findNodeCondition(getNode(), corev1.NodeReady)).NotTo(BeNil())
	})
})

var _ = Describe("getNodeResourceTopology", func() {
	var ctx context.Context

	newController := func(objs ...client.Object) *NodePoolController {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(nrtv1alpha2.AddToScheme(scheme)).To(Succeed())
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		return &NodePoolController{Client: cl}
	}

	BeforeEach(func() { ctx = context.Background() })

	It("returns the NRT object named after the node with found=true", func() {
		npc := newController(nrtWith("node-a", NrtZoneTypeNode, "single-numa-node"))
		nrt, found, err := npc.getNodeResourceTopology(ctx, "node-a")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(nrt).NotTo(BeNil())
		Expect(nrt.Name).To(Equal("node-a"))
	})

	It("returns found=false (no error) when no NRT object exists for the node", func() {
		npc := newController()
		nrt, found, err := npc.getNodeResourceTopology(ctx, "node-a")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		Expect(nrt).To(BeNil())
	})
})

// Sanity checks that reconcileNrtHealthCondition actually mutates the node (the path taken
// during Reconcile). Exhaustive health cases live in the evaluateNrtHealth specs above.
var _ = Describe("reconcileNrtHealthCondition (node mutation)", func() {
	var ctx context.Context

	newController := func(objs ...client.Object) (*NodePoolController, *corev1.Node) {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(nrtv1alpha2.AddToScheme(scheme)).To(Succeed())
		node := nodeNamed("node-a")
		all := append([]client.Object{node}, objs...)
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(all...).WithStatusSubresource(node).Build()
		npc := &NodePoolController{Client: cl}
		npc.nrtEnabled = true
		return npc, node
	}

	numaNodePool := func(enabled bool) *v1alpha1.NodePool {
		return &v1alpha1.NodePool{Spec: v1alpha1.NodePoolSpec{
			SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{NumaPluginName: {Enabled: ptr.To(enabled)}},
			},
		}}
	}

	getNode := func(npc *NodePoolController) *corev1.Node {
		fetched := &corev1.Node{}
		Expect(npc.Client.Get(ctx, client.ObjectKey{Name: "node-a"}, fetched)).To(Succeed())
		return fetched
	}

	BeforeEach(func() { ctx = context.Background() })

	DescribeTable("detects whether NUMA-aware scheduling is enabled",
		func(nodePool *v1alpha1.NodePool, expected bool) {
			Expect(isNumaEnabled(nodePool)).To(Equal(expected))
		},
		Entry("without scheduling shard configuration", &v1alpha1.NodePool{}, false),
		Entry("without the NUMA plugin", &v1alpha1.NodePool{Spec: v1alpha1.NodePoolSpec{
			SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{},
		}}, false),
		Entry("with a NUMA plugin whose enabled value is absent", &v1alpha1.NodePool{Spec: v1alpha1.NodePoolSpec{
			SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{NumaPluginName: {}},
			},
		}}, false),
		Entry("with the NUMA plugin disabled", numaNodePool(false), false),
		Entry("with the NUMA plugin enabled", numaNodePool(true), true),
	)

	It("sets the condition on a numa-enabled node whose NRT is missing", func() {
		npc, node := newController() // no NRT object
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).NotTo(BeNil(),
			"nodepool status calculation later in the reconcile must see the condition")

		cond := findNodeCondition(getNode(npc), NrtHealthyConditionType)
		Expect(cond).NotTo(BeNil())
		Expect(cond.Reason).To(Equal(NrtHealthReasonMissing))
	})

	It("clears an existing condition when NUMA-aware scheduling is disabled", func() {
		npc, node := newController() // no NRT object
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).NotTo(BeNil())

		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(false))).To(Succeed())

		Expect(findNodeCondition(node, NrtHealthyConditionType)).To(BeNil())
		Expect(findNodeCondition(getNode(npc), NrtHealthyConditionType)).To(BeNil())
	})

	It("leaves a healthy numa-enabled node without the condition", func() {
		npc, node := newController(nrtWith("node-a", NrtZoneTypeNode, "single-numa-node"))
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())

		Expect(findNodeCondition(getNode(npc), NrtHealthyConditionType)).To(BeNil())
	})

	It("sets the Missing condition when the NRT CRD is not installed (nrtEnabled=false)", func() {
		npc, node := newController() // no NRT object, and CRD not installed
		npc.nrtEnabled = false
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())

		cond := findNodeCondition(getNode(npc), NrtHealthyConditionType)
		Expect(cond).NotTo(BeNil())
		Expect(cond.Reason).To(Equal(NrtHealthReasonMissing))
	})

	It("handles a healthy to unhealthy to healthy NRT transition", func() {
		healthyNrt := nrtWith("node-a", NrtZoneTypeNode, "single-numa-node")
		npc, node := newController(healthyNrt)

		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).To(BeNil())

		Expect(npc.Client.Delete(ctx, healthyNrt)).To(Succeed())
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).NotTo(BeNil())
		Expect(findNodeCondition(getNode(npc), NrtHealthyConditionType)).NotTo(BeNil())

		Expect(npc.Client.Create(ctx, nrtWith("node-a", NrtZoneTypeNode, "single-numa-node"))).To(Succeed())
		Expect(npc.reconcileNrtHealthCondition(ctx, node, numaNodePool(true))).To(Succeed())
		Expect(findNodeCondition(node, NrtHealthyConditionType)).To(BeNil())
		Expect(findNodeCondition(getNode(npc), NrtHealthyConditionType)).To(BeNil())
	})
})

// fakeManager is a ctrl.Manager whose only usable method is GetCache — enough for watchForNrtCRD.
type fakeManager struct {
	ctrl.Manager
	cache cache.Cache
}

func (m *fakeManager) GetCache() cache.Cache { return m.cache }

var _ = Describe("watchForNrtCRD", func() {
	It("restarts the process only when the NodeResourceTopology CRD is installed", func() {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())
		fakeCache := &informertest.FakeInformers{Scheme: scheme}
		npc := &NodePoolController{}

		exitCode := -1
		defer func(orig func(int)) { osExit = orig }(osExit)
		osExit = func(code int) { exitCode = code }

		Expect(npc.watchForNrtCRD(ctx, &fakeManager{cache: fakeCache})).To(Succeed())

		crdGVK := apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition")
		crdInformer, err := fakeCache.FakeInformerForKind(ctx, crdGVK)
		Expect(err).NotTo(HaveOccurred())
		addCRD := func(name string) *apiextensionsv1.CustomResourceDefinition {
			crd := &apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: name}}
			crdInformer.Add(crd)
			return crd
		}

		otherCRD := addCRD("other.example.com")
		establishedOtherCRD := otherCRD.DeepCopy()
		establishedOtherCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		}}
		crdInformer.Update(otherCRD, establishedOtherCRD)
		Expect(exitCode).To(Equal(-1))

		nrtCRD := addCRD(nrtCRDName)
		Expect(exitCode).To(Equal(-1))

		establishedNRTCRD := nrtCRD.DeepCopy()
		establishedNRTCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{{
			Type:   apiextensionsv1.Established,
			Status: apiextensionsv1.ConditionTrue,
		}}
		crdInformer.Update(nrtCRD, establishedNRTCRD)
		Expect(exitCode).To(Equal(0))
	})
})

var _ = Describe("topologyManagerPolicy", func() {
	It("returns the lowercase policy value and found=true when the attribute is present", func() {
		policy, found := topologyManagerPolicy(nrtWith("node-a", NrtZoneTypeNode, "Single-NUMA-Node"))
		Expect(found).To(BeTrue())
		Expect(policy).To(Equal("single-numa-node"))
	})

	It("returns found=false when the topologyManagerPolicy attribute is absent", func() {
		policy, found := topologyManagerPolicy(nrtWith("node-a", NrtZoneTypeNode, ""))
		Expect(found).To(BeFalse())
		Expect(policy).To(BeEmpty())
	})
})

var _ = Describe("getTopologyManagerPolicyFromNodeNRT / reconcileNodeAnnotations", func() {
	var ctx context.Context

	newController := func(objs ...client.Object) (*NodePoolController, *corev1.Node) {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(nrtv1alpha2.AddToScheme(scheme)).To(Succeed())
		node := nodeNamed("node-a")
		all := append([]client.Object{node}, objs...)
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(all...).Build()
		return &NodePoolController{Client: cl, nrtEnabled: true}, node
	}

	numaNodePool := func(enabled bool) *v1alpha1.NodePool {
		return &v1alpha1.NodePool{Spec: v1alpha1.NodePoolSpec{
			SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{NumaPluginName: {Enabled: ptr.To(enabled)}},
			},
		}}
	}

	policyAnnotation := func(npc *NodePoolController) (string, bool) {
		fetched := &corev1.Node{}
		Expect(npc.Client.Get(ctx, client.ObjectKey{Name: "node-a"}, fetched)).To(Succeed())
		value, found := fetched.Annotations[v1alpha1.AnnotationTopologyManagerPolicy]
		return value, found
	}

	BeforeEach(func() { ctx = context.Background() })

	DescribeTable("getTopologyManagerPolicyFromNodeNRT",
		func(numaEnabled bool, nrt *nrtv1alpha2.NodeResourceTopology, expected string) {
			var objs []client.Object
			if nrt != nil {
				objs = append(objs, nrt)
			}
			npc, node := newController(objs...)
			policy, err := npc.getTopologyManagerPolicyFromNodeNRT(ctx, node, numaNodePool(numaEnabled))
			Expect(err).NotTo(HaveOccurred())
			Expect(policy).To(Equal(expected))
		},
		Entry("numa enabled + policy present => that policy", true, nrtWith("node-a", NrtZoneTypeNode, "single-numa-node"), "single-numa-node"),
		Entry("numa enabled + mixed-case policy => lowercase policy", true, nrtWith("node-a", NrtZoneTypeNode, "Best-Effort"), "best-effort"),
		Entry("numa enabled + NRT missing => empty", true, nil, ""),
		Entry("numa enabled + no policy attribute => empty", true, nrtWith("node-a", NrtZoneTypeNode, ""), ""),
		Entry("numa disabled => empty", false, nrtWith("node-a", NrtZoneTypeNode, "single-numa-node"), ""),
	)

	It("annotates a numa-enabled node with the lowercase Topology Manager policy", func() {
		npc, node := newController(nrtWith("node-a", NrtZoneTypeNode, "Best-Effort"))
		Expect(npc.reconcileNodeAnnotations(ctx, node, numaNodePool(true))).To(Succeed())

		value, found := policyAnnotation(npc)
		Expect(found).To(BeTrue())
		Expect(value).To(Equal("best-effort"))
	})

	It("removes the policy annotation when NUMA-aware scheduling is disabled", func() {
		node := nodeNamed("node-a")
		node.Annotations = map[string]string{v1alpha1.AnnotationTopologyManagerPolicy: "single-numa-node"}
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(nrtv1alpha2.AddToScheme(scheme)).To(Succeed())
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).Build()
		npc := &NodePoolController{Client: cl, nrtEnabled: true}

		Expect(npc.reconcileNodeAnnotations(ctx, node, numaNodePool(false))).To(Succeed())

		_, found := policyAnnotation(npc)
		Expect(found).To(BeFalse())
	})

	It("is idempotent: a second reconcile with the same policy makes no change", func() {
		npc, node := newController(nrtWith("node-a", NrtZoneTypeNode, "restricted"))
		Expect(npc.reconcileNodeAnnotations(ctx, node, numaNodePool(true))).To(Succeed())

		fetched := &corev1.Node{}
		Expect(npc.Client.Get(ctx, client.ObjectKey{Name: "node-a"}, fetched)).To(Succeed())
		before := fetched.ResourceVersion

		Expect(npc.reconcileNodeAnnotations(ctx, fetched, numaNodePool(true))).To(Succeed())
		after, _ := policyAnnotation(npc)
		Expect(after).To(Equal("restricted"))

		reReadNode := &corev1.Node{}
		Expect(npc.Client.Get(ctx, client.ObjectKey{Name: "node-a"}, reReadNode)).To(Succeed())
		Expect(reReadNode.ResourceVersion).To(Equal(before), "no patch should be issued when already up to date")
	})
})
