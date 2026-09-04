/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"fmt"
)

type AuthType string

const (
	ATBasic AuthType = "basic"
)

type MetricStoreEndpointAuth struct {
	Type  AuthType   `gorm:"not null;"`
	Basic *AuthBasic `gorm:"type:jsonb;not null;"`
}

type AuthBasic struct {
	Username string `gorm:"not null;"`
	Password string `gorm:"not null;"`
}

func (m *MetricStoreEndpointAuth) Scan(value interface{}) error {
	if value == nil {
		*m = MetricStoreEndpointAuth{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan MetricStoreEndpointAuth: %v", value)
	}

	return json.Unmarshal(bytes, m)
}
