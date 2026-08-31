/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
)

const (
	DatavolumeTableName       = "datavolumes"
	DatavolumeTenantIdColumn  = "tenant_id"
	DatavolumeProjectIdColumn = "project_id"
	DatavolumePvcNameColumn   = "origin_pvc_name"

	SharedScopesPreloadKey = "SharedScopes"

	DatavolumeClusterNameUniqueConstraintStr = "(cluster_id, name)"
	// This constraint is enforced only on project scope datavolume, the same project scope pvc can not be used in the same project twice.
	DatavolumeClusterProjectPvcUniqueConstraintStr = "(cluster_id, project_id, origin_pvc_name)"
	// This constraint is enforced only on department scope datavolume, the same department scope pvc can not be used in the same department twice.
	DatavolumeClusterDepartmentPvcUniqueConstraintStr = "(cluster_id, department_id, origin_pvc_name)"
	// This constraint is enforced only on cluster scope datavolume, the same cluster scope pvc can not be used in the same cluster twice.
	DatavolumeClusterPvcUniqueConstraintStr = "(cluster_id, origin_pvc_name)"
)

type Datavolume struct {
	BaseModel
	Name                       string
	Description                string
	ProjectId                  string
	OriginPvcName              string
	ShouldDeleteOriginalVolume bool
	UseOriginalVolume          bool `gorm:"column:should_use_original_volume"`
	Scope                      api.ScopeType
	ProjectName                string
	DepartmentName             string
	DepartmentId               string
	TenantId                   uint
	ClusterId                  uuid.UUID
	Internal                   bool
	SharedScopes               []SharedScope
	DatavolumeStatus
}

type DatavolumeStatus struct {
	Phase             api.DatavolumePhase
	PhaseMessage      string
	Conditions        ConditionArray `gorm:"type:jsonb"`
	DatavolumePvcName string
	DatavolumePvName  string
}

func (Datavolume) TableName() string {
	return DatavolumeTableName
}

func (Datavolume) AncestorsAndPhaseFields() []string {
	return []string{"phase", "project_id", "department_id", "cluster_id", "tenant_id"}
}

func (Datavolume) OmitStatusFieldsExceptPhase() []string {
	return []string{"phase_message", "conditions", "datavolume_pvc_name", "datavolume_pv_name"}
}
