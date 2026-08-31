/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"time"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
)

// MappingLocking is a per-tenant mutex for preventing multiple updates of mapping (formatters, field groups,
// spec selectors). the locking mechanism is based on success adding record for my tenant. the locked at
// is for protecting against service creating record and panic before successully removing the lock.
type MappingLocking struct {
	TenantId int32     `gorm:"primaryKey"`
	LockedAt time.Time `gorm:"not null"`
}

func (MappingLocking) TableName() string {
	return dbtables.MappingLockingTable
}
