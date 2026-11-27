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

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
// Failed, once a capture of a key in use failed, shows the workload Degraded: it
// serves, but its replicas start cold (§8.6).
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
	failed, err := recordKeys(ctx, c, runner, status)
	if err != nil {
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
	if failed != nil {
		status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
		_ = state.AddPluginIssue(ctx, p, template, captureFailure(failed), false, "", issue.Compute)
	}
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
	// Events are the capture events emitted for each key, so a restarted
	// controller neither loses nor repeats one.
	Events map[string]keyEvents `json:"events,omitempty"`
}

// keyEvents are the capture events emitted for one key.
type keyEvents struct {
	// SnapshotJob is the capture they report. One deleted by hand and created
	// again is a new capture, which reports afresh.
	SnapshotJob types.UID `json:"snapshotJob,omitempty"`
	Reasons     []string  `json:"reasons,omitempty"`
}

// errCaptureFailed is reported while a key in use has a failed capture.
var errCaptureFailed = errors.New(constants.SnapshotCaptureFailedMessage)

// recordKeys records the key of the runner's most recently created replica, and
// leaves it alone while it has none. It forgets the events of a key with neither
// a capture nor a place in the keep-set, so the record doesn't grow with every
// edit. It returns the failed capture of a key in use, if any.
func recordKeys(ctx context.Context, c client.Reader, runner *dromosv1alpha1.WorkloadRunner,
	status framework.PluginStatusSetter) (*snapshotv1alpha1.SnapshotJob, error) {
	replicas, err := keyedReplicas(ctx, c, runner)
	if err != nil {
		return nil, err
	}
	jobs, err := runnerSnapshotJobs(ctx, c, runner)
	if err != nil {
		return nil, err
	}
	// An unreadable record starts over from a replica's key, the one part that
	// can't be derived again. With no replica it stays as it is, so the deletion
	// keeps failing and the snapshot the next replica restores from survives.
	info, readErr := recordedStatus(runner)
	newest := newestKey(replicas)
	if readErr != nil && newest == "" {
		return nil, nil //nolint:nilerr // The deletion reports it.
	}
	changed := readErr != nil
	// With no replica the recorded key stays, so it survives a scale to zero.
	if newest != "" && newest != info.NewestKey {
		info.NewestKey = newest
		changed = true
	}
	keep := keysInUse(replicas, info.NewestKey)
	captured := datastructs.NewSet[string]()
	var failed *snapshotv1alpha1.SnapshotJob
	for i := range jobs {
		key := jobs[i].Labels[constants.SnapshotKeyLabel]
		captured.Add(key)
		// Deletion is dropping the rest, which would show Degraded for a moment.
		if failed == nil && keep.Has(key) && jobs[i].DeletionTimestamp == nil &&
			snapshotv1alpha1.IsSnapshotJobFailed(&jobs[i]) {
			failed = &jobs[i]
		}
	}
	// A key that comes back, as on a rollback, then reports afresh.
	for key := range info.Events {
		if !keep.Has(key) && !captured.Has(key) {
			delete(info.Events, key)
			changed = true
		}
	}
	// Only then, so the status isn't rewritten on every reconcile.
	if changed {
		return failed, writeStatusInfo(status, info)
	}
	return failed, nil
}

func newestKey(replicas []corev1.Pod) string {
	var newest *corev1.Pod
	for i := range replicas {
		if newest == nil || replicas[i].CreationTimestamp.After(newest.CreationTimestamp.Time) {
			newest = &replicas[i]
		}
	}
	if newest == nil {
		return ""
	}
	key, _ := snapshotkey.SnapshotKeyOf(newest)
	return key
}

// captureFailure is why a failed capture failed, from Snapshot's condition. If
// the condition is missing, it says only which capture failed.
func captureFailure(job *snapshotv1alpha1.SnapshotJob) error {
	failed := meta.FindStatusCondition(job.Status.Conditions, snapshotv1alpha1.SnapshotJobConditionFailed)
	if failed == nil {
		return fmt.Errorf("%w: SnapshotJob %s", errCaptureFailed, job.Name)
	}
	return fmt.Errorf("%w: SnapshotJob %s: %s: %s", errCaptureFailed, job.Name, failed.Reason, failed.Message)
}

// recordedStatus reads what Status and the capture recorded on earlier
// reconciles; the framework hands the plugin its own status back on every one.
func recordedStatus(runner *dromosv1alpha1.WorkloadRunner) (statusInfo, error) {
	var info statusInfo
	plugin := runner.Status.Plugins[PluginName]
	if plugin == nil || plugin.AdditionalInfo == nil {
		return info, nil
	}
	if err := json.Unmarshal(plugin.AdditionalInfo.Raw, &info); err != nil {
		return statusInfo{}, fmt.Errorf("read the plugin's recorded status: %w", err)
	}
	return info, nil
}

// pluginStatus is the plugin's status on runner, created as the framework would.
func pluginStatus(runner *dromosv1alpha1.WorkloadRunner) *dromosv1alpha1.PluginStatus {
	if runner.Status.Plugins == nil {
		runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{}
	}
	if runner.Status.Plugins[PluginName] == nil {
		runner.Status.Plugins[PluginName] = &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}
	}
	return runner.Status.Plugins[PluginName]
}

func writeStatusInfo(status framework.PluginStatusSetter, info statusInfo) error {
	raw, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("record the plugin's status: %w", err)
	}
	status.SetAdditionalInfo(&runtime.RawExtension{Raw: raw})
	return nil
}
