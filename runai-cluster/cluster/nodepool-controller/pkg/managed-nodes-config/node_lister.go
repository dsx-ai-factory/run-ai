/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managed_nodes_config

import (
	"context"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	helper "k8s.io/component-helpers/scheduling/corev1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (mncc *ManagedNodesConfigController) getNodesMatchingExcludingNodeConfigAndNoLabel(mnc *v1alpha1.ManagedNodesConfig, includedNodes []corev1.Node) (nodes *corev1.NodeList, err error) {
	nodesMatchingExclusion := filterNodesBasedOnNodeSelector(includedNodes, mnc.Spec.InclusionCriteria, false)

	return &corev1.NodeList{Items: nodesMatchingExclusion}, nil
}

func (mncc *ManagedNodesConfigController) getNodesNotMatchingExcludingNodeConfigAndWithLabel(mnc *v1alpha1.ManagedNodesConfig, nodes []corev1.Node) (*corev1.NodeList, error) {
	nodesMatchingInclusion := filterNodesBasedOnNodeSelector(nodes, mnc.Spec.InclusionCriteria, true)

	return &corev1.NodeList{Items: nodesMatchingInclusion}, nil
}

func filterNodesBasedOnNodeSelector(nodes []corev1.Node, selectorTerm corev1.NodeSelector, keepNodesThatMatch bool) []corev1.Node {
	var filtered []corev1.Node
	if len(selectorTerm.NodeSelectorTerms) == 0 {
		if keepNodesThatMatch {
			return nodes
		} else {
			return []corev1.Node{}
		}
	}

	for _, node := range nodes {
		matches, err := helper.MatchNodeSelectorTerms(&node, &selectorTerm)
		if err != nil {
			log.Error().Msgf("Got an error running match node selector term on node %s. error: %v", node.Name, err)
			matches = true
		}

		if matches && keepNodesThatMatch {
			filtered = append(filtered, node)
		}
		if !matches && !keepNodesThatMatch {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func (mncc *ManagedNodesConfigController) getManagedNodesConfig(ctx context.Context) (managedNodesConfig *v1alpha1.ManagedNodesConfig, err error) {
	managedNodesConfig = &v1alpha1.ManagedNodesConfig{}
	err = mncc.Client.Get(ctx, client.ObjectKey{Name: config.Get().ManagedNodesConfigName}, managedNodesConfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info().Msgf("Managed nodes config <%s> not found, handling an empty config instead", config.Get().ManagedNodesConfigName)
			managedNodesConfig = &v1alpha1.ManagedNodesConfig{
				Spec: v1alpha1.ManagedNodesConfigSpec{
					InclusionCriteria: corev1.NodeSelector{},
				},
			}
		} else {
			log.Error().Msgf("Failed to get managed nodes config, err: %v", err.Error())
			return nil, err
		}
	}

	return managedNodesConfig, nil
}
