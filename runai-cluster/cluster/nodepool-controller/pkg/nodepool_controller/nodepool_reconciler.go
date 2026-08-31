/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func (npc *NodePoolController) reconcileNodePool(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	log.Info().Msgf("Reconciling Node Pool <%v>", nodePool.Name)

	resultNodes, innerErr := npc.reconcileNodes(ctx, nodePool)
	err = utils.AppendErrIfNotNil(err, innerErr)

	innerErr = npc.reconcileScheduler(ctx, nodePool)
	err = utils.AppendErrIfNotNil(err, innerErr)

	innerErr = npc.reconcileNodePoolStatus(ctx, nodePool, resultNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	return
}

func (npc *NodePoolController) reconcileNodes(ctx context.Context, nodePool *v1alpha1.NodePool) ([]*corev1.Node, error) {
	log.Info().Msgf("Reconciling Nodes for Node Pool <%v>", nodePool.Name)

	var err error

	resultNodes := []*corev1.Node{}
	if nodePool.Name != config.Get().DefaultNodepoolName {
		alteredNodes, innerErr := npc.reconcileNodesMatchingNodePoolNameNoLabel(ctx, nodePool)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, alteredNodes...)

		alteredNodes, innerErr = npc.reconcileNodesInDefaultNodePoolMatchingOtherNodePoolLabel(ctx, nodePool)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, alteredNodes...)

		alteredNodes, innerErr = npc.reconcileNodesMatchingNodePoolNameAndLabel(ctx, nodePool)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, alteredNodes...)
	} else {
		alteredNodes, innerErr := npc.reconcileNodesInDefaultNodePool(ctx)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, alteredNodes...)

		alteredNodes, innerErr = npc.reconcileNodesNotInDefaultNodePool(ctx)
		err = utils.AppendErrIfNotNil(err, innerErr)
		resultNodes = append(resultNodes, alteredNodes...)
	}

	resultNodes, innerErr := npc.handleNodesOfNodePool(ctx, nodePool, resultNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	return resultNodes, err
}

func (npc *NodePoolController) reconcileNodePoolStatus(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	log.Info().Msgf("Reconciling Node Pool Status for <%v>", nodePool.Name)
	nodePoolNodes = npc.filterNodePoolNodes(nodePoolNodes, nodePool)

	nodePoolWithUpdatedStatus := nodePool.DeepCopy()
	var err error
	statusErr := npc.calculateNodePoolStatus(ctx, nodePoolWithUpdatedStatus, nodePoolNodes)
	err = utils.AppendErrIfNotNil(err, statusErr)

	nodepoolErr := npc.updateNodePoolStatusIfNeeded(ctx, nodePoolWithUpdatedStatus, nodePool)
	err = utils.AppendErrIfNotNil(err, nodepoolErr)

	annotErr := npc.updateAnnotationStatusFields(ctx, nodePoolWithUpdatedStatus, nodePoolNodes)
	err = utils.AppendErrIfNotNil(err, annotErr)
	return err
}
