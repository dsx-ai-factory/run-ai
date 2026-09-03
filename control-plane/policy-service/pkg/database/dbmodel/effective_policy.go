/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/google/uuid"
)

type EffectivePolicy struct {
	PolicyId      uuid.UUID `gorm:"primaryKey"`
	TenantId      int32
	DefaultsV2    dbmodels.JsonBMap
	RulesV2       dbmodels.JsonBMap
	ImposedAssets *api.DistributedImposedAssets `gorm:"serializer:json; type:jsonb;"`
	UpdatedAt     time.Time
	UpdateCount   int64

	DefaultFromOrder *api.DefaultFromOrder `gorm:"serializer:json; type:jsonb;"`
	ValidationError  *string

	// soft delete support implemented as ordinary field, as primaryKey + gorm deletion
	// is not implemented well by GORM, and for this table we actually do not need it
	DeletedAt *time.Time `gorm:"primaryKey"`
}

func (EffectivePolicy) TableName() string {
	return dbtables.EffectivePoliciesTable
}

func (ep EffectivePolicy) IsDeleted() bool {
	return ep.PolicyId == uuid.Nil || ep.DeletedAt != nil
}
