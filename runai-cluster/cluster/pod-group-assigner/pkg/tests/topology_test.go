/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assigner"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/testbuilders"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These specs cover pod-group-assigner adding a default network
// topology onto the PodGroup it assigns, based on the picked nodepool's
// NetworkTopologyName. They use unique project/nodepool names so they
// don't collide with the main assigner suite that shares the same fake client.
var _ = Describe("Pod Group Assigner - default network topology", Ordered, func() {
	const (
		topoProject   = "topo-team"
		topologyName  = "topo-net"
		lowestLevel   = "kubernetes.io/hostname"
		highestLevel  = "topology.kubernetes.io/zone"
		userTopology  = "user-chosen-topology"
		userTopoLevel = "topology.kubernetes.io/region"
	)

	var (
		topoNamespace = fmt.Sprintf("runai-%s", topoProject)
		// nodePoolWithTopology has a default network topology; nodePoolNoTopology does not.
		nodePoolWithTopology = TestNodePool{"topo-np-with", "topo-with-key", "topo-with-val", v1alpha1.NodePoolReady}
		nodePoolNoTopology   = TestNodePool{"topo-np-without", "topo-without-key", "topo-without-val", v1alpha1.NodePoolReady}
		topoNodePools        = []TestNodePool{nodePoolWithTopology, nodePoolNoTopology, defaultNodePool}
	)

	BeforeAll(func() {
		controllerSetup()

		for i := range topoNodePools {
			createTestNodePool("topology BeforeAll", &topoNodePools[i], k8sClient)
		}
		createNamespaceIfNeeded(topoNamespace, k8sClient)
		createProjectWithNodePoolsQueues(topoProject, topoNodePools, k8sClient)

		// Give the topology-bearing nodepool a default network topology.
		npWithTopology := &v1alpha1.NodePool{}
		Expect(k8sClient.Get(apiCtx,
			types.NamespacedName{Name: nodePoolWithTopology.Name}, npWithTopology)).To(Succeed())
		npWithTopology.Spec.PreferredNetworkTopologyName = topologyName
		Expect(k8sClient.Update(apiCtx, npWithTopology)).To(Succeed())
		// Levels are ordered highest-to-lowest; the assigner adds the lowest one.
		expectCreateResource(k8sClient, testbuilders.BuildTopology(topologyName, highestLevel, lowestLevel))
	})

	AfterAll(func() {
		deleteProject(topoProject, k8sClient)
		deleteNamespace(topoNamespace, k8sClient)
		for _, np := range topoNodePools {
			deleteNodePool("topology AfterAll", np.Name, k8sClient)
		}
	})

	AfterEach(afterEachCleanup)

	It("stamps the picked nodepool's topology as a system-sourced preferred constraint", func() {
		pg := assignPodGroupToNodePool("set-from-nodepool", topoProject, topoNamespace, nodePoolWithTopology)

		pgFromClient := getPodGroupFromClient(pg, k8sClient)
		Expect(pgFromClient).NotTo(BeNil())
		Expect(pgFromClient.Spec.TopologyConstraint).To(Equal(kaiv2alpha2.TopologyConstraint{
			Topology:               topologyName,
			PreferredTopologyLevel: lowestLevel,
		}))
		Expect(pgFromClient.Annotations).To(HaveKeyWithValue(assigner.TopologySourceAnnotationKey, assigner.TopologySourceSystem))
	})

	It("leaves a user-set topology constraint untouched", func() {
		pg := newTopologyPodGroupPods("user-set", topoProject, topoNamespace, nodePoolWithTopology)

		// The user requested a topology themselves (no system-source annotation).
		userConstraint := kaiv2alpha2.TopologyConstraint{
			Topology:               userTopology,
			PreferredTopologyLevel: userTopoLevel,
		}
		setPodGroupTopology(pg, userConstraint, false, k8sClient)

		reconcile(pg)

		pgFromClient := getPodGroupFromClient(pg, k8sClient)
		Expect(pgFromClient).NotTo(BeNil())
		Expect(pgFromClient.Spec.TopologyConstraint).To(Equal(userConstraint), "user topology must not be overridden")
		Expect(pgFromClient.Annotations).NotTo(HaveKey(assigner.TopologySourceAnnotationKey), "PGA must not claim ownership of a user constraint")
	})

	It("clears a previously system-sourced constraint when the nodepool has no topology", func() {
		pg := newTopologyPodGroupPods("clear-system", topoProject, topoNamespace, nodePoolNoTopology)

		// Simulate a constraint stamped by a previous PGA pass.
		setPodGroupTopology(pg, kaiv2alpha2.TopologyConstraint{
			Topology:               topologyName,
			PreferredTopologyLevel: lowestLevel,
		}, true, k8sClient)

		reconcile(pg)

		pgFromClient := getPodGroupFromClient(pg, k8sClient)
		Expect(pgFromClient).NotTo(BeNil())
		Expect(pgFromClient.Spec.TopologyConstraint).To(Equal(kaiv2alpha2.TopologyConstraint{}), "system constraint must be cleared")
		Expect(pgFromClient.Annotations).NotTo(HaveKey(assigner.TopologySourceAnnotationKey), "source annotation must be removed when cleared")
	})

	It("leaves the constraint empty when the nodepool has no topology and none was set", func() {
		pg := assignPodGroupToNodePool("no-op", topoProject, topoNamespace, nodePoolNoTopology)

		pgFromClient := getPodGroupFromClient(pg, k8sClient)
		Expect(pgFromClient).NotTo(BeNil())
		Expect(pgFromClient.Spec.TopologyConstraint).To(Equal(kaiv2alpha2.TopologyConstraint{}))
		Expect(pgFromClient.Annotations).NotTo(HaveKey(assigner.TopologySourceAnnotationKey))
	})

	// Depends on the KAI pod-grouper preserving Spec.TopologyConstraint in its
	// ignoreFields (a separate KAI-Scheduler PR + dependency bump). Until then the
	// grouper rebuilds the PodGroup spec from pods and clobbers a PGA-written
	// constraint, so this cannot be exercised against the current KAI dependency.
	It("keeps a PGA-written TopologyConstraint across a pod-grouper reconcile", func() {
		Skip("Requires KAI pod-grouper ignoreFields change (separate KAI-Scheduler PR + dep bump)")
	})
})

// newTopologyPodGroupPods creates the pods + PodGroup requesting a single nodepool
// (via pod node-affinity), without reconciling. The returned PodGroup is up for
// assignment to the given nodepool.
func newTopologyPodGroupPods(testName, projectName, namespace string, nodePool TestNodePool) types.NamespacedName {
	pg := types.NamespacedName{Namespace: namespace, Name: generatePodGroupName()}
	createdPodGroups = append(createdPodGroups, pg)

	matchExpressions := [][]corev1.NodeSelectorRequirement{{{
		Key:      nodePool.LabelKey,
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{nodePool.LabelValue},
	}}}
	createPodsForPodGroup(pg.Name, pg.Namespace, 1, matchExpressions)

	podGroup := getPodGroupObj(pg, []string{nodePool.Name}, "", projectName)
	createPodGroup(testName, podGroup, k8sClient)

	return pg
}

// assignPodGroupToNodePool creates a PodGroup requesting the given nodepool and
// reconciles once so the assigner picks it.
func assignPodGroupToNodePool(testName, projectName, namespace string, nodePool TestNodePool) types.NamespacedName {
	pg := newTopologyPodGroupPods(testName, projectName, namespace, nodePool)
	reconcile(pg)
	return pg
}

// setPodGroupTopology writes a topology constraint (and optionally the
// system-source annotation) onto an existing PodGroup, modelling either a
// user-requested constraint or one left by a previous PGA pass.
func setPodGroupTopology(pg types.NamespacedName, constraint kaiv2alpha2.TopologyConstraint, systemSourced bool, k8sClient client.Client) {
	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	ExpectWithOffset(1, pgFromClient).NotTo(BeNil())

	pgFromClient.Spec.TopologyConstraint = constraint
	if systemSourced {
		if pgFromClient.Annotations == nil {
			pgFromClient.Annotations = map[string]string{}
		}
		pgFromClient.Annotations[assigner.TopologySourceAnnotationKey] = assigner.TopologySourceSystem
	}

	ExpectWithOffset(1, k8sClient.Update(apiCtx, pgFromClient)).To(Succeed())
}
