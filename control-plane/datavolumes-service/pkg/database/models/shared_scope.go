/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
)

const (
	SharedScopeTableName = "shared_scopes"
)

type SharedScope struct {
	DatavolumeId uuid.UUID
	ScopeType    api.ScopeType
	ScopeId      string
	ScopeName    string
	TenantId     uint
	CreatedAt    time.Time
	CreatedBy    string

	ProjectId    string
	DepartmentId string
}

func (SharedScope) TableName() string {
	return SharedScopeTableName
}
