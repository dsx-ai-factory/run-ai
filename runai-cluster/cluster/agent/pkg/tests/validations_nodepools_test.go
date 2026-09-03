/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateNodePoolExistsEventually(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *v1alpha1.NodePool {
		return getNodePoolFromClient(nodePoolName, k8sClient)
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateNodePoolDeletedEventually(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		np := getNodePoolFromClient(nodePoolName, k8sClient)
		g.Expect(np).To(BeNil())
	}, DefaultTimeout).Should(Succeed())
}

func validateExistingNodePoolMatchesBackendSpec(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingNodePoolWithBackendSpec(g, nodePoolName, k8sClient)
	}, DefaultTimeout).Should(Succeed())
}

func validateExistingNodePoolDoesNotMatchBackendSpec(nodePoolName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingNodePoolWithBackendSpec(g, nodePoolName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingNodePoolWithBackendSpec(g Gomega, nodePoolName string, k8sClient client.Client) {
	clusterNodePool := getNodePoolFromClient(nodePoolName, k8sClient)
	g.ExpectWithOffset(1, clusterNodePool).ToNot(BeNil())

	backendNPSpec := OpenApiNodepoolsMap[nodePoolName]
	g.ExpectWithOffset(1, backendNPSpec).ToNot(BeNil(), "Backend nodepool spec should exist")

	g.ExpectWithOffset(1, clusterNodePool.Spec.LabelKey).To(Equal(backendNPSpec.LabelKey), "LabelKey")
	g.ExpectWithOffset(1, clusterNodePool.Spec.LabelValue).To(Equal(backendNPSpec.LabelValue), "LabelValue")

	g.ExpectWithOffset(1, clusterNodePool.Status.Phase).To(Equal(v1alpha1.NodePoolPhase("")), "NodePoolPhase")
	g.ExpectWithOffset(1, clusterNodePool.Status.Message).To(Equal(""), "StatusMessage")
	g.ExpectWithOffset(1, clusterNodePool.Status.Nodes).To(BeEmpty(), "StatusNodes")
}
