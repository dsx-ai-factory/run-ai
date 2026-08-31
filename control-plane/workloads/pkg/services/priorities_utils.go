/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"k8s.io/utils/ptr"
)

func PriorityModelToApi(priorityModel models.Priority) api.WorkloadPriority {
	status := []api.PriorityStatus{}
	statusMap := map[string]bool{}
	for _, modelStatus := range priorityModel.Status {
		if modelStatus.Phase != "" {
			status = append(status, api.PriorityStatus{
				ClusterId: modelStatus.ClusterId.String(),
				Phase:     modelStatus.Phase,
			})
			statusMap[modelStatus.Phase] = true
		}
	}
	return api.WorkloadPriority{
		Id:              priorityModel.ID.String(),
		Source:          priorityModel.Source,
		Name:            priorityModel.Name,
		Value:           int32(priorityModel.Value),
		ScopeType:       priorityModel.ScopeType,
		ScopeId:         priorityModel.ScopeId,
		Status:          status,
		AggregatedPhase: calculateAggregatedPhase(statusMap),
	}
}

func PriorityModelToClusterApi(priorityModel models.PriorityClusterView) api.ClusterPriority {
	return api.ClusterPriority{
		Id:     priorityModel.ID.String(),
		Source: priorityModel.Source,
		Name:   priorityModel.Name,
		Value:  int32(priorityModel.Value),
		Phase:  priorityModel.Phase,
	}
}

func PriorityModelToSyncApi(priorityModel models.Priority) api.SyncPriority {
	priority := api.SyncPriority{
		Id:        priorityModel.ID.String(),
		Source:    priorityModel.Source,
		Name:      priorityModel.Name,
		Value:     int32(priorityModel.Value),
		CreatedAt: priorityModel.CreatedAt,
		UpdatedAt: priorityModel.UpdatedAt,
	}
	if priorityModel.DeletedAt.Valid {
		priority.DeletedAt = ptr.To(priorityModel.DeletedAt.Time)
	}
	return priority
}

func calculateAggregatedPhase(statusMap map[string]bool) string {
	hasFailedStatus := statusMap[StatusFailed]
	if hasFailedStatus {
		return StatusFailed
	}
	hasDeletingStatus := statusMap[StatusDeleting]
	if hasDeletingStatus {
		return StatusDeleting
	}
	hasUpdatingStatus := statusMap[StatusUpdating]
	if hasUpdatingStatus {
		return StatusUpdating
	}
	hasCreatingStatus := statusMap[StatusCreating]
	if hasCreatingStatus {
		return StatusCreating
	}
	return StatusReady
}
