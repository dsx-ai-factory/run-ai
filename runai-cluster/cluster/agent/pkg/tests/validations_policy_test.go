/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"

	. "github.com/onsi/gomega"
	policy_openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func _validatePolicyLabelsExistsEventually(name, namespace string, k8sClient client.Client, scope policy_openapi.ScopeType, policyType string) { //nolint:unused
	var policy interface{}
	var policyLabels map[string]string
	expectedLabels := _getExpectedLabels(scope)

	EventuallyWithOffset(1, func() []string {
		var keys []string

		switch policyType {
		case common.PolicyTypeInteractive:
			policy = getInteractivePolicyFromClient(name, namespace, k8sClient)
			if policy != nil {
				policyLabels = policy.(*ipv1.InteractivePolicy).Labels
			}

		case common.PolicyTypeTraining:
			policy = getTrainingPolicyFromClient(name, namespace, k8sClient)
			if policy != nil {
				policyLabels = policy.(*tpv1.TrainingPolicy).Labels
			}

		case common.PolicyTypeDistributed:
			policy = getDistributedPolicyFromClient(name, namespace, k8sClient)
			if policy != nil {
				policyLabels = policy.(*dpv1.DistributedPolicy).Labels
			}

		case common.PolicyTypeInference:
			policy = getInferencePolicyFromClient(name, namespace, k8sClient)
			if policy != nil {
				policyLabels = policy.(*infpv1.InferencePolicy).Labels
			}
		}

		for key := range policyLabels {
			keys = append(keys, key)
		}
		return keys
	}, DefaultTimeout).Should(Equal(expectedLabels))
}

func _getExpectedLabels(scope policy_openapi.ScopeType) []string { //nolint:unused
	expectedLabels := []string{clusterCommonConstants.PolicyLabel}
	switch scope {
	case policy_openapi.SCOPETYPE_TENANT:
		expectedLabels = append(expectedLabels, clusterCommonConstants.TenantScopeLabel)
	case policy_openapi.SCOPETYPE_PROJECT:
		expectedLabels = append(expectedLabels, clusterCommonConstants.ProjectScopeLabel)
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		expectedLabels = append(expectedLabels, clusterCommonConstants.DepartmentScopeLabel)
	}
	return expectedLabels
}

func validateTPCounterValueExistsEventually(name, namespace string, k8sClient client.Client, updatedCounterValue string) {
	var ip *tpv1.TrainingPolicy = nil
	EventuallyWithOffset(1, func() string {
		ip = getTrainingPolicyFromClient(name, namespace, k8sClient)
		return ip.Annotations[clusterCommonConstants.AnnotationsBackendPolicyCounter]
	}, DefaultTimeout).Should(Equal(updatedCounterValue))
}

func _validateTPImageValueExistsEventually(name, namespace string, k8sClient client.Client, specImageValue *string) { //nolint:unused
	var tp *tpv1.TrainingPolicy = nil
	EventuallyWithOffset(1, func() *string {
		tp = getTrainingPolicyFromClient(name, namespace, k8sClient)
		if tp.Spec.Image != nil {
			return tp.Spec.Image.Value
		}
		return nil
	}, DefaultTimeout).Should(Equal(specImageValue))
}

func _validateDPImageValueExistsEventually(name, namespace string, k8sClient client.Client, specImageValue *string) { //nolint:unused
	var dp *dpv1.DistributedPolicy = nil
	EventuallyWithOffset(1, func() *string {
		dp = getDistributedPolicyFromClient(name, namespace, k8sClient)
		if dp != nil {
			if dp.Spec.Image != nil {
				return dp.Spec.Image.Value
			}
		}
		return nil
	}, DefaultTimeout).Should(Equal(specImageValue))
}

func _validateDPMasterImageValueExistsEventually(name, namespace string, k8sClient client.Client, masterSpecImageValue *string) { //nolint:unused
	var dp *dpv1.DistributedPolicy = nil
	EventuallyWithOffset(1, func() *string {
		dp = getDistributedPolicyFromClient(name, namespace, k8sClient)
		if dp != nil {
			if dp.MasterSpec != nil && dp.MasterSpec.Image != nil {
				return dp.MasterSpec.Image.Value
			}
		}
		return nil
	}, DefaultTimeout).Should(Equal(masterSpecImageValue))
}

func validateIPCounterValueExistsEventually(name, namespace string, k8sClient client.Client, updatedCounterValue string) {
	var ip *ipv1.InteractivePolicy = nil
	EventuallyWithOffset(1, func() string {
		ip = getInteractivePolicyFromClient(name, namespace, k8sClient)
		return ip.Annotations[clusterCommonConstants.AnnotationsBackendPolicyCounter]
	}, DefaultTimeout).Should(Equal(updatedCounterValue))
}

func _validateIPExistsEventually(name, namespace string, k8sClient client.Client) { //nolint:unused
	var ip *ipv1.InteractivePolicy = nil
	EventuallyWithOffset(1, func() *ipv1.InteractivePolicy {
		ip = getInteractivePolicyFromClient(name, namespace, k8sClient)
		return ip
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateINFPCounterValueExistsEventually(name, namespace string, k8sClient client.Client, updatedCounterValue string) {
	var infp *infpv1.InferencePolicy = nil
	EventuallyWithOffset(1, func() string {
		infp = getInferencePolicyFromClient(name, namespace, k8sClient)
		return infp.Annotations[clusterCommonConstants.AnnotationsBackendPolicyCounter]
	}, DefaultTimeout).Should(Equal(updatedCounterValue))
}

func _validateINFPImageValueExistsEventually(name, namespace string, k8sClient client.Client, specImageValue *string) { //nolint:unused
	var infp *infpv1.InferencePolicy = nil
	EventuallyWithOffset(1, func() *string {
		infp = getInferencePolicyFromClient(name, namespace, k8sClient)
		if infp.Spec.Image != nil {
			return infp.Spec.Image.Value
		}
		return nil
	}, DefaultTimeout).Should(Equal(specImageValue))
}

func _validateTPExistsEventually(name, namespace string, k8sClient client.Client) { //nolint:unused
	var ip *tpv1.TrainingPolicy = nil
	EventuallyWithOffset(1, func() *tpv1.TrainingPolicy {
		ip = getTrainingPolicyFromClient(name, namespace, k8sClient)
		return ip
	}, DefaultTimeout).ShouldNot(BeNil())
}

func _validateDPExistsEventually(name, namespace string, k8sClient client.Client) { //nolint:unused
	var dp *dpv1.DistributedPolicy = nil
	EventuallyWithOffset(1, func() *dpv1.DistributedPolicy {
		dp = getDistributedPolicyFromClient(name, namespace, k8sClient)
		return dp
	}, DefaultTimeout).ShouldNot(BeNil())
}

func _validateINFPExistsEventually(name, namespace string, k8sClient client.Client) { //nolint:unused
	var infp *infpv1.InferencePolicy = nil
	EventuallyWithOffset(1, func() *infpv1.InferencePolicy {
		infp = getInferencePolicyFromClient(name, namespace, k8sClient)
		return infp
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateTPDeletedEventually(name, namespace string, k8sClient client.Client) {
	var ip *tpv1.TrainingPolicy = nil
	EventuallyWithOffset(1, func() *tpv1.TrainingPolicy {
		ip = getTrainingPolicyFromClient(name, namespace, k8sClient)
		return ip
	}, DefaultTimeout).Should(BeNil())
}

func validateDPDeletedEventually(name, namespace string, k8sClient client.Client) {
	var dp *dpv1.DistributedPolicy = nil
	EventuallyWithOffset(1, func() *dpv1.DistributedPolicy {
		dp = getDistributedPolicyFromClient(name, namespace, k8sClient)
		return dp
	}, DefaultTimeout).Should(BeNil())
}

func validateINFPDeletedEventually(name, namespace string, k8sClient client.Client) {
	var infp *infpv1.InferencePolicy = nil
	EventuallyWithOffset(1, func() *infpv1.InferencePolicy {
		infp = getInferencePolicyFromClient(name, namespace, k8sClient)
		return infp
	}, DefaultTimeout).Should(BeNil())
}

func validateAgentConfigExistEventually(k8sClient client.Client) {
	configMap := &corev1.ConfigMap{}

	EventuallyWithOffset(1, func() *corev1.ConfigMap {
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name:      configmaps.AgentConfigName,
			Namespace: clusterCommonConstants.NamespaceRunAi,
		}, configMap)
		if err != nil {
			return nil
		}
		return configMap
	}, DefaultTimeout).ShouldNot(BeNil())
}
