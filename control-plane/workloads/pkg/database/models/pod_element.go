/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "github.com/google/uuid"

const PodElementsTableName = "pod_elements"

type PodElement struct {
	PodID     uuid.UUID `gorm:"column:pod_id;primaryKey"`
	ElementID string    `gorm:"column:element_id"`
	Level     int       `gorm:"column:level;primaryKey"`
}

func (PodElement) TableName() string {
	return PodElementsTableName
}
