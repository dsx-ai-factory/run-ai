/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"time"

	org_unit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

const (
	// LastSyncTimestampAnnotation is the annotation cluster-sync writes on the new
	// kai.resources Project CR to record the last time its status was synced to the
	// control-plane.
	LastSyncTimestampAnnotation = "run.ai/last-sync-timestamp"
	DeletingPhase               = "Deleting"
)

func ParseKaiProjectToUpdateStatusRequest(project *kaiv1alpha1.Project) org_unit_openapi.ProjectStatusUpdateRequest {
	return org_unit_openapi.ProjectStatusUpdateRequest{
		Id:        project.Labels[ProjectIdLabel],
		ClusterId: config.Config().ClusterUuid,
		Status: org_unit_openapi.ProjectStatus{
			Namespace:             *org_unit_openapi.NewNullableString(ptr.To(project.Status.Namespace)),
			Phase:                 convertKaiPhase(project.Status.Phase),
			PhaseMessage:          *org_unit_openapi.NewNullableString(ptr.To(project.Status.Message)),
			LastUpdatedTime:       *org_unit_openapi.NewNullableTime(ptr.To(lastSyncTimeFromAnnotations(project.Annotations))),
			NodePoolQuotaStatuses: convertKaiNodePoolQuotaStatus(project.Status.NodePoolsQuotaStatuses),
			QuotaStatus:           convertKaiQuotaStatus(project.Status.QuotaStatus),
			AdditionalStatusData: map[string]any{
				"conditions": project.Status.Conditions,
			},
		},
	}
}

func ParseKaiProjectToDeleteRequest(project *kaiv1alpha1.Project) org_unit_openapi.ProjectStatusUpdateRequest {
	return org_unit_openapi.ProjectStatusUpdateRequest{
		Id:        project.Labels[ProjectIdLabel],
		ClusterId: config.Config().ClusterUuid,
		Status: org_unit_openapi.ProjectStatus{
			Phase: ptr.To(org_unit_openapi.PROJECTPHASE_DELETED),
		},
	}
}

// lastSyncTimeFromAnnotations reads the last sync timestamp cluster-sync previously
// wrote onto the org unit. A missing or malformed annotation yields the zero time,
// matching the legacy behavior when status.syncStatus.lastSyncTimestamp was unset.
func lastSyncTimeFromAnnotations(annotations map[string]string) time.Time {
	raw, ok := annotations[LastSyncTimestampAnnotation]
	if !ok {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func convertKaiPhase(phase kaiv1alpha1.ProjectPhase) *org_unit_openapi.ProjectPhase {
	switch phase {
	case "":
		return ptr.To(org_unit_openapi.PROJECTPHASE_INITIALIZING)
	case DeletingPhase:
		return ptr.To(org_unit_openapi.PROJECTPHASE_DELETING)
	case kaiv1alpha1.NotReady:
		return ptr.To(org_unit_openapi.PROJECTPHASE_NOT_READY)
	case kaiv1alpha1.Ready:
		return ptr.To(org_unit_openapi.PROJECTPHASE_READY)
	}
	return ptr.To(org_unit_openapi.PROJECTPHASE_UNKNOWN)
}

func convertKaiNodePoolQuotaStatus(nodePoolQuotaStatus []kaiv1alpha1.QuotaStatus) []org_unit_openapi.QuotaStatusNodePool {
	apiNodePoolStatus := []org_unit_openapi.QuotaStatusNodePool{}
	for _, nodePoolStatus := range nodePoolQuotaStatus {
		quotaNodePoolStatus := org_unit_openapi.QuotaStatusNodePool{
			Allocated:               convertResourcesListToQuotaStatusResourceList(nodePoolStatus.QueueStatus.Allocated),
			AllocatedNonPreemptible: convertResourcesListToQuotaStatusResourceList(nodePoolStatus.QueueStatus.AllocatedNonPreemptible),
			Requested:               convertResourcesListToQuotaStatusResourceList(nodePoolStatus.QueueStatus.Requested),
			NodePoolName:            nodePoolStatus.NodePoolName,
		}
		apiNodePoolStatus = append(apiNodePoolStatus, quotaNodePoolStatus)
	}

	return apiNodePoolStatus
}

func convertKaiQuotaStatus(quotaStatus kaiv1alpha1.ProjectQuotaStatus) *org_unit_openapi.QuotaStatus {
	return &org_unit_openapi.QuotaStatus{
		Allocated:               convertResourcesListToQuotaStatusResourceList(quotaStatus.Allocated),
		AllocatedNonPreemptible: convertResourcesListToQuotaStatusResourceList(quotaStatus.AllocatedNonPreemptible),
		Requested:               convertResourcesListToQuotaStatusResourceList(quotaStatus.Requested),
	}
}
