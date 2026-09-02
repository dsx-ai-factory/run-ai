/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
)

type Permission struct {
	CreatedAt              time.Time
	UpdatedAt              time.Time
	RoleID                 uint             `gorm:"primarykey"`
	ResourceType           api.ResourceType `gorm:"primarykey;index"`
	Create                 bool
	Read                   bool
	Update                 bool
	Delete                 bool
	Sync                   bool
	ReadOnlyForDGXCManaged bool `gorm:"-"`
	ReadOnlyForCSP         bool `gorm:"-"`
}
