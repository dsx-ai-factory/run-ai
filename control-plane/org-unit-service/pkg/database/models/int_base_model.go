/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"gorm.io/gorm"
)

type IntBaseModel struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	CreatedBy string
	UpdatedBy string
	DeletedBy string
}

func convertRawToIntBaseModel(raw map[string]interface{}) IntBaseModel {
	data := IntBaseModel{}

	if raw["id"] != nil {
		data.ID = int64(raw["id"].(float64))
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
