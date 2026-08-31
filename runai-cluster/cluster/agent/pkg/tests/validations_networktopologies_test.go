/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/adapters"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getNetworkTopologyFromClient(name string, k8sClient client.Client) *kaiv1alpha1.Topology {
	networkTopology := &kaiv1alpha1.Topology{}
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: name}, networkTopology)
	if err != nil {
		return nil
	}
	return networkTopology
}

func validateNetworkTopologyExistsEventually(networkTopologyName string, k8sClient client.Client) {
	Eventually(func(g Gomega) {
		networkTopology := getNetworkTopologyFromClient(networkTopologyName, k8sClient)
		fmt.Printf("Network Topology: Expecting to exist %s, %v\n", networkTopologyName, networkTopology)
		g.Expect(networkTopology).ToNot(BeNil())
	}, DefaultTimeout).Should(Succeed())
}

func validateNetworkTopologyDeletedEventually(networkTopologyName string, k8sClient client.Client) {
	Eventually(func(g Gomega) {
		networkTopology := getNetworkTopologyFromClient(networkTopologyName, k8sClient)
		fmt.Printf("Network Topology: Expecting to BE NIL %s, %v\n", networkTopologyName, networkTopology)
		g.Expect(networkTopology).To(BeNil())
	}, DefaultTimeout).Should(Succeed())
}

func validateExistingNetworkTopologyMatchesBackendSpec(networkTopologyName, source string, k8sClient client.Client) {
	networkTopology := getNetworkTopologyFromClient(networkTopologyName, k8sClient)
	Expect(networkTopology).ToNot(BeNil())

	backendNetworkTopology := BackendNetworkTopologiesMap[networkTopologyName]
	Expect(backendNetworkTopology).ToNot(BeNil())

	// Validate levels match
	Expect(len(networkTopology.Spec.Levels)).To(Equal(len(backendNetworkTopology.Levels)))
	for i, level := range networkTopology.Spec.Levels {
		Expect(level.NodeLabel).To(Equal(backendNetworkTopology.Levels[i]))
	}
	Expect(networkTopology.Labels[adapters.NetworkTopologySourceLabel]).To(Equal(source))
}
