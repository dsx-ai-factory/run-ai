/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
)

type WorkloadResourceRef struct {
	UID  string `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type WorkloadResourceRefArray []WorkloadResourceRef

func (r WorkloadResourceRefArray) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	valueString, err := json.Marshal(r)
	return string(valueString), err
}

func (r *WorkloadResourceRefArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &r); err != nil {
		return err
	}
	return nil
}
