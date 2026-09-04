/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workloadPendingStatus     = "Pending"
	workloadSuspendedStatus   = "Stopped"
	workloadProgressingStatus = "Progressing"
)

var workloadBoundStatuses = map[string]bool{
	"Initializing": true,
	"Running":      true,
	"Ready":        true,
	"Degraded":     true,
	"Stopping":     true,
	"Terminating":  true}

func GetWorkloadId(workloadUid types.UID, annotations map[string]string) string {
	return topowner.GetRunaiWorkloadId(workloadUid, annotations)
}

func GetNodePoolOfWorkload(ctx context.Context,
	podsOfWorkload []corev1.Pod,
	readerClient client.Reader) (string, error) {
	if len(podsOfWorkload) == 0 {
		return "", nil
	}

	podDetailedPhase := GetPodDetailedPhase(&podsOfWorkload[0])
	return GetNodePoolOfPod(ctx, &podsOfWorkload[0], podDetailedPhase, readerClient)
}

func getWorkloadTypeFromPodGroups(ctx context.Context, podGroupsOfWorkload []*kaiv2alpha2.PodGroup,
	workloadKind, workloadNamespace, workloadName string,
	readerClient client.Reader) MetricsWorkloadType {
	if len(podGroupsOfWorkload) == 0 {
		log.Warn().Msgf("No pod groups found for workload %s <%s/%s>, will use default Train workload type",
			workloadKind, workloadNamespace, workloadName)
		return JobTypeTrain
	}
	return GetWorkloadTypeFromPodGroup(ctx, podGroupsOfWorkload[0], readerClient)
}

func GetWorkloadCategory(ctx context.Context, readerClient client.Reader, workloadKind string, workloadGroup string, labels map[string]string) (string, error) {
	workloadCategory := labels[constants.WorkloadCategoryLabel]

	if err := workloadTypeConfig.ValidateWorkloadCategory(ctx, readerClient, workloadCategory); err == nil {
		return workloadCategory, nil
	}
	cfg, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(ctx, readerClient, workloadGroup, workloadKind)
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("category label <%s> is not valid, replacing with <%s>", workloadCategory, cfg.CategoryName)
	return cfg.CategoryName, nil
}

// GetWorkloadTypeMetricLabelValueBackwards -
// returns the metric workload type label value of the given workload - for backwards compatibility.
// can be only one of: MetricsWorkloadType - Train, Interactive, Inference, and Interactive-Preemptible.
// does that either from the workload's kind - if is a runai workload;
// or using the PriorityClassName of the external workload;
// or by listing the pod groups of the workload and getting the workload type from the first pod group.
// podsOfWorkload - if previously fetched, should be the pods of the workload, otherwise - nil.
func GetWorkloadTypeMetricLabelValueBackwards(ctx context.Context,
	workloadKind, workloadId, workloadNamespace, workloadName string,
	isPreemptible *bool, priorityClassName *string,
	podsOfWorkload *[]corev1.Pod, readerClient client.Reader) MetricsWorkloadType {
	workloadTypeForRunaiWorkloads := getWorkloadTypeForRunaiWorkloads(workloadKind, isPreemptible)
	if workloadTypeForRunaiWorkloads != "" {
		return workloadTypeForRunaiWorkloads
	}

	if priorityClassName != nil && *priorityClassName != "" {
		return GetWorkloadTypeFromPriorityClassName(ctx, *priorityClassName, workloadNamespace, workloadName, readerClient)
	}

	podGroups, err := utils.GetPodGroupsOfWorkload(ctx,
		workloadId, workloadNamespace,
		podsOfWorkload, readerClient)
	if err != nil {
		log.Error().Msgf("Failed listing pod groups of workload %s <%s/%s> (id <%s>), err: %s ; will use default Train workload type",
			workloadKind, workloadNamespace, workloadName, workloadId, err.Error())
		return JobTypeTrain
	}

	return getWorkloadTypeFromPodGroups(ctx, podGroups,
		workloadKind, workloadNamespace, workloadName,
		readerClient)
}

func GetUserOfWorkload(workloadFields *field_groups.WorkloadFields) string {
	if workloadFields == nil || workloadFields.UserName == nil {
		return ""
	}
	return workloadFields.UserName.Deref()
}

// GetPhaseFromWorkloadStatus
// the 'status' param should be the workload's status.
// the returned value - phase label - is one of - Pending, Suspended, Running, "" .
func GetPhaseFromWorkloadStatus(status string) PhaseLabel {
	phase := PhaseLabel("")
	if status == workloadPendingStatus || status == workloadProgressingStatus {
		phase = pendingPhase
	} else if status == workloadSuspendedStatus {
		phase = suspendedPhase
	} else if isWorkloadBound(status) {
		phase = runningPhase
	}

	return phase
}

func isWorkloadBound(detailedStatus string) bool {
	if _, found := workloadBoundStatuses[detailedStatus]; found {
		return true
	}
	return false
}

func getWorkloadTypeForRunaiWorkloads(workloadKind string, isPreemptible *bool) MetricsWorkloadType {
	switch workloadKind {
	case commonconstants.KindTrainingWorkload, commonconstants.KindDistributedWorkload:
		return JobTypeTrain
	case commonconstants.KindInteractiveWorkload:
		if isPreemptible != nil && *isPreemptible {
			return JobTypeInteractivePreemptible
		}
		return JobTypeInteractive
	case commonconstants.KindInferenceWorkload:
		return JobTypeInference
	case commonconstants.KindDistributedInferenceWorkload:
		return JobTypeDistributedInference
	}

	return ""
}
