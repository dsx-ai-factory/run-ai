/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type Container struct {
	StartedAt *time.Time
	ClusterID uuid.UUID
	PodID     uuid.UUID
	Pod       Pod `gorm:"foreignKey:PodID"`
	Name      string
	Image     string
}

func (Container) TableName() string {
	return "containers"
}
