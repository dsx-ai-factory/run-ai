/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNodePoolNameFromLabels reads the assignment label from the given labels
// map and returns the node pool name found there. When the label is missing
// or empty, the caller-supplied defaultNodepoolName is returned. The label
// key and default name are supplied by the caller (rather than hardcoded)
// so this utility can be shared across binaries with different label
// vocabularies (runai vs OSS).
func GetNodePoolNameFromLabels(labels map[string]string,
	nodePoolAssignmentLabelKey, defaultNodepoolName string) string {
	nodePoolName, found := labels[nodePoolAssignmentLabelKey]
	if !found || nodePoolName == "" {
		nodePoolName = defaultNodepoolName
	}

	return nodePoolName
}

// GetAllNodePools lists every KAI node pool in the cluster. An empty cluster is an error:
// the caller resolves node affinity against these pools and cannot do so without any.
func GetAllNodePools(ctx context.Context, readerClient client.Reader) (*kaiv1alpha1.NodePoolList, error) {
	nodePoolList := &kaiv1alpha1.NodePoolList{}
	err := readerClient.List(ctx, nodePoolList)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing node pools, error: %s", err.Error())
		return nil, err
	}
	if len(nodePoolList.Items) == 0 {
		err = fmt.Errorf("didn't find any node pools in the cluster")
		log.Ctx(ctx).Error().Msg(err.Error())
		return nil, err
	}
	return nodePoolList, nil
}

// GetNodePoolsMap returns:
// * a map of: nodePoolKey to a map of: nodePoolValue to nodePoolName
// * a map for deleting node pools: nodePoolName to Phase
func GetNodePoolsMap(allNodePools []kaiv1alpha1.NodePool) (
	map[string]map[string]string, map[string]string) {
	nodePoolKeyToValueToName := make(map[string]map[string]string)
	deletingNodePools := make(map[string]string)

	for _, nodePool := range allNodePools {
		key := nodePool.Spec.LabelKey
		if _, found := nodePoolKeyToValueToName[key]; !found {
			nodePoolKeyToValueToName[key] = make(map[string]string)
		}

		nodePoolKeyToValueToName[key][nodePool.Spec.LabelValue] = nodePool.Name

		if nodePool.Status.Phase == kaiv1alpha1.NodePoolDeleting {
			deletingNodePools[nodePool.Name] = string(nodePool.Status.Phase)
		}
	}

	return nodePoolKeyToValueToName, deletingNodePools
}

func RemoveDuplicates[T string | int](list []T) []T {
	allKeys := make(map[T]bool)
	result := []T{}
	for _, item := range list {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			result = append(result, item)
		}
	}
	return result
}
