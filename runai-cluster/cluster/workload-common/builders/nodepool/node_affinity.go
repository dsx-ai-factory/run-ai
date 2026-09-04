/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package nodepool resolves node pools and turns them into node affinity.
//
// It is deliberately a separate package from workload-common/builders: reading the
// kai.resources NodePool CRD pulls KAI-scheduler's pkg/apis/kai/v1 into the import
// graph, and that package reaches KAI's scheduler internals. Keeping node pool resolution
// out of builders keeps the control plane clear of chain imports.
package nodepool

import (
	"context"
	"fmt"
	"strings"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateNodeAffinityForNodePools(ctx context.Context, k8sClient client.Client, nodePools string, nodePool string, affinity **corev1.Affinity) error {
	nodepoolList, err := GetNodePools(ctx, k8sClient, nodePools, nodePool)
	if err != nil {
		return err
	}
	if len(nodepoolList) == 0 {
		return nil
	}
	return NodeAffinityFromNodePools(nodepoolList, affinity)
}

func GetNodePools(ctx context.Context, k8sClient client.Client, nodePools string, nodePool string) ([]kaiv1alpha1.NodePool, error) {
	logger := log.FromContext(ctx)
	nodePoolsNames, err := getNodePoolsNamesFromParams(ctx, nodePools, nodePool)
	if err != nil {
		return nil, err
	}
	logger.V(4).Info("found node pools", "nodePools", nodePoolsNames)
	nodepoolList := make([]kaiv1alpha1.NodePool, 0)
	for _, nodePoolName := range nodePoolsNames {
		var nodePoolObj *kaiv1alpha1.NodePool
		if nodePoolName == "" {
			continue
		}
		nodePoolObj, err = getNodePool(ctx, k8sClient, nodePoolName)
		if err != nil {
			return nil, err
		}
		nodepoolList = append(nodepoolList, *nodePoolObj)

	}
	return nodepoolList, nil
}

func NodeAffinityFromNodePools(nodePools []kaiv1alpha1.NodePool, affinity **corev1.Affinity) error {
	matchExpressions := make([]corev1.NodeSelectorTerm, 0)
	for _, nodePool := range nodePools {
		matchExpression := corev1.NodeSelectorRequirement{}

		CreateMatchExpressionForNodePool(&nodePool, &matchExpression)
		matchExpressions = append(matchExpressions, corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{matchExpression},
		})
	}
	// when using affinity should be careful to not override existing affinity
	// both pod affinity and node affinity are part of affinity
	nodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: matchExpressions,
		},
	}

	if (*affinity) == nil {
		*affinity = &corev1.Affinity{
			NodeAffinity: nodeAffinity,
		}
	} else if (*affinity).NodeAffinity == nil {
		(*affinity).NodeAffinity = nodeAffinity
	} else if (*affinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		(*affinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	} else {
		(*affinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(
			(*affinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			matchExpressions...,
		)
	}

	return nil
}

func CreateMatchExpressionForNodePool(nodePool *kaiv1alpha1.NodePool, matchExpression *corev1.NodeSelectorRequirement) {
	if nodePool == nil {
		return
	}
	if nodePool.Name == commonconstants.DefaultNodePoolName {
		// the default node pool is different from any other nodepool
		// its a pools of nodes that not belong to any existing node pool
		*matchExpression = corev1.NodeSelectorRequirement{
			Key:      commonconstants.NodePoolNameLabel,
			Operator: corev1.NodeSelectorOpDoesNotExist,
		}
	} else {
		*matchExpression = corev1.NodeSelectorRequirement{
			Key:      nodePool.Spec.LabelKey,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{nodePool.Spec.LabelValue},
		}
	}
}

func getNodePool(ctx context.Context, k8sClient client.Client, nodePoolName string) (*kaiv1alpha1.NodePool, error) {
	nodePool := &kaiv1alpha1.NodePool{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: nodePoolName}, nodePool)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("error: node pool %s does not exist %w", nodePoolName, err)
		}
		return nil, err
	}

	if nodePool.Status.Phase == kaiv1alpha1.NodePoolDeleting {
		return nil, fmt.Errorf("error: node pool %s is %s - can't submit to this node pool", nodePoolName, kaiv1alpha1.NodePoolDeleting)
	}
	return nodePool, nil
}

func getNodePoolsNamesFromParams(ctx context.Context, nodePools, nodePool string) ([]string, error) {
	logger := log.FromContext(ctx)

	if nodePools == "" && nodePool == "" {
		return nil, nil
	}
	if nodePools == "" {
		return []string{nodePool}, nil
	}

	nodePoolsList := strings.Split(nodePools, " ")

	nodePoolsSet := make(map[string]bool)
	for _, np := range nodePoolsList {
		nodePoolsSet[np] = true
	}
	if len(nodePoolsList) != len(nodePoolsSet) {
		// do not change error string - it is a UX requirement
		err := fmt.Errorf("node pools names must be unique. Remove duplicated names")
		logger.Error(err, fmt.Sprintf("got nodePools: %s", nodePools))
		return []string{}, err
	}
	return nodePoolsList, nil
}
