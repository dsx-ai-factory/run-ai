/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"strings"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	knativeapis "knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	InferenceType = "Inference"

	MinScaleAnnotation = "autoscaling.knative.dev/min-scale"
	MaxScaleAnnotation = "autoscaling.knative.dev/max-scale"
)

type KNativeTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewKNativeTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *KNativeTopOwnerHandler {
	return &KNativeTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *KNativeTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	knativeService, ok := object.(*servingv1.Service)
	if !ok {
		return fmt.Errorf("failed converting object to servingv1.Service")
	}

	if knativeService.Kind == "" {
		// If the Kind is not set, we assume it's a Knative Service
		knativeService.Kind = "Service"
	}
	if knativeService.APIVersion == "" {
		knativeService.APIVersion = "serving.knative.dev/v1"
	}

	templateAnnotations := knativeService.Spec.Template.Annotations
	podTemplate := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: templateAnnotations,
		},
		Spec: knativeService.Spec.Template.Spec.PodSpec,
	}

	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, knativeService.ObjectMeta, knativeService.GroupVersionKind().GroupKind(),
		podTemplate, nil, h.cachedClient, h.draParser, externalWorkload)

	minReplicas := getAnnotationIntOrNil(templateAnnotations, MinScaleAnnotation)
	if minReplicas == nil {
		minReplicas = ptr.To(int32(1))
	}

	externalWorkload.Spec.RequestedPods.MinReplicas = minReplicas
	externalWorkload.Spec.RequestedPods.MaxReplicas = getAnnotationIntOrNil(templateAnnotations, MaxScaleAnnotation)

	commonparsers.MultiplyRequestedResourcesPerNumberOfRequestedPods(*minReplicas, externalWorkload)

	setKnativeExternalWorkloadAdditionalFields(knativeService, externalWorkload)
	setKnativeTypeIfNeeded(knativeService, externalWorkload)

	h.parseKNativeStatus(ctx, knativeService, &externalWorkload.Status.ObservedStatus)
	return nil
}

func (h *KNativeTopOwnerHandler) parseKNativeStatus(ctx context.Context, knativeService *servingv1.Service, observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	status := &knativeService.Status
	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range status.Conditions {
		if condition.Type == knativeapis.ConditionReady {
			if condition.Status == corev1.ConditionTrue {
				observedStatus.Phase = v1alpha1.ExWRunningPhase
				observedStatus.K8sStatus = "Ready"
			} else {
				observedStatus.Phase = v1alpha1.ExWPendingPhase
				observedStatus.Message = condition.Message
				observedStatus.K8sStatus = "NotReady"
			}
		}

		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime.Inner,
		})
	}

	observedStatus.Conditions = observedConditions

	if observedStatus.Phase == "" {
		observedStatus.Phase = v1alpha1.ExWUnknownPhase
	}

	if observedStatus.Phase == v1alpha1.ExWUnknownPhase || observedStatus.Phase == v1alpha1.ExWPendingPhase {
		// if the knative is still in progress, we need to check the pods
		commonparsers.CalculateStatusByPods(ctx, h.cachedClient, knativeService, observedStatus)
	}
}

// ShouldTriggerTopOwnerHandlerAsWell - since we look at pods to calculate status of the knative,
// in case of a pod event, we need to trigger the knative handler as well
func (h *KNativeTopOwnerHandler) ShouldTriggerTopOwnerHandlerAsWell(triggeredOwnedObject client.Object) bool {
	switch triggeredOwnedObject.(type) {
	case *corev1.Pod:
		return true
	}
	return false
}

func (h *KNativeTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &servingv1.Service{} })
}

func getAnnotationIntOrNil(annotations map[string]string, annotationKey string) *int32 {
	value, err := utils.ParseAnnotationInt32(annotations, annotationKey)
	if err != nil {
		return nil
	}
	return &value
}

func setKnativeExternalWorkloadAdditionalFields(knativeService *servingv1.Service,
	externalWorkload *v1alpha1.ExternalWorkload) {
	if knativeService.Status.URL == nil {
		return
	}
	url := knativeService.Status.URL.String()
	if url == "" {
		return
	}

	// Handle JobUrl annotation (if present)
	if jobURLs, exists := knativeService.Annotations[clusterconstants.AnnotationJobUrl]; exists {
		externalWorkload.Spec.Urls = []string{jobURLs, url}
	} else {
		externalWorkload.Spec.Urls = []string{url}
	}
}

func setKnativeTypeIfNeeded(knativeService *servingv1.Service, externalWorkload *v1alpha1.ExternalWorkload) {
	_, workloadNameFound := knativeService.Labels[commonconsts.ControlPlaneWorkloadNameLabelForWorkload]
	if !workloadNameFound {
		return
	}

	workloadKind, workloadKindFound := knativeService.Labels[commonconsts.ControlPlaneWorkloadKindLabelForWorkload]
	if workloadKindFound && strings.Contains(workloadKind, InferenceType) {
		externalWorkload.Spec.Type = InferenceType
	}
}

func (h *KNativeTopOwnerHandler) Filter(object client.Object) bool {
	knativeService, ok := object.(*servingv1.Service)
	if !ok {
		return false
	}

	return knativeService.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}

func (h *KNativeTopOwnerHandler) ShouldUpdateNetworkStatus() bool {
	return true
}
