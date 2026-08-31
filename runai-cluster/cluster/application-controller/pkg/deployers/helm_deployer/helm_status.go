/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type AppStatus struct {
	operation *v1alpha1.ActionStatus // the status of an action that is running (install, upgrade or uninatall)
	status    *v1alpha1.HelmStatus   // the status from the helm itself
}

const (
	ActionQueued      = "queued"
	ActionStarted     = "started"
	ActionRunning     = "running"
	ActionDownloading = "downloading"
	ActionFailed      = "failed"
	ActionCompleted   = "completed"
)

// PhaseValue maps statuses to the order we are expecting them to appear.
// the purpose is that new status can only replace status with less/equal significance
var PhaseValue = map[string]int{
	// both completion statuses have the same significance
	ActionCompleted: 5,
	ActionFailed:    5,

	ActionRunning:     4,
	ActionDownloading: 3,
	ActionStarted:     2,
	ActionQueued:      1,
}

func NewAppStatus() *AppStatus {
	return &AppStatus{
		// as long as we do not have another information, we assume we are preparing for action
		status:    &v1alpha1.HelmStatus{},
		operation: &v1alpha1.ActionStatus{},
	}
}

func (as *AppStatus) WithReleaseInfo(releaseInfo *release.Release) *AppStatus {
	if releaseInfo != nil {
		as.status.ReleaseVersion = ptr.To(releaseInfo.Version)
		as.status.Phase = string(releaseInfo.Info.Status)
	}
	return as
}

func (as *AppStatus) WithDeployedChart(chartUrl string) *AppStatus {
	if chartUrl != "" {
		as.status.ChartUrl = chartUrl
	}
	return as
}

func (as *AppStatus) TaskQueued(task Task) *AppStatus {
	log.Info().Msgf("HelmTask --> %s task queued", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionQueued
	as.operation.Message = fmt.Sprintf("%s task is waiting for execution", task.Action())

	return as
}

func (as *AppStatus) TaskStarted(task Task) *AppStatus {
	log.Info().Msgf("HelmTask --> %s task started", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionStarted
	as.operation.StartedAt = metav1.Now()
	as.operation.Message = fmt.Sprintf("%s task started", task.Action())

	return as
}

func (as *AppStatus) TaskDownloading(task Task) *AppStatus {
	log.Info().Msgf("HelmTask --> %s task is downloading", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionDownloading
	as.operation.Message = "download in progress"

	return as
}

func (as *AppStatus) TaskRunning(task Task) *AppStatus {
	log.Info().Msgf("HelmTask --> %s task is in progress", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionRunning
	as.operation.Message = fmt.Sprintf("%s task is running", task.Action())

	return as
}

func (as *AppStatus) TaskFailed(task Task, err error) *AppStatus {
	log.Info().Err(err).Msgf("HelmTask --> %s task failed", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionFailed
	as.operation.CompletedAt = metav1.Now()
	as.operation.Message = fmt.Sprintf("%s failed", task.Action())
	as.operation.Error = err.Error()

	return as
}

func (as *AppStatus) TaskCompleted(task Task) *AppStatus {
	log.Info().Msgf("HelmTask --> %s task completed", task.Action())

	as.operation.Type = string(task.Action())
	as.operation.Phase = ActionCompleted
	as.operation.CompletedAt = metav1.Now()
	as.operation.Message = fmt.Sprintf("%s completed", task.Action())

	return as
}

// RecoveryCompleted marks the action as completed without a live Task object.
// Used by the reconciler to recover applications whose ActionStatus was left in a
// non-terminal phase because the worker's final status write was lost to K8s conflicts.
func (as *AppStatus) RecoveryCompleted() *AppStatus {
	log.Info().Msg("HelmTask --> recovering stuck action status to completed")

	as.operation.Phase = ActionCompleted
	as.operation.CompletedAt = metav1.Now()
	as.operation.Message = "completed"

	return as
}

func (as *AppStatus) WithError(err error) *AppStatus {
	if err != nil {
		as.status.Error = err.Error()
	}
	return as
}

// ApplyOnApplication applies everything we know -> to the status of the application resource
func (as *AppStatus) ApplyOnApplication(app *v1alpha1.Application) bool {
	var changed bool

	if as.operation != nil {
		if app.Status.ActionStatus == nil {
			app.Status.ActionStatus = &v1alpha1.ActionStatus{}
		}

		if as.operation.Type != "" && as.operation.Type != app.Status.ActionStatus.Type {
			app.Status.ActionStatus.Type = as.operation.Type
			changed = true
		}

		if as.operation.Phase != "" && as.operation.Phase != app.Status.ActionStatus.Phase {
			if PhaseValue[as.operation.Phase] >= PhaseValue[app.Status.ActionStatus.Phase] {
				app.Status.ActionStatus.Phase = as.operation.Phase
				changed = true
			}
		}

		if !as.operation.StartedAt.IsZero() && !as.operation.StartedAt.Equal(&app.Status.ActionStatus.StartedAt) {
			app.Status.ActionStatus.StartedAt = as.operation.StartedAt
			changed = true
		}

		if !as.operation.CompletedAt.IsZero() && !as.operation.CompletedAt.Equal(&app.Status.ActionStatus.CompletedAt) {
			app.Status.ActionStatus.CompletedAt = as.operation.CompletedAt
			changed = true
		}

		if as.operation.Message != "" && as.operation.Message != app.Status.ActionStatus.Message {
			app.Status.ActionStatus.Message = as.operation.Message
			changed = true
		}

		// make sure to clean error in case the process has completed successfully
		if as.operation.Phase == ActionCompleted && app.Status.ActionStatus.Error != "" {
			app.Status.ActionStatus.Error = ""
			changed = true
		}

		if as.operation.Error != "" && as.operation.Error != app.Status.ActionStatus.Error {
			app.Status.ActionStatus.Error = as.operation.Error
			changed = true
		}
	}

	if as.status != nil {
		if app.Status.HelmStatus == nil {
			app.Status.HelmStatus = &v1alpha1.HelmStatus{}
		}

		if as.status.Error != "" && app.Status.HelmStatus.Error != as.status.Error {
			app.Status.HelmStatus.Error = as.status.Error
			changed = true
		}

		if as.status.Phase != "" && app.Status.HelmStatus.Phase != as.status.Phase {
			app.Status.HelmStatus.Phase = as.status.Phase
			changed = true
		}

		if ptr.Deref(app.Status.HelmStatus.ReleaseVersion, 0) != ptr.Deref(as.status.ReleaseVersion, 0) {
			app.Status.HelmStatus.ReleaseVersion = ptr.To(ptr.Deref(as.status.ReleaseVersion, 0))
		}

		if as.status.ChartUrl != "" && as.status.ChartUrl != app.Status.HelmStatus.ChartUrl {
			app.Status.HelmStatus.ChartUrl = as.status.ChartUrl
			changed = true
		}

		// Keep k8sPhase aligned with the Helm release status on controller-driven installs.
		// status.action / helmStatus are updated here, but k8sPhase was only set via the secret
		// informer path; concurrent status updates can drop that write after conflicts.
		if as.status.Phase != "" && app.Status.K8sPhase != as.status.Phase {
			app.Status.K8sPhase = as.status.Phase
			changed = true
		}
	}

	return changed
}

// IsHelmWorkingOnThisRelease returns true when a Helm operation is already in progress on the release.
// Used to detect in-flight operations after a controller restart, when runningTasks is empty.
func IsHelmWorkingOnThisRelease(rel *release.Info) bool {
	if rel == nil {
		return false
	}

	switch rel.Status {
	case release.StatusPendingInstall,
		release.StatusPendingUpgrade,
		release.StatusPendingRollback:
		return true
	}

	// when uninstalling, helm is working on it only if it is not deleted
	if rel.Status == release.StatusUninstalling {
		return rel.Deleted.IsZero()
	}

	return false
}
