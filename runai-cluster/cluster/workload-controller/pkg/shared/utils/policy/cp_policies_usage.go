/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"fmt"

	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logger.Log.WithName("policies.usage")

const (
	OldPolicyExistsConfigMapKey        = "old-policies-exists"
	ExistingPoliciesConfigMapName      = "wc-existing-policies"
	ExistingPoliciesConfigMapNamespace = "runai"
)

func GetExistingPoliciesIndicatorConfigMap(ctx context.Context, k8sClient client.Client) (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      ExistingPoliciesConfigMapName,
		Namespace: ExistingPoliciesConfigMapNamespace,
	}, &configMap)
	return configMap, err
}

func createExistingPoliciesIndicatorConfigMap(ctx context.Context, k8sClient client.Client, configMap *corev1.ConfigMap) error {
	configMap.Name = ExistingPoliciesConfigMapName
	configMap.Namespace = ExistingPoliciesConfigMapNamespace
	return k8sClient.Create(ctx, configMap)
}

func CheckOldPoliciesExistance(ctx context.Context, k8sClient client.Client) (bool, error) {
	configMap, err := GetExistingPoliciesIndicatorConfigMap(ctx, k8sClient)

	if err == nil && configMap.Data != nil {
		if value, ok := configMap.Data[OldPolicyExistsConfigMapKey]; ok {
			log.Info("found existing policy indicator in configmap")
			return value == "true", nil
		}
	} else {
		log.Error(err, "failed to get existing policy indicator from configmap, running existing policies checks to determine if cluster policy is present")
	}

	defer func() {
		err = createExistingPoliciesIndicatorConfigMap(ctx, k8sClient, &configMap)
		if err != nil {
			log.Error(err, "failed to create existing policy indicator in configmap")
		} else {
			log.Info("created existing policy indicator in configmap")
		}
	}()

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	listInteractive := &ipv1.InteractivePolicyList{}
	listTraining := &tpv1.TrainingPolicyList{}
	listDistributed := &dpv1.DistributedPolicyList{}

	listPoliciesTypes := []client.ObjectList{
		listInteractive,
		listTraining,
		listDistributed,
	}

	for _, listPolicy := range listPoliciesTypes {
		amount := 0
		labelSelector, labelErr := createNotExistingLabelSelector(clusterCommonConstants.PolicyLabel)
		if labelErr != nil {
			log.Error(labelErr, "failed to create a label selector")
			return false, err
		}

		err := k8sClient.List(ctx, listPolicy, labelSelector, client.Limit(1))
		if err != nil {
			return false, err
		}

		switch list := listPolicy.(type) {
		case *ipv1.InteractivePolicyList:
			amount = len(list.Items)
		case *tpv1.TrainingPolicyList:
			amount = len(list.Items)
		case *dpv1.DistributedPolicyList:
			amount = len(list.Items)
		default:
			return false, fmt.Errorf("unsupported policy type %T", listPolicy)
		}

		if amount > 0 {
			configMap.Data[OldPolicyExistsConfigMapKey] = clusterCommonConstants.BoolTrueAsString
			return true, nil
		}

	}
	configMap.Data[OldPolicyExistsConfigMapKey] = clusterCommonConstants.BoolFalseAsString
	return false, nil

}

func SetupControlPlanePoliciesUsage(ctx context.Context, k8sClient client.Client) bool {
	useControlPlanePolicies := true

	clusterPoliciesExist, err := CheckOldPoliciesExistance(ctx, k8sClient)

	if err != nil {
		log.Error(err, "Failed to check for existing policies, control plane policies will be enabled")
	}

	if err == nil && clusterPoliciesExist {
		if utils.IsApplyControlPlanePoliciesEnabled() {
			log.Info("Control plane policies enabled")
		} else {
			useControlPlanePolicies = false
			log.Info("control plane policies disabled")
		}
	} else {
		log.Info("No existing policies, control plane policies enabled")
	}

	return useControlPlanePolicies
}

func createNotExistingLabelSelector(key string) (client.ListOption, error) {
	r, err := labels.NewRequirement(key, selection.DoesNotExist, nil)
	if err != nil {
		return nil, err
	}
	return client.MatchingLabelsSelector{Selector: labels.NewSelector().Add(*r)}, nil
}
