/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mncName = constants.ManagedNodesConfigName

	appliedConditionType    = string(kaiv1alpha1.MNCConditionTypeApplied)
	toBeExcludedNodesReason = string(kaiv1alpha1.MNCConditionReasonToBeExcludedNodes)
)

func ManagedNodesConfigForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {

	return &kaiv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: mncName,
		},
		Spec: kaiv1alpha1.ManagedNodesConfigSpec{
			InclusionCriteria: *runaiConfig.Spec.Global.ManagedNodes.InclusionCriteria,
		},
	}, nil
}

func ManagedNodesConfigStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {

	conditions, err := managedNodesConfigConditions(ctx, k8sReader)
	if err != nil {
		if errors.IsNotFound(err) {
			return operands.NotReadyStatus("ManagedNodesConfig is missing"), nil
		}
		return operands.NotReadyStatus(), err
	}

	appliedCondition := utils.GetCondition(conditions, appliedConditionType)
	if appliedCondition == nil {
		return operands.NotReadyStatus("ManagedNodesConfig has no applied condition"), nil
	}

	if appliedCondition.Status == metav1.ConditionFalse && appliedCondition.Reason != toBeExcludedNodesReason {
		return operands.NotReadyStatus(fmt.Sprintf("ManagedNodesConfig is not applied: %s", appliedCondition.Message)), nil
	}

	return operands.ReadyStatus(), nil
}

func managedNodesConfigConditions(ctx context.Context, k8sReader client.Reader) ([]metav1.Condition, error) {
	mnc := &kaiv1alpha1.ManagedNodesConfig{}
	if err := k8sReader.Get(ctx, types.NamespacedName{Name: mncName}, mnc); err != nil {
		return nil, err
	}
	return mnc.Status.Conditions, nil
}
