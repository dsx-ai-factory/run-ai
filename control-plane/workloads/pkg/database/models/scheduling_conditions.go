/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

type SchedulingCondition struct {
	NodePool     string                               `json:"nodePool,omitempty"`
	Reason       api.PhaseReason                      `json:"reason,omitempty"`
	QueueDetails *api.SchedulingConditionQueueDetails `json:"queueDetails,omitempty"`
}

type SchedulingConditionArray []SchedulingCondition

func (c SchedulingCondition) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *SchedulingCondition) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}

func (c SchedulingConditionArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *SchedulingConditionArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
