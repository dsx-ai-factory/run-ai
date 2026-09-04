/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managed_nodes_config

import (
	"context"
	"fmt"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (mncc *ManagedNodesConfigController) reconcileStatus(ctx context.Context, req MNCReconcileRequest, mnc *v1alpha1.ManagedNodesConfig, toBeExcludedNodes []v1.Node) error {
	changed, err := mncc.calculateStatus(ctx, req, mnc, toBeExcludedNodes)
	if err != nil {
		log.Error().Msgf("Failed to calculate ManagedNodesConfig status. error %v", err)
		return err
	}

	if changed {
		return mncc.updateStatus(ctx, mnc)
	}

	return nil
}

func (mncc *ManagedNodesConfigController) calculateStatus(ctx context.Context, req MNCReconcileRequest, mnc *v1alpha1.ManagedNodesConfig, nodes []v1.Node) (changed bool, err error) {
	if req.isNode {
		toBeExcludedRequirment, err := labels.NewRequirement(config.Get().ShouldBeExcludedLabelKey, selection.Equals, []string{"true"})
		if err != nil {
			return false, err
		}

		nodesList, err := mncc.NodePoolController.ListNodesWithRequirements(ctx, *toBeExcludedRequirment)
		if err != nil {
			return false, err
		}

		nodes = nodesList.Items
	}

	if len(nodes) == 0 {
		changed = mncc.setCondition(mnc, metav1.Condition{
			Type:               string(v1alpha1.MNCConditionTypeApplied),
			Status:             metav1.ConditionTrue,
			Reason:             string(v1alpha1.MNCConditionReasonAllNodesIncludedCorrectly),
			LastTransitionTime: metav1.Now(),
		})
	} else {
		nodeNames := ""
		for i, node := range nodes {
			if i == 5 {
				nodeNames += fmt.Sprintf(" and %d more", len(nodes)-5)
				break
			}

			nodeNames += fmt.Sprintf(",%s", node.Name)
		}
		nodeNames = nodeNames[1:]

		changed = mncc.setCondition(mnc, metav1.Condition{
			Type:               string(v1alpha1.MNCConditionTypeApplied),
			Status:             metav1.ConditionFalse,
			Reason:             string(v1alpha1.MNCConditionReasonToBeExcludedNodes),
			Message:            fmt.Sprintf("Some nodes need to be drained before gracefull exclusion: %s", nodeNames),
			LastTransitionTime: metav1.Now(),
		})
	}

	if mnc.Status.ObservedGeneration != mnc.Generation {
		mnc.Status.ObservedGeneration = mnc.Generation
		changed = true
	}

	return changed, nil
}

func (mncc *ManagedNodesConfigController) setCondition(mnc *v1alpha1.ManagedNodesConfig, condition metav1.Condition) bool {
	for i := range mnc.Status.Conditions {
		existCondition := mnc.Status.Conditions[i]
		if existCondition.Type == condition.Type {
			mnc.Status.Conditions[i] = condition

			return existCondition.Reason != condition.Reason ||
				existCondition.Status != condition.Status
		}
	}

	mnc.Status.Conditions = append(mnc.Status.Conditions, condition)

	return true
}

func (mncc *ManagedNodesConfigController) updateStatus(ctx context.Context, mnc *v1alpha1.ManagedNodesConfig) error {
	err := mncc.Client.Status().Update(ctx, mnc)
	if err != nil {
		log.Error().Msgf("Error updating status of ManagedNodesConfig. error %v", err)
		return err
	}

	log.Info().Msgf("Updated status of ManagedNodesConfig")
	return nil
}
