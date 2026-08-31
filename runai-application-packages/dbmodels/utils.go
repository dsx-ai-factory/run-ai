/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodels

import (
	"time"

	"gorm.io/gorm"
)

func GormDeletedAtToNullableTime(time gorm.DeletedAt) *time.Time {
	if !time.Valid {
		return nil
	}

	return &time.Time
}
