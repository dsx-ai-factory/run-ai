/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ApplicationChartUploadTableName       = "application_chart_uploads"
	ApplicationChartUploadColumnId        = "id"
	ApplicationChartUploadColumnTenantId  = "tenant_id"
	ApplicationChartUploadColumnCreatedAt = "created_at"
)

// ApplicationChartUpload is a transiently stored Helm chart archive uploaded from the UI.
// It is staged only until the cluster installs the release, then removed (see chart cleanup).
type ApplicationChartUpload struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId   string
	ChartBytes []byte
	CreatedAt  time.Time
}

func (ApplicationChartUpload) TableName() string {
	return ApplicationChartUploadTableName
}
