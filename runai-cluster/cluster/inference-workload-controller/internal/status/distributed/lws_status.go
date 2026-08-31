/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_status

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	commondistributedinference "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/distributedinference"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

type LwsStatusCalculatorInterface interface {
	GetStatus(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
		lws *lwsv1.LeaderWorkerSet) (string, int32, int32, string, error)
}

type LwsStatusCalculator struct {
	k8sClient client.Client
}

func NewLwsStatusCalculator(k8sClient client.Client) *LwsStatusCalculator {
	return &LwsStatusCalculator{k8sClient: k8sClient}
}

func (c *LwsStatusCalculator) GetStatus(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	lws *lwsv1.LeaderWorkerSet,
) (string, int32, int32, string, error) {
	pods, err := c.listPods(ctx, workload)
	if err != nil {
		return "", 0, 0, "", err
	}

	replicas := int32(1)
	if lws.Spec.Replicas != nil {
		replicas = *lws.Spec.Replicas
	}
	groupSize := 1
	if lws.Spec.LeaderWorkerTemplate.Size != nil {
		groupSize = int(*lws.Spec.LeaderWorkerTemplate.Size)
	}
	result, err := commondistributedinference.CalculateRuntime(commondistributedinference.RuntimeInput{
		Replicas: replicas, GroupSize: groupSize, Pods: pods,
	})
	if err != nil {
		return "", 0, 0, "", err
	}
	return string(result.Phase), result.Replicas, result.ReadyReplicas, result.Message, nil
}

func (c *LwsStatusCalculator) listPods(ctx context.Context,
	workload *dinfw.DistributedInferenceWorkload,
) ([]corev1.Pod, error) {
	if workload == nil {
		return nil, nil
	}
	selector := labels.Set{
		lwsv1.SetNameLabelKey:            builders.GetUniqueLwsName(workload),
		commonconstants.TopOwnerUidLabel: topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations),
	}.AsSelector()
	pods := &corev1.PodList{}
	if err := c.k8sClient.List(ctx, pods, &client.ListOptions{
		Namespace: workload.Namespace, LabelSelector: selector,
	}); err != nil {
		return nil, err
	}
	return pods.Items, nil
}
