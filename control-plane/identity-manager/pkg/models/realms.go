/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type Realm struct {
	BaseModel
	TenantId string `gorm:"not null;unique_index"`
	Realm    string `gorm:"not null;unique_index"`
}

type DeleteRealm struct {
	DeleteBaseModel
	TenantId string `gorm:"not null;unique_index"`
	Realm    string `gorm:"not null;unique_index"`
}
