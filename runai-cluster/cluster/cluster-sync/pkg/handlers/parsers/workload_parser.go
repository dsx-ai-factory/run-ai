/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kubeaischedv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AnnotationWorkloadId         = "workloadId"
	lastStartTimestampAnnotation = "kai.scheduler/last-start-timestamp"
)

func WorkloadToStatusChangeRequest(workloadObjectMeta *v1.ObjectMeta, workloadStatus jobs.JobStatus) *swagger.WorkloadStatus {
	workloadId := workloadObjectMeta.Annotations[AnnotationWorkloadId]
	if workloadId == "" {
		return nil
	}
	if workloadStatus.CreationStatus == "" {
		// do not notify backend before the workload receives status.
		return nil
	}
	return &swagger.WorkloadStatus{
		Id:     workloadId,
		Status: workloadStatus.CreationStatus,
	}
}

// WorkloadDistributedToStatusChangeRequest converts a distributed workload status to a backend request.
// TODO: check this later.
func WorkloadDistributedToStatusChangeRequest(workloadObjectMeta *v1.ObjectMeta, workloadStatus diwv2.DistributedWorkloadStatus) *swagger.WorkloadStatus {
	workloadId := workloadObjectMeta.Annotations[AnnotationWorkloadId]
	if workloadId == "" {
		return nil
	}
	if workloadStatus.CreationStatus == "" {
		// do not notify backend before the workload receives status.
		return nil
	}
	return &swagger.WorkloadStatus{
		Id:     workloadId,
		Status: workloadStatus.CreationStatus,
	}
}

// getNodepoolMinGuaranteedRuntimeOfPod returns the minimum guaranteed runtime configured on
// the node pool the pod is assigned to.
func (p *WorkloadParser) getNodepoolMinGuaranteedRuntimeOfPod(pod corev1.Pod) (string, error) {
	nodepoolName := getPodCurrentNodePool(pod.Labels, pod.Status.Phase)
	if nodepoolName == "" {
		return "", errors.New(fmt.Sprintf("failed to get nodepool for pod in phase %s", pod.Status.Phase))
	}
	key := client.ObjectKey{
		Name: nodepoolName,
	}

	nodePool := &kaiv1alpha1.NodePool{}
	if err := p.cachedReader.Get(context.Background(), key, nodePool); err != nil {
		zap.S().Warnf(err.Error())
		return "", err
	}
	shardConfig := nodePool.Spec.SchedulingShardConfig
	if shardConfig == nil || shardConfig.MinRuntime == nil || shardConfig.MinRuntime.PreemptMinRuntime == nil {
		return "", nil
	}
	return *shardConfig.MinRuntime.PreemptMinRuntime, nil
}

func (p *WorkloadParser) getGuaranteedRuntimeTimestamp(podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup, workloadPods *[]corev1.Pod) *time.Time {
	if len(podGroupsOfWorkload) != 1 || workloadPods == nil || len(*workloadPods) == 0 {
		return nil
	}
	aWorkloadPodGroup := podGroupsOfWorkload[0]
	aWorkloadPod := (*workloadPods)[0]

	lastStartTimestampStr, ok := aWorkloadPodGroup.Annotations[lastStartTimestampAnnotation]
	if !ok {
		return nil
	}
	lastStartTimestamp, err := time.Parse(time.RFC3339, lastStartTimestampStr)
	if err != nil {
		zap.S().Debugf("failed to parse %s = '%s', error: %s", lastStartTimestampAnnotation, lastStartTimestampStr, err.Error())
		return nil
	}

	minGuaranteedRuntimeStr, err := p.getNodepoolMinGuaranteedRuntimeOfPod(aWorkloadPod)
	if err != nil {
		return nil
	}

	if minGuaranteedRuntimeStr == "" {
		return nil
	}
	minGuaranteedRuntimeDuration, err := time.ParseDuration(minGuaranteedRuntimeStr)
	if err != nil {
		zap.S().Warnf("failed to parse MinGuaranteedRuntime '%s' for pod <%s/%s>, error: %s",
			minGuaranteedRuntimeStr, aWorkloadPod.Namespace, aWorkloadPod.Name, err.Error())
		return nil
	}
	guaranteedRuntimeEndsAt := lastStartTimestamp.Add(minGuaranteedRuntimeDuration)
	return &guaranteedRuntimeEndsAt
}
