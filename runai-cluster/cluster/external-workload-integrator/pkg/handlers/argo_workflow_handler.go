/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	workflow "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/argo/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ArgoWorkflowTopOwnerHandler struct {
	cachedClient client.Client
}

func NewArgoWorkflowTopOwnerHandler(cachedClient client.Client) *ArgoWorkflowTopOwnerHandler {
	return &ArgoWorkflowTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *ArgoWorkflowTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &workflow.Workflow{} })
}

func (h *ArgoWorkflowTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	marshaled, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed marshaling workflow.Workflow object: %v", err)
	}
	argoWorkflow := &workflow.Workflow{}
	err = json.Unmarshal(marshaled, argoWorkflow)
	if err != nil {
		return fmt.Errorf("failed unmarshaling workflow.Workflow object: %v", err)
	}

	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, argoWorkflow.Labels,
		argoWorkflow.GroupVersionKind().GroupKind(), argoWorkflow.Spec.PodPriorityClassName)

	//TODO: add more parsing. also commit change to ewi so integration tests CI runs.

	externalWorkload.Status.ObservedStatus.Phase = translateWorkflowStatus(argoWorkflow.Status.Phase)
	externalWorkload.Status.ObservedStatus.K8sStatus = string(argoWorkflow.Status.Phase)
	externalWorkload.Status.ObservedStatus.Message = argoWorkflow.Status.Message
	if externalWorkload.Spec.ImageName == "" && len(argoWorkflow.Spec.Templates) > 0 && argoWorkflow.Spec.Templates[0].Container != nil {
		externalWorkload.Spec.ImageName = argoWorkflow.Spec.Templates[0].Container.Image
	}

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range argoWorkflow.Status.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:    string(condition.Type),
			Status:  corev1.ConditionStatus(condition.Status),
			Reason:  string(condition.Type),
			Message: condition.Message,
		})
	}
	externalWorkload.Status.ObservedStatus.Conditions = observedConditions

	return nil
}

func translateWorkflowStatus(argoWorkflowPhase workflow.WorkflowPhase) v1alpha1.ExternalWorkloadPhase {
	switch argoWorkflowPhase {
	case workflow.WorkflowUnknown:
		return v1alpha1.ExWUnknownPhase
	case workflow.WorkflowPending:
		return v1alpha1.ExWPendingPhase
	case workflow.WorkflowRunning:
		return v1alpha1.ExWRunningPhase
	case workflow.WorkflowFailed:
		return v1alpha1.ExWFailedPhase
	case workflow.WorkflowSucceeded:
		return v1alpha1.ExWCompletedPhase
	case workflow.WorkflowError:
		return v1alpha1.ExWFailedPhase
	default:
		return v1alpha1.ExWUnknownPhase
	}
}
