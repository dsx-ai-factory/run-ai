/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	org_unit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

func ParseKaiDepartmentToUpdateStatusRequest(department *kaiv1alpha1.Department) org_unit_openapi.DepartmentStatusUpdateRequest {
	return org_unit_openapi.DepartmentStatusUpdateRequest{
		ClusterId: &config.Config().ClusterUuid,
		Status: org_unit_openapi.DepartmentStatus{
			LastUpdatedTime: *org_unit_openapi.NewNullableTime(ptr.To(lastSyncTimeFromAnnotations(department.Annotations))),
		},
	}
}
