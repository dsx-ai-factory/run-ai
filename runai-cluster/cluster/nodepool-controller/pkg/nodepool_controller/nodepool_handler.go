/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	"k8s.io/apimachinery/pkg/types"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (npc *NodePoolController) reconcileNodesMatchingNodePoolNameNoLabel(ctx context.Context,
	nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	nodes, err := npc.getNodesMatchingNodePoolNameNoLabel(ctx, nodePool)
	if err != nil {
		log.Error().Msgf("For nodepool <%v>, failed getting nodes that are assigned to nodepool but don't match nodepool's labels, err: %v",
			nodePool.Name, err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	nodePools, err := npc.getAvailableNodePools(ctx)
	if err != nil {
		return []*corev1.Node{}, err
	}

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]

		log.Info().Msgf("Found node <%v> with wrong nodepool assignment: <%v> (labels don't match)",
			node.Name, nodePool.Name)
		schedulable, innerErr := npc.findAndChangeNodePoolForNode(ctx, node, nodePool.Name, nodePools)
		err = utils.AppendErrIfNotNil(err, innerErr)
		if !schedulable {
			unschedulableErr := errors.New(fmt.Sprintf("Node <%v> is unschedulable", node.Name))
			err = utils.AppendErrIfNotNil(err, unschedulableErr)
		}

		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

func (npc *NodePoolController) reconcileNodesInDefaultNodePoolMatchingOtherNodePoolLabel(ctx context.Context,
	nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	nodes, err := npc.getNodesInDefaultNodePoolMatchingOtherNodePoolLabel(ctx, nodePool)
	if err != nil {
		log.Error().Msgf("For nodepool <%v>, failed getting nodes that are assigned to default nodepool and label matches nodepool <%v>, err: %v",
			nodePool.Name, nodePool.Name, err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		log.Info().Msgf("Found node <%v> that is assigned to default node pool: <%v>, but labels match nodepool <%v>",
			node.Name, config.Get().DefaultNodepoolName, nodePool.Name)
		schedulable, innerErr := npc.ChangeNodePoolForNode(ctx, node, nodePool.Name, config.Get().DefaultNodepoolName)
		err = utils.AppendErrIfNotNil(err, innerErr)
		if !schedulable {
			unschedulableErr := errors.New(fmt.Sprintf("Node <%v> is unschedulable", node.Name))
			err = utils.AppendErrIfNotNil(err, unschedulableErr)
		}
		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

// addOrUpdateNodePoolTopologyMismatchCondition In case of hasMismatch true - add new or update existing condition.
// In case of hasMismatch is false - updates existing condition in case of change of Status
// Does not add new condition on hasMismatch false
func (npc *NodePoolController) addOrUpdateNodePoolTopologyMismatchCondition(nodePool *v1alpha1.NodePool, mismatchNodes []*corev1.Node) {
	hasMismatch := len(mismatchNodes) > 0
	condition := v1alpha1.NodePoolCondition{
		Reason: v1alpha1.NodeTopologyMismatchReason,
		Type:   v1alpha1.NodeTopologyMismatch,
	}
	condition.SetConditionStatusValue(hasMismatch)
	nodePool.SetNodePoolCondition(condition)
	return
}

func (npc *NodePoolController) reconcileNodesMatchingNodePoolNameAndLabel(ctx context.Context,
	nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	nodes, err := npc.getNodesMatchingNodePoolNameAndLabel(ctx, nodePool)
	if err != nil {
		log.Error().Msgf("For nodepool <%v>, failed getting nodes that are assigned to nodepool and label matches, err: %v",
			nodePool.Name, err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		innerErr := npc.validateNodeStatusInNodePoolWrapper(ctx, node, nodePool.Name)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

func (npc *NodePoolController) reconcileNodesInDefaultNodePool(ctx context.Context) ([]*corev1.Node, error) {
	nodes, err := npc.getDefaultNodePoolNodes(ctx)
	if err != nil {
		log.Error().Msgf("Failed getting nodes assigned to default nodepool, err: %v", err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	nodePools, err := npc.getAvailableNodePools(ctx)
	if err != nil {
		return []*corev1.Node{}, err
	}

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]

		log.Debug().Msgf("Found node <%v> that is assigned to default node pool: <%v>, will try to find a new nodepool for it...",
			node.Name, config.Get().DefaultNodepoolName)
		foundNodePoolName := findMatchingNodePoolFromList(node, nodePools)

		if foundNodePoolName == config.Get().DefaultNodepoolName {
			innerErr := npc.validateDefaultNodePoolLabel(ctx, node)
			err = utils.AppendErrIfNotNil(err, innerErr)
			innerErr = npc.validateNodeStatusInNodePoolWrapper(ctx, node, config.Get().DefaultNodepoolName)
			err = utils.AppendErrIfNotNil(err, innerErr)
			resultNodes = append(resultNodes, node)
			continue
		}

		log.Info().Msgf("Found node <%v> with wrong nodepool assignment. Was: <%v>, should be: <%v>",
			node.Name, config.Get().DefaultNodepoolName, foundNodePoolName)
		schedulable, innerErr := npc.ChangeNodePoolForNode(ctx, node, foundNodePoolName, config.Get().DefaultNodepoolName)
		err = utils.AppendErrIfNotNil(err, innerErr)
		if !schedulable {
			unschedulableErr := errors.New(fmt.Sprintf("Node <%v> is unschedulable", node.Name))
			err = utils.AppendErrIfNotNil(err, unschedulableErr)
		}
		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

func (npc *NodePoolController) reconcileNodesNotInDefaultNodePool(ctx context.Context) ([]*corev1.Node, error) {
	nodes, err := npc.getNodesNotInDefaultNodePool(ctx)
	if err != nil {
		log.Error().Msgf("Failed getting nodes assigned to not-default nodepool, err: %v", err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	nodePools, err := npc.getAvailableNodePools(ctx)
	if err != nil {
		return []*corev1.Node{}, err
	}

	availableNodePoolsMap := map[string]string{
		// fake nodepool, need to consider it as always existing.
		config.Get().ExcludedNodepoolName: config.Get().ExcludedNodepoolName,
	}
	for _, nodePool := range nodePools {
		availableNodePoolsMap[nodePool.Name] = nodePool.Name
	}

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		nodeAssignedNodePool := utils.GetNodePoolNameFromLabels(node.Labels)

		if _, found := availableNodePoolsMap[nodeAssignedNodePool]; found {
			resultNodes = append(resultNodes, node)
			continue
		}

		log.Info().Msgf("Found node <%v> that is assigned to node pool: <%v>, but node pool doesn't exist",
			node.Name, nodeAssignedNodePool)
		schedulable, innerErr := npc.findAndChangeNodePoolForNode(ctx, node, nodeAssignedNodePool, nodePools)
		err = utils.AppendErrIfNotNil(err, innerErr)
		if !schedulable {
			unschedulableErr := errors.New(fmt.Sprintf("Node <%v> is unschedulable", node.Name))
			err = utils.AppendErrIfNotNil(err, unschedulableErr)
		}
		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

func (npc *NodePoolController) getAvailableNodePools(ctx context.Context) ([]v1alpha1.NodePool, error) {
	fieldSelector := fields.OneTermEqualSelector(common.IsDeletingPhaseField, "false")
	nodePoolsList, err := npc.listNodePoolsWithFieldSelector(ctx, fieldSelector)
	if err != nil {
		log.Error().Msgf("Failed listing available nodepools, err: %v", err.Error())
		return nil, err
	}

	return nodePoolsList.Items, nil
}

func (npc *NodePoolController) handleNodePoolDeletion(ctx context.Context, nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	log.Info().Msgf("Deleting NodePool <%v>; re-assigning nodes if needed", nodePool.Name)
	return npc.reassignNodeForDeletingNodePool(ctx, nodePool)
}

func (npc *NodePoolController) reassignNodeForDeletingNodePool(ctx context.Context, nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	nodePoolName := nodePool.Name

	nodes, err := npc.getNodePoolNodes(ctx, nodePoolName)
	if err != nil {
		log.Error().Msgf("While reassigning nodes for nodepool, failed getting nodes for nodepool <%v>, err: %v",
			nodePoolName, err.Error())
		return []*corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []*corev1.Node{}, nil
	}

	nodePools, err := npc.getAvailableNodePools(ctx)
	if err != nil {
		return []*corev1.Node{}, err
	}
	nodePools = utils.DeleteFromNodePoolList(nodePools, *nodePool)

	resultNodes := []*corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		log.Info().Msgf("Node <%v> is assigned to nodepool <%v> that is being deleted, will try to re-assign",
			node.Name, nodePoolName)

		_, innerErr := npc.findAndChangeNodePoolForNode(ctx, node, nodePoolName, nodePools)
		metrics.RemoveNodeNodePool(node.Name, nodePoolName)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, node)
	}

	return resultNodes, err
}

func (npc *NodePoolController) cleanupDeletedNodePool(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	nodes, err := npc.getNodePoolNodes(ctx, nodePool.Name)
	if err != nil {
		log.Error().Msgf("While deleting nodepool, failed getting nodes for nodepool <%v>, err: %v",
			nodePool.Name, err.Error())
		return
	} else if len(nodes.Items) == 0 {
		return
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]

		log.Info().Msgf("Node <%v> is assigned to nodepool <%v> that is being deleted during cluster delete, removing nodepool name label",
			node.Name, nodePool.Name)
		innerErr := npc.UpdateNodeLabels(ctx, node, map[string]string{}, config.Get().NodePoolNameLabel)
		err = utils.AppendErrIfNotNil(err, innerErr)
	}

	return
}

func (npc *NodePoolController) listNodePoolsWithFieldSelector(ctx context.Context, fieldSelector fields.Selector) (
	*v1alpha1.NodePoolList, error) {
	nodePools := &v1alpha1.NodePoolList{}
	err := npc.Client.List(ctx, nodePools, &client.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Error().Msgf("Failed listing nodepools with field selector <%v>, err: %v",
			fieldSelector.String(), err.Error())
		return nil, err
	}

	return nodePools, nil
}

func (npc *NodePoolController) getTopologyByName(ctx context.Context, topologyName string) (*kaiv1alpha1.Topology, error) {
	if topologyName == "" {
		return nil, nil
	}
	topology := &kaiv1alpha1.Topology{}
	err := npc.Client.Get(ctx, types.NamespacedName{Name: topologyName}, topology)
	if err != nil {
		return nil, err
	}

	return topology, nil
}

// getNodePoolCondition gets a condition on a nodepool by conditionType
func (npc *NodePoolController) getNodePoolCondition(nodepool *v1alpha1.NodePool, conditionType v1alpha1.NodePoolConditionType) *v1alpha1.NodePoolCondition {
	predicate := func(condition v1alpha1.NodePoolCondition, conditionType v1alpha1.NodePoolConditionType) bool {
		return condition.Type == conditionType
	}
	return getByPredicate(nodepool.Status.Conditions, conditionType, predicate)
}

type conditionTypePredicate[T any, D any] func(a T, b D) bool

// getByPredicate gets object by a predicate
func getByPredicate[T, D any](objsList []T, conditionType D, predicate conditionTypePredicate[T, D]) *T {
	for i := range objsList {
		obj := objsList[i]
		if predicate(obj, conditionType) {
			return &obj
		}
	}
	return nil
}
