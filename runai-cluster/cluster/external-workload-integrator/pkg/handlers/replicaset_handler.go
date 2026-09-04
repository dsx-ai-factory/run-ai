/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReplicaSetTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewReplicaSetTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *ReplicaSetTopOwnerHandler {
	return &ReplicaSetTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *ReplicaSetTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	replicaSet, ok := object.(*appsv1.ReplicaSet)
	if !ok {
		return fmt.Errorf("failed converting object to appsv1.ReplicaSet")
	}

	numberOfRequestedPods := commonparsers.HandleK8sPodControllerSpec(ctx, replicaSet.ObjectMeta, replicaSet.GroupVersionKind().GroupKind(),
		replicaSet.Spec.Replicas, &replicaSet.Spec.Template, h.cachedClient, h.draParser, externalWorkload)

	setReplicaSetObservedStatus(&replicaSet.Status, numberOfRequestedPods, &externalWorkload.Status.ObservedStatus)

	return nil
}

func setReplicaSetObservedStatus(replicaSetStatus *appsv1.ReplicaSetStatus,
	numberOfRequestedPods int32,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	if numberOfRequestedPods == 0 {
		observedStatus.K8sStatus = commonparsers.SuspendedStatus
		if replicaSetStatus.AvailableReplicas == 0 {
			observedStatus.Phase = v1alpha1.ExWStoppedPhase
		} else {
			observedStatus.Phase = v1alpha1.ExWStoppingPhase
		}
	} else {
		if replicaSetStatus.AvailableReplicas == numberOfRequestedPods {
			observedStatus.Phase = v1alpha1.ExWRunningPhase
			observedStatus.K8sStatus = commonparsers.ReadyStatus
		} else {
			observedStatus.K8sStatus = commonparsers.NotReadyStatus
			if replicaSetStatus.AvailableReplicas > 0 {
				observedStatus.Phase = v1alpha1.ExWDegradedPhase
			} else if replicaSetStatus.Replicas > 0 {
				observedStatus.Phase = v1alpha1.ExWPendingPhase
			} else {
				observedStatus.Phase = v1alpha1.ExWCreatingPhase
			}
		}
	}

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range replicaSetStatus.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})
	}

	observedStatus.Conditions = observedConditions
}

func (h *ReplicaSetTopOwnerHandler) Filter(object client.Object) bool {
	replicaSet, ok := object.(*appsv1.ReplicaSet)
	if !ok {
		return false
	}
	return replicaSet.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}

func (h *ReplicaSetTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &appsv1.ReplicaSet{} })
}
