/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type keyValuePair struct {
	Key   string
	Value string
}

// mutateNodeAffinityForDefaultNodePools resolves the node pools a pod should run
// on and adds them as required node affinity:
//  1. an explicit node-pool label on the pod wins;
//  2. otherwise the project's defaultNodePools are applied, but only if the pod
//     does not already reference a node pool in its node affinity.
//
// It is fail-open: any resolution error is logged and the pod is left untouched,
// so a misconfiguration never blocks pod creation.
func (pm *PodMutator) mutateNodeAffinityForDefaultNodePools(ctx context.Context, pod *corev1.Pod, namespace string) {
	projectName, err := utils.GetProjectNameOfNamespace(ctx, pm.client, namespace)
	if err != nil {
		log.Ctx(ctx).Info().Msgf(
			"could not identify the project related to pod <%s/%s> (err: <%s>); skipping default node pools mutation",
			namespace, pod.Name, err.Error())
		return
	}

	if nodePoolName, found := pod.Labels[config.Config().NodePoolLabelKey]; found {
		log.Ctx(ctx).Info().Msgf("found node pool label on pod <%s/%s>, node pool: <%s>",
			namespace, pod.Name, nodePoolName)

		if err = pm.setNodeAffinityForNodePools(ctx, pod, []string{nodePoolName}); err != nil {
			log.Ctx(ctx).Error().Msgf(
				"failed to add nodepool <%s> from label to the pod's node affinity, err: <%s>",
				nodePoolName, err.Error())
		}
		return
	}

	project := &v1alpha1.Project{}
	if err = pm.client.Get(ctx, types.NamespacedName{Name: projectName}, project); err != nil {
		log.Ctx(ctx).Error().Msgf("could not get kai project <%s>, err: <%s>", projectName, err.Error())
		return
	}
	defaultNodePools := project.Spec.DefaultNodePools

	if len(defaultNodePools) == 0 {
		log.Ctx(ctx).Info().Msgf("no default node pools defined for project <%s>, do nothing", projectName)
		return
	}

	hasNodePool, err := pm.hasNodePoolInPodsNodeAffinity(ctx, pod)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("failed to calculate current pod's node pools, err: <%s>", err.Error())
		return
	}
	if hasNodePool {
		log.Ctx(ctx).Info().Msgf(
			"there are existing node pools in pod's node affinity, pod <%s/%s> should not be mutated - do nothing",
			namespace, pod.Name)
		return
	}

	if err = pm.setNodeAffinityForNodePools(ctx, pod, defaultNodePools); err != nil {
		log.Ctx(ctx).Error().Msgf(
			"failed to add project <%s> default node pools to the pod's node affinity, err: <%s>",
			projectName, err.Error())
	}
}

func (pm *PodMutator) hasNodePoolInPodsNodeAffinity(ctx context.Context, pod *corev1.Pod) (bool, error) {
	if pod.Spec.Affinity == nil ||
		pod.Spec.Affinity.NodeAffinity == nil ||
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return false, nil
	}
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

	nodePoolsMapping, err := pm.getNodePoolsMapping(ctx)
	if err != nil {
		return false, err
	}

	return hasNodePoolInNodeAffinity(terms, nodePoolsMapping), nil
}

func hasNodePoolInNodeAffinity(terms []corev1.NodeSelectorTerm, nodePoolsMapping map[keyValuePair]bool) bool {
	for _, term := range terms {
		for _, matchExpression := range term.MatchExpressions {
			if matchExpression.Operator != corev1.NodeSelectorOpIn {
				if isMatchExpressionOfDefaultNodePool(matchExpression) {
					return true
				}
				continue
			}

			if isMatchExpressionOfCustomNodePool(matchExpression, nodePoolsMapping) {
				return true
			}
		}
	}

	return false
}

func isMatchExpressionOfCustomNodePool(
	matchExpression corev1.NodeSelectorRequirement,
	nodePoolsMapping map[keyValuePair]bool) bool {
	for _, value := range matchExpression.Values {
		if nodePoolsMapping[keyValuePair{Key: matchExpression.Key, Value: value}] {
			return true
		}
	}
	return false
}

func isMatchExpressionOfDefaultNodePool(expression corev1.NodeSelectorRequirement) bool {
	return expression.Operator == corev1.NodeSelectorOpDoesNotExist &&
		expression.Key == config.Config().NodePoolLabelKey
}

// getNodePoolsMapping lists the node pools and returns their affinity label key/value pairs.
func (pm *PodMutator) getNodePoolsMapping(ctx context.Context) (map[keyValuePair]bool, error) {
	mapping := map[keyValuePair]bool{}
	defaultNodePoolName := config.Config().DefaultNodepoolName

	nodePoolList := &v1alpha1.NodePoolList{}
	if err := pm.client.List(ctx, nodePoolList); err != nil {
		return nil, err
	}
	for i := range nodePoolList.Items {
		nodePool := &nodePoolList.Items[i]
		if nodePool.Name == defaultNodePoolName {
			continue
		}
		mapping[keyValuePair{Key: nodePool.Spec.LabelKey, Value: nodePool.Spec.LabelValue}] = true
	}

	return mapping, nil
}

func (pm *PodMutator) setNodeAffinityForNodePools(ctx context.Context, pod *corev1.Pod, nodePoolsNames []string) error {
	matchExpressionsToAdd := []corev1.NodeSelectorRequirement{}
	for _, nodePoolName := range nodePoolsNames {
		matchExpression, created, err := pm.createMatchExpressionForNodePool(ctx, nodePoolName)
		if err != nil {
			log.Ctx(ctx).Error().Msgf("failed to create a match expression for node pool <%s>, err: <%s>",
				nodePoolName, err.Error())
			return err
		}
		if created {
			matchExpressionsToAdd = append(matchExpressionsToAdd, matchExpression)
		}
	}

	if len(matchExpressionsToAdd) > 0 {
		AddNodeAffinityToPodSpec(&pod.Spec, matchExpressionsToAdd)
	}

	return nil
}

func (pm *PodMutator) createMatchExpressionForNodePool(
	ctx context.Context, nodePoolName string) (corev1.NodeSelectorRequirement, bool, error) {
	if nodePoolName == config.Config().DefaultNodepoolName {
		return corev1.NodeSelectorRequirement{
			Key:      config.Config().NodePoolLabelKey,
			Operator: corev1.NodeSelectorOpDoesNotExist,
		}, true, nil
	}

	phase, labelKey, labelValue, err := utils.GetNodePoolFields(ctx, pm.client, nodePoolName)
	if err != nil {
		if errors.IsNotFound(err) {
			return corev1.NodeSelectorRequirement{}, false, fmt.Errorf("node pool %s does not exist", nodePoolName)
		}
		return corev1.NodeSelectorRequirement{}, false, err
	}

	if phase == string(v1alpha1.NodePoolDeleting) {
		log.Ctx(ctx).Error().Msgf("node pool <%s> is <%s> - can't submit to this node pool",
			nodePoolName, v1alpha1.NodePoolDeleting)
		return corev1.NodeSelectorRequirement{}, false, nil
	}

	return corev1.NodeSelectorRequirement{
		Key:      labelKey,
		Operator: corev1.NodeSelectorOpIn,
		Values:   []string{labelValue},
	}, true, nil
}
