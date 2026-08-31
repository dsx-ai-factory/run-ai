/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/gomega"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadTypeConfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/workload-type-config"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateAgentConfigWithWorkloadCounters(k8sClient client.Client, counterKey string, counterValue string) {
	EventuallyWithOffset(1, func() string {
		configMap := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name:      configmaps.AgentConfigName,
			Namespace: clusterCommonConstants.NamespaceRunAi,
		}, configMap)
		if err != nil {
			return ""
		}
		counter, ok := configMap.Data[counterKey]
		if !ok {
			return ""
		}
		return counter
	}, DefaultTimeout).Should(Equal(counterValue))
}

func validateWorkloadsConfigMap(k8sClient client.Client) {
	// Verify config map exists with correct data
	Eventually(func(g Gomega) {
		cm := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      clusterCommonConstants.WorkloadsConfigMapName,
			Namespace: clusterCommonConstants.WorkloadsConfigMapNamespace,
		}, cm)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(cm.Data).ToNot(BeNil())
		g.Expect(cm.Data[clusterCommonConstants.WorkloadsConfigMapTypesKey]).ToNot(BeEmpty())
	}, EventuallyTimeout).Should(Succeed())

	// Verify config map data contains expected workload types
	Eventually(func(g Gomega) {
		cm := &corev1.ConfigMap{}
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      clusterCommonConstants.WorkloadsConfigMapName,
			Namespace: clusterCommonConstants.WorkloadsConfigMapNamespace,
		}, cm)
		g.Expect(err).ToNot(HaveOccurred())

		var configs []workloadTypeConfig.WorkloadTypeConfig
		err = json.Unmarshal([]byte(cm.Data[clusterCommonConstants.WorkloadsConfigMapTypesKey]), &configs)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(configs).ToNot(BeEmpty())
		g.Expect(len(configs)).To(Equal(3))
		g.Expect(configs[0].TypeName).To(Equal(backendWorkloadTypeConfigs[0].Name))
		g.Expect(configs[0].TypeName).To(Equal(backendWorkloadTypeConfigs[0].Name))
		g.Expect(configs[1].Group).To(Equal(backendWorkloadTypeConfigs[1].Group))
		g.Expect(configs[1].Group).To(Equal(backendWorkloadTypeConfigs[1].Group))
		g.Expect(configs[2].TypeName).To(Equal(backendWorkloadTypeConfigs[2].Name))
		g.Expect(configs[2].TypeName).To(Equal(backendWorkloadTypeConfigs[2].Name))
	}, EventuallyTimeout).Should(Succeed())
}

func validateWorkloadsPriorities(name string, value int32, id string, source string, k8sClient client.Client) {
	Eventually(func(g Gomega) {
		pc := getPriorityClassByName(name, k8sClient)
		g.Expect(pc).ToNot(BeNil())
		g.Expect(pc.Value).To(Equal(value))
		g.Expect(pc.Labels[clusterCommonConstants.PriorityClassIdLabel]).To(Equal(id))
		g.Expect(pc.Labels[clusterCommonConstants.PriorityClassSourceLabel]).To(Equal(source))
	}, DefaultTimeout).Should(Succeed())
}

func validateWorkloadsPrioritiesNotExisting(name string, k8sClient client.Client) {
	Eventually(func(g Gomega) {
		pc := getPriorityClassByName(name, k8sClient)
		g.Expect(pc).To(BeNil())
	}, DefaultTimeout).Should(Succeed())
}

func getPriorityClassByName(name string, k8sClient client.Client) *schedulingv1.PriorityClass {
	pcKey := types.NamespacedName{Name: name}
	pcObj := &schedulingv1.PriorityClass{}
	err := k8sClient.Get(context.Background(), pcKey, pcObj)
	if err != nil {
		return nil
	}
	return pcObj
}
