/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type PodResourceClaim struct {
	Name                       string  `json:"name"`
	ResourceClaimTemplateName  *string `json:"resource_claim_template_name,omitempty"`
	ResourceClaimName          *string `json:"resource_claim_name,omitempty"`
	AllocatedResourceClaimName *string `json:"allocated_resource_claim_name,omitempty"`
}

func (c PodResourceClaim) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

// Scan resets the receiver before decoding because json.Unmarshal merges into
// an existing struct, leaving fields absent from the payload (the omitempty
// pointers) holding stale values from a reused destination.
func (c *PodResourceClaim) Scan(value interface{}) error {
	*c = PodResourceClaim{}
	if value == nil {
		return nil
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("scan PodResourceClaim: unsupported type %T", value)
	}
	if err := json.Unmarshal(valueBytes, c); err != nil {
		return fmt.Errorf("scan PodResourceClaim: %w", err)
	}
	return nil
}
