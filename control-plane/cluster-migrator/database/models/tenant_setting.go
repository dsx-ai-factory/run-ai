/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type TenantSetting struct {
	TenantId int32
	Key      string
	Value    string
}

func (TenantSetting) TableName() string {
	return "tenant_settings"
}
