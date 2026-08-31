/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "time"

type ProvisioningRunStatus string

const (
	ProvisioningRunRunning ProvisioningRunStatus = "running"
	ProvisioningRunSuccess ProvisioningRunStatus = "success"
	ProvisioningRunFailed  ProvisioningRunStatus = "failed"
)

type ProvisioningRun struct {
	ID          uint
	TenantId    uint
	StartedAt   time.Time
	EndedAt     *time.Time
	Status      ProvisioningRunStatus
	Diagnostics string `gorm:"type:text"`
}
