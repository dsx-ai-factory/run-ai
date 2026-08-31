/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeaffinity

import (
	"context"
	"encoding/json"
	"fmt"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	affinityutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/affinity"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// NodeAffinityAppliedConditionType reports whether the instruction reached the child.
	NodeAffinityAppliedConditionType = "NodeAffinityApplied"

	ReasonDisabled   = "Disabled"
	ReasonApplied    = "Applied"
	ReasonNotApplied = "NotApplied"
	ReasonFailed     = "Failed"

	msgNoInstruction = "no node affinity instruction is set"
	msgApplied       = "node affinity merged into the workload's pod affinity"
	msgNotApplied    = "reconciliation did not apply the workload this round, so the node affinity has not reached its pods"
)

// Status publishes the applied affinity, so the constraint a workload runs under
// is visible without reading the child.
func (p *Plugin) Status(
	_ context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ client.Client,
	status framework.PluginStatusSetter,
	state *reconcile.State,
) error {
	desired := extractNodeAffinity(runner)
	if desired == nil {
		status.SetCondition(metav1.Condition{
			Type:    NodeAffinityAppliedConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonDisabled,
			Message: msgNoInstruction,
		})
		status.SetAdditionalInfo(nil)
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	for _, iss := range state.PluginIssues(p) {
		if iss.IsCritical {
			status.SetCondition(metav1.Condition{
				Type:    NodeAffinityAppliedConditionType,
				Status:  metav1.ConditionFalse,
				Reason:  ReasonFailed,
				Message: iss.Error(),
			})
			status.SetAdditionalInfo(nil)
			status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
			return nil
		}
	}

	if state.HasCriticalControllerIssues() {
		status.SetCondition(metav1.Condition{
			Type:    NodeAffinityAppliedConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonNotApplied,
			Message: msgNotApplied,
		})
		status.SetAdditionalInfo(nil)
		status.SetPhase(dromosv1alpha1.PluginPhasePending)
		return nil
	}

	applied := desired.DeepCopy()
	affinityutils.NormalizeNodeAffinity(applied)
	raw, err := json.Marshal(&dromosv1alpha1.NodeAffinityConfig{NodeAffinity: applied})
	if err != nil {
		return fmt.Errorf("marshal applied node affinity: %w", err)
	}

	status.SetAdditionalInfo(&runtime.RawExtension{Raw: raw})
	status.SetCondition(metav1.Condition{
		Type:    NodeAffinityAppliedConditionType,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonApplied,
		Message: msgApplied,
	})
	status.SetPhase(dromosv1alpha1.PluginPhaseReady)

	return nil
}
