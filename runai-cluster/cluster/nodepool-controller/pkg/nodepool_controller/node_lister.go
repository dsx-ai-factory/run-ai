/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (npc *NodePoolController) getNodePoolNodes(ctx context.Context, nodePoolName string) (nodes *corev1.NodeList, err error) {
	matchNodePool, err := utils.GetNodePoolLabelSelector(nodePoolName)
	if err != nil {
		return nil, err
	}

	return npc.ListNodesWithRequirements(ctx, *matchNodePool)
}

func (npc *NodePoolController) filterNodePoolNodes(alteredNodes []*corev1.Node,
	nodePool *v1alpha1.NodePool) []*corev1.Node {
	filteredNodesMap := map[string]*corev1.Node{}
	for _, node := range alteredNodes {
		nodePoolOfNode := utils.GetNodePoolNameFromLabels(node.Labels)
		if nodePoolOfNode == nodePool.Name {
			filteredNodesMap[node.Name] = node
		}
	}

	// avoided duplicates using map - now copy to array
	filteredNodes := []*corev1.Node{}
	for _, node := range filteredNodesMap {
		filteredNodes = append(filteredNodes, node)
	}

	return filteredNodes
}

// getNodesAssignedToNodePoolWaitingForDrain -
// we want to list the nodes that are:
//   - matching the nodepool label key and value (meaning, they are assigned to the nodepool)
//   - not matching the nodepool name label (meaning, they are not yet migrated to the nodepool)
//   - unschedulable + have the "unschedulable by us" label set to true
//     (meaning, we marked them as unschedulable - they are waiting for drain and not just unschedulable for other reasons)
func (npc *NodePoolController) getNodesAssignedToNodePoolWaitingForDrain(ctx context.Context, nodePool *v1alpha1.NodePool) ([]string, error) {
	return npc.getNodesUnschedulableByUsMatchingNodePoolLabelNotMatchingNodePoolName(ctx, nodePool)
}

func (npc *NodePoolController) getNodesMatchingNodePoolNameNoLabel(ctx context.Context, nodePool *v1alpha1.NodePool) (nodes *corev1.NodeList, err error) {
	matchNodePoolName, err := utils.RequirementNodePoolNameMatching(nodePool.Name)
	if err != nil {
		return nil, err
	}
	noNodePoolLabel, err := utils.RequirementNodePoolLabelNotMatching(nodePool.Spec.LabelKey, nodePool.Spec.LabelValue)
	if err != nil {
		return nil, err
	}

	return npc.ListNodesWithRequirements(ctx, *matchNodePoolName, *noNodePoolLabel)
}

func (npc *NodePoolController) getNodesInDefaultNodePoolMatchingOtherNodePoolLabel(ctx context.Context, nodePool *v1alpha1.NodePool) (nodes *corev1.NodeList, err error) {
	matchDefaultNodePool, err := utils.GetNodePoolLabelSelector(config.Get().DefaultNodepoolName)
	if err != nil {
		return nil, err
	}
	matchNodePoolLabel, err := utils.RequirementNodePoolLabelMatching(nodePool.Spec.LabelKey, nodePool.Spec.LabelValue)
	if err != nil {
		return nil, err
	}

	return npc.ListNodesWithRequirements(ctx, *matchDefaultNodePool, *matchNodePoolLabel)
}

func (npc *NodePoolController) getNodesMatchingNodePoolNameAndLabel(ctx context.Context, nodePool *v1alpha1.NodePool) (nodes *corev1.NodeList, err error) {
	matchNodePoolName, err := utils.RequirementNodePoolNameMatching(nodePool.Name)
	if err != nil {
		return nil, err
	}
	matchNodePoolLabel, err := utils.RequirementNodePoolLabelMatching(nodePool.Spec.LabelKey, nodePool.Spec.LabelValue)
	if err != nil {
		return nil, err
	}

	return npc.ListNodesWithRequirements(ctx, *matchNodePoolName, *matchNodePoolLabel)
}

// getNodesUnschedulableByUsMatchingNodePoolLabelNotMatchingNodePoolName
// The function name is intentionally descriptive to convey the exact matching conditions.
func (npc *NodePoolController) getNodesUnschedulableByUsMatchingNodePoolLabelNotMatchingNodePoolName(ctx context.Context, nodePool *v1alpha1.NodePool) ([]string, error) {
	matchNotNodePoolNameLabel, err := utils.GetNotNodePoolLabelSelector(nodePool.Name)
	if err != nil {
		return nil, err
	}

	matchUnschedulableByUsLabel, err := utils.RequirementNodePoolLabelMatching(config.Get().UnschedulableLabelKey, MarkedUnschedulableByNodePoolControllerLabelValue)
	if err != nil {
		return nil, err
	}

	requirements := []labels.Requirement{*matchNotNodePoolNameLabel, *matchUnschedulableByUsLabel}

	if nodePool.Spec.LabelKey != "" && nodePool.Spec.LabelValue != "" {
		matchNodePoolLabel, err := utils.RequirementNodePoolLabelMatching(nodePool.Spec.LabelKey, nodePool.Spec.LabelValue)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, *matchNodePoolLabel)
	}

	nodes, err := npc.ListNodesWithRequirements(ctx, requirements...)
	if err != nil {
		return nil, err
	}

	// we listed using label selectors,
	// we still need to validate the node has Unschedulable in spec.
	unschedulableNodeNames := []string{}
	for _, node := range nodes.Items {
		if node.Spec.Unschedulable {
			unschedulableNodeNames = append(unschedulableNodeNames, node.Name)
		}
	}
	return unschedulableNodeNames, nil
}

func (npc *NodePoolController) getNodesNotInDefaultNodePool(ctx context.Context) (nodes *corev1.NodeList, err error) {
	matchNotDefaultNodePool, err := utils.GetNotNodePoolLabelSelector(config.Get().DefaultNodepoolName)
	if err != nil {
		return nil, err
	}
	return npc.ListNodesWithRequirements(ctx, *matchNotDefaultNodePool)
}

func (npc *NodePoolController) getDefaultNodePoolNodes(ctx context.Context) (nodes *corev1.NodeList, err error) {
	return npc.getNodePoolNodes(ctx, config.Get().DefaultNodepoolName)
}

func (npc *NodePoolController) ListNodesWithRequirements(ctx context.Context, requirements ...labels.Requirement) (nodes *corev1.NodeList, err error) {
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(requirements...)

	return npc.listNodesWithSelector(ctx, labelSelector)
}

func (npc *NodePoolController) listNodesWithSelector(ctx context.Context, labelSelector labels.Selector) (nodes *corev1.NodeList, err error) {
	nodes = &corev1.NodeList{}
	err = npc.Client.List(ctx, nodes, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Error().Msgf("Failed listing nodes with label selector <%v>, err: %v",
			labelSelector.String(), err.Error())
		return nil, err
	}

	return nodes, err
}
