/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
//
// For an opted-in workload it also records the newest key, which the deletion
// keeps while no replica runs.
func (p *Plugin) Status(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	c client.Client,
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
	if err := recordNewestKey(ctx, c, runner, status); err != nil {
		_ = state.AddPluginIssue(ctx, p, template, err, false, "", issue.Compute)
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

// statusInfo is the plugin's structured status.
type statusInfo struct {
	// NewestKey is the most recently created replica's key, kept while no replica runs.
	NewestKey string `json:"newestKey,omitempty"`
}

// recordNewestKey records the key of the runner's most recently created replica,
// and leaves the record alone while it has none.
func recordNewestKey(ctx context.Context, c client.Reader, runner *dromosv1alpha1.WorkloadRunner,
	status framework.PluginStatusSetter) error {
	replicas, err := keyedReplicas(ctx, c, runner)
	if err != nil {
		return err
	}
	var newest *corev1.Pod
	for i := range replicas {
		if newest == nil || replicas[i].CreationTimestamp.After(newest.CreationTimestamp.Time) {
			newest = &replicas[i]
		}
	}
	if newest == nil {
		return nil
	}
	key, _ := snapshotkey.SnapshotKeyOf(newest)
	raw, err := json.Marshal(statusInfo{NewestKey: key})
	if err != nil {
		return fmt.Errorf("record the newest key: %w", err)
	}
	status.SetAdditionalInfo(&runtime.RawExtension{Raw: raw})
	return nil
}

// recordedNewestKey reads the key Status recorded on an earlier reconcile; the
// framework hands the plugin its own status back on every one.
func recordedNewestKey(runner *dromosv1alpha1.WorkloadRunner) (string, error) {
	plugin := runner.Status.Plugins[PluginName]
	if plugin == nil || plugin.AdditionalInfo == nil {
		return "", nil
	}
	var info statusInfo
	if err := json.Unmarshal(plugin.AdditionalInfo.Raw, &info); err != nil {
		return "", fmt.Errorf("read the recorded newest key: %w", err)
	}
	return info.NewestKey, nil
}
