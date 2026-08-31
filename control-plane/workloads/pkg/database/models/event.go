/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	TenantID                uint
	ID                      uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClusterID               uuid.UUID
	Type                    string
	Message                 string
	Reason                  string
	CreatedAt               time.Time
	Source                  string
	InvolvedObjectUid       string
	InvolvedObjectKind      string
	InvolvedObjectName      string
	InvolvedObjectNamespace string
}

func (Event) TableName() string {
	return "events"
}
