/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
)

type WorkloadEndpoints struct {
	WorkloadId string        `json:"workloadId,omitempty"`
	Endpoints  EndpointArray `json:"endpoints,omitempty"`
}

type WorkloadEndpointsArray []WorkloadEndpoints

func (e WorkloadEndpointsArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(e)
	return string(valueString), err
}

func (e *WorkloadEndpointsArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &e); err != nil {
		return err
	}
	return nil
}
