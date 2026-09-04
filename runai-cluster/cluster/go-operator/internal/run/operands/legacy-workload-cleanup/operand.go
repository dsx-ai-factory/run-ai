/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacy_workload_cleanup

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const name = "legacy-workload-cleanup"

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

func (o *operand) Enabled() bool { return true }

func Operand(reader kc.HybridClient, config *runaiv1.RunaiConfig) operands.Operand {
	job := resources.NewJobResource(reader)
	return &operand{resources: resourcesForCondition(
		config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete), job,
	)}
}

func resourcesForCondition(condition *metav1.Condition, job operands.Resource) []operands.Resource {
	if condition == nil || condition.Status == metav1.ConditionTrue {
		return nil
	}
	switch runaiv1.ConditionReason(condition.Reason) {
	case runaiv1.LegacyWorkloadCleanupCreating,
		runaiv1.LegacyWorkloadCleanupRunning,
		runaiv1.LegacyWorkloadCleanupFailed:
		return []operands.Resource{job}
	default:
		return nil
	}
}
