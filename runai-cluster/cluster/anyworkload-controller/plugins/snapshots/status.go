/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SnapshotEnabled says whether snapshots are on for the workload. Its reason
// tells apart what the Disabled phase covers: not opted in (Disabled), or an
// opt-in that can't be honoured (NotInstalled, Ineligible).
const (
	SnapshotEnabledConditionType = "SnapshotEnabled"

	ReasonEnabled      = "Enabled"
	ReasonDisabled     = "Disabled"
	ReasonNotInstalled = "NotInstalled"
	ReasonIneligible   = "Ineligible"
	ReasonCheckFailed  = "CheckFailed"

	msgEnabled      = "the workload can be captured and restored"
	msgDisabled     = "the workload does not opt in with the " + optInAnnotation + " annotation"
	msgNotInstalled = "Snapshot is not installed on the cluster"
)

// Status reports Ready for an opted-in workload that can use snapshots, and
// Disabled otherwise, with a warning issue when an opt-in can't be honoured.
// Neither moves the workload's phase: an ineligible workload just starts cold.
func (p *Plugin) Status(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ client.Client,
	status framework.PluginStatusSetter,
	state *reconcile.State,
) error {
	// Every path but an eligible workload returns without a phase, and the
	// framework starts a plugin at Unknown, which would hold a Running workload
	// in Initializing.
	status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)

	template, err := p.parser.Parse(runner.Spec.Template)
	if err != nil {
		return checkFailed(status, fmt.Errorf("parse template: %w", err))
	}
	// First, so a workload that never opted in gets no lookup and no issue.
	if !optedIn(template) {
		setEnabled(status, metav1.ConditionFalse, ReasonDisabled, msgDisabled)
		return nil
	}
	// Opting in has no effect until the admin installs Snapshot (REQ-02-2).
	if !p.snapshotInstalled {
		setEnabled(status, metav1.ConditionFalse, ReasonNotInstalled, msgNotInstalled)
		return nil
	}
	gvk := template.GroupVersionKind()
	karta, err := p.kartas.GetKartaForGVK(ctx, gvk)
	if err != nil {
		return checkFailed(status, fmt.Errorf("look up karta for %s: %w", gvk, err))
	}

	err = checkEligibility(ctx, template, karta)
	switch {
	case errors.Is(err, errIneligible):
		setEnabled(status, metav1.ConditionFalse, ReasonIneligible, err.Error())
		_ = state.AddPluginIssue(ctx, p, template, err, false, "", issue.Compute)
		return nil
	case err != nil:
		return checkFailed(status, err)
	}

	status.SetPhase(dromosv1alpha1.PluginPhaseReady)
	setEnabled(status, metav1.ConditionTrue, ReasonEnabled, msgEnabled)
	return nil
}

// setEnabled sets the SnapshotEnabled condition. Every path sets it: the
// framework keeps a plugin's conditions between reconciles.
func setEnabled(status framework.PluginStatusSetter, enabled metav1.ConditionStatus, reason, message string) {
	status.SetCondition(metav1.Condition{
		Type:    SnapshotEnabledConditionType,
		Status:  enabled,
		Reason:  reason,
		Message: message,
	})
}

// checkFailed marks the condition Unknown, so an earlier answer doesn't stay.
func checkFailed(status framework.PluginStatusSetter, err error) error {
	setEnabled(status, metav1.ConditionUnknown, ReasonCheckFailed, err.Error())
	return err
}
