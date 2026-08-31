/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"gorm.io/gorm"
)

func isDuplicateKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
