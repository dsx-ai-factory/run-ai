/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managed_nodes_config

import (
	"context"
	"time"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	rateLimiterBaseDelay = 500 * time.Millisecond
	rateLimiterMaxDelay  = 30 * time.Second
)

func (mncc *ManagedNodesConfigController) reconcileIncludedNodesThatShouldBeExcluded(ctx context.Context, mnc *v1alpha1.ManagedNodesConfig, includedNodes []corev1.Node) ([]corev1.Node, error) {
	nodes, err := mncc.getNodesMatchingExcludingNodeConfigAndNoLabel(mnc, includedNodes)
	if err != nil {
		log.Error().Msgf("Failed getting nodes that match excluding node config, err: %v", err.Error())
		return []corev1.Node{}, err
	} else if len(nodes.Items) == 0 {
		return []corev1.Node{}, nil
	}

	toBeExcludedNodes := []corev1.Node{}
	for i := range nodes.Items {
		node := &nodes.Items[i]

		log.Info().Msgf("Found node <%v> that should be excluded but it isnt", node.Name)
		schedulable, innerErr := mncc.ChangeNodePoolForNode(ctx, node, config.Get().ExcludedNodepoolName, utils.GetNodePoolNameFromLabels(node.Labels))
		err = utils.AppendErrIfNotNil(err, innerErr)

		if err == nil {
			if !schedulable {
				toBeExcludedNodes = append(toBeExcludedNodes, *node)
				innerErr = mncc.UpdateNodeLabels(ctx, node, map[string]string{config.Get().ShouldBeExcludedLabelKey: "true"})
			} else {
				if _, found := node.Labels[config.Get().ShouldBeExcludedLabelKey]; found {
					innerErr = mncc.UpdateNodeLabels(ctx, node, map[string]string{}, config.Get().ShouldBeExcludedLabelKey)
				}
			}
			err = utils.AppendErrIfNotNil(err, innerErr)
		}

	}

	return toBeExcludedNodes, err
}

func (mncc *ManagedNodesConfigController) reconcileExcludedNodesThatShouldBeIncluded(ctx context.Context, mnc *v1alpha1.ManagedNodesConfig, excludedNodes []corev1.Node) error {
	nodes, err := mncc.getNodesNotMatchingExcludingNodeConfigAndWithLabel(mnc, excludedNodes)
	if err != nil {
		log.Error().Msgf("Failed getting nodes that not match excluding node config, err: %v", err.Error())
		return err
	} else if len(nodes.Items) == 0 {
		return nil
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]

		log.Info().Msgf("Found node <%v> that should be included but it isnt", node.Name)
		_, innerErr := mncc.ChangeNodePoolForNode(ctx, node, config.Get().DefaultNodepoolName, config.Get().ExcludedNodepoolName)
		err = utils.AppendErrIfNotNil(err, innerErr)

		if _, found := node.Labels[config.Get().ShouldBeExcludedLabelKey]; found {
			innerErr = mncc.UpdateNodeLabels(ctx, node, map[string]string{}, config.Get().ShouldBeExcludedLabelKey)
		}
		err = utils.AppendErrIfNotNil(err, innerErr)

	}

	return err
}

func (mncc *ManagedNodesConfigController) reconcileMarkedToBeExcludedAndRevertedToBeIncluded(ctx context.Context, mnc *v1alpha1.ManagedNodesConfig, includedNodes []corev1.Node) error {
	actuallyIncludedWithLabel, err := mncc.getNodesNotMatchingExcludingNodeConfigAndWithLabel(mnc, includedNodes)
	if err != nil {
		return err
	}

	for _, includedWithLabelNode := range actuallyIncludedWithLabel.Items {
		if _, found := includedWithLabelNode.Labels[config.Get().ShouldBeExcludedLabelKey]; found {
			log.Info().Msgf("Found node <%v> that is going to be excluded but it shouldn't", includedWithLabelNode.Name)
			innerErr := mncc.NodePoolController.UpdateNodeLabels(ctx, &includedWithLabelNode, map[string]string{}, config.Get().ShouldBeExcludedLabelKey)
			err = utils.AppendErrIfNotNil(err, innerErr)
		}
	}

	if err != nil {
		return err
	}
	return nil
}
