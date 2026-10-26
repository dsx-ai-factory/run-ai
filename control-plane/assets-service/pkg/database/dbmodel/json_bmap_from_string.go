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

// ToJsonBMap converts an arbitrary JSON-marshalable value into the map-shaped
// storage representation used by jsonb columns backed by JsonBMapFromString.
// Round-tripping through JSON is intentional: it normalizes typed structs,
// pointers, and any existing map into the exact `map[string]any` shape the
// storage layer stores.
func ToJsonBMap(v any) (JsonBMapFromString, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal value to json: %w", err)
	}
	var m JsonBMapFromString
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, fmt.Errorf("unmarshal json into JsonBMapFromString: %w", err)
	}
	return m, nil
}

// FromJsonBMap is the inverse of ToJsonBMap: it materializes the caller's
// typed destination from the map-shaped stored value via a JSON round-trip.
// Used by per-kind read DAOs to unmarshal a stored spec into its typed
// swagger struct.
func FromJsonBMap(m JsonBMapFromString, out any) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal JsonBMapFromString to json: %w", err)
	}
	if err := json.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("unmarshal json into %T: %w", out, err)
	}
	return nil
}
