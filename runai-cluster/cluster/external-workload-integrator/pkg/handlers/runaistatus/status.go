/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package runaistatus computes the Run:ai phase + phaseMessage for v2 wlr-owned Jobs, matching v1 semantics.
package runaistatus

import (
	"encoding/json"
	"fmt"
	"sort"

	podphase "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/pod"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// podPhaseCalc is v1's pod-phase calculator, reused so EWI's pod classification can't drift from v1.
var podPhaseCalc = &podphase.RunaiPhaseCalculator{}

// minReadyReplicas mirrors workloadCommon.MinimumReadyReplicas (hardcoded 1 in v1).
const minReadyReplicas = 1

// Follow-up (https://github.com/dsx-ai-factory/run-ai/pull/22136) adds IsCritical
// issue filtering and makes this package its permanent home.

// ComputeStatus derives the Run:ai (phase, phaseMessage) from the Job + pods; wlr may be nil.
// Only batch Job workloads (Training/Workspace) for now; other wlr-backed kinds come later.
func ComputeStatus(job *batchv1.Job, pods []corev1.Pod,
	wlr *dromosv1alpha1.WorkloadRunner) (v1alpha1.ExternalWorkloadPhase, string) {
	if job == nil {
		return v1alpha1.ExWUnknownPhase, ""
	}

	// 1. Terminal / explicit states from the Job itself.
	if jobConditionTrue(job, batchv1.JobSuspended) {
		return v1alpha1.ExWStoppedPhase, ""
	}
	if jobConditionTrue(job, batchv1.JobComplete) {
		return v1alpha1.ExWCompletedPhase, ""
	}
	if jobConditionTrue(job, batchv1.JobFailed) {
		return v1alpha1.ExWFailedPhase, failedMessage(job, wlr)
	}

	// 2. Failure / degradation surfaced only by the WorkloadRunner.
	switch wlrPhase(wlr) {
	case dromosv1alpha1.WorkloadRunnerPhaseFailed:
		return v1alpha1.ExWFailedPhase, failedMessage(job, wlr)
	case dromosv1alpha1.WorkloadRunnerPhaseDegraded:
		return v1alpha1.ExWDegradedPhase, degradedMessage(wlr)
	}

	// 3. Lifecycle from the Job + its pods.
	if readyCount(job, pods) >= minReadyReplicas {
		return v1alpha1.ExWRunningPhase, ""
	}
	if hasActivePods(job, pods) {
		if anyPodInitializing(pods) {
			return v1alpha1.ExWInitializingPhase, initializingMessage(job)
		}
		return v1alpha1.ExWPendingPhase, ""
	}
	return v1alpha1.ExWCreatingPhase, ""
}

func initializingMessage(job *batchv1.Job) string {
	if isSinglePod(job) {
		return statuscontract.MsgInitializingSinglePod
	}
	return fmt.Sprintf(statuscontract.MsgInitializingMultiPods, minReadyReplicas)
}

// isSinglePod mirrors v1; parallelism<=1 also covers Workspaces (always single-pod).
func isSinglePod(job *batchv1.Job) bool {
	return parallelism(job) <= 1
}

func parallelism(job *batchv1.Job) int32 {
	if job.Spec.Parallelism == nil {
		return 1
	}
	return *job.Spec.Parallelism
}

func failedMessage(job *batchv1.Job, wlr *dromosv1alpha1.WorkloadRunner) string {
	// Like v1: when the Job itself failed, its own message wins over the wlr issue groups.
	if jobConditionTrue(job, batchv1.JobFailed) {
		if msg := jobConditionMessage(job, batchv1.JobFailed); msg != "" {
			return msg
		}
		return statuscontract.MsgGenericJobFailed
	}
	if cause := issueCause(wlr, statuscontract.MsgFailedDueToIssueSingular,
		statuscontract.MsgFailedDueToIssuesPlural); cause != "" {
		return cause
	}
	return statuscontract.MsgGenericJobFailed
}

func degradedMessage(wlr *dromosv1alpha1.WorkloadRunner) string {
	if cause := issueCause(wlr, statuscontract.MsgDegradedDueToIssueSingular,
		statuscontract.MsgDegradedDueToIssuesPlural); cause != "" {
		return cause
	}
	return statuscontract.MsgGenericDegraded
}

func issueCause(wlr *dromosv1alpha1.WorkloadRunner, singular, plural string) string {
	return statuscontract.FormatIssueCause(issueGroups(wlr), singular, plural)
}

// issueGroups returns the wlr's distinct issue group names, sorted, verbatim as v1 reports them.
func issueGroups(wlr *dromosv1alpha1.WorkloadRunner) []string {
	if wlr == nil {
		return nil
	}
	set := map[string]struct{}{}
	add := func(issues []dromosv1alpha1.Issue) {
		for _, is := range issues {
			// Only critical issues affect the phase message (v1 parity).
			if is.IsCritical && is.Group != "" {
				set[is.Group] = struct{}{}
			}
		}
	}
	add(wlr.Status.Issues)
	for _, plugin := range wlr.Status.Plugins {
		if plugin != nil {
			add(plugin.Issues)
		}
	}
	groups := make([]string, 0, len(set))
	for g := range set {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

func wlrPhase(wlr *dromosv1alpha1.WorkloadRunner) dromosv1alpha1.WorkloadRunnerPhase {
	if wlr == nil {
		return ""
	}
	return wlr.Status.Phase
}

func MnnvlRequired(wlr *dromosv1alpha1.WorkloadRunner) bool {
	return wlr.Spec.Instructions != nil &&
		wlr.Spec.Instructions.GpuOptimizations != nil &&
		wlr.Spec.Instructions.GpuOptimizations.Mnnvl != nil &&
		wlr.Spec.Instructions.GpuOptimizations.Mnnvl.Mode == dromosv1alpha1.MnnvlModeRequired
}

func SetAdditionalFields(externalWorkload *v1alpha1.ExternalWorkload, ownedKeys []string, updates map[string]any) {
	additional := make(map[string]any)
	if externalWorkload.Spec.AdditionalFields != nil {
		if err := json.Unmarshal([]byte(*externalWorkload.Spec.AdditionalFields), &additional); err != nil {
			log.Warn().Msgf("runaistatus: failed parsing existing additional fields: %s", err.Error())
			additional = make(map[string]any)
		}
	}
	if additional == nil {
		additional = make(map[string]any)
	}
	for _, key := range ownedKeys {
		delete(additional, key)
	}
	for key, value := range updates {
		switch typed := value.(type) {
		case string:
			if typed != "" {
				additional[key] = typed
			}
		case int32:
			if typed > 0 {
				additional[key] = typed
			}
		case int:
			if typed > 0 {
				additional[key] = typed
			}
		}
	}
	if len(additional) == 0 {
		externalWorkload.Spec.AdditionalFields = nil
		return
	}
	raw, err := json.Marshal(additional)
	if err != nil {
		log.Warn().Msgf("runaistatus: failed serializing additional fields: %s", err.Error())
		return
	}
	s := string(raw)
	externalWorkload.Spec.AdditionalFields = &s
}

func jobConditionTrue(job *batchv1.Job, condType batchv1.JobConditionType) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == condType {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func jobConditionMessage(job *batchv1.Job, condType batchv1.JobConditionType) string {
	for _, c := range job.Status.Conditions {
		if c.Type == condType {
			return c.Message
		}
	}
	return ""
}

func readyCount(job *batchv1.Job, pods []corev1.Pod) int32 {
	if job.Status.Ready != nil {
		return *job.Status.Ready
	}
	var ready int32
	for i := range pods {
		if podPhaseCalc.Calculate(&pods[i]).GetPhase() == podphase.RunaiPhaseReady {
			ready++
		}
	}
	return ready
}

func hasActivePods(job *batchv1.Job, pods []corev1.Pod) bool {
	return job.Status.Active > 0 || len(pods) > 0
}

// anyPodInitializing reports whether any pod is in v1's Initializing phase (scheduled and starting up).
func anyPodInitializing(pods []corev1.Pod) bool {
	for i := range pods {
		if podPhaseCalc.Calculate(&pods[i]).GetPhase() == podphase.RunaiPhaseInitializing {
			return true
		}
	}
	return false
}
