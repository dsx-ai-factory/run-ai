/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const (
	SyncStatusTableName = "sync_status"

	SyncStatusColumnProcessedByAgent = "processed_by_agent"
	SyncStatusColumnSentToAgent      = "sent_to_agent"
	SyncStatusColumnSyncMessage      = "sync_message"

	SyncStatusColumnWorkloadId   = "workload_id"
	SyncStatusColumnDeleting     = "deleting"
	SyncStatusColumnTenantId     = "tenant_id"
	SyncStatusColumnClusterId    = "cluster_id"
	SyncStatusColumnDepartmentId = "department_id"
	SyncStatusColumnProjectId    = "project_id"
)

type SyncStatusRecord struct {
	WorkloadId string `gorm:"primaryKey"`

	TenantId     int32  `gorm:"not null;"`
	ClusterId    string `gorm:"not null;"`
	DepartmentId string
	ProjectId    string

	Deleting bool `gorm:"not null"`

	SentToAgent      int64 `gorm:"not null"`
	ProcessedByAgent int64 `gorm:"not null"`

	SyncMessage *string
}

func (*SyncStatusRecord) TableName() string {
	return SyncStatusTableName
}

func (q *SyncStatusRecord) ToSyncInfo(updateCount int64) api.WorkloadV2SyncInfo {
	var result api.WorkloadV2SyncInfo

	if updateCount > q.SentToAgent {
		result.Status = api.WORKLOADV2SYNCSTATUS_QUEUED
	} else if q.SentToAgent > q.ProcessedByAgent {
		if q.Deleting {
			result.Status = api.WORKLOADV2SYNCSTATUS_DELETING
		} else {
			result.Status = api.WORKLOADV2SYNCSTATUS_APPLYING
		}
	} else if convutils.IsEmptyOrNull(q.SyncMessage) {
		if q.Deleting {
			// typically once a workload is deleted, it is removed from the DB but
			// just in case, indicate this situation with a sensible status
			result.Status = api.WORKLOADV2SYNCSTATUS_DELETED
		} else {
			result.Status = api.WORKLOADV2SYNCSTATUS_SYNCED
		}
	} else {
		result.Status = api.WORKLOADV2SYNCSTATUS_FAILED
		result.Message = convutils.Deref(q.SyncMessage, "")
	}

	return result
}
