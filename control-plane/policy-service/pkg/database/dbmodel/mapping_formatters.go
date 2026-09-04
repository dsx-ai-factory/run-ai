/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
)

type MappingFormatters struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string
	Description *string

	ScopeType api.MappingScope
	TenantId  int32

	Fields map[string]api.FormatterField `gorm:"serializer:json;type:jsonb;not null"`

	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func (MappingFormatters) TableName() string {
	return dbtables.MappingFormattersTable
}

func (pf MappingFormatters) ToApiFormatter() *api.Formatter {
	return &api.Formatter{
		Id: pf.ID.String(),

		Name:        pf.Name,
		Description: pf.Description,

		Scope: pf.ScopeType,

		Fields: pf.Fields,

		CreatedBy: pf.CreatedBy,
		CreatedAt: pf.CreatedAt,
		UpdatedBy: pf.UpdatedBy,
		UpdatedAt: pf.UpdatedAt,
	}
}

const BuiltinFormatterKeyValue = "KeyValue"

const BuiltinFormatterKeyValueMap = "KeyValueMap"
