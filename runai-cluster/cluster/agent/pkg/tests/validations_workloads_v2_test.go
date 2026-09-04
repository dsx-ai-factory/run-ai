/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"
	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateWorkloadV2ExistsEventually(name, namespace string, gvk schema.GroupVersionKind, k8sClient client.Client) {
	var u *unstructured.Unstructured = nil
	EventuallyWithOffset(1, func() *unstructured.Unstructured {
		u = getWorkloadV2FromClient(name, namespace, gvk, k8sClient)
		return u
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateWorkloadV2DoesNotExistEventually(name, namespace string, gvk schema.GroupVersionKind, k8sClient client.Client) {
	var u *unstructured.Unstructured = nil
	EventuallyWithOffset(1, func() *unstructured.Unstructured {
		u = getWorkloadV2FromClient(name, namespace, gvk, k8sClient)
		return u
	}, DefaultTimeout).Should(BeNil())
}

func validateWorkloadV2CounterExistsEventually(name, namespace string, gvk schema.GroupVersionKind, updatedCounterValue int64, k8sClient client.Client) {
	updateCounterAsString := fmt.Sprintf("%d", updatedCounterValue)

	var u *unstructured.Unstructured = nil
	EventuallyWithOffset(1, func() string {
		u = getWorkloadV2FromClient(name, namespace, gvk, k8sClient)
		return client.Object(u).GetAnnotations()[clusterCommonConstants.AnnotationsBackendWorkloadCounter]
	}, DefaultTimeout).Should(Equal(updateCounterAsString))
}

func validateAgentConfigWithWorkloadV2Counters(k8sClient client.Client, counterValue string) {
	EventuallyWithOffset(1, func() string {
		configMap := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name:      configmaps.AgentConfigName,
			Namespace: clusterCommonConstants.NamespaceRunAi,
		}, configMap)
		if err != nil {
			return ""
		}
		counter, ok := configMap.Data[configmaps.WorkloadsV2SyncCounter]
		if !ok {
			return ""
		}
		return counter
	}, DefaultTimeout).Should(Equal(counterValue))
}
