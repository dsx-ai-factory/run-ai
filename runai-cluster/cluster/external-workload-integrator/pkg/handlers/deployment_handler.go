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
	"github.com/run-ai/runai/runai-cluster/common/priority"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProgressingDeploymentStatus = "Progressing"
	newReplicaSetCreated        = "NewReplicaSetCreated"
	foundNewReplicaSet          = "FoundNewReplicaSet"
	replicaSetUpdated           = "ReplicaSetUpdated"

	CompleteDeploymentStatus = "Complete"
	newReplicaSetAvailable   = "NewReplicaSetAvailable"

	DegradedDeploymentStatus = "Degraded"
	progressDeadlineExceeded = "ProgressDeadlineExceeded"

	progressingConditionNotFound = "Progressing condition type not found"
)

type DeploymentTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewDeploymentTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *DeploymentTopOwnerHandler {
	return &DeploymentTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *DeploymentTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	deployment, ok := object.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("failed converting object to appsv1.Deployment")
	}

	commonparsers.HandleK8sPodControllerSpec(ctx, deployment.ObjectMeta, deployment.GroupVersionKind().GroupKind(),
		deployment.Spec.Replicas, &deployment.Spec.Template, h.cachedClient, h.draParser, externalWorkload)

	if externalWorkload.Spec.PriorityClassName == "" {
		externalWorkload.Spec.PriorityClassName = priority.InferencePriorityClassName
	}

	h.setDeploymentObservedStatus(ctx, deployment, &externalWorkload.Status.ObservedStatus)

	return nil
}

// setDeploymentObservedStatus calculates the status of the deployment, as described here:
// https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#deployment-status
func (h *DeploymentTopOwnerHandler) setDeploymentObservedStatus(ctx context.Context,
	deployment *appsv1.Deployment, observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	// will be changed later if needed
	observedStatus.Phase = v1alpha1.ExWUnknownPhase
	observedStatus.Message = progressingConditionNotFound

	observedConditions := []v1alpha1.ObservedCondition{}
	deploymentStatus := deployment.Status
	for _, condition := range deploymentStatus.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})

		if condition.Type != appsv1.DeploymentProgressing {
			continue
		}

		observedStatus.Message = condition.Message

		reason := condition.Reason
		if condition.Status == corev1.ConditionTrue {
			if reason == newReplicaSetCreated || reason == foundNewReplicaSet || reason == replicaSetUpdated {
				observedStatus.Phase = v1alpha1.ExWPendingPhase
				observedStatus.K8sStatus = ProgressingDeploymentStatus
				continue
			}

			if reason == newReplicaSetAvailable {
				observedStatus.Phase = v1alpha1.ExWRunningPhase
				observedStatus.K8sStatus = CompleteDeploymentStatus
				continue
			}
		}

		if condition.Status == corev1.ConditionFalse && reason == progressDeadlineExceeded {
			observedStatus.Phase = v1alpha1.ExWDegradedPhase
			observedStatus.K8sStatus = DegradedDeploymentStatus
			continue
		}

		// k8s status cannot be parsed
		observedStatus.Phase = v1alpha1.ExWUnknownPhase
	}

	observedStatus.Conditions = observedConditions

	if observedStatus.Phase == v1alpha1.ExWUnknownPhase || observedStatus.Phase == v1alpha1.ExWPendingPhase {
		// if the deployment is still in progress, we need to check the pods
		commonparsers.CalculateStatusByPods(ctx, h.cachedClient, deployment, observedStatus)
	}
}

// ShouldTriggerTopOwnerHandlerAsWell - since we look at pods to calculate status of the deployment,
// in case of a pod event, we need to trigger the deployment handler as well
func (h *DeploymentTopOwnerHandler) ShouldTriggerTopOwnerHandlerAsWell(triggeredOwnedObject client.Object) bool {
	switch triggeredOwnedObject.(type) {
	case *corev1.Pod:
		return true
	}
	return false
}

func (h *DeploymentTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &appsv1.Deployment{} })
}

func (h *DeploymentTopOwnerHandler) Filter(object client.Object) bool {
	deployment, ok := object.(*appsv1.Deployment)
	if !ok {
		return false
	}
	return deployment.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
