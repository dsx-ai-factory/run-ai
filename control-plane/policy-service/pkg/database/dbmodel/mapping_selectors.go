/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// MappingSelectors is one scope-specific row (system or tenant) for a workload GVK mapping.
// The same logical GVK may exist twice (system + tenant); the service merges specSelectors by key.
type MappingSelectors struct {
	ID uuid.UUID `gorm:"primaryKey"`

	GvkGroup   string
	GvkKind    string
	GvkVersion string

	Description *string

	ScopeType api.MappingScope
	TenantId  int32

	SpecSelectors map[string]api.SpecSelectorsWithScopeValue `gorm:"column:spec_selectors;serializer:json;type:jsonb;not null"`

	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func (MappingSelectors) TableName() string {
	return dbtables.MappingSelectorsTable
}

func (ms MappingSelectors) DbmodelToApi() *api.Mapping {
	return &api.Mapping{
		Id:          ms.ID.String(),
		Scope:       ms.ScopeType,
		Description: ms.Description,

		WorkloadType: api.Gvk{
			Group:   convutils.Ref(ms.GvkGroup),
			Kind:    ms.GvkKind,
			Version: ms.GvkVersion,
		},

		SpecSelectors: ms.SpecSelectors,

		CreatedBy: ms.CreatedBy,
		CreatedAt: ms.CreatedAt,
		UpdatedBy: ms.UpdatedBy,
		UpdatedAt: ms.UpdatedAt,
	}
}

// FilterFieldsToColumn maps API filter field names (OpenAPI MappingFilter: group, kind, version) to mapping_selectors columns.
func (MappingSelectors) FilterFieldsToColumn() map[string]string {
	return map[string]string{
		"group":   dbtables.MappingSelectorsColumnGvkGroup,
		"kind":    dbtables.MappingSelectorsColumnGvkKind,
		"version": dbtables.MappingSelectorsColumnGvkVersion,
	}
}
