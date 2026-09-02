/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_group_assigner

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaitopologyv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	commonconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/config"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Topology levels are ordered highest-to-lowest; the pod-group-assigner stamps the lowest one.
	topologyName = "topo-e2e-network-topology"
	highestLevel = "topology.kubernetes.io/zone"
	lowestLevel  = "kubernetes.io/hostname"

	// Written by the pod-group-assigner (kept in sync with its assigner package, which is a separate
	// module and so cannot be imported here).
	topologySourceAnnotationKey = "topology.kai/source"
	topologySourceSystem        = "system"
)

// Verifies the outcome: applying a workload whose picked nodepool has a default network
// topology results in that topology being stamped automatically onto the workload's PodGroup, with
// no topology specified by the user. The pod-group-assigner reads the picked nodepool's
// NetworkTopologyName and writes a preferred constraint at the topology's lowest level.
var _ = Describe("Pod Group Assigner - default network topology", Ordered, func() {
	var (
		testConfig         *config.TestConfig
		kubeClient         client.Client
		ctx                context.Context
		testCreatedObjects = map[string]client.Object{}

		topology *kaitopologyv1alpha1.Topology
	)

	BeforeAll(func() {
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx

		// The network topology the nodepool will reference. Levels are highest-to-lowest, so the
		// lowest level (last) is what the assigner stamps as the preferred placement.
		topology = &kaitopologyv1alpha1.Topology{
			ObjectMeta: metav1.ObjectMeta{Name: topologyName},
			Spec: kaitopologyv1alpha1.TopologySpec{
				Levels: []kaitopologyv1alpha1.TopologyLevel{
					{NodeLabel: highestLevel},
					{NodeLabel: lowestLevel},
				},
			},
		}
		err := kubeClient.Create(ctx, topology)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred(), "failed to create Topology")
		}

		// Give the default nodepool a default network topology. The test project already has a queue
		// for the default nodepool, so a workload submitted without a nodepool lands there.
		defaultNodePool := &v1alpha1.NodePool{}
		Expect(kubeClient.Get(ctx, client.ObjectKey{Name: "default"}, defaultNodePool)).To(Succeed())
		defaultNodePool.Spec.PreferredNetworkTopologyName = topologyName
		Expect(kubeClient.Update(ctx, defaultNodePool)).To(Succeed(), "failed to set PreferredNetworkTopologyName on the default nodepool")
	})

	AfterAll(func() {
		// Restore the default nodepool to having no topology (its state before this suite). Only the
		// topology this suite set is cleared, and the change must be persisted with an Update.
		defaultNodePool := &v1alpha1.NodePool{}
		if err := kubeClient.Get(ctx, client.ObjectKey{Name: "default"}, defaultNodePool); err != nil {
			fmt.Printf("failed to get default nodepool for cleanup: %v\n", err)
		} else if defaultNodePool.Spec.PreferredNetworkTopologyName == topologyName {
			defaultNodePool.Spec.PreferredNetworkTopologyName = ""
			if err := kubeClient.Update(ctx, defaultNodePool); err != nil {
				fmt.Printf("failed to restore default nodepool topology: %v\n", err)
			}
		}

		if err := kubeClient.Delete(ctx, topology); err != nil {
			fmt.Printf("failed to delete topology: %v\n", err)
		}
		for _, obj := range testCreatedObjects {
			if err := kubeClient.Delete(ctx, obj); err != nil {
				fmt.Printf("failed to delete resource %v: %v\n", obj, err)
			}
		}
		testConfig.Cleanup()
	})

	It("stamps the picked nodepool's network topology onto the workload's PodGroup", func() {
		podNamePrefix := common.GetRandomName("topo")
		workload, err := common.LoadWorkload(
			"../../resources/workspace.yaml",
			common.GetRandomName("topo-wl"),
			testConfig.TestNamespace.Name,
			podNamePrefix,
			&iwv2.InteractiveWorkload{},
			nil,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to load InteractiveWorkload")
		testCreatedObjects[podNamePrefix] = workload
		Expect(kubeClient.Create(ctx, workload)).To(Succeed(), "failed to create workload")

		// The pod-grouper creates the PodGroup from the workload's pod and annotates the pod with its
		// name; resolve that name first (the pod only needs to exist, not be scheduled).
		var podGroupName string
		Eventually(func(g Gomega) {
			pod, err := common.GetPodByPrefix(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(pod).ToNot(BeNil())
			podGroupName = pod.Annotations[commonconstants.PodGroupAnnotationForPod]
			g.Expect(podGroupName).ToNot(BeEmpty(), "pod is not yet grouped into a PodGroup")
		}, time.Second*60, time.Second*2).Should(Succeed(),
			fmt.Sprintf("expected a grouped pod with prefix %q in namespace %q", podNamePrefix, testConfig.TestNamespace.Name))

		// The pod-group-assigner should have stamped the default nodepool's topology onto the PodGroup.
		Eventually(func(g Gomega) {
			podGroup := &kaiv2alpha2.PodGroup{}
			g.Expect(kubeClient.Get(ctx, client.ObjectKey{Namespace: testConfig.TestNamespace.Name, Name: podGroupName}, podGroup)).To(Succeed())
			g.Expect(podGroup.Spec.TopologyConstraint).To(Equal(kaiv2alpha2.TopologyConstraint{
				Topology:               topologyName,
				PreferredTopologyLevel: lowestLevel,
			}), "expected the nodepool's topology to be stamped at the lowest level")
			g.Expect(podGroup.Annotations).To(HaveKeyWithValue(topologySourceAnnotationKey, topologySourceSystem),
				"expected the constraint to be marked system-sourced")
		}, time.Second*60, time.Second*2).Should(Succeed())
	})
})
