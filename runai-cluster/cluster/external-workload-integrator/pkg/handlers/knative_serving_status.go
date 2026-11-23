/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"strconv"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *KNativeTopOwnerHandler) setServingWorkloadUpdateCounter(ctx context.Context,
	service *servingv1.Service, externalWorkload *v1alpha1.ExternalWorkload) {
	// A Ready old revision must not authorize a change to a newer, unready submission.
	runaistatus.SetAdditionalFields(externalWorkload, []string{commonconstants.ServingWorkloadUpdateCounterField}, nil)
	if service == nil || !service.DeletionTimestamp.IsZero() || !service.IsReady() {
		return
	}

	counter := service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter]
	parsedCounter, err := strconv.ParseInt(counter, 10, 64)
	if err != nil || parsedCounter < 0 || strconv.FormatInt(parsedCounter, 10) != counter {
		return
	}
	owner := metav1.GetControllerOf(service)
	if owner == nil || owner.Kind != dromosv1alpha1.WorkloadRunnerKind ||
		owner.APIVersion != dromosv1alpha1.SchemeGroupVersion.String() || owner.UID == "" {
		return
	}

	revision := service.Status.LatestReadyRevisionName
	if revision == "" || revision != service.Status.LatestCreatedRevisionName || len(service.Status.Traffic) != 1 {
		return
	}
	traffic := service.Status.Traffic[0]
	if traffic.RevisionName != revision || traffic.Percent == nil || *traffic.Percent != 100 {
		return
	}

	runner := &dromosv1alpha1.WorkloadRunner{}
	runner.SetGroupVersionKind(dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind))
	if err := h.cachedClient.Get(ctx, client.ObjectKey{Namespace: service.Namespace, Name: owner.Name}, runner); err != nil {
		log.Warn().Err(err).Msg("knative serving status: failed getting WorkloadRunner")
		return
	}
	if runner.UID != owner.UID || !runner.DeletionTimestamp.IsZero() ||
		runner.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] != counter {
		return
	}

	minimum := int32(1)
	if value, exists := service.Spec.Template.Annotations[MinScaleAnnotation]; exists {
		parsedMinimum, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsedMinimum < 0 {
			return
		}
		minimum = int32(parsedMinimum)
	}
	if minimum > 0 && !h.hasReadyServingPod(ctx, service, revision) {
		return
	}
	runaistatus.SetAdditionalFields(externalWorkload, nil,
		map[string]any{commonconstants.ServingWorkloadUpdateCounterField: counter})
}

func (h *KNativeTopOwnerHandler) hasReadyServingPod(ctx context.Context, service *servingv1.Service, revision string) bool {
	pods := &corev1.PodList{}
	if err := h.cachedClient.List(ctx, pods, client.InNamespace(service.Namespace), client.MatchingLabels{
		serving.ServiceLabelKey: service.Name, serving.RevisionLabelKey: revision,
	}); err != nil {
		log.Warn().Err(err).Msg("knative serving status: failed listing revision pods")
		return false
	}
	for _, pod := range pods.Items {
		if !pod.DeletionTimestamp.IsZero() || pod.Status.Phase != corev1.PodRunning {
			continue
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	return false
}
