/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectStatusResourceHandler interface {
	ReconcileStatus(ctx context.Context, project *kaiv1alpha1.Project, conditionsChanged bool) error
	ReconcileStatusOnDeletion(ctx context.Context, project *kaiv1alpha1.Project, conditionsChanged bool) error
}

type ProjectStatusHandler struct {
	common.WithLoggerAndCli
}

type ProjectConditionReason string

const (
	NamespaceHandlerFailed ProjectConditionReason = "NamespaceHandlerFailed"
	NamespaceGetFailed     ProjectConditionReason = "NamespaceGetFailed"
	NamespaceNotFound      ProjectConditionReason = "NamespaceNotFound"
	NamespaceLabelMissing  ProjectConditionReason = "NamespaceLabelMissing"
	NamespaceUpdateFailed  ProjectConditionReason = "NamespaceUpdateFailed"
	NamespaceCreateFailed  ProjectConditionReason = "NamespaceCreateFailed"

	LimitRangeHandlerFailed   ProjectConditionReason = "LimitRangeHandlerFailed"
	QueuesHandlerFailed       ProjectConditionReason = "QueuesHandlerFailed"
	RoleBindingsHandlerFailed ProjectConditionReason = "RoleBindingsHandlerFailed"
	SecretsHandlerFailed      ProjectConditionReason = "SecretsHandlerFailed"
	PvcsHandlerFailed         ProjectConditionReason = "PvcsHandlerFailed"

	PhaseDeleting = "Deleting"
)

func NewProjectStatusHandler(client client.Client) ProjectStatusHandler {
	return ProjectStatusHandler{
		WithLoggerAndCli: common.WithLoggerAndCli{
			Client: client,
			Log:    ctrl.Log.WithName("resource_handlers").WithName(common.LogProjectStatusTag),
		},
	}
}

func (handler ProjectStatusHandler) ReconcileStatus(ctx context.Context,
	project *kaiv1alpha1.Project, conditionsChanged bool) error {
	handler.Log.Info("Reconciling Project Status", common.LogProjectTag, project.Name)

	var err error

	projectStatusCopy := project.Status.DeepCopy()
	innerErr := handler.getProjectStatus(ctx, project, conditionsChanged)
	err = common.AppendErrIfNotNil(err, innerErr)
	// even if failed - continue to update status anyway, because it might have partially succeeded

	innerErr = handler.updateProjectStatusIfNeeded(ctx, project, projectStatusCopy, conditionsChanged)
	err = common.AppendErrIfNotNil(err, innerErr)

	return err
}

func (handler ProjectStatusHandler) ReconcileStatusOnDeletion(ctx context.Context,
	project *kaiv1alpha1.Project, conditionsChanged bool) error {
	projectStatusCopy := project.Status.DeepCopy()
	handler.setPhase(&project.Status, conditionsChanged, true)

	return handler.updateProjectStatusIfNeeded(ctx, project, projectStatusCopy, conditionsChanged)
}

func (handler ProjectStatusHandler) getProjectStatus(ctx context.Context,
	project *kaiv1alpha1.Project, conditionsChanged bool) error {
	err := handler.setNamespace(project)

	handler.setPhase(&project.Status, conditionsChanged, false)

	innerErr := handler.setQuotaStatuses(ctx, &project.Status, project.Name)
	err = common.AppendErrIfNotNil(err, innerErr)

	return err
}

func (handler ProjectStatusHandler) setNamespace(project *kaiv1alpha1.Project) error {
	namespace, err := handler.KaiProjectToNamespace(project)
	if err != nil {
		handler.Log.Error(err, "To get project status, failed getting namespace",
			common.LogProjectTag, project.Name)
		return err
	}

	project.Status.Namespace = namespace
	return nil
}

func (handler ProjectStatusHandler) setPhase(status *kaiv1alpha1.ProjectStatus, conditionsChanged bool, onDeletion bool) {
	if !conditionsChanged {
		return
	}
	failedHandlersReasons := []string{}
	phase := kaiv1alpha1.Ready

	for _, condition := range status.Conditions {
		if condition.Status == corev1.ConditionFalse {
			if onDeletion {
				phase = PhaseDeleting
			} else {
				phase = kaiv1alpha1.NotReady
			}

			failedHandlersReasons = append(failedHandlersReasons, condition.Reason)
		}
	}

	status.Phase = phase
	status.Message = strings.Join(failedHandlersReasons, ", ")
}

func (handler ProjectStatusHandler) setQuotaStatuses(ctx context.Context,
	status *kaiv1alpha1.ProjectStatus, projectName string) error {
	queues, err := ListProjectQueues(ctx, handler.Client, projectName)
	if err != nil {
		handler.Log.Error(err, "Failed listing all queues for project", common.LogProjectTag, projectName)
		return err
	}

	status.NodePoolsQuotaStatuses = []kaiv1alpha1.QuotaStatus{}

	for _, queue := range queues {
		nodePoolName := GetObjectNodePoolFromLabels(queue.Labels)
		status.NodePoolsQuotaStatuses = append(status.NodePoolsQuotaStatuses,
			kaiv1alpha1.QuotaStatus{NodePoolName: nodePoolName, QueueStatus: queue.Status})
	}

	return nil
}

func (handler ProjectStatusHandler) updateProjectStatusIfNeeded(ctx context.Context,
	project *kaiv1alpha1.Project, projectStatusCopy *kaiv1alpha1.ProjectStatus, conditionsChanged bool) error {
	if conditionsChanged ||
		project.Status.Namespace != projectStatusCopy.Namespace ||
		project.Status.Phase != projectStatusCopy.Phase ||
		project.Status.Message != projectStatusCopy.Message {
		return handler.updateStatus(ctx, project)
	}

	newQuotaStatuses, err := json.Marshal(project.Status.NodePoolsQuotaStatuses)
	if err != nil {
		handler.Log.Error(err, fmt.Sprintf("Failed to json marshal newQuotaStatuses <%+v>",
			newQuotaStatuses))
		return err
	}
	previousQuotaStatuses, err := json.Marshal(projectStatusCopy.NodePoolsQuotaStatuses)
	if err != nil {
		handler.Log.Error(err,
			fmt.Sprintf("Failed to json marshal previousQuotaStatuses <%+v> - will update quota status anyway",
				previousQuotaStatuses))
		return handler.updateStatus(ctx, project)
	}

	if !reflect.DeepEqual(newQuotaStatuses, previousQuotaStatuses) {
		return handler.updateStatus(ctx, project)
	}
	return nil
}

func (handler ProjectStatusHandler) updateStatus(ctx context.Context, project *kaiv1alpha1.Project) error {
	sumNodePoolsQuotaStatusesToProjectQuotaStatus(&project.Status)

	patchBytes, err := updateProjectStatusPatchBytes(project.Status)
	if err != nil {
		handler.Log.Error(err, "Error creating patch bytes for project status update",
			common.LogProjectTag, project.Name)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = handler.Client.Status().Patch(ctx, project, patch)

	if err != nil {
		handler.Log.Error(err, "Error updating status of project",
			common.LogProjectTag, project.Name,
			common.LogNamespaceTag, project.Status.Namespace)
		return err
	}

	handler.Log.Info("Updated status of project",
		common.LogProjectTag, project.Name,
		common.LogNamespaceTag, project.Status.Namespace)
	return nil
}

func sumNodePoolsQuotaStatusesToProjectQuotaStatus(projectStatus *kaiv1alpha1.ProjectStatus) {
	projectQuotaStatus := kaiv1alpha1.ProjectQuotaStatus{
		Allocated:               corev1.ResourceList{},
		AllocatedNonPreemptible: corev1.ResourceList{},
		Requested:               corev1.ResourceList{},
	}

	for _, nodePoolQuotaStatus := range projectStatus.NodePoolsQuotaStatuses {
		addResourceList(&projectQuotaStatus.Allocated, &nodePoolQuotaStatus.QueueStatus.Allocated)
		addResourceList(&projectQuotaStatus.AllocatedNonPreemptible,
			&nodePoolQuotaStatus.QueueStatus.AllocatedNonPreemptible)
		addResourceList(&projectQuotaStatus.Requested, &nodePoolQuotaStatus.QueueStatus.Requested)
	}

	projectStatus.QuotaStatus = projectQuotaStatus
}

func addResourceList(result *corev1.ResourceList, resourceListToAdd *corev1.ResourceList) {
	for resourceName, quantityToAdd := range *resourceListToAdd {
		existingValue, found := (*result)[resourceName]
		if found {
			existingValue.Add(quantityToAdd)
			(*result)[resourceName] = existingValue
		} else {
			(*result)[resourceName] = quantityToAdd
		}
	}
}

func updateProjectStatusPatchBytes(projectStatus kaiv1alpha1.ProjectStatus) ([]byte, error) {
	// Build status map explicitly to handle merge patch semantics correctly -
	// send "nil" when wanting to clear a map field, send empty slice when wanting to clear an array field.
	quotaStatus := map[string]any{
		"allocated":               resourceListForPatch(projectStatus.QuotaStatus.Allocated),
		"allocatedNonPreemptible": resourceListForPatch(projectStatus.QuotaStatus.AllocatedNonPreemptible),
		"requested":               resourceListForPatch(projectStatus.QuotaStatus.Requested),
	}

	status := map[string]any{
		"namespace":              projectStatus.Namespace,
		"phase":                  projectStatus.Phase,
		"message":                projectStatus.Message,
		"conditions":             sliceForPatch(projectStatus.Conditions),
		"nodePoolsQuotaStatuses": sliceForPatch(projectStatus.NodePoolsQuotaStatuses),
		"quotaStatus":            quotaStatus,
	}
	statusMap := map[string]any{"status": status}
	return json.Marshal(statusMap)
}

func resourceListForPatch(rl corev1.ResourceList) corev1.ResourceList {
	if len(rl) == 0 {
		return nil
	}
	return rl
}

func sliceForPatch[T any](slice []T) []T {
	if slice == nil {
		return []T{}
	}
	return slice
}
