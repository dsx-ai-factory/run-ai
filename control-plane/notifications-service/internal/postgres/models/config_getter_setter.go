/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "gorm.io/datatypes"

type ConfigGetterSetter interface {
	GetConfig() datatypes.JSON
	SetConfig(datatypes.JSON)
}
