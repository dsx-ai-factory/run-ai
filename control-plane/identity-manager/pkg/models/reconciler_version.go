/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type ReconcilerVersions struct {
	DeleteBaseModel
	TenantId                string `gorm:"not null;unique_index"`
	ServiceAccounts         string `gorm:"column:service_accounts"`
	InternalServiceAccounts string `gorm:"column:internal_service_accounts"`
	AccessKeys              string `gorm:"column:access_keys"`
}
