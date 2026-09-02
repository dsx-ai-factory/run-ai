/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"strings"

	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	distcalc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/distributed"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var distributedCountKeys = []string{
	statuscontract.AdditionalFieldReadyPods,
	statuscontract.AdditionalFieldRunningPods,
	statuscontract.AdditionalFieldPendingPods,
	statuscontract.AdditionalFieldInitializingPods,
	statuscontract.AdditionalFieldSucceededPods,
	statuscontract.AdditionalFieldFailedPods,
}

const (
	computeIssueGroup = "compute"
	mnnvlPluginName   = "mnnvl"
)

func (h *KubeflowJobTopOwnerHandler) applyRunaiDistributedStatus(
	ctx context.Context,
	object client.Object,
	fields *KubeflowJobFieldsForParsing,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	delete(externalWorkload.Annotations, commonconstants.EwiStatusAnnotation)
	runaistatus.SetAdditionalFields(externalWorkload, distributedCountKeys, nil)

	wlrName := externalWorkload.Labels[commonconstants.WorkloadRunnerNameLabel]
	if wlrName == "" {
		return
	}

	ns := object.GetNamespace()
	workloadID := topowner.GetRunaiWorkloadId(object.GetUID(), object.GetAnnotations())
	pods, err := getters.GetPodsOfWorkload(ctx, workloadID, ns, h.cachedClient)
	if err != nil {
		log.Warn().Msgf("kubeflow runaistatus: failed fetching pods for <%s/%s>: %s", ns, object.GetName(), err)
		return
	}

	wlr := h.getWorkloadRunner(ctx, wlrName, ns)
	if wlr == nil {
		return
	}

	jobStatus := enrichJobStatusIfNeeded(object, fields)
	totalReplicas := totalReplicaCount(fields.replicaSpecs)

	in := distcalc.Input{
		JobStatus:           jobStatus,
		Pods:                pods,
		Suspend:             isSuspended(fields),
		Terminating:         !object.GetDeletionTimestamp().IsZero(),
		CriticalIssues:      extractCriticalIssues(wlr),
		PriorCompletionTime: nil,
		MultiReplica:        totalReplicas > 1,
	}

	result := distcalc.Calculate(in)

	phase, message := legacyToExWPhase(result.Phase, result.Message)
	externalWorkload.Status.ObservedStatus.Phase = phase
	externalWorkload.Status.ObservedStatus.Message = message

	applyRuntimeIssueCondition(externalWorkload, result.Phase == statuscontract.Failed, pods, wlr)
	applyMNNVLConditions(externalWorkload, wlr)
	mergeDistributedCounts(externalWorkload, result)

	if externalWorkload.Annotations == nil {
		externalWorkload.Annotations = map[string]string{}
	}
	externalWorkload.Annotations[commonconstants.EwiStatusAnnotation] = "true"
}

func (h *KubeflowJobTopOwnerHandler) getWorkloadRunner(ctx context.Context, name, namespace string) *dromosv1alpha1.WorkloadRunner {
	wlr := &dromosv1alpha1.WorkloadRunner{}
	wlr.SetGroupVersionKind(dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind))
	if err := h.cachedClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, wlr); err != nil {
		log.Warn().Msgf("kubeflow runaistatus: failed getting WorkloadRunner <%s/%s>: %s", namespace, name, err)
		return nil
	}
	return wlr
}

func applyRuntimeIssueCondition(
	externalWorkload *v1alpha1.ExternalWorkload,
	failed bool,
	pods []corev1.Pod,
	wlr *dromosv1alpha1.WorkloadRunner,
) {
	group, desc, ok := runaistatus.ComputeRuntimeIssueFromPods(failed, pods, wlr)
	if !ok {
		return
	}
	if desc != "" {
		externalWorkload.Status.ObservedStatus.Message = desc
	} else {
		desc = statuscontract.MsgGenericJobFailed
	}
	externalWorkload.Status.ObservedStatus.Conditions = append(externalWorkload.Status.ObservedStatus.Conditions,
		v1alpha1.ObservedCondition{
			Type:    commonconstants.WorkloadIssueConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  group,
			Message: desc,
		})
}

func applyMNNVLConditions(externalWorkload *v1alpha1.ExternalWorkload, wlr *dromosv1alpha1.WorkloadRunner) {
	if !runaistatus.MnnvlRequired(wlr) {
		return
	}

	plugin := wlr.Status.Plugins[mnnvlPluginName]
	if plugin != nil && plugin.Phase == dromosv1alpha1.PluginPhaseReady {
		upsertObservedCondition(&externalWorkload.Status.ObservedStatus.Conditions, v1alpha1.ObservedCondition{
			Type:   statuscontract.ConditionTypeGPUNetworkAccelerated,
			Status: corev1.ConditionTrue,
		})
		return
	}

	upsertObservedCondition(&externalWorkload.Status.ObservedStatus.Conditions, v1alpha1.ObservedCondition{
		Type:    statuscontract.ConditionTypeReady,
		Status:  corev1.ConditionFalse,
		Reason:  statuscontract.ReasonMnnvlNotReady,
		Message: statuscontract.MsgMnnvlNotReady,
	})
	upsertObservedCondition(&externalWorkload.Status.ObservedStatus.Conditions, v1alpha1.ObservedCondition{
		Type:    statuscontract.ConditionTypeGPUNetworkAccelerated,
		Status:  corev1.ConditionFalse,
		Reason:  statuscontract.ReasonMnnvlNotReady,
		Message: statuscontract.MsgMnnvlNotReady,
	})
}

func upsertObservedCondition(conditions *[]v1alpha1.ObservedCondition, condition v1alpha1.ObservedCondition) {
	for i := range *conditions {
		if (*conditions)[i].Type == condition.Type {
			(*conditions)[i] = condition
			return
		}
	}
	*conditions = append(*conditions, condition)
}

func enrichJobStatusIfNeeded(obj client.Object, fields *KubeflowJobFieldsForParsing) *kubeflowv1.JobStatus {
	mpiV2, ok := obj.(*kubeflowv2beta1.MPIJob)
	if !ok {
		return fields.jobStatus
	}
	return &kubeflowv1.JobStatus{
		Conditions:      fields.jobStatus.Conditions,
		StartTime:       mpiV2.Status.StartTime,
		CompletionTime:  mpiV2.Status.CompletionTime,
		ReplicaStatuses: convertMPIReplicaStatuses(mpiV2.Status.ReplicaStatuses),
	}
}

func convertMPIReplicaStatuses(
	in map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaStatus,
) map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaStatus {
	out := make(map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaStatus, len(in))
	for rt, rs := range in {
		if rs == nil {
			continue
		}
		out[kubeflowv1.ReplicaType(rt)] = &kubeflowv1.ReplicaStatus{
			Active:    rs.Active,
			Succeeded: rs.Succeeded,
			Failed:    rs.Failed,
		}
	}
	return out
}

func isSuspended(fields *KubeflowJobFieldsForParsing) bool {
	if fields.runPolicy != nil && fields.runPolicy.Suspend != nil {
		return *fields.runPolicy.Suspend
	}
	if fields.runPolicyV2beta1 != nil && fields.runPolicyV2beta1.Suspend != nil {
		return *fields.runPolicyV2beta1.Suspend
	}
	return false
}

func totalReplicaCount(specs map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec) int32 {
	var total int32
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if spec.Replicas != nil {
			total += *spec.Replicas
		} else {
			total++ // Kubeflow Replicas default is 1
		}
	}
	return total
}

func extractCriticalIssues(wlr *dromosv1alpha1.WorkloadRunner) []distcalc.CriticalIssue {
	out := appendCriticalIssues(nil, wlr.Status.Issues)
	for _, plugin := range wlr.Status.Plugins {
		if plugin != nil {
			out = appendCriticalIssues(out, plugin.Issues)
		}
	}

	return out
}

func appendCriticalIssues(out []distcalc.CriticalIssue, issues []dromosv1alpha1.Issue) []distcalc.CriticalIssue {
	for _, iss := range issues {
		if !iss.IsCritical || iss.Group == "" {
			continue
		}
		out = append(out, distcalc.CriticalIssue{
			Group:   iss.Group,
			Compute: strings.EqualFold(iss.Group, computeIssueGroup),
		})
	}
	return out
}

func legacyToExWPhase(legacyPhase statuscontract.Phase, msg string) (v1alpha1.ExternalWorkloadPhase, string) {
	switch legacyPhase {
	case statuscontract.Completed:
		return v1alpha1.ExWCompletedPhase, msg
	case statuscontract.Failed:
		return v1alpha1.ExWFailedPhase, msg
	case statuscontract.Running:
		return v1alpha1.ExWRunningPhase, msg
	case statuscontract.Degraded:
		return v1alpha1.ExWDegradedPhase, msg
	case statuscontract.Initializing:
		return v1alpha1.ExWInitializingPhase, msg
	case statuscontract.Pending, statuscontract.Resuming:
		return v1alpha1.ExWPendingPhase, msg
	case statuscontract.Suspending:
		return v1alpha1.ExWStoppingPhase, msg
	case statuscontract.Suspended:
		return v1alpha1.ExWStoppedPhase, msg
	case statuscontract.Terminating:
		return v1alpha1.ExWDeletingPhase, msg
	default:
		return v1alpha1.ExWCreatingPhase, msg
	}
}

func mergeDistributedCounts(ew *v1alpha1.ExternalWorkload, r distcalc.Result) {
	runaistatus.SetAdditionalFields(ew, distributedCountKeys, map[string]any{
		statuscontract.AdditionalFieldReadyPods:        r.ReadyPods,
		statuscontract.AdditionalFieldRunningPods:      r.RunningPods,
		statuscontract.AdditionalFieldPendingPods:      r.PendingPods,
		statuscontract.AdditionalFieldInitializingPods: r.InitializingPods,
		statuscontract.AdditionalFieldSucceededPods:    r.SucceededPods,
		statuscontract.AdditionalFieldFailedPods:       r.FailedPods,
	})
}
