/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
)

type NodePoolQuotaStatus struct {
	NodePoolName            string                  `json:"node_pool_name,omitempty"`
	NodePoolId              string                  `json:"node_pool_id,omitempty"`
	Allocated               api.QuotaStatusResource `json:"allocated,omitempty"`
	AllocatedNonPreemptible api.QuotaStatusResource `json:"allocated_non_preemptible,omitempty"`
	Requested               api.QuotaStatusResource `json:"requested,omitempty"`
}

func (c NodePoolQuotaStatus) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *NodePoolQuotaStatus) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}

type NodePoolQuotaStatusArray []NodePoolQuotaStatus

func (c NodePoolQuotaStatusArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *NodePoolQuotaStatusArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
