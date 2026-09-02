/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

// JsonBMapFromString is only needed because we do not run the real migrations
// in our unit test, and so the json column is created as json and not as jsonb
// once we remove d
type JsonBMapFromString dbmodels.JsonBMap

func (m JsonBMapFromString) Value() (driver.Value, error) {
	if m == nil {
		return "null", nil
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

func (m *JsonBMapFromString) Scan(value interface{}) error {
	var b []byte

	switch realValue := value.(type) {
	case []byte:
		b = realValue
	case string:
		b = []byte(realValue)
	default:
		return fmt.Errorf("invalid scan type for JsonBMap")
	}

	tmpMap, err := gojsonq.LoadJsonBytes(b)
	if err != nil {
		return err
	}
	*m = tmpMap
	return nil
}
