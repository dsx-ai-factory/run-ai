/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
)

type PodResourceClaimArray []PodResourceClaim

func (c PodResourceClaimArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

// Scan resets the receiver for both SQL NULL (pre-upgrade rows, since
// resource_claims was added to the live pods table) and JSON null (what Value
// writes for a nil array), so a reused destination never keeps stale claims.
// Decoding into c rather than &c is what lets JSON null clear *c.
func (c *PodResourceClaimArray) Scan(value interface{}) error {
	valueBytes, ok := value.([]byte)
	if !ok || len(valueBytes) == 0 {
		*c = nil
		return nil
	}
	if err := json.Unmarshal(valueBytes, c); err != nil {
		return err
	}
	return nil
}
