/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"reflect"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const DefaultNodePoolName = "default"

type OrgUnitData struct {
	dbmodels.BaseModel
	Resources                         []Resources
	TotalGpuQuota                     *float64 `gorm:"-"`
	TotalCpuQuota                     *float64 `gorm:"-"`
	TotalMemoryQuota                  *float64 `gorm:"-"`
	InteractiveJobTimeLimit           *int32
	InteractiveJobMaxIdleDuration     *int32
	InteractiveJobPreemptIdleDuration *int32
	TrainingJobTimeLimit              *int32
	TrainingJobMaxIdleDuration        *int32
	DefaultNodePoolList               *pq.StringArray `gorm:"type:text[]"`
	Children                          []Child         `gorm:"-"`
	TrainingNodeTypes                 *[]NodeType     `gorm:"many2many:org_unit_train_node_type;"`
	WorkspaceNodeTypes                *[]NodeType     `gorm:"many2many:org_unit_workspace_node_type;"`
	OvertimeData                      OvertimeData
	EffectiveData
}

type OrgUnitTrainNodeType struct {
	CreatedAt     time.Time
	OrgUnitDataId uuid.UUID
	NodeTypeId    int64
}

type OrgUnitWorkspaceNodeType struct {
	CreatedAt     time.Time
	OrgUnitDataId uuid.UUID
	NodeTypeId    int64
}

type EffectiveData struct {
	InteractiveJobTimeLimitEffective                 *int32
	InteractiveJobTimeLimitEffectiveOrigin           *uuid.UUID
	InteractiveJobMaxIdleDurationEffective           *int32
	InteractiveJobMaxIdleDurationEffectiveOrigin     *uuid.UUID
	InteractiveJobPreemptIdleDurationEffective       *int32
	InteractiveJobPreemptIdleDurationEffectiveOrigin *uuid.UUID
	TrainingJobTimeLimitEffective                    *int32
	TrainingJobTimeLimitEffectiveOrigin              *uuid.UUID
	TrainingJobMaxIdleDurationEffective              *int32
	TrainingJobMaxIdleDurationEffectiveOrigin        *uuid.UUID
	DefaultNodePoolListEffective                     *pq.StringArray `gorm:"type:text[]"`
	DefaultNodePoolListEffectiveOrigin               *uuid.UUID
	TrainingNodeTypesEffective                       *[]NodeType `gorm:"-"`
	TrainingNodeTypesEffectiveOrigin                 *uuid.UUID
	WorkspaceNodeTypesEffective                      *[]NodeType `gorm:"-"`
	WorkspaceNodeTypesEffectiveOrigin                *uuid.UUID
}

type OrgUnitDataView struct {
	ResourcesJson                     interface{} `gorm:"serializer:json;->"`
	TotalGpuQuota                     *float64
	TotalCpuQuota                     *float64
	TotalMemoryQuota                  *float64
	InteractiveJobTimeLimit           *int32
	InteractiveJobMaxIdleDuration     *int32
	InteractiveJobPreemptIdleDuration *int32
	TrainingJobTimeLimit              *int32
	TrainingJobMaxIdleDuration        *int32
	DefaultNodePoolList               *pq.StringArray `gorm:"type:text[]"`
	ChildrenView                      interface{}     `gorm:"serializer:json;->"`
	TrainingNodeTypesView             interface{}     `gorm:"serializer:json;->"`
	WorkspaceNodeTypesView            interface{}     `gorm:"serializer:json;->"`
	EffectiveDataView
	OvertimeDataView
}

type OrgUnitDataSyncView struct {
	ResourcesJson                     interface{} `gorm:"serializer:json;->"`
	InteractiveJobTimeLimit           *int32
	InteractiveJobMaxIdleDuration     *int32
	InteractiveJobPreemptIdleDuration *int32
	TrainingJobTimeLimit              *int32
	TrainingJobMaxIdleDuration        *int32
	DefaultNodePoolList               *pq.StringArray `gorm:"type:text[]"`
	TrainingNodeTypesView             interface{}     `gorm:"serializer:json;->"`
	WorkspaceNodeTypesView            interface{}     `gorm:"serializer:json;->"`
}

type EffectiveDataView struct {
	InteractiveJobTimeLimitEffective                 *int32
	InteractiveJobTimeLimitEffectiveOrigin           *uuid.UUID
	InteractiveJobMaxIdleDurationEffective           *int32
	InteractiveJobMaxIdleDurationEffectiveOrigin     *uuid.UUID
	InteractiveJobPreemptIdleDurationEffective       *int32
	InteractiveJobPreemptIdleDurationEffectiveOrigin *uuid.UUID
	TrainingJobTimeLimitEffective                    *int32
	TrainingJobTimeLimitEffectiveOrigin              *uuid.UUID
	TrainingJobMaxIdleDurationEffective              *int32
	TrainingJobMaxIdleDurationEffectiveOrigin        *uuid.UUID
	DefaultNodePoolListEffective                     *pq.StringArray `gorm:"type:text[]"`
	DefaultNodePoolListEffectiveOrigin               *uuid.UUID
	TrainingNodeTypesViewEffective                   interface{} `gorm:"serializer:json;->"`
	TrainingNodeTypesViewEffectiveOrigin             *uuid.UUID
	WorkspaceNodeTypesViewEffective                  interface{} `gorm:"serializer:json;->"`
	WorkspaceNodeTypesViewEffectiveOrigin            *uuid.UUID
}

type OvertimeDataView struct {
	OvertimeDataID             uuid.UUID
	AvgGpuAllocation24h        *float64
	AvgGpuUtilization24h       *float64
	AvgGpuMemoryUtilization24h *float64
	Avg24hRangeUpdatedAt       *time.Time
	AvgGpuAllocation7d         *float64
	AvgGpuUtilization7d        *float64
	AvgGpuMemoryUtilization7d  *float64
	Avg7dRangeUpdatedAt        *time.Time
	AvgGpuAllocation30d        *float64
	AvgGpuUtilization30d       *float64
	AvgGpuMemoryUtilization30d *float64
	Avg30dRangeUpdatedAt       *time.Time
}

func (OrgUnitData) TableName() string {
	return OrgUnitDataTable
}

func (o OrgUnitData) Clone() OrgUnitData {
	resources := []Resources{}
	if o.Resources != nil {
		for _, resource := range o.Resources {
			resources = append(resources, resource.Clone())
		}
	}

	var trainNodeTypes *[]NodeType = nil
	if o.TrainingNodeTypes != nil {
		nodeTypes := []NodeType{}
		for _, nodeType := range *o.TrainingNodeTypes {
			nodeTypes = append(nodeTypes, nodeType.Clone())
		}
		trainNodeTypes = ptr.To(nodeTypes)
	}

	var workspaceNodeTypes *[]NodeType = nil
	if o.WorkspaceNodeTypes != nil {
		nodeTypes := []NodeType{}
		for _, nodeType := range *o.WorkspaceNodeTypes {
			nodeTypes = append(nodeTypes, nodeType.Clone())
		}
		workspaceNodeTypes = ptr.To(nodeTypes)
	}

	return OrgUnitData{
		BaseModel: dbmodels.BaseModel{
			ID:        o.ID,
			CreatedAt: o.CreatedAt,
			UpdatedAt: o.UpdatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  o.DeletedAt.Time,
				Valid: o.DeletedAt.Valid,
			},
			CreatedBy: o.CreatedBy,
			UpdatedBy: o.UpdatedBy,
			DeletedBy: o.DeletedBy,
		},
		Resources:                         resources,
		InteractiveJobTimeLimit:           CopyPtr(o.InteractiveJobTimeLimit),
		InteractiveJobMaxIdleDuration:     CopyPtr(o.InteractiveJobMaxIdleDuration),
		InteractiveJobPreemptIdleDuration: CopyPtr(o.InteractiveJobPreemptIdleDuration),
		TrainingJobTimeLimit:              CopyPtr(o.TrainingJobTimeLimit),
		TrainingJobMaxIdleDuration:        CopyPtr(o.TrainingJobMaxIdleDuration),
		DefaultNodePoolList:               CopyPtr(o.DefaultNodePoolList),
		TrainingNodeTypes:                 trainNodeTypes,
		WorkspaceNodeTypes:                workspaceNodeTypes,
	}
}

func OrgUnitDataEqual(ou1, ou2 OrgUnitData) bool {
	// Compare basic fields and handle nil pointers
	if !ptr.Equal(ou1.InteractiveJobTimeLimit, ou2.InteractiveJobTimeLimit) ||
		!ptr.Equal(ou1.InteractiveJobMaxIdleDuration, ou2.InteractiveJobMaxIdleDuration) ||
		!ptr.Equal(ou1.InteractiveJobPreemptIdleDuration, ou2.InteractiveJobPreemptIdleDuration) ||
		!ptr.Equal(ou1.TrainingJobTimeLimit, ou2.TrainingJobTimeLimit) ||
		!ptr.Equal(ou1.TrainingJobMaxIdleDuration, ou2.TrainingJobMaxIdleDuration) {
		return false
	}

	if !reflect.DeepEqual(ou1.DefaultNodePoolList, ou2.DefaultNodePoolList) {
		return false
	}

	// Compare slices of structs
	if !ResourcesArrayEqual(ou1.Resources, ou2.Resources) ||
		!NodeTypesArrayEqual(ou1.TrainingNodeTypes, ou2.TrainingNodeTypes) ||
		!NodeTypesArrayEqual(ou1.WorkspaceNodeTypes, ou2.WorkspaceNodeTypes) {
		return false
	}

	return true
}

func OrgUnitDataV1Equal(ou1, ou2 OrgUnitData) bool {
	// Compare basic fields and handle nil pointers
	if !ptr.Equal(ou1.InteractiveJobTimeLimit, ou2.InteractiveJobTimeLimit) ||
		!ptr.Equal(ou1.InteractiveJobMaxIdleDuration, ou2.InteractiveJobMaxIdleDuration) ||
		!ptr.Equal(ou1.InteractiveJobPreemptIdleDuration, ou2.InteractiveJobPreemptIdleDuration) ||
		!ptr.Equal(ou1.TrainingJobTimeLimit, ou2.TrainingJobTimeLimit) ||
		!ptr.Equal(ou1.TrainingJobMaxIdleDuration, ou2.TrainingJobMaxIdleDuration) {
		return false
	}

	if !reflect.DeepEqual(ou1.DefaultNodePoolList, ou2.DefaultNodePoolList) {
		if !defaultNodePoolListEquals(ou1.DefaultNodePoolList, ou2.DefaultNodePoolList) && !defaultNodePoolListEquals(ou2.DefaultNodePoolList, ou1.DefaultNodePoolList) {
			return false
		}
	}

	// Compare slices of structs
	if !ResourcesArrayEqual(ou1.Resources, ou2.Resources) ||
		!NodeTypesArrayEqual(ou1.TrainingNodeTypes, ou2.TrainingNodeTypes) ||
		!NodeTypesArrayEqual(ou1.WorkspaceNodeTypes, ou2.WorkspaceNodeTypes) {
		return false
	}

	return true
}

func defaultNodePoolListEquals(l1, l2 *pq.StringArray) bool {
	defaultNodePoolArray := pq.StringArray([]string{DefaultNodePoolName})
	l1Empty := l1 == nil || len(*l1) == 0
	l2IsDefault := l2 != nil && reflect.DeepEqual(&defaultNodePoolArray, l2)
	if l1Empty && l2IsDefault {
		return true
	}
	return false
}

func convertRawToBaseModel(raw map[string]interface{}) dbmodels.BaseModel {
	data := dbmodels.BaseModel{}

	if raw["id"] != nil {
		data.ID = uuid.MustParse(raw["id"].(string))
	}
	if raw["created_at"] != nil {
		data.CreatedAt, _ = time.Parse(time.RFC3339, raw["created_at"].(string))
	}
	if raw["updated_at"] != nil {
		data.UpdatedAt, _ = time.Parse(time.RFC3339, raw["updated_at"].(string))
	}
	if raw["deleted_at"] != nil {
		deletedAt, _ := time.Parse(time.RFC3339, raw["deleted_at"].(string))
		data.DeletedAt = gorm.DeletedAt{Time: deletedAt, Valid: true}
	}
	if raw["created_by"] != nil {
		data.CreatedBy = raw["created_by"].(string)
	}
	if raw["updated_by"] != nil {
		data.UpdatedBy = raw["updated_by"].(string)
	}
	if raw["deleted_by"] != nil {
		data.DeletedBy = raw["deleted_by"].(string)
	}

	return data
}

func ConvertRawToResources(raw interface{}) []Resources {
	resources := []Resources{}
	if raw == nil {
		return resources
	}
	seenResources := map[string]bool{}
	for _, v := range raw.([]interface{}) {
		if v == nil {
			continue
		}

		resourceMap := v.(map[string]interface{})

		if resourceMap["deleted_at"] != nil {
			continue
		}

		if _, ok := seenResources[resourceMap[ResourcesNodePoolIdColumn].(string)]; ok {
			continue
		}

		resource := convertRawToResources(resourceMap)

		seenResources[resourceMap[ResourcesNodePoolIdColumn].(string)] = true
		resources = append(resources, resource)
	}
	return resources
}

func ConvertRawToNodeTypes(raw interface{}) *[]NodeType {
	if raw == nil {
		return nil
	}
	nodeTypes := []NodeType{}
	seenNodeTypes := map[string]bool{}
	for _, v := range raw.([]interface{}) {
		if v == nil {
			continue
		}

		nodeTypeMap := v.(map[string]interface{})

		if nodeTypeMap["id"] == nil {
			continue
		}

		if nodeTypeMap["deleted_at"] != nil {
			continue
		}

		if _, ok := seenNodeTypes[nodeTypeMap["name"].(string)]; ok {
			continue
		}
		nodeType := convertRawToNodeType(nodeTypeMap)
		seenNodeTypes[nodeTypeMap["name"].(string)] = true

		nodeTypes = append(nodeTypes, nodeType)
	}
	return ptr.To(nodeTypes)
}

func ConvertRawToChildren(raw interface{}) []Child {
	children := []Child{}
	if raw == nil {
		return children
	}
	seenChildren := map[int64]bool{}
	for _, v := range raw.([]interface{}) {
		if v == nil {
			continue
		}

		childMap := v.(map[string]interface{})

		if id, ok := childMap["id"]; !ok || id == nil {
			continue
		}

		id, ok := childMap["id"].(float64)
		if !ok {
			continue
		}

		if _, ok := seenChildren[int64(id)]; ok {
			continue
		}
		child := convertRawToChild(childMap)
		seenChildren[int64(id)] = true

		children = append(children, child)
	}
	return children
}
