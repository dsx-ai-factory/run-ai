/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nimWorkloadType          = "NVIDIA-NIM"
	helmPackageLabelName     = "app.kubernetes.io/name"
	nimHelmPackageLabelValue = "nemollm-inference"
)

type StatefulSetTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewStatefulSetTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *StatefulSetTopOwnerHandler {
	return &StatefulSetTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *StatefulSetTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &appsv1.StatefulSet{} })
}

func (h *StatefulSetTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	statefulSet, ok := object.(*appsv1.StatefulSet)
	if !ok {
		return fmt.Errorf("failed converting object to appsv1.StatefulSet")
	}
	if isNimLabelExists(statefulSet.Labels) {
		externalWorkload.Spec.Type = nimWorkloadType
	}

	numberOfRequestedPods := commonparsers.HandleK8sPodControllerSpec(ctx, statefulSet.ObjectMeta, statefulSet.GroupVersionKind().GroupKind(),
		statefulSet.Spec.Replicas, &statefulSet.Spec.Template, h.cachedClient, h.draParser, externalWorkload)

	setStatefulSetObservedStatus(&statefulSet.Status, numberOfRequestedPods, &externalWorkload.Status.ObservedStatus)

	return nil
}

func setStatefulSetObservedStatus(statefulSetStatus *appsv1.StatefulSetStatus,
	numberOfRequestedPods int32,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	if numberOfRequestedPods == 0 {
		observedStatus.K8sStatus = commonparsers.SuspendedStatus
		if statefulSetStatus.ReadyReplicas == 0 {
			observedStatus.Phase = v1alpha1.ExWStoppedPhase
		} else {
			observedStatus.Phase = v1alpha1.ExWStoppingPhase
		}
	} else {
		if statefulSetStatus.ReadyReplicas == numberOfRequestedPods {
			observedStatus.Phase = v1alpha1.ExWRunningPhase
			observedStatus.K8sStatus = commonparsers.ReadyStatus
		} else {
			observedStatus.K8sStatus = commonparsers.NotReadyStatus
			if statefulSetStatus.ReadyReplicas > 0 {
				observedStatus.Phase = v1alpha1.ExWDegradedPhase
			} else if statefulSetStatus.Replicas > 0 {
				observedStatus.Phase = v1alpha1.ExWPendingPhase
			} else {
				observedStatus.Phase = v1alpha1.ExWCreatingPhase
			}
		}
	}

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range statefulSetStatus.Conditions {
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

func isNimLabelExists(labels map[string]string) bool {
	value, exists := labels[helmPackageLabelName]
	if !exists {
		return false
	}
	return value == nimHelmPackageLabelValue
}

func (h *StatefulSetTopOwnerHandler) Filter(object client.Object) bool {
	statefulSet, ok := object.(*appsv1.StatefulSet)
	if !ok {
		return false
	}
	return statefulSet.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
