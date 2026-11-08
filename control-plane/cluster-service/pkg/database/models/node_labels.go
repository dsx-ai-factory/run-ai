/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
)

func (labels NodeLabels) Value() (driver.Value, error) {
	if labels == nil {
		return "{}", nil
	}
	encoded, err := json.Marshal(labels)
	return string(encoded), err
}
