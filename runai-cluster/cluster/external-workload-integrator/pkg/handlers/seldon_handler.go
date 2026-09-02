/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	seldonv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1"
	seldonv1alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// we cannot import "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1" because
// their operator is using a very old version of controller-runtime, so it's incompatible with our module.
// all code under "external-workload-integrator/pkg/thirdpartyapi/seldon" is copied from their repo.

var seldonStatusToExternalWorkloadPhase = map[seldonv1.StatusState]v1alpha1.ExternalWorkloadPhase{
	seldonv1.StatusStateAvailable: v1alpha1.ExWRunningPhase,
	seldonv1.StatusStateCreating:  v1alpha1.ExWCreatingPhase,
	seldonv1.StatusStateFailed:    v1alpha1.ExWFailedPhase,
	seldonv1.StatusState(""):      v1alpha1.ExWUnknownPhase,
}

type SeldonTopOwnerHandler struct {
	cachedClient client.Client
}

func NewSeldonTopOwnerHandler(cachedClient client.Client) *SeldonTopOwnerHandler {
	return &SeldonTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *SeldonTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	seldonDeployment, ok := object.(*seldonv1.SeldonDeployment)
	if !ok {
		seldonDeploymentv1alpha2, ok2 := object.(*seldonv1alpha2.SeldonDeployment)
		if !ok2 {
			return fmt.Errorf("failed converting object to seldonv1/seldonv1alpha2.SeldonDeployment")
		}

		seldonDeployment = &seldonv1.SeldonDeployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SeldonDeployment",
				APIVersion: "machinelearning.seldon.io/v1alpha2",
			},
			ObjectMeta: seldonDeploymentv1alpha2.ObjectMeta,
			Spec:       seldonDeploymentv1alpha2.Spec,
			Status:     seldonDeploymentv1alpha2.Status,
		}
	}

	h.parseSeldonDeploymentSpec(ctx, seldonDeployment.ObjectMeta, &seldonDeployment.Spec, seldonDeployment.GroupVersionKind().GroupKind(), externalWorkload)

	commonparsers.FillDefaultNodePoolIfNeeded(externalWorkload)

	externalWorkload.Status.ObservedStatus.K8sStatus = string(seldonDeployment.Status.State)
	externalWorkload.Status.ObservedStatus.Phase = convertSeldonStatusPhase(seldonDeployment.Status.State)
	externalWorkload.Status.ObservedStatus.Message = seldonDeployment.Status.Description
	externalWorkload.Status.ObservedStatus.Conditions = convertConditions(seldonDeployment.Status.Conditions)

	return nil
}

func (h *SeldonTopOwnerHandler) parseSeldonDeploymentSpec(ctx context.Context, objectMeta metav1.ObjectMeta,
	seldonDeploymentSpec *seldonv1.SeldonDeploymentSpec, seldonGroupKind schema.GroupKind, externalWorkload *v1alpha1.ExternalWorkload) {
	specReplicas := utils.DerefInt32(seldonDeploymentSpec.Replicas, 1)

	var totalReplicas int32 = 0
	var totalMinReplicas int32 = 0
	var totalMaxReplicas int32 = 0
	externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{}
	var podSpecPriorityClassName string

	for _, predictor := range seldonDeploymentSpec.Predictors {
		predictorReplicas := utils.DerefInt32(predictor.Replicas, specReplicas)

		for _, seldonPodSpec := range predictor.ComponentSpecs {
			var componentReplicas int32
			if seldonPodSpec.Replicas != nil {
				componentReplicas = *seldonPodSpec.Replicas
			} else if seldonPodSpec.HpaSpec != nil && seldonPodSpec.HpaSpec.MinReplicas != nil {
				componentReplicas = *seldonPodSpec.HpaSpec.MinReplicas
			} else {
				componentReplicas = predictorReplicas
			}

			totalReplicas += componentReplicas
			if seldonPodSpec.HpaSpec != nil {
				totalMinReplicas += utils.DerefInt32(seldonPodSpec.HpaSpec.MinReplicas, 0)
				totalMaxReplicas += seldonPodSpec.HpaSpec.MaxReplicas
			}

			podTemplate := &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: utils.MergeAnnotations(seldonPodSpec.Metadata.Annotations, objectMeta.Annotations,
						predictor.Annotations),
				},
				Spec: seldonPodSpec.Spec,
			}
			podRequestedResources := commonparsers.GetPodRequestedResources(podTemplate)
			workerRequestedResources := commonparsers.MultiplyResources(componentReplicas, podRequestedResources)

			externalWorkload.Spec.RequestedResources = resources.SumResourceRequirements(
				externalWorkload.Spec.RequestedResources, workerRequestedResources)

			commonparsers.FillRequestedNodePools(ctx, objectMeta, "SeldonDeployment", seldonPodSpec.Spec.Affinity,
				h.cachedClient, externalWorkload)

			if podSpecPriorityClassName == "" && seldonPodSpec.Spec.PriorityClassName != "" {
				podSpecPriorityClassName = seldonPodSpec.Spec.PriorityClassName
			}
		}
	}

	externalWorkload.Spec.RequestedPods.Replicas = utils.Int32PtrOrNil(totalReplicas)
	externalWorkload.Spec.RequestedPods.MinReplicas = utils.Int32PtrOrNil(totalMinReplicas)
	externalWorkload.Spec.RequestedPods.MaxReplicas = utils.Int32PtrOrNil(totalMaxReplicas)
	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, objectMeta.Labels,
		seldonGroupKind, podSpecPriorityClassName)
}

func convertSeldonStatusPhase(status seldonv1.StatusState) v1alpha1.ExternalWorkloadPhase {
	phase, found := seldonStatusToExternalWorkloadPhase[status]
	if found {
		return phase
	}
	return v1alpha1.ExWUnknownPhase
}

func convertConditions(conditions duckv1.Conditions) []v1alpha1.ObservedCondition {
	result := []v1alpha1.ObservedCondition{}
	for _, condition := range conditions {
		result = append(result, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime.Inner,
		})
	}

	return result
}

func (h *SeldonTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &seldonv1.SeldonDeployment{} })
}
