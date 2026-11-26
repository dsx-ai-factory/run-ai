/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	workflow "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/argo/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
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
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object {
		// the scheme knows Workflow only as unstructured
		object := &unstructured.Unstructured{}
		object.SetGroupVersionKind(workflow.GroupVersion.WithKind("Workflow"))
		return object
	})
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

	if err = h.setActivePodsAsRequested(ctx, argoWorkflow, externalWorkload); err != nil {
		return err
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

// setActivePodsAsRequested sets the requested pods and resources to those of the workflow's pending and running
// pods, since the workflow's tasks don't all run at the same time.
func (h *ArgoWorkflowTopOwnerHandler) setActivePodsAsRequested(ctx context.Context, argoWorkflow *workflow.Workflow,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	workloadId := topowner.GetRunaiWorkloadId(argoWorkflow.UID, argoWorkflow.Annotations)
	pods, err := getters.GetPodsOfWorkload(ctx, workloadId, argoWorkflow.Namespace, h.cachedClient)
	if err != nil {
		return fmt.Errorf("failed listing pods of workflow <%s/%s>: %w", argoWorkflow.Namespace, argoWorkflow.Name, err)
	}

	activePods := int32(0)
	requestedResources := corev1.ResourceRequirements{}
	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodPending && pod.Status.Phase != corev1.PodRunning {
			continue
		}
		activePods++
		requestedResources = resources.SumResourceRequirements(requestedResources,
			commonparsers.GetPodRequestedResources(&corev1.PodTemplateSpec{ObjectMeta: pod.ObjectMeta, Spec: pod.Spec}))
	}

	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(activePods)
	externalWorkload.Spec.RequestedResources = requestedResources
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
