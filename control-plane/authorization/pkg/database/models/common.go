/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"
)

type Common struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
}
